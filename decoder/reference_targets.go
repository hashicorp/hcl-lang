package decoder

import (
	"bytes"
	"errors"
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

// ReferenceTargetForOrigin returns the first ReferenceTarget
// with matching ReferenceOrigin Address, if one exists, else nil
func (d *Decoder) ReferenceTargetForOrigin(refOrigin lang.ReferenceOrigin) (*lang.ReferenceTarget, error) {
	if d.refTargetReader == nil {
		return nil, nil
	}

	allTargets := ReferenceTargets(d.refTargetReader())

	ref, err := allTargets.FirstTargetableBy(refOrigin)
	if err != nil {
		if _, ok := err.(*NoRefTargetFound); ok {
			return nil, nil
		}
		return nil, err
	}

	return &ref, nil
}

func (d *Decoder) ReferenceTargetsInFile(file string) (lang.ReferenceTargets, error) {
	if d.refTargetReader == nil {
		return nil, nil
	}

	allTargets := ReferenceTargets(d.refTargetReader())

	targets := make(lang.ReferenceTargets, 0)

	// It is practically impossible for nested targets to be placed
	// in a separate file from their parent target, so we save
	// some cycles here by limiting walk just to the top level.
	depth := 0

	allTargets.DeepWalk(func(target lang.ReferenceTarget) error {
		if target.RangePtr == nil {
			return nil
		}
		if target.RangePtr.Filename == file {
			targets = append(targets, target)
		}
		return nil
	}, depth)

	return targets, nil
}

func (d *Decoder) OutermostReferenceTargetsAtPos(file string, pos hcl.Pos) (lang.ReferenceTargets, error) {
	if d.refTargetReader == nil {
		return nil, nil
	}

	allTargets := ReferenceTargets(d.refTargetReader())

	matchingTargets := make(lang.ReferenceTargets, 0)
	for _, target := range allTargets {
		if target.RangePtr == nil {
			continue
		}
		if target.RangePtr.Filename != file {
			continue
		}
		if target.RangePtr.ContainsPos(pos) {
			matchingTargets = append(matchingTargets, target)
		}
	}

	return matchingTargets, nil
}

func (d *Decoder) InnermostReferenceTargetsAtPos(file string, pos hcl.Pos) (lang.ReferenceTargets, error) {
	if d.refTargetReader == nil {
		return nil, nil
	}

	targets, _ := d.innermostReferenceTargetsAtPos(d.refTargetReader(), file, pos)

	return targets, nil
}

func (d *Decoder) innermostReferenceTargetsAtPos(targets lang.ReferenceTargets, file string, pos hcl.Pos) (lang.ReferenceTargets, bool) {
	allTargets := ReferenceTargets(targets)

	matchingTargets := make(lang.ReferenceTargets, 0)

	for _, target := range allTargets {
		if target.RangePtr == nil {
			continue
		}
		if target.RangePtr.Filename != file {
			continue
		}
		if target.RangePtr.ContainsPos(pos) {
			matchingTargets = append(matchingTargets, target)
		}
	}

	var innermostTargets lang.ReferenceTargets

	for _, target := range matchingTargets {
		if target.DefRangePtr != nil {
			if target.DefRangePtr.Filename == file &&
				target.DefRangePtr.ContainsPos(pos) {
				innermostTargets = append(innermostTargets, target)
				continue
			}
		}

		nestedTargets, ok := d.innermostReferenceTargetsAtPos(target.NestedTargets, file, pos)
		if ok {
			innermostTargets = nestedTargets
			continue
		}

		innermostTargets = append(innermostTargets, target)
	}

	return innermostTargets, len(innermostTargets) > 0
}

type ReferenceTarget lang.ReferenceTarget

func (ref ReferenceTarget) MatchesConstraint(te schema.TraversalExpr) bool {
	return ref.MatchesScopeId(te.OfScopeId) && ref.ConformsToType(te.OfType)
}

func (ref ReferenceTarget) MatchesScopeId(scopeId lang.ScopeId) bool {
	return scopeId == "" || ref.ScopeId == scopeId
}

func (ref ReferenceTarget) ConformsToType(typ cty.Type) bool {
	conformsToType := false
	if typ != cty.NilType && ref.Type != cty.NilType {
		if ref.Type == cty.DynamicPseudoType {
			// anything conforms with dynamic
			conformsToType = true
		}
		if errs := ref.Type.TestConformance(typ); len(errs) == 0 {
			conformsToType = true
		}
	}

	return conformsToType || (typ == cty.NilType && ref.Type == cty.NilType)
}

func (target ReferenceTarget) IsTargetableBy(origin lang.ReferenceOrigin) bool {
	if len(target.Addr) > len(origin.Addr) {
		return false
	}

	originAddr := Address(origin.Addr)

	matchesCons := false

	if len(origin.Constraints) == 0 && target.Type != cty.NilType {
		matchesCons = true
	}

	for _, cons := range origin.Constraints {
		if !target.MatchesScopeId(cons.OfScopeId) {
			continue
		}

		if target.Type == cty.DynamicPseudoType {
			originAddr = Address(origin.Addr).FirstSteps(uint(len(target.Addr)))
			matchesCons = true
			continue
		}
		if cons.OfType != cty.NilType && target.ConformsToType(cons.OfType) {
			matchesCons = true
		}
		if cons.OfType == cty.NilType && target.Type == cty.NilType {
			// This just simplifies testing
			matchesCons = true
		}
	}

	return Address(target.Addr).Equals(originAddr) && matchesCons
}

type ReferenceTargets lang.ReferenceTargets

type RefTargetWalkFunc func(lang.ReferenceTarget) error

var StopWalking error = errors.New("stop walking")

const InfiniteDepth = -1

func (refs ReferenceTargets) DeepWalk(f RefTargetWalkFunc, depth int) {
	w := refTargetDeepWalker{
		WalkFunc: f,
		Depth:    depth,
	}
	w.Walk(refs)
}

type refTargetDeepWalker struct {
	WalkFunc RefTargetWalkFunc
	Depth    int

	currentDepth int
}

func (w refTargetDeepWalker) Walk(refTargets ReferenceTargets) {
	for _, ref := range refTargets {
		err := w.WalkFunc(ref)
		if err == StopWalking {
			return
		}

		if len(ref.NestedTargets) > 0 && (w.Depth == InfiniteDepth || w.Depth > w.currentDepth) {
			irefs := ReferenceTargets(ref.NestedTargets)
			w.currentDepth++
			w.Walk(irefs)
			w.currentDepth--
		}
	}
}

func (refs ReferenceTargets) MatchWalk(te schema.TraversalExpr, prefix string, f RefTargetWalkFunc) {
	for _, ref := range refs {
		if strings.HasPrefix(ref.Addr.String(), string(prefix)) {
			nestedMatches := ReferenceTargets(ref.NestedTargets).ContainsMatch(te, prefix)
			if ReferenceTarget(ref).MatchesConstraint(te) || nestedMatches {
				f(ref)
				continue
			}
		}

		ReferenceTargets(ref.NestedTargets).MatchWalk(te, prefix, f)
	}
}

func (refs ReferenceTargets) ContainsMatch(te schema.TraversalExpr, prefix string) bool {
	for _, ref := range refs {
		if strings.HasPrefix(ref.Addr.String(), string(prefix)) &&
			ReferenceTarget(ref).MatchesConstraint(te) {
			return true
		}
		if len(ref.NestedTargets) > 0 {
			if match := ReferenceTargets(ref.NestedTargets).ContainsMatch(te, prefix); match {
				return true
			}
		}
	}
	return false
}

func (refs ReferenceTargets) FirstTargetableBy(origin lang.ReferenceOrigin) (lang.ReferenceTarget, error) {
	var matchingReference *lang.ReferenceTarget

	refs.DeepWalk(func(ref lang.ReferenceTarget) error {
		if ReferenceTarget(ref).IsTargetableBy(origin) {
			matchingReference = &ref
			return StopWalking
		}
		return nil
	}, InfiniteDepth)

	if matchingReference == nil {
		return lang.ReferenceTarget{}, &NoRefTargetFound{}
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

func (a Address) FirstSteps(steps uint) Address {
	return a[0:steps]
}

func (d *Decoder) CollectReferenceTargets() (lang.ReferenceTargets, error) {
	d.rootSchemaMu.RLock()
	defer d.rootSchemaMu.RUnlock()
	if d.rootSchema == nil {
		// unable to collect reference targets without schema
		return nil, &NoSchemaError{}
	}

	refs := make(lang.ReferenceTargets, 0)
	files := d.Filenames()
	for _, filename := range files {
		f, err := d.fileByName(filename)
		if err != nil {
			// skip unparseable file
			continue
		}
		refs = append(refs, d.decodeReferenceTargetsForBody(f.Body, nil, d.rootSchema)...)
	}

	return refs, nil
}

func (d *Decoder) decodeReferenceTargetsForBody(body hcl.Body, parentBlock *blockContent, bodySchema *schema.BodySchema) lang.ReferenceTargets {
	refs := make(lang.ReferenceTargets, 0)

	if bodySchema == nil {
		return lang.ReferenceTargets{}
	}

	content := decodeBody(body, bodySchema)

	for _, attr := range content.Attributes {
		attrSchema, ok := bodySchema.Attributes[attr.Name]
		if !ok {
			if bodySchema.AnyAttribute == nil {
				// unknown attribute (no schema)
				continue
			}
			attrSchema = bodySchema.AnyAttribute
		}

		refs = append(refs, decodeReferenceTargetsForAttribute(attr, attrSchema)...)
	}

	for _, blk := range content.Blocks {
		bSchema, ok := bodySchema.Blocks[blk.Type]
		if !ok {
			// unknown block (no schema)
			continue
		}

		mergedSchema, err := mergeBlockBodySchemas(blk.Block, bSchema)
		if err != nil {
			mergedSchema = bSchema.Body
		}

		iRefs := d.decodeReferenceTargetsForBody(blk.Body, blk, mergedSchema)
		refs = append(refs, iRefs...)

		addr, ok := resolveBlockAddress(blk.Block, bSchema)
		if !ok {
			// skip unresolvable address
			continue
		}

		if bSchema.Address.AsReference {
			ref := lang.ReferenceTarget{
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

		var bodyRef lang.ReferenceTarget

		if bSchema.Address.BodyAsData {
			bodyRef = lang.ReferenceTarget{
				Addr:        addr,
				ScopeId:     bSchema.Address.ScopeId,
				DefRangePtr: blk.DefRange.Ptr(),
				RangePtr:    blk.Range.Ptr(),
			}

			if bSchema.Body != nil {
				bodyRef.Description = bSchema.Body.Description
			}

			if bSchema.Address.InferBody && bSchema.Body != nil {
				bodyRef.NestedTargets = append(bodyRef.NestedTargets,
					d.collectInferredReferenceTargetsForBody(addr, bSchema.Address.ScopeId, blk.Body, bSchema.Body)...)
			}

			bodyRef.Type = bodyToDataType(bSchema.Type, bSchema.Body)

			refs = append(refs, bodyRef)
		}

		if bSchema.Address.DependentBodyAsData {
			if !bSchema.Address.BodyAsData {
				bodyRef = lang.ReferenceTarget{
					Addr:        addr,
					ScopeId:     bSchema.Address.ScopeId,
					DefRangePtr: blk.DefRange.Ptr(),
					RangePtr:    blk.Range.Ptr(),
				}
			}

			depSchema, _, ok := NewBlockSchema(bSchema).DependentBodySchema(blk.Block)
			if ok {
				fullSchema := depSchema
				if bSchema.Address.BodyAsData {
					mergedSchema, err := mergeBlockBodySchemas(blk.Block, bSchema)
					if err != nil {
						continue
					}
					bodyRef.NestedTargets = make(lang.ReferenceTargets, 0)
					fullSchema = mergedSchema
				}

				bodyRef.Type = bodyToDataType(bSchema.Type, fullSchema)

				if bSchema.Address.InferDependentBody && len(bSchema.DependentBody) > 0 {
					bodyRef.NestedTargets = append(bodyRef.NestedTargets,
						d.collectInferredReferenceTargetsForBody(addr, bSchema.Address.ScopeId, blk.Body, fullSchema)...)
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

func decodeTargetableBody(body hcl.Body, parentBlock *blockContent, tt *schema.Targetable) lang.ReferenceTarget {
	target := lang.ReferenceTarget{
		Addr:        tt.Address.Copy(),
		ScopeId:     tt.ScopeId,
		RangePtr:    parentBlock.Range.Ptr(),
		DefRangePtr: parentBlock.DefRange.Ptr(),
		Type:        tt.AsType,
		Description: tt.Description,
	}

	if tt.NestedTargetables != nil {
		target.NestedTargets = make(lang.ReferenceTargets, len(tt.NestedTargetables))
		for i, ntt := range tt.NestedTargetables {
			target.NestedTargets[i] = decodeTargetableBody(body, parentBlock, ntt)
		}
	}

	return target
}

func decodeReferenceTargetsForAttribute(attr *hcl.Attribute, attrSchema *schema.AttributeSchema) lang.ReferenceTargets {
	refs := make(lang.ReferenceTargets, 0)

	attrAddr, ok := resolveAttributeAddress(attr, attrSchema.Address)
	if ok {
		if attrSchema.Address.AsReference {
			ref := lang.ReferenceTarget{
				Addr:        attrAddr,
				ScopeId:     attrSchema.Address.ScopeId,
				DefRangePtr: &attr.NameRange,
				RangePtr:    attr.Range.Ptr(),
				Name:        attrSchema.Address.FriendlyName,
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

				ref := lang.ReferenceTarget{
					Addr:        attrAddr,
					Type:        t,
					ScopeId:     scopeId,
					DefRangePtr: attr.NameRange.Ptr(),
					RangePtr:    attr.Range.Ptr(),
					Name:        attrSchema.Address.FriendlyName,
				}

				if attr.Expr != nil && !t.IsPrimitiveType() {
					ref.NestedTargets = make(lang.ReferenceTargets, 0)
					ref.NestedTargets = append(ref.NestedTargets, decodeReferenceTargetsForComplexTypeExpr(attrAddr, attr.Expr, t, scopeId)...)
				}

				refs = append(refs, ref)
			}
		}
	}

	ec := ExprConstraints(attrSchema.Expr)
	refs = append(refs, referenceTargetsForExpr(attr.Expr, ec)...)
	return refs
}

func decodeReferenceTargetsForComplexTypeExpr(addr lang.Address, expr hcl.Expression, t cty.Type, scopeId lang.ScopeId) lang.ReferenceTargets {
	refs := make(lang.ReferenceTargets, 0)

	if expr == nil {
		return refs
	}

	// Nested expressions are not addressable in JSON yet
	// as accessing expression in JSON generally requires
	// some upstream HCL changes.
	// TODO: See https: //github.com/hashicorp/terraform-ls/issues/675
	switch e := expr.(type) {
	case *hclsyntax.TupleConsExpr:
		if t.IsListType() {
			for i, item := range e.Exprs {
				elemAddr := append(addr.Copy(), lang.IndexStep{Key: cty.NumberIntVal(int64(i))})
				elemType := t.ElementType()

				ref := lang.ReferenceTarget{
					Addr:     elemAddr,
					Type:     elemType,
					ScopeId:  scopeId,
					RangePtr: item.Range().Ptr(),
				}
				if !elemType.IsPrimitiveType() {
					ref.NestedTargets = make(lang.ReferenceTargets, 0)
					ref.NestedTargets = append(ref.NestedTargets, decodeReferenceTargetsForComplexTypeExpr(elemAddr, item, elemType, scopeId)...)
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
				attrAddr := append(addr.Copy(), lang.AttrStep{Name: key.AsString()})
				rng := hcl.RangeBetween(item.KeyExpr.Range(), item.ValueExpr.Range())

				ref := lang.ReferenceTarget{
					Addr:        attrAddr,
					Type:        attrType,
					ScopeId:     scopeId,
					DefRangePtr: item.KeyExpr.Range().Ptr(),
					RangePtr:    rng.Ptr(),
				}
				if !attrType.IsPrimitiveType() {
					ref.NestedTargets = make(lang.ReferenceTargets, 0)
					ref.NestedTargets = append(ref.NestedTargets, decodeReferenceTargetsForComplexTypeExpr(attrAddr, item.ValueExpr, attrType, scopeId)...)
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

				elemAddr := append(addr.Copy(), lang.IndexStep{Key: key})
				rng := hcl.RangeBetween(item.KeyExpr.Range(), item.ValueExpr.Range())

				ref := lang.ReferenceTarget{
					Addr:        elemAddr,
					Type:        elemType,
					ScopeId:     scopeId,
					DefRangePtr: item.KeyExpr.Range().Ptr(),
					RangePtr:    rng.Ptr(),
				}
				if !elemType.IsPrimitiveType() {
					ref.NestedTargets = make(lang.ReferenceTargets, 0)
					ref.NestedTargets = append(ref.NestedTargets, decodeReferenceTargetsForComplexTypeExpr(elemAddr, item.ValueExpr, elemType, scopeId)...)
				}

				refs = append(refs, ref)
			}
		}
	}

	return refs
}

func referenceAsTypeOf(block *hcl.Block, rngPtr *hcl.Range, bSchema *schema.BlockSchema, addr lang.Address) lang.ReferenceTargets {
	ref := lang.ReferenceTarget{
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
		return lang.ReferenceTargets{ref}
	}

	if bSchema.Address.AsTypeOf.AttributeExpr != "" {
		typeDecl, ok := asTypeOfAttrExpr(attrs, bSchema)
		if !ok && bSchema.Address.AsTypeOf.AttributeValue == "" {
			// nothing to fall back to, exit early
			return lang.ReferenceTargets{ref}
		}
		ref.Type = typeDecl
	}

	if bSchema.Address.AsTypeOf.AttributeValue != "" {
		attr, ok := attrs[bSchema.Address.AsTypeOf.AttributeValue]
		if !ok {
			return lang.ReferenceTargets{ref}
		}
		value, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return lang.ReferenceTargets{ref}
		}
		val, err := convert.Convert(value, ref.Type)
		if err != nil {
			// type does not comply with type constraint
			return lang.ReferenceTargets{ref}
		}
		ref.Type = val.Type()
	}

	return lang.ReferenceTargets{ref}
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

func referenceTargetsForExpr(expr hcl.Expression, ec ExprConstraints) lang.ReferenceTargets {
	refs := make(lang.ReferenceTargets, 0)

	switch e := expr.(type) {
	// TODO: Support all expression types (list/set/map literals)
	case *hclsyntax.ScopeTraversalExpr:
		tes, ok := ec.TraversalExprs()
		if !ok {
			// unknown traversal
			return lang.ReferenceTargets{}
		}

		addr, err := lang.TraversalToAddress(e.AsTraversal())
		if err != nil {
			return lang.ReferenceTargets{}
		}

		for _, te := range tes {
			if te.Address == nil {
				// skip traversals which are not addressable by themselves
				continue
			}

			refs = append(refs, lang.ReferenceTarget{
				Addr:     addr,
				ScopeId:  te.Address.ScopeId,
				RangePtr: e.SrcRange.Ptr(),
				Name:     te.Name,
			})
		}
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

				refs = append(refs, referenceTargetsForExpr(item.ValueExpr, ExprConstraints(attr.Expr))...)
			}
		}
		me, ok := ec.MapExpr()
		if ok {
			for _, item := range e.Items {
				refs = append(refs, referenceTargetsForExpr(item.ValueExpr, ExprConstraints(me.Elem))...)
			}
		}
	case *hclsyntax.TupleConsExpr:
		le, ok := ec.ListExpr()
		if ok {
			for _, itemExpr := range e.Exprs {
				refs = append(refs, referenceTargetsForExpr(itemExpr, ExprConstraints(le.Elem))...)
			}
		}
		se, ok := ec.SetExpr()
		if ok {
			for _, itemExpr := range e.Exprs {
				refs = append(refs, referenceTargetsForExpr(itemExpr, ExprConstraints(se.Elem))...)
			}
		}
		te, ok := ec.TupleExpr()
		if ok {
			for i, itemExpr := range e.Exprs {
				if i >= len(te.Elems) {
					break
				}
				refs = append(refs, referenceTargetsForExpr(itemExpr, ExprConstraints(te.Elems[i]))...)
			}
		}
		tce, ok := ec.TupleConsExpr()
		if ok {
			for _, itemExpr := range e.Exprs {
				refs = append(refs, referenceTargetsForExpr(itemExpr, ExprConstraints(tce.AnyElem))...)
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

func (d *Decoder) collectInferredReferenceTargetsForBody(addr lang.Address, scopeId lang.ScopeId, body hcl.Body, bodySchema *schema.BodySchema) lang.ReferenceTargets {
	refs := make(lang.ReferenceTargets, 0)

	content := decodeBody(body, bodySchema)

	for name, aSchema := range bodySchema.Attributes {
		attrType, ok := exprConstraintToDataType(aSchema.Expr)
		if !ok {
			// unknown type
			continue
		}

		attrAddr := append(addr.Copy(), lang.AttrStep{Name: name})

		ref := lang.ReferenceTarget{
			Addr:        attrAddr,
			ScopeId:     scopeId,
			Type:        attrType,
			Description: aSchema.Description,
			RangePtr:    body.MissingItemRange().Ptr(),
		}

		var attrExpr hcl.Expression
		if attr, ok := content.Attributes[name]; ok {
			ref.RangePtr = attr.Range.Ptr()
			ref.DefRangePtr = attr.NameRange.Ptr()
			attrExpr = attr.Expr
		}

		if attrExpr != nil && !attrType.IsPrimitiveType() {
			ref.NestedTargets = make(lang.ReferenceTargets, 0)
			ref.NestedTargets = append(ref.NestedTargets, decodeReferenceTargetsForComplexTypeExpr(attrAddr, attrExpr, attrType, scopeId)...)
		}

		refs = append(refs, ref)
	}

	bTypes := blocksTypesWithSchema(body, bodySchema)

	for bType, bCollection := range bTypes.OfSchemaType(schema.BlockTypeObject) {
		blockAddr := make(lang.Address, len(addr))
		copy(blockAddr, addr)
		blockAddr = append(blockAddr, lang.AttrStep{Name: bType})

		blk := bCollection.Blocks[0]

		blockRef := lang.ReferenceTarget{
			Addr:        blockAddr,
			ScopeId:     scopeId,
			Type:        cty.Object(bodySchemaAsAttrTypes(bCollection.Schema.Body)),
			Description: bCollection.Schema.Description,
			DefRangePtr: blk.DefRange.Ptr(),
			RangePtr:    blk.Range.Ptr(),
			NestedTargets: d.collectInferredReferenceTargetsForBody(
				blockAddr, scopeId, blk.Body, bCollection.Schema.Body),
		}
		sort.Sort(blockRef.NestedTargets)
		refs = append(refs, blockRef)
	}

	for bType, bCollection := range bTypes.OfSchemaType(schema.BlockTypeList) {
		blockAddr := make(lang.Address, len(addr))
		copy(blockAddr, addr)
		blockAddr = append(blockAddr, lang.AttrStep{Name: bType})

		blockRef := lang.ReferenceTarget{
			Addr:        blockAddr,
			ScopeId:     scopeId,
			Type:        cty.List(cty.Object(bodySchemaAsAttrTypes(bCollection.Schema.Body))),
			Description: bCollection.Schema.Description,
			RangePtr:    body.MissingItemRange().Ptr(),
		}

		for i, b := range bCollection.Blocks {
			elemAddr := make(lang.Address, len(blockAddr))
			copy(elemAddr, blockAddr)
			elemAddr = append(elemAddr, lang.IndexStep{
				Key: cty.NumberIntVal(int64(i)),
			})

			elemRef := lang.ReferenceTarget{
				Addr:        elemAddr,
				ScopeId:     scopeId,
				Type:        cty.Object(bodySchemaAsAttrTypes(bCollection.Schema.Body)),
				Description: bCollection.Schema.Description,
				DefRangePtr: b.DefRange.Ptr(),
				RangePtr:    b.Range.Ptr(),
				NestedTargets: d.collectInferredReferenceTargetsForBody(
					elemAddr, scopeId, b.Body, bCollection.Schema.Body),
			}
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
		blockAddr := make(lang.Address, len(addr))
		copy(blockAddr, addr)
		blockAddr = append(blockAddr, lang.AttrStep{Name: bType})

		blockRef := lang.ReferenceTarget{
			Addr:        blockAddr,
			ScopeId:     scopeId,
			Type:        cty.Set(cty.Object(bodySchemaAsAttrTypes(bCollection.Schema.Body))),
			Description: bCollection.Schema.Description,
			RangePtr:    body.MissingItemRange().Ptr(),
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
		blockAddr := make(lang.Address, len(addr))
		copy(blockAddr, addr)
		blockAddr = append(blockAddr, lang.AttrStep{Name: bType})

		blockRef := lang.ReferenceTarget{
			Addr:        blockAddr,
			ScopeId:     scopeId,
			Type:        cty.Map(cty.Object(bodySchemaAsAttrTypes(bCollection.Schema.Body))),
			Description: bCollection.Schema.Description,
			RangePtr:    body.MissingItemRange().Ptr(),
		}

		for i, b := range bCollection.Blocks {
			elemAddr := make(lang.Address, len(blockAddr))
			copy(elemAddr, blockAddr)
			elemAddr = append(elemAddr, lang.IndexStep{
				Key: cty.StringVal(b.Labels[0]),
			})

			refType := cty.Object(bodySchemaAsAttrTypes(bCollection.Schema.Body))

			elemRef := lang.ReferenceTarget{
				Addr:        elemAddr,
				ScopeId:     scopeId,
				Type:        refType,
				Description: bCollection.Schema.Description,
				RangePtr:    b.Range.Ptr(),
				DefRangePtr: b.DefRange.Ptr(),
				NestedTargets: d.collectInferredReferenceTargetsForBody(
					elemAddr, scopeId, b.Body, bCollection.Schema.Body),
			}
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
	Blocks []*blockContent
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

	content := decodeBody(body, bodySchema)

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
				Blocks: make([]*blockContent, 0),
			}
		}

		blockTypes[block.Type].Blocks = append(blockTypes[block.Type].Blocks, block)
	}

	return blockTypes
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

func resolveAttributeAddress(attr *hcl.Attribute, addr *schema.AttributeAddrSchema) (lang.Address, bool) {
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
				return lang.Address{}, false
			}
			stepName = block.Labels[step.Index]
		case schema.AttrValueStep:
			content := decodeBody(block.Body, blockSchema.Body)

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
