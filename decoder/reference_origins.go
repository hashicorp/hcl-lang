package decoder

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// ReferenceOriginAtPos returns the ReferenceOrigin
// enclosing the position in a file, if one exists, else nil
func (d *Decoder) ReferenceOriginAtPos(filename string, pos hcl.Pos) (*lang.ReferenceOrigin, error) {
	f, err := d.fileByName(filename)
	if err != nil {
		return nil, err
	}

	rootBody, err := d.bodyForFileAndPos(filename, f, pos)
	if err != nil {
		return nil, err
	}

	return d.referenceOriginAtPos(rootBody, pos)
}

func (d *Decoder) referenceOriginAtPos(body *hclsyntax.Body, pos hcl.Pos) (*lang.ReferenceOrigin, error) {
	for _, attr := range body.Attributes {
		if d.isPosInsideAttrExpr(attr, pos) {
			traversal, ok := d.traversalAtPos(attr.Expr, pos)
			if ok {
				addr, err := lang.TraversalToAddress(traversal)
				if err == nil {
					return &lang.ReferenceOrigin{
						Addr:  addr,
						Range: traversal.SourceRange(),
					}, nil
				}
			}

			return nil, nil
		}
	}

	for _, block := range body.Blocks {
		if block.Range().ContainsPos(pos) {
			if block.Body != nil && block.Body.Range().ContainsPos(pos) {
				return d.referenceOriginAtPos(block.Body, pos)
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
