package decoder

import (
	"sort"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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

		tes, ok := d.findTraversalContraintsForExpr(aSchema.Expr)
		if !ok {
			continue
		}
		for _, traversal := range attr.Expr.Variables() {
			origin, err := TraversalToReferenceOrigin(traversal, tes)
			if err != nil {
				continue
			}

			origins = append(origins, origin)
		}
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

func (d *Decoder) findTraversalContraintsForExpr(ec schema.ExprConstraints) (schema.TraversalExprs, bool) {
	tes, ok := ExprConstraints(ec).TraversalExprs()
	if ok {
		return tes, true
	}

	tce, ok := ExprConstraints(ec).TupleConsExpr()
	if ok {
		return d.findTraversalContraintsForExpr(tce.AnyElem)
	}

	le, ok := ExprConstraints(ec).ListExpr()
	if ok {
		return d.findTraversalContraintsForExpr(le.Elem)
	}

	se, ok := ExprConstraints(ec).SetExpr()
	if ok {
		return d.findTraversalContraintsForExpr(se.Elem)
	}

	tue, ok := ExprConstraints(ec).TupleExpr()
	if ok {
		for _, elem := range tue.Elems {
			te, ok := d.findTraversalContraintsForExpr(elem)
			if ok {
				return te, true
			}
		}
	}

	oe, ok := ExprConstraints(ec).ObjectExpr()
	if ok {
		for _, val := range oe.Attributes {
			te, ok := d.findTraversalContraintsForExpr(val.Expr)
			if ok {
				return te, true
			}
		}
	}

	me, ok := ExprConstraints(ec).MapExpr()
	if ok {
		te, ok := d.findTraversalContraintsForExpr(me.Elem)
		if ok {
			return te, true
		}
	}

	return nil, false
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

			traversal, ok := d.traversalAtPos(attr.Expr, pos)
			if ok {
				var tExprs schema.TraversalExprs
				switch attr.Expr.(type) {
				case *hclsyntax.ScopeTraversalExpr:
					tes, ok := ExprConstraints(aSchema.Expr).TraversalExprs()
					if !ok {
						continue
					}
					tExprs = tes
				default:
					// TODO Reflect real expression at position
					tes, ok := d.findTraversalContraintsForExpr(aSchema.Expr)
					if !ok {
						continue
					}
					tExprs = tes
				}

				origin, err := TraversalToReferenceOrigin(traversal, tExprs)
				if err != nil {
					continue
				}
				return &origin, nil
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
