package decoder

import (
	"bytes"
	"sort"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

type Reference lang.Reference

func (ref Reference) MatchesConstraint(te schema.TraversalExpr) bool {
	if te.OfScopeId != "" && te.OfScopeId != ref.ScopeId {
		return false
	}

	conformsToType := false
	if te.OfType != cty.NilType && ref.Type != cty.NilType {
		if errs := ref.Type.TestConformance(te.OfType); len(errs) == 0 {
			conformsToType = true
		}
	}

	return conformsToType || (te.OfType == cty.NilType && ref.Type == cty.NilType)
}

func (ref Reference) AddrMatchesTraversal(t hcl.Traversal) bool {
	addr, err := lang.TraversalToAddress(t)
	if err != nil {
		return false
	}

	rAddr := Address(ref.Addr)
	return rAddr.Equals(Address(addr))
}

type References lang.References

type RefWalkFunc func(lang.Reference)

func (refs References) DeepWalk(f RefWalkFunc) {
	for _, ref := range refs {
		f(ref)
		if len(ref.InsideReferences) > 0 {
			irefs := References(ref.InsideReferences)
			irefs.DeepWalk(f)
		}
	}
}

type RefMatchWalkFunc func(lang.Reference, bool)

func (refs References) MatchWalk(te schema.TraversalExpr, prefix string, f RefWalkFunc) {
	for _, ref := range refs {
		if strings.HasPrefix(ref.Addr.String(), string(prefix)) {
			nestedMatches := References(ref.InsideReferences).ContainsMatch(te, prefix)
			if Reference(ref).MatchesConstraint(te) || nestedMatches {
				f(ref)
				continue
			}
		}

		References(ref.InsideReferences).MatchWalk(te, prefix, f)
	}
}

func (refs References) ContainsMatch(te schema.TraversalExpr, prefix string) bool {
	for _, ref := range refs {
		if strings.HasPrefix(ref.Addr.String(), string(prefix)) &&
			Reference(ref).MatchesConstraint(te) {
			return true
		}
		if len(ref.InsideReferences) > 0 {
			if match := References(ref.InsideReferences).ContainsMatch(te, prefix); match {
				return true
			}
		}
	}
	return false
}

func (refs References) FirstTraversalMatch(expr hcl.Traversal, tSchema schema.TraversalExpr) (lang.Reference, error) {
	var matchingReference *lang.Reference

	refs.DeepWalk(func(r lang.Reference) {
		ref := Reference(r)
		if ref.AddrMatchesTraversal(expr) && ref.MatchesConstraint(tSchema) {
			matchingReference = &r
			return
		}
	})

	if matchingReference == nil {
		return lang.Reference{}, &NoReferenceFound{}
	}

	return *matchingReference, nil
}

type Address lang.Address

func (a Address) Equals(addr Address) bool {
	if len(a) != len(addr) {
		return false
	}
	for i, step := range a {
		if step.String() != addr[i].String() {
			return false
		}
	}

	return true
}

func (d *Decoder) DecodeReferences() (lang.References, error) {
	d.rootSchemaMu.RLock()
	defer d.rootSchemaMu.RUnlock()
	if d.rootSchema == nil {
		// unable to collect references without schema
		return nil, &NoSchemaError{}
	}

	refs := make(lang.References, 0)
	files := d.Filenames()
	for _, filename := range files {
		f, err := d.fileByName(filename)
		if err != nil {
			// skip unparseable file
			continue
		}

		body, ok := f.Body.(*hclsyntax.Body)
		if !ok {
			// skip JSON or other body format
			continue
		}

		refs = append(refs, d.decodeReferencesForBody(body, d.rootSchema)...)
	}

	return refs, nil
}

func (d *Decoder) decodeReferencesForBody(body *hclsyntax.Body, bodySchema *schema.BodySchema) lang.References {
	refs := make(lang.References, 0)

	if bodySchema == nil {
		return lang.References{}
	}

	for _, attr := range body.Attributes {
		attrSchema, ok := bodySchema.Attributes[attr.Name]
		if !ok {
			if bodySchema.AnyAttribute == nil {
				// unknown attribute (no schema)
				continue
			}
			attrSchema = bodySchema.AnyAttribute
		}

		refs = append(refs, decodeReferencesForAttribute(attr, attrSchema)...)
	}

	for _, block := range body.Blocks {
		bSchema, ok := bodySchema.Blocks[block.Type]
		if !ok {
			// unknown block (no schema)
			continue
		}

		iRefs := d.decodeReferencesForBody(block.Body, bSchema.Body)
		refs = append(refs, iRefs...)

		addr, ok := resolveBlockAddress(block, bSchema.Address)
		if !ok {
			// skip unresolvable address
			continue
		}

		if bSchema.Address.AsReference {
			ref := lang.Reference{
				Addr:     addr,
				ScopeId:  bSchema.Address.ScopeId,
				RangePtr: block.Range().Ptr(),
				Name:     bSchema.Address.FriendlyName,
			}
			refs = append(refs, ref)
		}

		if bSchema.Address.AsTypeOf != nil {
			refs = append(refs, referenceAsTypeOf(block, bSchema, addr)...)
		}

		var bodyRef lang.Reference

		if bSchema.Address.BodyAsData {
			bodyRef = lang.Reference{
				Addr:     addr,
				ScopeId:  bSchema.Address.ScopeId,
				RangePtr: block.Range().Ptr(),
			}

			if bSchema.Body != nil {
				bodyRef.Description = bSchema.Body.Description
			}

			if bSchema.Address.InferBody && bSchema.Body != nil {
				bodyRef.InsideReferences = append(bodyRef.InsideReferences,
					d.collectInferredReferencesForBody(addr, bSchema.Address.ScopeId, block.Body, bSchema.Body)...)
			}

			bodyRef.Type = bodyToDataType(bSchema.Type, bSchema.Body)

			refs = append(refs, bodyRef)
		}

		if bSchema.Address.DependentBodyAsData {
			if !bSchema.Address.BodyAsData {
				bodyRef = lang.Reference{
					Addr:     addr,
					ScopeId:  bSchema.Address.ScopeId,
					RangePtr: block.Range().Ptr(),
				}
			}

			depSchema, _, ok := NewBlockSchema(bSchema).DependentBodySchema(block)
			if ok {
				fullSchema := depSchema
				if bSchema.Address.BodyAsData {
					mergedSchema, err := mergeBlockBodySchemas(block, bSchema)
					if err != nil {
						continue
					}
					bodyRef.InsideReferences = make(lang.References, 0)
					fullSchema = mergedSchema
				}

				bodyRef.Type = bodyToDataType(bSchema.Type, fullSchema)

				if bSchema.Address.InferDependentBody && len(bSchema.DependentBody) > 0 {
					bodyRef.InsideReferences = append(bodyRef.InsideReferences,
						d.collectInferredReferencesForBody(addr, bSchema.Address.ScopeId, block.Body, fullSchema)...)
				}

				if !bSchema.Address.BodyAsData {
					refs = append(refs, bodyRef)
				}
			}
		}

		sort.Sort(bodyRef.InsideReferences)
	}

	sort.Sort(refs)

	return refs
}

func decodeReferencesForAttribute(attr *hclsyntax.Attribute, attrSchema *schema.AttributeSchema) lang.References {
	refs := make(lang.References, 0)

	attrAddr, ok := resolveAttributeAddress(attr, attrSchema.Address)
	if ok {
		if attrSchema.Address.AsReference {
			ref := lang.Reference{
				Addr:     attrAddr,
				ScopeId:  attrSchema.Address.ScopeId,
				RangePtr: attr.SrcRange.Ptr(),
				Name:     attrSchema.Address.FriendlyName,
			}
			refs = append(refs, ref)
		}

		if attrSchema.Address.AsExprType {
			t, ok := exprConstraintToDataType(attrSchema.Expr)
			if ok {
				if t == cty.DynamicPseudoType && attr.Expr != nil {
					// attempt to make the type more specific
					exprVal, diags := attr.Expr.Value(nil)
					if !diags.HasErrors() {
						t = exprVal.Type()
					}
				}

				scopeId := attrSchema.Address.ScopeId

				ref := lang.Reference{
					Addr:     attrAddr,
					Type:     t,
					ScopeId:  scopeId,
					RangePtr: attr.SrcRange.Ptr(),
					Name:     attrSchema.Address.FriendlyName,
				}

				if attr.Expr != nil && !t.IsPrimitiveType() {
					ref.InsideReferences = make(lang.References, 0)
					ref.InsideReferences = append(ref.InsideReferences, decodeReferencesForComplexTypeExpr(attrAddr, attr.Expr, t, scopeId)...)
				}

				refs = append(refs, ref)
			}
		}
	}

	ec := ExprConstraints(attrSchema.Expr)
	refs = append(refs, referencesForExpr(attr.Expr, ec)...)
	return refs
}

func decodeReferencesForComplexTypeExpr(addr lang.Address, expr hclsyntax.Expression, t cty.Type, scopeId lang.ScopeId) lang.References {
	refs := make(lang.References, 0)

	if expr == nil {
		return refs
	}

	switch e := expr.(type) {
	case *hclsyntax.TupleConsExpr:
		if t.IsListType() {
			for i, item := range e.Exprs {
				elemAddr := append(addr, lang.IndexStep{Key: cty.NumberIntVal(int64(i))})
				elemType := t.ElementType()

				ref := lang.Reference{
					Addr:     elemAddr,
					Type:     elemType,
					ScopeId:  scopeId,
					RangePtr: item.Range().Ptr(),
				}
				if !elemType.IsPrimitiveType() {
					ref.InsideReferences = make(lang.References, 0)
					ref.InsideReferences = append(refs, decodeReferencesForComplexTypeExpr(elemAddr, item, elemType, scopeId)...)
				}

				refs = append(refs, ref)
			}
		}
	case *hclsyntax.ObjectConsExpr:
		if t.IsObjectType() {
			for _, item := range e.Items {
				key, _ := item.KeyExpr.Value(nil)
				if key.IsNull() || !key.IsWhollyKnown() || key.Type() != cty.String {
					// skip items keys that can't be interpolated
					// without further context
					continue
				}
				attrType, ok := t.AttributeTypes()[key.AsString()]
				if !ok {
					continue
				}
				attrAddr := append(addr, lang.AttrStep{Name: key.AsString()})
				rng := hcl.RangeBetween(item.KeyExpr.Range(), item.ValueExpr.Range())

				ref := lang.Reference{
					Addr:     attrAddr,
					Type:     attrType,
					ScopeId:  scopeId,
					RangePtr: rng.Ptr(),
				}
				if !attrType.IsPrimitiveType() {
					ref.InsideReferences = make(lang.References, 0)
					ref.InsideReferences = append(refs, decodeReferencesForComplexTypeExpr(attrAddr, item.ValueExpr, attrType, scopeId)...)
				}

				refs = append(refs, ref)
			}
		}
		if t.IsMapType() {
			for _, item := range e.Items {
				key, _ := item.KeyExpr.Value(nil)
				if key.IsNull() || !key.IsWhollyKnown() || key.Type() != cty.String {
					// skip items keys that can't be interpolated
					// without further context
					continue
				}
				elemTypePtr := t.MapElementType()
				if elemTypePtr == nil {
					continue
				}
				elemType := *elemTypePtr

				elemAddr := append(addr, lang.IndexStep{Key: key})
				rng := hcl.RangeBetween(item.KeyExpr.Range(), item.ValueExpr.Range())

				ref := lang.Reference{
					Addr:     elemAddr,
					Type:     elemType,
					ScopeId:  scopeId,
					RangePtr: rng.Ptr(),
				}
				if !elemType.IsPrimitiveType() {
					ref.InsideReferences = make(lang.References, 0)
					ref.InsideReferences = append(refs, decodeReferencesForComplexTypeExpr(elemAddr, item.ValueExpr, elemType, scopeId)...)
				}

				refs = append(refs, ref)
			}
		}
	}

	return refs
}

func referenceAsTypeOf(block *hclsyntax.Block, bSchema *schema.BlockSchema, addr lang.Address) lang.References {
	ref := lang.Reference{
		Addr:     addr,
		ScopeId:  bSchema.Address.ScopeId,
		RangePtr: block.Range().Ptr(),
		Type:     cty.DynamicPseudoType,
	}

	if bSchema.Body != nil {
		ref.Description = bSchema.Body.Description
	}

	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return lang.References{ref}
	}

	if bSchema.Address.AsTypeOf.AttributeExpr != "" {
		typeDecl, ok := asTypeOfAttrExpr(attrs, bSchema)
		if !ok && bSchema.Address.AsTypeOf.AttributeValue == "" {
			// nothing to fall back to, exit early
			return lang.References{ref}
		}
		ref.Type = typeDecl
	}

	if bSchema.Address.AsTypeOf.AttributeValue != "" {
		attr, ok := attrs[bSchema.Address.AsTypeOf.AttributeValue]
		if !ok {
			return lang.References{ref}
		}
		value, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return lang.References{ref}
		}
		val, err := convert.Convert(value, ref.Type)
		if err != nil {
			// type does not comply with type constraint
			return lang.References{ref}
		}
		ref.Type = val.Type()
	}

	return lang.References{ref}
}

