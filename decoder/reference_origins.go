package decoder

import (
	"sort"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// ReferenceOriginAtPos returns the ReferenceOrigin
// enclosing the position in a file, if one exists, else nil
func (d *Decoder) ReferenceOriginAtPos(filename string, pos hcl.Pos) (*lang.ReferenceOrigin, error) {
	// TODO: Filter d.refOriginReader instead here

	f, err := d.fileByName(filename)
	if err != nil {
		return nil, err
	}

	rootBody, err := d.bodyForFileAndPos(filename, f, pos)
	if err != nil {
		return nil, err
	}

	d.rootSchemaMu.RLock()
	defer d.rootSchemaMu.RUnlock()
	if d.rootSchema == nil {
		return nil, &NoSchemaError{}
	}

	return d.referenceOriginAtPos(rootBody, d.rootSchema, pos)
}

func (d *Decoder) ReferenceOriginsTargeting(refTarget lang.ReferenceTarget) (lang.ReferenceOrigins, error) {
	if d.refOriginReader == nil {
		return nil, nil
	}

	allOrigins := ReferenceOrigins(d.refOriginReader())

	return allOrigins.Targeting(refTarget), nil
}

func (d *Decoder) CollectReferenceOrigins() (lang.ReferenceOrigins, error) {
	refOrigins := make(lang.ReferenceOrigins, 0)

	d.rootSchemaMu.RLock()
	defer d.rootSchemaMu.RUnlock()

	if d.rootSchema == nil {
		// unable to collect reference origins without schema
		return refOrigins, &NoSchemaError{}
	}

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

		refOrigins = append(refOrigins, d.referenceOriginsInBody(body, d.rootSchema)...)
	}

	sort.SliceStable(refOrigins, func(i, j int) bool {
		return refOrigins[i].Range.Filename <= refOrigins[i].Range.Filename &&
			refOrigins[i].Range.Start.Byte < refOrigins[j].Range.Start.Byte
	})

	return refOrigins, nil
}

func (d *Decoder) referenceOriginsInBody(body *hclsyntax.Body, bodySchema *schema.BodySchema) lang.ReferenceOrigins {
	origins := make(lang.ReferenceOrigins, 0)

	if bodySchema == nil {
		return origins
	}

	for _, attr := range body.Attributes {
		aSchema, ok := bodySchema.Attributes[attr.Name]
		if !ok {
			if bodySchema.AnyAttribute == nil {
				// skip unknown attribute
				continue
			}
			aSchema = bodySchema.AnyAttribute
		}

		origins = append(origins, d.findOriginsInExpression(attr.Expr, aSchema.Expr)...)
	}

	for _, block := range body.Blocks {
		if block.Body != nil {
			bSchema, ok := bodySchema.Blocks[block.Type]
			if !ok {
				// skip unknown blocks
				continue
			}
			mergedSchema, err := mergeBlockBodySchemas(block, bSchema)
			if err != nil {
				continue
			}
			origins = append(origins, d.referenceOriginsInBody(block.Body, mergedSchema)...)
		}
	}

	return origins
}

func (d *Decoder) findOriginsInExpression(expr hcl.Expression, ec schema.ExprConstraints) lang.ReferenceOrigins {
	// TODO Review once nested expressions are supported

	origins := make(lang.ReferenceOrigins, 0)

	switch eType := expr.(type) {
	case *hclsyntax.TupleConsExpr:
		tce, ok := ExprConstraints(ec).TupleConsExpr()
		if ok {
			for _, elemExpr := range eType.ExprList() {
				origins = append(origins, d.findOriginsInExpression(elemExpr, tce.AnyElem)...)
			}
			break
		}

		le, ok := ExprConstraints(ec).ListExpr()
		if ok {
			for _, elemExpr := range eType.ExprList() {
				origins = append(origins, d.findOriginsInExpression(elemExpr, le.Elem)...)
			}
			break
		}

		se, ok := ExprConstraints(ec).SetExpr()
		if ok {
			for _, elemExpr := range eType.ExprList() {
				origins = append(origins, d.findOriginsInExpression(elemExpr, se.Elem)...)
			}
			break
		}

		tue, ok := ExprConstraints(ec).TupleExpr()
		if ok {
			for i, elemExpr := range eType.ExprList() {
				if len(tue.Elems) < i+1 {
					break
				}
				origins = append(origins, d.findOriginsInExpression(elemExpr, tue.Elems[i])...)
			}
		}
	case *hclsyntax.ObjectConsExpr:
		oe, ok := ExprConstraints(ec).ObjectExpr()
		if ok {
			for _, item := range eType.Items {
				key, _ := item.KeyExpr.Value(nil)
				if key.IsNull() || !key.IsWhollyKnown() || key.Type() != cty.String {
					// skip items keys that can't be interpolated
					// without further context
					continue
				}

				attr, ok := oe.Attributes[key.AsString()]
				if !ok {
					// skip unknown attribute
					continue
				}

				origins = append(origins, d.findOriginsInExpression(item.ValueExpr, attr.Expr)...)
			}
		}

		me, ok := ExprConstraints(ec).MapExpr()
		if ok {
			for _, item := range eType.Items {
				origins = append(origins, d.findOriginsInExpression(item.ValueExpr, me.Elem)...)
			}
		}
	default:
		tes, ok := ExprConstraints(ec).TraversalExprs()
		if ok {
			origins = append(origins, traversalsToReferenceOrigins(expr.Variables(), tes)...)
		}
	}

	return origins
}

