package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (list List) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	eType, ok := list.expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return reference.Targets{}
	}

	if len(eType.Exprs) == 0 || list.cons.Elem == nil {
		return reference.Targets{}
	}

	targets := make(reference.Targets, 0)

	// TODO: collect parent target for the whole list

	for i, elemExpr := range eType.Exprs {
		expr := newExpression(list.pathCtx, elemExpr, list.cons.Elem)
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
