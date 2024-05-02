// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"bytes"
	"context"
	"sort"

	"github.com/hashicorp/hcl-lang/decoder/internal/ast"
	"github.com/hashicorp/hcl-lang/decoder/internal/schemahelper"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/zclconf/go-cty/cty"
)

type ReferenceTargets []*ReferenceTarget

func (d *Decoder) ReferenceTargetsForOriginAtPos(path lang.Path, file string, pos hcl.Pos) (ReferenceTargets, error) {
	pathCtx, err := d.pathReader.PathContext(path)
	if err != nil {
		return nil, err
	}

	matchingTargets := make(ReferenceTargets, 0)

	origins, ok := pathCtx.ReferenceOrigins.AtPos(file, pos)
	if !ok {
		return matchingTargets, &reference.NoOriginFound{}
	}

	for _, origin := range origins {
		targetCtx := pathCtx
		targetPath := path

		if directOrigin, ok := origin.(reference.DirectOrigin); ok {
			matchingTargets = append(matchingTargets, &ReferenceTarget{
				OriginRange: origin.OriginRange(),
				Path:        directOrigin.TargetPath,
				Range:       directOrigin.TargetRange,
				DefRangePtr: nil,
			})
			continue
		}
		if pathOrigin, ok := origin.(reference.PathOrigin); ok {
			ctx, err := d.pathReader.PathContext(pathOrigin.TargetPath)
			if err != nil {
				continue
			}
			targetCtx = ctx
			targetPath = pathOrigin.TargetPath
		}

		matchableOrigin, ok := origin.(reference.MatchableOrigin)
		if !ok {
			continue
		}
		targets, ok := targetCtx.ReferenceTargets.Match(matchableOrigin)
		if !ok {
			// target not found
			continue
		}
		for _, target := range targets {
			if target.RangePtr == nil {
				// target is not addressable
				continue
			}
			matchingTargets = append(matchingTargets, &ReferenceTarget{
				OriginRange: origin.OriginRange(),
				Path:        targetPath,
				Range:       *target.RangePtr,
				DefRangePtr: target.DefRangePtr,
			})
		}
	}

	return matchingTargets, nil
}

func (d *PathDecoder) CollectReferenceTargets() (reference.Targets, error) {
	if d.pathCtx.Schema == nil {
		// unable to collect reference targets without schema
		return nil, &NoSchemaError{}
	}

	refs := make(reference.Targets, 0)
	files := d.filenames()
	for _, filename := range files {
		f, err := d.fileByName(filename)
		if err != nil {
			// skip unparseable file
			continue
		}
		refs = append(refs, d.decodeReferenceTargetsForBody(f.Body, nil, d.pathCtx.Schema)...)
	}

	return refs, nil
}

