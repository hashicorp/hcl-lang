package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (tuple Tuple) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	eType, ok := tuple.expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return reference.Targets{}
	}

	if len(eType.Exprs) == 0 || len(tuple.cons.Elems) == 0 {
		return reference.Targets{}
	}

	targets := make(reference.Targets, 0)

	// TODO: collect parent target for the whole tuple

	for i, elemExpr := range eType.Exprs {
		if i+1 > len(tuple.cons.Elems) {
			break
		}

		expr := newExpression(tuple.pathCtx, elemExpr, tuple.cons.Elems[i])
		if e, ok := expr.(ReferenceTargetsExpression); ok {
			elemCtx := targetCtx.Copy()
			elemCtx.ParentAddress = append(elemCtx.ParentAddress, lang.IndexStep{
				Key: cty.NumberIntVal(int64(i)),
			})
			if elemCtx.ParentLocalAddress != nil {
				elemCtx.ParentLocalAddress = append(elemCtx.ParentLocalAddress, lang.IndexStep{
					Key: cty.NumberIntVal(int64(i)),
				})
			}
			targets = append(targets, e.ReferenceTargets(ctx, elemCtx)...)
		}
	}

	return targets
}