func asTypeOfAttrExpr(attrs hcl.Attributes, bSchema *schema.BlockSchema) (cty.Type, bool) {
	attrName := bSchema.Address.AsTypeOf.AttributeExpr
	attr, ok := attrs[attrName]
	if !ok {
		return cty.DynamicPseudoType, false
	}

	ec := ExprConstraints(bSchema.Body.Attributes[attrName].Expr)
	_, ok = ec.TypeDeclarationExpr()
	if !ok {
		return cty.DynamicPseudoType, false
	}

	typeDecl, diags := typeexpr.TypeConstraint(attr.Expr)
	if diags.HasErrors() {
		return cty.DynamicPseudoType, false
	}

	return typeDecl, true
}

func exprConstraintToDataType(expr schema.ExprConstraints) (cty.Type, bool) {
	ec := ExprConstraints(expr)

	lt, ok := ec.LiteralType()
	if ok {
		return lt, true
	}

	le, ok := ec.ListExpr()
	if ok {
		elemType, elemOk := exprConstraintToDataType(le.Elem)
		if elemOk {
			return cty.List(elemType), true
		}
	}

	se, ok := ec.SetExpr()
	if ok {
		elemType, elemOk := exprConstraintToDataType(se.Elem)
		if elemOk {
			return cty.Set(elemType), true
		}
	}

	te, ok := ec.TupleExpr()
	if ok {
		elems := make([]cty.Type, len(te.Elems))
		elemsOk := true
		for i, elem := range te.Elems {
			elemType, ok := exprConstraintToDataType(elem)
			if ok {
				elems[i] = elemType
			} else {
				elemsOk = false
				break
			}
		}
		if elemsOk {
			return cty.Tuple(elems), true
		}
	}

	oe, ok := ec.ObjectExpr()
	if ok {
		attributes := make(map[string]cty.Type, 0)
		for name, attr := range oe.Attributes {
			attrType, ok := exprConstraintToDataType(attr.Expr)
			if ok {
				attributes[name] = attrType
			}
		}
		return cty.Object(attributes), true
	}

	me, ok := ec.MapExpr()
	if ok {
		elemType, elemOk := exprConstraintToDataType(me.Elem)
		if elemOk {
			return cty.Map(elemType), true
		}
	}

	return cty.NilType, false
}