func (d *PathDecoder) decodeReferenceTargetsForBody(body hcl.Body, parentBlock *ast.BlockContent, bodySchema *schema.BodySchema) reference.Targets {
	refs := make(reference.Targets, 0)

	if bodySchema == nil {
		return reference.Targets{}
	}

	content := ast.DecodeBody(body, bodySchema)

	for _, attr := range content.Attributes {
		if bodySchema.Extensions != nil {
			if bodySchema.Extensions.Count && attr.Name == "count" && content.RangePtr != nil {
				refs = append(refs, countIndexReferenceTarget(attr, *content.RangePtr))
				continue
			}
			if bodySchema.Extensions.ForEach && attr.Name == "for_each" && content.RangePtr != nil {
				refs = append(refs, forEachReferenceTargets(attr, *content.RangePtr)...)
				continue
			}
		}
		attrSchema, ok := bodySchema.Attributes[attr.Name]
		if !ok {
			if bodySchema.AnyAttribute == nil {
				// unknown attribute (no schema)
				continue
			}
			attrSchema = bodySchema.AnyAttribute
		}

		refs = append(refs, d.decodeReferenceTargetsForAttribute(attr, attrSchema)...)
	}

	for _, blk := range content.Blocks {
		bSchema, ok := bodySchema.Blocks[blk.Type]
		if !ok {
			// unknown block (no schema)
			continue
		}

		mergedSchema, _ := schemahelper.MergeBlockBodySchemas(blk.Block, bSchema)

		iRefs := d.decodeReferenceTargetsForBody(blk.Body, blk, mergedSchema)
		refs = append(refs, iRefs...)

		addr, ok := resolveBlockAddress(blk.Block, bSchema)
		if !ok {
			// skip unresolvable address
			continue
		}

		if bSchema.Address.AsReference {
			ref := reference.Target{
				Addr:        addr,
				ScopeId:     bSchema.Address.ScopeId,
				DefRangePtr: blk.DefRange.Ptr(),
				RangePtr:    blk.Range.Ptr(),
				Name:        bSchema.Address.FriendlyName,
			}
			refs = append(refs, ref)
		}

		if bSchema.Address.AsTypeOf != nil {
			refs = append(refs, referenceAsTypeOf(blk.Block, blk.Range.Ptr(), bSchema, addr)...)
		}

		var bodyRef reference.Target

		if bSchema.Address.BodyAsData {
			bodyRef = reference.Target{
				Addr:        addr,
				ScopeId:     bSchema.Address.ScopeId,
				DefRangePtr: blk.DefRange.Ptr(),
				RangePtr:    blk.Range.Ptr(),
			}

			if bSchema.Body != nil {
				bodyRef.Description = bSchema.Body.Description
			}

			if bSchema.Address.InferBody && bSchema.Body != nil {
				var localAddr lang.Address
				if bSchema.Address.BodySelfRef {
					localAddr = lang.Address{
						lang.RootStep{Name: "self"},
					}
					bodyRef.TargetableFromRangePtr = blk.Range.Ptr()
				}
				bodyRef.NestedTargets = append(bodyRef.NestedTargets,
					d.collectInferredReferenceTargetsForBody(addr, bSchema.Address, blk.Body, bSchema.Body, nil, localAddr)...)
			}

			bodyRef.Type = bodyToDataType(bSchema.Type, bSchema.Body)

			refs = append(refs, bodyRef)
		}

		if bSchema.Address.DependentBodyAsData {
			if !bSchema.Address.BodyAsData {
				bodyRef = reference.Target{
					Addr:        addr,
					ScopeId:     bSchema.Address.ScopeId,
					DefRangePtr: blk.DefRange.Ptr(),
					RangePtr:    blk.Range.Ptr(),
				}
			}

			depSchema, _, result := schemahelper.NewBlockSchema(bSchema).DependentBodySchema(blk.Block)
			if result == schemahelper.LookupSuccessful {
				fullSchema := depSchema
				if bSchema.Address.BodyAsData {
					mergedSchema, _ := schemahelper.MergeBlockBodySchemas(blk.Block, bSchema)
					bodyRef.NestedTargets = make(reference.Targets, 0)
					fullSchema = mergedSchema
				}

				bodyRef.Type = bodyToDataType(bSchema.Type, fullSchema)

				if bSchema.Address.InferDependentBody && len(bSchema.DependentBody) > 0 {
					if bSchema.Address.DependentBodySelfRef {
						bodyRef.LocalAddr = lang.Address{
							lang.RootStep{Name: "self"},
						}
						bodyRef.TargetableFromRangePtr = blk.Range.Ptr()
					} else {
						bodyRef.LocalAddr = lang.Address{}
					}

					bodyRef.NestedTargets = append(bodyRef.NestedTargets,
						d.collectInferredReferenceTargetsForBody(addr, bSchema.Address, blk.Body, fullSchema, nil, bodyRef.LocalAddr)...)
				}

				if !bSchema.Address.BodyAsData {
					refs = append(refs, bodyRef)
				}
			}
		}

		sort.Sort(bodyRef.NestedTargets)
	}

	for _, tb := range bodySchema.TargetableAs {
		refs = append(refs, decodeTargetableBody(body, parentBlock, tb))
	}

	sort.Sort(refs)

	return refs
}

