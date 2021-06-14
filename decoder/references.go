package decoder

import (
	"sort"

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

func (refs References) Walk(f RefWalkFunc) {
	for _, ref := range refs {
		f(ref)
		if len(ref.InsideReferences) > 0 {
			irefs := References(ref.InsideReferences)
			irefs.Walk(f)
		}
	}
}

func (refs References) FirstTraversalMatch(expr hcl.Traversal, tSchema schema.TraversalExpr) (lang.Reference, error) {
	var matchingReference *lang.Reference

	refs.Walk(func(r lang.Reference) {
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

		refs = append(refs, decodeReferencesForBody(body, d.rootSchema)...)
	}

	return refs, nil
}

func decodeReferencesForBody(body *hclsyntax.Body, bodySchema *schema.BodySchema) lang.References {
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

					ref := lang.Reference{
						Addr:     attrAddr,
						Type:     t,
						ScopeId:  attrSchema.Address.ScopeId,
						RangePtr: attr.SrcRange.Ptr(),
						Name:     attrSchema.Address.FriendlyName,
					}
					refs = append(refs, ref)
				}
			}
		}

		ec := ExprConstraints(attrSchema.Expr)
		refs = append(refs, referencesForExpr(attr.Expr, ec)...)
	}

	for _, block := range body.Blocks {
		bSchema, ok := bodySchema.Blocks[block.Type]
		if !ok {
			// unknown block (no schema)
			continue
		}

		iRefs := decodeReferencesForBody(block.Body, bSchema.Body)
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
					collectInferredReferencesForBody(addr, bSchema.Address.ScopeId, block.Body, bSchema.Body)...)
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
						collectInferredReferencesForBody(addr, bSchema.Address.ScopeId, block.Body, fullSchema)...)
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
	// TODO: Support all expression types
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

func collectInferredReferencesForBody(addr lang.Address, scopeId lang.ScopeId, body *hclsyntax.Body, bodySchema *schema.BodySchema) lang.References {
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

		if body != nil {
			if attr, ok := body.Attributes[name]; ok {
				ref.RangePtr = attr.Range().Ptr()
			}
		}

		refs = append(refs, ref)
	}

	for name, bSchema := range bodySchema.Blocks {
		blockAddr := make(lang.Address, len(addr))
		copy(blockAddr, addr)
		blockAddr = append(blockAddr, lang.AttrStep{Name: name})

		ref := lang.Reference{
			Addr:        blockAddr,
			ScopeId:     scopeId,
			Type:        bodyToDataType(bSchema.Type, bSchema.Body),
			Description: bSchema.Description,
			RangePtr:    body.EndRange.Ptr(),
		}

		if body != nil {
			for i, block := range body.Blocks {
				if name == block.Type {
					switch bSchema.Type {
					case schema.BlockTypeObject:
						ref.RangePtr = block.Range().Ptr()
						insideRefs := collectInferredReferencesForBody(blockAddr, scopeId, block.Body, bSchema.Body)
						ref.InsideReferences = append(ref.InsideReferences, insideRefs...)
						break
					case schema.BlockTypeList:
						elemAddr := make(lang.Address, len(blockAddr))
						copy(elemAddr, blockAddr)
						elemAddr = append(elemAddr, lang.IndexStep{
							Key: cty.NumberIntVal(int64(i)),
						})
						insideRefs := collectInferredReferencesForBody(elemAddr, scopeId, block.Body, bSchema.Body)
						ref.InsideReferences = append(ref.InsideReferences, insideRefs...)

					case schema.BlockTypeMap:
						if len(block.Labels) != 1 {
							// this should never happen
							continue
						}
						elemAddr := make(lang.Address, len(blockAddr))
						copy(elemAddr, blockAddr)
						elemAddr = append(elemAddr, lang.IndexStep{
							Key: cty.StringVal(block.Labels[0]),
						})
						insideRefs := collectInferredReferencesForBody(elemAddr, scopeId, block.Body, bSchema.Body)
						ref.InsideReferences = append(ref.InsideReferences, insideRefs...)
					}
				}
			}
		}

		sort.Sort(ref.InsideReferences)

		refs = append(refs, ref)
	}

	return refs
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
