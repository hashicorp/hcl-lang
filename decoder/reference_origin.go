package decoder

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// ReferenceOriginAtPos returns ReferenceOrigin for the enclosing
// traversal in the given file and position, if one exists, otherwise nil
func (d *Decoder) ReferenceOriginAtPos(filename string, pos hcl.Pos) (*lang.ReferenceOrigin, error) {
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

	ref, err := d.referenceOriginAtPos(rootBody, d.rootSchema, pos)
	if err != nil {
		return nil, err
	}

	return ref, nil
}

func (d *Decoder) referenceOriginAtPos(body *hclsyntax.Body, bodySchema *schema.BodySchema, pos hcl.Pos) (*lang.ReferenceOrigin, error) {
	for _, attr := range body.Attributes {
		if d.isPosInsideAttrExpr(attr, pos) {
			aSchema, ok := bodySchema.Attributes[attr.Name]
			if ok {
				traversal, ok := d.traversalAtPos(attr.Expr, aSchema.Expr, pos)
				if ok {
					addr, err := lang.TraversalToAddress(traversal)
					if err == nil {
						return &lang.ReferenceOrigin{
							Addr:  addr,
							Range: traversal.SourceRange(),
						}, nil
					}
				}
			}
			if bodySchema.AnyAttribute != nil {
				traversal, ok := d.traversalAtPos(attr.Expr, bodySchema.AnyAttribute.Expr, pos)
				if ok {
					addr, err := lang.TraversalToAddress(traversal)
					if err == nil {
						return &lang.ReferenceOrigin{
							Addr:  addr,
							Range: traversal.SourceRange(),
						}, nil
					}
				}
			}

			return nil, nil
		}
	}

	for _, block := range body.Blocks {
		if block.Range().ContainsPos(pos) {
			bSchema, ok := bodySchema.Blocks[block.Type]
			if !ok {
				return nil, nil
			}

			if block.Body != nil && block.Body.Range().ContainsPos(pos) {
				mergedSchema, err := mergeBlockBodySchemas(block, bSchema)
				if err != nil {
					return nil, err
				}

				return d.referenceOriginAtPos(block.Body, mergedSchema, pos)
			}
		}
	}

	return nil, nil
}

func (d *Decoder) traversalAtPos(expr hclsyntax.Expression, ec schema.ExprConstraints, pos hcl.Pos) (hcl.Traversal, bool) {
	for _, traversal := range expr.Variables() {
		if traversal.SourceRange().ContainsPos(pos) {
			return traversal, true
		}
	}

	return nil, false
}