func decodeTargetableBody(body hcl.Body, parentBlock *ast.BlockContent, tt *schema.Targetable) reference.Target {
	target := reference.Target{
		Addr:        tt.Address.Copy(),
		ScopeId:     tt.ScopeId,
		RangePtr:    parentBlock.Range.Ptr(),
		DefRangePtr: parentBlock.DefRange.Ptr(),
		Type:        tt.AsType,
		Description: tt.Description,
	}

	if tt.NestedTargetables != nil {
		target.NestedTargets = make(reference.Targets, len(tt.NestedTargetables))
		for i, ntt := range tt.NestedTargetables {
			target.NestedTargets[i] = decodeTargetableBody(body, parentBlock, ntt)
		}
	}

	return target
}

func (d *PathDecoder) decodeReferenceTargetsForAttribute(attr *hcl.Attribute, attrSchema *schema.AttributeSchema) reference.Targets {
	refs := make(reference.Targets, 0)

	ctx := context.Background()

	expr := d.newExpression(attr.Expr, attrSchema.Constraint)
	if eType, ok := expr.(ReferenceTargetsExpression); ok {
		var targetCtx *TargetContext
		if attrSchema.Address != nil {
			attrAddr, ok := resolveAttributeAddress(attr, attrSchema.Address.Steps)
			if ok && (attrSchema.Address.AsExprType || attrSchema.Address.AsReference) {
				targetCtx = &TargetContext{
					FriendlyName:      attrSchema.Address.FriendlyName,
					ScopeId:           attrSchema.Address.ScopeId,
					AsExprType:        attrSchema.Address.AsExprType,
					AsReference:       attrSchema.Address.AsReference,
					ParentAddress:     attrAddr,
					ParentRangePtr:    attr.Range.Ptr(),
					ParentDefRangePtr: attr.NameRange.Ptr(),
				}
			}

			if attrSchema.Address.AsReference {
				ref := reference.Target{
					Addr:          attrAddr,
					ScopeId:       attrSchema.Address.ScopeId,
					DefRangePtr:   attr.NameRange.Ptr(),
					RangePtr:      attr.Range.Ptr(),
					Name:          attrSchema.Address.FriendlyName,
					NestedTargets: reference.Targets{},
				}
				refs = append(refs, ref)
			}
		}

		refs = append(refs, eType.ReferenceTargets(ctx, targetCtx)...)
	}

	return refs
}

func referenceAsTypeOf(block *hcl.Block, rngPtr *hcl.Range, bSchema *schema.BlockSchema, addr lang.Address) reference.Targets {
	ref := reference.Target{
		Addr:        addr,
		ScopeId:     bSchema.Address.ScopeId,
		DefRangePtr: block.DefRange.Ptr(),
		RangePtr:    rngPtr,
		Type:        cty.DynamicPseudoType,
	}

	if bSchema.Body != nil {
		ref.Description = bSchema.Body.Description
	}

	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return reference.Targets{ref}
	}

	if bSchema.Address.AsTypeOf.AttributeExpr != "" {
		typeDecl, ok := asTypeOfAttrExpr(attrs, bSchema)
		if !ok {
			// nothing to fall back to, exit early
			return reference.Targets{ref}
		}
		ref.Type = typeDecl
	}

	return reference.Targets{ref}
}

func asTypeOfAttrExpr(attrs hcl.Attributes, bSchema *schema.BlockSchema) (cty.Type, bool) {
	attrName := bSchema.Address.AsTypeOf.AttributeExpr
	attr, ok := attrs[attrName]
	if !ok {
		return cty.DynamicPseudoType, false
	}

	aSchema := bSchema.Body.Attributes[attrName]
	_, ok = aSchema.Constraint.(schema.TypeDeclaration)
	if !ok {
		return cty.DynamicPseudoType, false
	}

	// TODO: TypeConstraintWithDefaults
	typeDecl, diags := typeexpr.TypeConstraint(attr.Expr)
	if diags.HasErrors() {
		return cty.DynamicPseudoType, false
	}

	return typeDecl, true
}