func traversalsToReferenceOrigins(traversals []hcl.Traversal, tes schema.TraversalExprs) lang.ReferenceOrigins {
	origins := make(lang.ReferenceOrigins, 0)
	for _, traversal := range traversals {
		origin, err := TraversalToReferenceOrigin(traversal, tes)
		if err != nil {
			continue
		}
		origins = append(origins, origin)
	}

	return origins
}

func (d *Decoder) referenceOriginAtPos(body *hclsyntax.Body, bodySchema *schema.BodySchema, pos hcl.Pos) (*lang.ReferenceOrigin, error) {
	for _, attr := range body.Attributes {
		if d.isPosInsideAttrExpr(attr, pos) {
			aSchema, ok := bodySchema.Attributes[attr.Name]
			if !ok {
				if bodySchema.AnyAttribute == nil {
					// skip unknown attribute
					continue
				}
				aSchema = bodySchema.AnyAttribute
			}

			for _, origin := range d.findOriginsInExpression(attr.Expr, aSchema.Expr) {
				if origin.Range.ContainsPos(pos) {
					return &origin, nil
				}
			}

			return nil, nil
		}
	}

	for _, block := range body.Blocks {
		if block.Range().ContainsPos(pos) {
			if block.Body != nil && block.Body.Range().ContainsPos(pos) {
				bSchema, ok := bodySchema.Blocks[block.Type]
				if !ok {
					// skip unknown block
					continue
				}

				mergedSchema, err := mergeBlockBodySchemas(block, bSchema)
				if err != nil {
					continue
				}

				return d.referenceOriginAtPos(block.Body, mergedSchema, pos)
			}
		}
	}

	return nil, nil
}

func (d *Decoder) traversalAtPos(expr hclsyntax.Expression, pos hcl.Pos) (hcl.Traversal, bool) {
	for _, traversal := range expr.Variables() {
		if traversal.SourceRange().ContainsPos(pos) {
			return traversal, true
		}
	}

	return nil, false
}

type ReferenceOrigins lang.ReferenceOrigins

func (ro ReferenceOrigins) Targeting(refTarget lang.ReferenceTarget) lang.ReferenceOrigins {
	origins := make(lang.ReferenceOrigins, 0)

	target := ReferenceTarget(refTarget)

	for _, refOrigin := range ro {
		if target.IsTargetableBy(refOrigin) {
			origins = append(origins, refOrigin)
		}
	}

	for _, iTarget := range refTarget.NestedTargets {
		origins = append(origins, ro.Targeting(iTarget)...)
	}

	return origins
}

func TraversalToReferenceOrigin(traversal hcl.Traversal, tes []schema.TraversalExpr) (lang.ReferenceOrigin, error) {
	addr, err := lang.TraversalToAddress(traversal)
	if err != nil {
		return lang.ReferenceOrigin{}, err
	}

	return lang.ReferenceOrigin{
		Addr:        addr,
		Range:       traversal.SourceRange(),
		Constraints: traversalExpressionsToOriginConstraints(tes),
	}, nil
}

func traversalExpressionsToOriginConstraints(tes []schema.TraversalExpr) lang.ReferenceOriginConstraints {
	if tes == nil {
		return nil
	}

	roc := make(lang.ReferenceOriginConstraints, 0)
	for _, te := range tes {
		if te.Address != nil {
			// skip traversals which are targets by themselves (not origins)
			continue
		}
		roc = append(roc, lang.ReferenceOriginConstraint{
			OfType:    te.OfType,
			OfScopeId: te.OfScopeId,
		})
	}
	return roc
}