func referencesForExpr(expr hcl.Expression, ec ExprConstraints) lang.References {
	refs := make(lang.References, 0)

	switch e := expr.(type) {
	// TODO: Support all expression types (list/set/map literals)
	case *hclsyntax.ScopeTraversalExpr:
		te, ok := ec.TraversalExpr()
		if !ok {
			// unknown traversal
			return lang.References{}
		}
		if te.Address == nil {
			// traversal not addressable
			return lang.References{}
		}

		addr, err := lang.TraversalToAddress(e.AsTraversal())
		if err != nil {
			return lang.References{}
		}
		refs = append(refs, lang.Reference{
			Addr:     addr,
			ScopeId:  te.Address.ScopeId,
			RangePtr: e.SrcRange.Ptr(),
			Name:     te.Name,
		})
	case *hclsyntax.ObjectConsExpr:
		oe, ok := ec.ObjectExpr()
		if ok {
			for _, item := range e.Items {
				key, _ := item.KeyExpr.Value(nil)
				if key.IsNull() || !key.IsWhollyKnown() || key.Type() != cty.String {
					// skip items keys that can't be interpolated
					// without further context
					continue
				}
				attr, ok := oe.Attributes[key.AsString()]
				if !ok {
					continue
				}

				refs = append(refs, referencesForExpr(item.ValueExpr, ExprConstraints(attr.Expr))...)
			}
		}
		me, ok := ec.MapExpr()
		if ok {
			for _, item := range e.Items {
				refs = append(refs, referencesForExpr(item.ValueExpr, ExprConstraints(me.Elem))...)
			}
		}
	case *hclsyntax.TupleConsExpr:
		le, ok := ec.ListExpr()
		if ok {
			for _, itemExpr := range e.Exprs {
				refs = append(refs, referencesForExpr(itemExpr, ExprConstraints(le.Elem))...)
			}
		}
		se, ok := ec.SetExpr()
		if ok {
			for _, itemExpr := range e.Exprs {
				refs = append(refs, referencesForExpr(itemExpr, ExprConstraints(se.Elem))...)
			}
		}
		te, ok := ec.TupleExpr()
		if ok {
			for i, itemExpr := range e.Exprs {
				if i >= len(te.Elems) {
					break
				}
				refs = append(refs, referencesForExpr(itemExpr, ExprConstraints(te.Elems[i]))...)
			}
		}
		tce, ok := ec.TupleConsExpr()
		if ok {
			for _, itemExpr := range e.Exprs {
				refs = append(refs, referencesForExpr(itemExpr, ExprConstraints(tce.AnyElem))...)
			}
		}
	}

	return refs
}