func bodySchemaAsAttrTypes(bodySchema *schema.BodySchema) map[string]cty.Type {
	attrTypes := make(map[string]cty.Type, 0)

	if bodySchema == nil {
		return attrTypes
	}

	for name, attr := range bodySchema.Attributes {
		cons, ok := attr.Constraint.(schema.TypeAwareConstraint)
		if !ok {
			continue
		}
		typ, ok := cons.ConstraintType()
		if !ok {
			continue
		}
		attrTypes[name] = typ
	}

	for name, block := range bodySchema.Blocks {
		attrTypes[name] = bodyToDataType(block.Type, block.Body)
	}

	return attrTypes
}

func (d *PathDecoder) collectInferredReferenceTargetsForBody(addr lang.Address, bAddrSchema *schema.BlockAddrSchema, body hcl.Body, bodySchema *schema.BodySchema, selfRefBodyRangePtr *hcl.Range, selfRefAddr lang.Address) reference.Targets {
	var (
		refs             = make(reference.Targets, 0)
		collectLocalAddr = false
		content          = ast.DecodeBody(body, bodySchema)
	)
	if bAddrSchema.DependentBodySelfRef || bAddrSchema.BodySelfRef {
		if selfRefBodyRangePtr == nil {
			// We don't get body range for JSON here
			// TODO? calculate or implement upstream
			selfRefBodyRangePtr = content.RangePtr
		}
		collectLocalAddr = selfRefBodyRangePtr != nil
	}

	rawAttributes, _ := body.JustAttributes()

	for name, aSchema := range bodySchema.Attributes {
		var attrType cty.Type
		cons, ok := aSchema.Constraint.(schema.TypeAwareConstraint)
		if ok {
			typ, ok := cons.ConstraintType()
			if ok {
				attrType = typ
			}
		}

		rawAttr, ok := rawAttributes[name]
		if ok {
			// try to infer type if attribute is declared
			expr, ok := newExpression(d.pathCtx, rawAttr.Expr, aSchema.Constraint).(CanInferTypeExpression)
			if ok {
				typ, ok := expr.InferType()
				if ok {
					attrType = typ
				}
			}
		}

		if attrType == cty.NilType {
			continue
		}

		attrAddr := append(addr.Copy(), lang.AttrStep{Name: name})
		targetCtx := &TargetContext{
			ParentAddress: attrAddr,
			ScopeId:       bAddrSchema.ScopeId,
			AsExprType:    true,
		}

		if collectLocalAddr {
			localAddr := append(selfRefAddr.Copy(), lang.AttrStep{Name: name})
			targetCtx.ParentLocalAddress = localAddr
			targetCtx.TargetableFromRangePtr = selfRefBodyRangePtr.Ptr()
		}

		var attrExpr hcl.Expression
		if attr, ok := content.Attributes[name]; ok {
			targetCtx.ParentRangePtr = attr.Range.Ptr()
			targetCtx.ParentDefRangePtr = attr.NameRange.Ptr()
			attrExpr = attr.Expr
		}

		if attrExpr == nil {
			attrExpr = newEmptyExpressionAtPos(body.MissingItemRange().Filename, body.MissingItemRange().Start)
		}
		expr, ok := newExpression(d.pathCtx, attrExpr, aSchema.Constraint).(ReferenceTargetsExpression)
		if ok {
			ctx := context.Background()
			refs = append(refs, expr.ReferenceTargets(ctx, targetCtx)...)
		}
	}

	bTypes := blocksTypesWithSchema(body, bodySchema)

	for bType, bCollection := range bTypes.OfSchemaType(schema.BlockTypeObject) {
		blockAddr := append(addr.Copy(), lang.AttrStep{Name: bType})

		blk := bCollection.Blocks[0]

		blockRef := reference.Target{
			Addr:        blockAddr,
			LocalAddr:   make(lang.Address, 0),
			ScopeId:     bAddrSchema.ScopeId,
			Type:        cty.Object(bodySchemaAsAttrTypes(bCollection.Schema.Body)),
			Description: bCollection.Schema.Description,
			DefRangePtr: blk.DefRange.Ptr(),
			RangePtr:    blk.Range.Ptr(),
		}
		if collectLocalAddr {
			blockRef.LocalAddr = append(selfRefAddr.Copy(), lang.AttrStep{Name: bType})
			blockRef.TargetableFromRangePtr = selfRefBodyRangePtr.Ptr()
		}
		blockRef.NestedTargets = d.collectInferredReferenceTargetsForBody(
			blockAddr, bAddrSchema, blk.Body, bCollection.Schema.Body, selfRefBodyRangePtr, blockRef.LocalAddr)

		sort.Sort(blockRef.NestedTargets)
		refs = append(refs, blockRef)
	}

	for bType, bCollection := range bTypes.OfSchemaType(schema.BlockTypeList) {
		blockAddr := append(addr.Copy(), lang.AttrStep{Name: bType})

		blockRef := reference.Target{
			Addr:        blockAddr,
			LocalAddr:   make(lang.Address, 0),
			ScopeId:     bAddrSchema.ScopeId,
			Type:        cty.List(cty.Object(bodySchemaAsAttrTypes(bCollection.Schema.Body))),
			Description: bCollection.Schema.Description,
			RangePtr:    body.MissingItemRange().Ptr(),
		}
		if collectLocalAddr {
			blockRef.LocalAddr = append(selfRefAddr.Copy(), lang.AttrStep{Name: bType})
			blockRef.TargetableFromRangePtr = selfRefBodyRangePtr.Ptr()
		}

		for i, b := range bCollection.Blocks {
			elemAddr := append(blockAddr.Copy(), lang.IndexStep{
				Key: cty.NumberIntVal(int64(i)),
			})

			elemRef := reference.Target{
				Addr:        elemAddr,
				LocalAddr:   make(lang.Address, 0),
				ScopeId:     bAddrSchema.ScopeId,
				Type:        cty.Object(bodySchemaAsAttrTypes(bCollection.Schema.Body)),
				Description: bCollection.Schema.Description,
				DefRangePtr: b.DefRange.Ptr(),
				RangePtr:    b.Range.Ptr(),
			}

			if collectLocalAddr {
				elemRef.LocalAddr = append(blockRef.LocalAddr.Copy(), lang.IndexStep{
					Key: cty.NumberIntVal(int64(i)),
				})
				elemRef.TargetableFromRangePtr = selfRefBodyRangePtr.Ptr()
			}

			elemRef.NestedTargets = d.collectInferredReferenceTargetsForBody(
				elemAddr, bAddrSchema, b.Body, bCollection.Schema.Body, selfRefBodyRangePtr, elemRef.LocalAddr)

			sort.Sort(elemRef.NestedTargets)
			blockRef.NestedTargets = append(blockRef.NestedTargets, elemRef)

			if i == 0 {
				blockRef.RangePtr = elemRef.RangePtr
			} else {
				// try to expand the range of the "parent" (list) reference
				// if the individual blocks follow each other
				betweenBlocks, err := d.bytesInRange(hcl.Range{
					Filename: blockRef.RangePtr.Filename,
					Start:    blockRef.RangePtr.End,
					End:      elemRef.RangePtr.Start,
				})
				if err == nil && len(bytes.TrimSpace(betweenBlocks)) == 0 {
					blockRef.RangePtr.End = elemRef.RangePtr.End
				}
			}
		}
		sort.Sort(blockRef.NestedTargets)
		refs = append(refs, blockRef)
	}

	for bType, bCollection := range bTypes.OfSchemaType(schema.BlockTypeSet) {
		blockAddr := append(addr.Copy(), lang.AttrStep{Name: bType})

		blockRef := reference.Target{
			Addr:        blockAddr,
			LocalAddr:   make(lang.Address, 0),
			ScopeId:     bAddrSchema.ScopeId,
			Type:        cty.Set(cty.Object(bodySchemaAsAttrTypes(bCollection.Schema.Body))),
			Description: bCollection.Schema.Description,
			RangePtr:    body.MissingItemRange().Ptr(),
		}
		if collectLocalAddr {
			blockRef.LocalAddr = append(selfRefAddr.Copy(), lang.AttrStep{Name: bType})
			blockRef.TargetableFromRangePtr = selfRefBodyRangePtr.Ptr()
		}

		for i, b := range bCollection.Blocks {
			if i == 0 {
				blockRef.RangePtr = b.Range.Ptr()
			} else {
				// try to expand the range of the "parent" (set) reference
				// if the individual blocks follow each other
				betweenBlocks, err := d.bytesInRange(hcl.Range{
					Filename: blockRef.RangePtr.Filename,
					Start:    blockRef.RangePtr.End,
					End:      b.Range.Start,
				})
				if err == nil && len(bytes.TrimSpace(betweenBlocks)) == 0 {
					blockRef.RangePtr.End = b.Range.End
				}
			}
		}
		refs = append(refs, blockRef)
	}

	for bType, bCollection := range bTypes.OfSchemaType(schema.BlockTypeMap) {
		blockAddr := append(addr.Copy(), lang.AttrStep{Name: bType})

		blockRef := reference.Target{
			Addr:        blockAddr,
			LocalAddr:   make(lang.Address, 0),
			ScopeId:     bAddrSchema.ScopeId,
			Type:        cty.Map(cty.Object(bodySchemaAsAttrTypes(bCollection.Schema.Body))),
			Description: bCollection.Schema.Description,
			RangePtr:    body.MissingItemRange().Ptr(),
		}
		if collectLocalAddr {
			blockRef.LocalAddr = append(selfRefAddr.Copy(), lang.AttrStep{Name: bType})
			blockRef.TargetableFromRangePtr = selfRefBodyRangePtr.Ptr()
		}

		for i, b := range bCollection.Blocks {
			elemAddr := append(blockAddr.Copy(), lang.IndexStep{
				Key: cty.StringVal(b.Labels[0]),
			})

			refType := cty.Object(bodySchemaAsAttrTypes(bCollection.Schema.Body))

			elemRef := reference.Target{
				Addr:        elemAddr,
				LocalAddr:   make(lang.Address, 0),
				ScopeId:     bAddrSchema.ScopeId,
				Type:        refType,
				Description: bCollection.Schema.Description,
				RangePtr:    b.Range.Ptr(),
				DefRangePtr: b.DefRange.Ptr(),
			}
			if collectLocalAddr {
				elemRef.LocalAddr = append(blockRef.LocalAddr.Copy(), lang.IndexStep{
					Key: cty.StringVal(b.Labels[0]),
				})
				elemRef.TargetableFromRangePtr = selfRefBodyRangePtr.Ptr()
			}

			elemRef.NestedTargets = d.collectInferredReferenceTargetsForBody(
				elemAddr, bAddrSchema, b.Body, bCollection.Schema.Body, selfRefBodyRangePtr, elemRef.LocalAddr)
			sort.Sort(elemRef.NestedTargets)
			blockRef.NestedTargets = append(blockRef.NestedTargets, elemRef)

			if i == 0 {
				blockRef.RangePtr = elemRef.RangePtr
			} else {
				// try to expand the range of the "parent" (map) reference
				// if the individual blocks follow each other
				betweenBlocks, err := d.bytesInRange(hcl.Range{
					Filename: blockRef.RangePtr.Filename,
					Start:    blockRef.RangePtr.End,
					End:      elemRef.RangePtr.Start,
				})
				if err == nil && len(bytes.TrimSpace(betweenBlocks)) == 0 {
					blockRef.RangePtr.End = elemRef.RangePtr.End
				}
			}
		}
		sort.Sort(blockRef.NestedTargets)
		refs = append(refs, blockRef)
	}

	return refs
}

