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
func (d *PathDecoder) ReferenceOriginAtPos(filename string, pos hcl.Pos) (*lang.ReferenceOrigin, error) {
	// TODO: Filter d.refOriginReader instead here

	f, err := d.fileByName(filename)
	if err != nil {
		return nil, err
	}

	rootBody, err := d.bodyForFileAndPos(filename, f, pos)
	if err != nil {
		return nil, err
	}

	if d.pathCtx.Schema == nil {
		return nil, &NoSchemaError{}
	}

	return d.referenceOriginAtPos(rootBody, d.pathCtx.Schema, pos)
}

func (d *PathDecoder) ReferenceOriginsTargeting(refTarget lang.ReferenceTarget) (lang.ReferenceOrigins, error) {
	if d.pathCtx.ReferenceOrigins == nil {
		return nil, nil
	}

	allOrigins := ReferenceOrigins(d.pathCtx.ReferenceOrigins)

	return allOrigins.Targeting(refTarget), nil
}

func (d *PathDecoder) CollectReferenceOrigins() (lang.ReferenceOrigins, error) {
	refOrigins := make(lang.ReferenceOrigins, 0)

	if d.pathCtx.Schema == nil {
		// unable to collect reference origins without schema
		return refOrigins, &NoSchemaError{}
	}

	files := d.filenames()
	for _, filename := range files {
		f, err := d.fileByName(filename)
		if err != nil {
			// skip unparseable file
			continue
		}

		refOrigins = append(refOrigins, d.referenceOriginsInBody(f.Body, d.pathCtx.Schema)...)
	}

	sort.SliceStable(refOrigins, func(i, j int) bool {
		return refOrigins[i].Range.Filename <= refOrigins[i].Range.Filename &&
			refOrigins[i].Range.Start.Byte < refOrigins[j].Range.Start.Byte
	})

	return refOrigins, nil
}

func (d *PathDecoder) referenceOriginsInBody(body hcl.Body, bodySchema *schema.BodySchema) lang.ReferenceOrigins {
	origins := make(lang.ReferenceOrigins, 0)

	if bodySchema == nil {
		return origins
	}

	content := decodeBody(body, bodySchema)

	for _, attr := range content.Attributes {
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

	for _, block := range content.Blocks {
		if block.Body != nil {
			bSchema, ok := bodySchema.Blocks[block.Type]
			if !ok {
				// skip unknown blocks
				continue
			}
			mergedSchema, err := mergeBlockBodySchemas(block.Block, bSchema)
			if err != nil {
				continue
			}
			origins = append(origins, d.referenceOriginsInBody(block.Body, mergedSchema)...)
		}
	}

	return origins
}

func (d *PathDecoder) findOriginsInExpression(expr hcl.Expression, ec schema.ExprConstraints) lang.ReferenceOrigins {
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
	case *hclsyntax.AnonSymbolExpr,
		*hclsyntax.BinaryOpExpr,
		*hclsyntax.ConditionalExpr,
		*hclsyntax.ForExpr,
		*hclsyntax.FunctionCallExpr,
		*hclsyntax.IndexExpr,
		*hclsyntax.ParenthesesExpr,
		*hclsyntax.RelativeTraversalExpr,
		*hclsyntax.ScopeTraversalExpr,
		*hclsyntax.SplatExpr,
		*hclsyntax.TemplateExpr,
		*hclsyntax.TemplateJoinExpr,
		*hclsyntax.TemplateWrapExpr,
		*hclsyntax.UnaryOpExpr:

		// Constraints detected here may be inaccurate, but close enough
		// to be more useful for relevant completion than no constraints.
		// TODO: Review this when we support all expression types and nesting
		// see https://github.com/hashicorp/terraform-ls/issues/496
		tes, ok := ExprConstraints(ec).TraversalExprs()
		if ok {
			origins = append(origins, traversalsToReferenceOrigins(expr.Variables(), tes)...)
		}
	case *hclsyntax.LiteralValueExpr:
		// String constant may also be a traversal in some cases, but currently not recognized
		// TODO: https://github.com/hashicorp/terraform-ls/issues/674
	default:
		// Given that all hclsyntax.* expressions are listed above
		// this should only apply to (unexported) json.* expressions
		// for which we return no constraints as upstream doesn't provide
		// any way to map the schema to individual traversals.
		// This may result in less accurate decoding where even origins
		// which do not actually conform to the constraints are recognized.
		// TODO: https://github.com/hashicorp/terraform-ls/issues/675
		origins = append(origins, traversalsToReferenceOrigins(expr.Variables(), schema.TraversalExprs{})...)
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

func (d *PathDecoder) referenceOriginAtPos(body *hclsyntax.Body, bodySchema *schema.BodySchema, pos hcl.Pos) (*lang.ReferenceOrigin, error) {
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

				mergedSchema, err := mergeBlockBodySchemas(block.AsHCLBlock(), bSchema)
				if err != nil {
					continue
				}

				return d.referenceOriginAtPos(block.Body, mergedSchema, pos)
			}
		}
	}

	return nil, nil
}

func (d *PathDecoder) traversalAtPos(expr hclsyntax.Expression, pos hcl.Pos) (hcl.Traversal, bool) {
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
	if len(tes) == 0 {
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
