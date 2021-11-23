package decoder

import (
	"context"
	"sort"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (d *Decoder) ReferenceOriginsTargetingPos(path lang.Path, file string, pos hcl.Pos) ReferenceOrigins {
	origins := make(ReferenceOrigins, 0)

	ctx := context.Background()

	localCtx, err := d.pathReader.PathContext(path)
	if err != nil {
		return origins
	}

	targets, ok := localCtx.ReferenceTargets.InnermostAtPos(file, pos)
	if !ok {
		return ReferenceOrigins{}
	}

	for _, target := range targets {
		paths := d.pathReader.Paths(ctx)
		for _, p := range paths {
			pathCtx, err := d.pathReader.PathContext(p)
			if err != nil {
				continue
			}

			rawOrigins := pathCtx.ReferenceOrigins.Match(p, target, path)
			for _, origin := range rawOrigins {
				origins = append(origins, ReferenceOrigin{
					Path:  p,
					Range: origin.OriginRange(),
				})
			}
		}
	}

	return origins
}

func (d *PathDecoder) CollectReferenceOrigins() (reference.Origins, error) {
	refOrigins := make(reference.Origins, 0)

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
		return refOrigins[i].OriginRange().Filename <= refOrigins[i].OriginRange().Filename &&
			refOrigins[i].OriginRange().Start.Byte < refOrigins[j].OriginRange().Start.Byte
	})

	return refOrigins, nil
}

func (d *PathDecoder) referenceOriginsInBody(body hcl.Body, bodySchema *schema.BodySchema) reference.Origins {
	origins := make(reference.Origins, 0)

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

		if aSchema.OriginForTarget != nil {
			targetAddr, ok := resolveAttributeAddress(attr, aSchema.OriginForTarget.Address)
			if ok {
				origins = append(origins, reference.PathOrigin{
					Range:      attr.NameRange,
					TargetAddr: targetAddr,
					TargetPath: aSchema.OriginForTarget.Path,
					Constraints: reference.OriginConstraints{
						{
							OfScopeId: aSchema.OriginForTarget.Constraints.ScopeId,
							OfType:    aSchema.OriginForTarget.Constraints.Type,
						},
					},
				})
			}
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

func (d *PathDecoder) findOriginsInExpression(expr hcl.Expression, ec schema.ExprConstraints) reference.Origins {
	origins := make(reference.Origins, 0)

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
			origins = append(origins, reference.TraversalsToLocalOrigins(expr.Variables(), tes)...)
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
		origins = append(origins, reference.TraversalsToLocalOrigins(expr.Variables(), schema.TraversalExprs{})...)
	}

	return origins
}

func (d *PathDecoder) referenceOriginAtPos(body *hclsyntax.Body, bodySchema *schema.BodySchema, pos hcl.Pos) (*reference.Origin, error) {
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
				if origin.OriginRange().ContainsPos(pos) {
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