type blockCollection struct {
	Schema *schema.BlockSchema
	Blocks []*ast.BlockContent
}

type blockTypes map[string]*blockCollection

func (bt blockTypes) OfSchemaType(t schema.BlockType) blockTypes {
	blockTypes := make(blockTypes, 0)
	for blockType, blockCollection := range bt {
		if blockCollection.Schema.Type == t {
			blockTypes[blockType] = blockCollection
		}
	}
	return blockTypes
}

func blocksTypesWithSchema(body hcl.Body, bodySchema *schema.BodySchema) blockTypes {
	blockTypes := make(blockTypes, 0)

	content := ast.DecodeBody(body, bodySchema)

	for _, block := range content.Blocks {
		bSchema, ok := bodySchema.Blocks[block.Type]
		if !ok {
			// skip unknown block
			continue
		}

		_, ok = blockTypes[block.Type]
		if !ok {
			blockTypes[block.Type] = &blockCollection{
				Schema: bSchema,
				Blocks: make([]*ast.BlockContent, 0),
			}
		}

		blockTypes[block.Type].Blocks = append(blockTypes[block.Type].Blocks, block)
	}

	return blockTypes
}

func (d *PathDecoder) bytesInRange(rng hcl.Range) ([]byte, error) {
	f, err := d.fileByName(rng.Filename)
	if err != nil {
		return nil, err
	}
	return rng.SliceBytes(f.Bytes), nil
}