func bodySchemaAsAttrTypes(bodySchema *schema.BodySchema) map[string]cty.Type {
	attrTypes := make(map[string]cty.Type, 0)

	if bodySchema == nil {
		return attrTypes
	}

	for name, attr := range bodySchema.Attributes {
		attrType, ok := exprConstraintToDataType(attr.Expr)
		if ok {
			attrTypes[name] = attrType
		}
	}

	for name, block := range bodySchema.Blocks {
		attrTypes[name] = bodyToDataType(block.Type, block.Body)
	}

	return attrTypes
}

func (d *Decoder) collectInferredReferencesForBody(addr lang.Address, scopeId lang.ScopeId, body *hclsyntax.Body, bodySchema *schema.BodySchema) lang.References {
	refs := make(lang.References, 0)

	for name, aSchema := range bodySchema.Attributes {
		attrType, ok := exprConstraintToDataType(aSchema.Expr)
		if !ok {
			// unknown type
			continue
		}

		attrAddr := make(lang.Address, len(addr))
		copy(attrAddr, addr)
		attrAddr = append(attrAddr, lang.AttrStep{Name: name})

		ref := lang.Reference{
			Addr:        attrAddr,
			ScopeId:     scopeId,
			Type:        attrType,
			Description: aSchema.Description,
			RangePtr:    body.EndRange.Ptr(),
		}

		var attrExpr hclsyntax.Expression
		if body != nil {
			if attr, ok := body.Attributes[name]; ok {
				ref.RangePtr = attr.Range().Ptr()
				attrExpr = attr.Expr
			}
		}

		if attrExpr != nil && !attrType.IsPrimitiveType() {
			ref.InsideReferences = make(lang.References, 0)
			ref.InsideReferences = append(ref.InsideReferences, decodeReferencesForComplexTypeExpr(attrAddr, attrExpr, attrType, scopeId)...)
		}

		refs = append(refs, ref)
	}

	objectBlocks := make(map[string]*hclsyntax.Block, 0)
	listBlocks := make(map[string][]*hclsyntax.Block, 0)
	setBlocks := make(map[string][]*hclsyntax.Block, 0)
	mapBlocks := make(map[string][]*hclsyntax.Block, 0)

	for _, block := range body.Blocks {
		bSchema, ok := bodySchema.Blocks[block.Type]
		if !ok {
			// skip unknown block
			continue
		}

		switch bSchema.Type {
		case schema.BlockTypeObject:
			_, ok := objectBlocks[block.Type]
			if ok {
				// objects are expected to be singletons
				continue
			}
			objectBlocks[block.Type] = block
		case schema.BlockTypeList:
			if bSchema.MaxItems > 0 && uint64(len(listBlocks[block.Type])) >= bSchema.MaxItems {
				// skip item if limit was reached
				continue
			}

			_, ok := listBlocks[block.Type]
			if !ok {
				listBlocks[block.Type] = make([]*hclsyntax.Block, 0)
			}
			listBlocks[block.Type] = append(listBlocks[block.Type], block)
		case schema.BlockTypeSet:
			if bSchema.MaxItems > 0 && uint64(len(setBlocks[block.Type])) >= bSchema.MaxItems {
				// skip item if limit was reached
				continue
			}

			_, ok := setBlocks[block.Type]
			if !ok {
				setBlocks[block.Type] = make([]*hclsyntax.Block, 0)
			}
			setBlocks[block.Type] = append(setBlocks[block.Type], block)
		case schema.BlockTypeMap:
			if len(block.Labels) != 1 {
				// this should never happen
				continue
			}
			if bSchema.MaxItems > 0 && uint64(len(listBlocks[block.Type])) >= bSchema.MaxItems {
				// skip item if limit was reached
				continue
			}

			_, ok := mapBlocks[block.Type]
			if !ok {
				mapBlocks[block.Type] = make([]*hclsyntax.Block, 0)
			}
			mapBlocks[block.Type] = append(mapBlocks[block.Type], block)
		}
	}

	for blockType, block := range objectBlocks {
		bSchema, ok := bodySchema.Blocks[blockType]
		if !ok {
			// skip unknown block
			continue
		}

		blockAddr := make(lang.Address, len(addr))
		copy(blockAddr, addr)
		blockAddr = append(blockAddr, lang.AttrStep{Name: blockType})

		blockRef := lang.Reference{
			Addr:        blockAddr,
			ScopeId:     scopeId,
			Type:        cty.Object(bodySchemaAsAttrTypes(bSchema.Body)),
			Description: bSchema.Description,
			RangePtr:    block.Range().Ptr(),
			InsideReferences: d.collectInferredReferencesForBody(
				blockAddr, scopeId, block.Body, bSchema.Body),
		}
		sort.Sort(blockRef.InsideReferences)
		refs = append(refs, blockRef)
	}

	for blockType, blocks := range listBlocks {
		bSchema, ok := bodySchema.Blocks[blockType]
		if !ok {
			// skip unknown block
			continue
		}

		blockAddr := make(lang.Address, len(addr))
		copy(blockAddr, addr)
		blockAddr = append(blockAddr, lang.AttrStep{Name: blockType})

		blockRef := lang.Reference{
			Addr:        blockAddr,
			ScopeId:     scopeId,
			Type:        cty.List(cty.Object(bodySchemaAsAttrTypes(bSchema.Body))),
			Description: bSchema.Description,
			RangePtr:    body.MissingItemRange().Ptr(),
		}

		for i, block := range blocks {
			elemAddr := make(lang.Address, len(blockAddr))
			copy(elemAddr, blockAddr)
			elemAddr = append(elemAddr, lang.IndexStep{
				Key: cty.NumberIntVal(int64(i)),
			})

			elemRef := lang.Reference{
				Addr:        elemAddr,
				ScopeId:     scopeId,
				Type:        cty.Object(bodySchemaAsAttrTypes(bSchema.Body)),
				Description: bSchema.Description,
				RangePtr:    block.Range().Ptr(),
				InsideReferences: d.collectInferredReferencesForBody(
					elemAddr, scopeId, block.Body, bSchema.Body),
			}
			sort.Sort(elemRef.InsideReferences)
			blockRef.InsideReferences = append(blockRef.InsideReferences, elemRef)

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

		sort.Sort(blockRef.InsideReferences)
		refs = append(refs, blockRef)
	}

	for blockType, blocks := range setBlocks {
		bSchema, ok := bodySchema.Blocks[blockType]
		if !ok {
			// skip unknown block
			continue
		}

		blockAddr := make(lang.Address, len(addr))
		copy(blockAddr, addr)
		blockAddr = append(blockAddr, lang.AttrStep{Name: blockType})

		blockRef := lang.Reference{
			Addr:        blockAddr,
			ScopeId:     scopeId,
			Type:        cty.Set(cty.Object(bodySchemaAsAttrTypes(bSchema.Body))),
			Description: bSchema.Description,
			RangePtr:    body.MissingItemRange().Ptr(),
		}

		for i, block := range blocks {
			if i == 0 {
				blockRef.RangePtr = block.Range().Ptr()
			} else {
				// try to expand the range of the "parent" (set) reference
				// if the individual blocks follow each other
				betweenBlocks, err := d.bytesInRange(hcl.Range{
					Filename: blockRef.RangePtr.Filename,
					Start:    blockRef.RangePtr.End,
					End:      block.Range().Start,
				})
				if err == nil && len(bytes.TrimSpace(betweenBlocks)) == 0 {
					blockRef.RangePtr.End = block.Range().End
				}
			}
		}

		refs = append(refs, blockRef)
	}

	for blockType, blocks := range mapBlocks {
		bSchema, ok := bodySchema.Blocks[blockType]
		if !ok {
			// skip unknown block
			continue
		}

		blockAddr := make(lang.Address, len(addr))
		copy(blockAddr, addr)
		blockAddr = append(blockAddr, lang.AttrStep{Name: blockType})

		blockRef := lang.Reference{
			Addr:        blockAddr,
			ScopeId:     scopeId,
			Type:        cty.Map(cty.Object(bodySchemaAsAttrTypes(bSchema.Body))),
			Description: bSchema.Description,
			RangePtr:    body.MissingItemRange().Ptr(),
		}

		for i, block := range blocks {
			elemAddr := make(lang.Address, len(blockAddr))
			copy(elemAddr, blockAddr)
			elemAddr = append(elemAddr, lang.IndexStep{
				Key: cty.StringVal(block.Labels[0]),
			})

			refType := cty.Object(bodySchemaAsAttrTypes(bSchema.Body))

			elemRef := lang.Reference{
				Addr:        elemAddr,
				ScopeId:     scopeId,
				Type:        refType,
				Description: bSchema.Description,
				RangePtr:    block.Range().Ptr(),
				InsideReferences: d.collectInferredReferencesForBody(
					elemAddr, scopeId, block.Body, bSchema.Body),
			}
			sort.Sort(elemRef.InsideReferences)
			blockRef.InsideReferences = append(blockRef.InsideReferences, elemRef)

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

		sort.Sort(blockRef.InsideReferences)
		refs = append(refs, blockRef)
	}

	return refs
}

func (d *Decoder) bytesInRange(rng hcl.Range) ([]byte, error) {
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

func resolveAttributeAddress(attr *hclsyntax.Attribute, addr *schema.AttributeAddrSchema) (lang.Address, bool) {
	address := make(lang.Address, 0)

	if addr == nil {
		// attribute not addressable
		return lang.Address{}, false
	}

	for i, s := range addr.Steps {
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

func resolveBlockAddress(block *hclsyntax.Block, addr *schema.BlockAddrSchema) (lang.Address, bool) {
	address := make(lang.Address, 0)

	if addr == nil {
		// block not addressable
		return lang.Address{}, false
	}

	for i, s := range addr.Steps {
		var stepName string

		switch step := s.(type) {
		case schema.StaticStep:
			stepName = step.Name
		case schema.LabelStep:
			if uint(len(block.Labels)-1) < step.Index {
				// label not present
				return lang.Address{}, false
			}
			stepName = block.Labels[step.Index]
		case schema.AttrValueStep:
			attr, ok := block.Body.Attributes[step.Name]
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