func bodyToDataType(blockType schema.BlockType, body *schema.BodySchema) cty.Type {
	switch blockType {
	case schema.BlockTypeObject:
		return cty.Object(bodySchemaAsAttrTypes(body))
	case schema.BlockTypeList:
		return cty.List(cty.Object(bodySchemaAsAttrTypes(body)))
	case schema.BlockTypeMap:
		return cty.Map(cty.Object(bodySchemaAsAttrTypes(body)))
	case schema.BlockTypeSet:
		return cty.Set(cty.Object(bodySchemaAsAttrTypes(body)))
	}
	return cty.Object(bodySchemaAsAttrTypes(body))
}

func resolveAttributeAddress(attr *hcl.Attribute, addr schema.Address) (lang.Address, bool) {
	address := make(lang.Address, 0)

	if len(addr) == 0 {
		return lang.Address{}, false
	}

	for i, s := range addr {
		var stepName string

		switch step := s.(type) {
		case schema.StaticStep:
			stepName = step.Name
		case schema.AttrNameStep:
			stepName = attr.Name
		// TODO: AttrValueStep? Currently no use case for it
		default:
			// unknown step
			return lang.Address{}, false
		}

		if i == 0 {
			address = append(address, lang.RootStep{
				Name: stepName,
			})
			continue
		}
		address = append(address, lang.AttrStep{
			Name: stepName,
		})
	}

	return address, true
}

func resolveBlockAddress(block *hcl.Block, blockSchema *schema.BlockSchema) (lang.Address, bool) {
	address := make(lang.Address, 0)

	if blockSchema.Address == nil {
		// block not addressable
		return lang.Address{}, false
	}

	for i, s := range blockSchema.Address.Steps {
		var stepName string

		switch step := s.(type) {
		case schema.StaticStep:
			stepName = step.Name
		case schema.LabelStep:
			if len(block.Labels)-1 < int(step.Index) {
				// label not present
				if step.IsOptional {
					continue
				}
				return lang.Address{}, false
			}
			stepName = block.Labels[step.Index]
		case schema.AttrValueStep:
			content := ast.DecodeBody(block.Body, blockSchema.Body)

			attr, ok := content.Attributes[step.Name]
			if !ok && step.IsOptional {
				// skip step if not found and optional
				continue
			}

			if !ok {
				// attribute not present
				return lang.Address{}, false
			}
			if attr.Expr == nil {
				// empty attribute
				return lang.Address{}, false
			}
			val, _ := attr.Expr.Value(nil)
			if !val.IsWhollyKnown() {
				// unknown value
				return lang.Address{}, false
			}
			if val.Type() != cty.String {
				// non-string attributes are currently unsupported
				return lang.Address{}, false
			}
			stepName = val.AsString()
		default:
			// unknown step
			return lang.Address{}, false
		}

		if i == 0 {
			address = append(address, lang.RootStep{
				Name: stepName,
			})
			continue
		}
		address = append(address, lang.AttrStep{
			Name: stepName,
		})
	}

	return address, true
}
