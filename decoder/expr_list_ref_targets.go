package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
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

	elemTargets := make(reference.Targets, 0)

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

			elemTargets = append(elemTargets, e.ReferenceTargets(ctx, elemCtx)...)
		}
	}

	targets := make(reference.Targets, 0)

	if targetCtx != nil {
		// collect target for the whole list

		// type-aware
		elemCons, ok := list.cons.Elem.(schema.TypeAwareConstraint)
		if targetCtx.AsExprType && ok {
			elemType, ok := elemCons.ConstraintType()
			if ok {
				targets = append(targets, reference.Target{
					Addr:                   targetCtx.ParentAddress,
					Name:                   targetCtx.FriendlyName,
					Type:                   cty.List(elemType),
					ScopeId:                targetCtx.ScopeId,
					RangePtr:               list.expr.Range().Ptr(),
					NestedTargets:          elemTargets,
					LocalAddr:              targetCtx.ParentLocalAddress,
					TargetableFromRangePtr: targetCtx.TargetableFromRangePtr,
				})
			}
		}

		// type-unaware
		if targetCtx.AsReference {
			targets = append(targets, reference.Target{
				Addr:                   targetCtx.ParentAddress,
				Name:                   targetCtx.FriendlyName,
				ScopeId:                targetCtx.ScopeId,
				RangePtr:               list.expr.Range().Ptr(),
				NestedTargets:          elemTargets,
				LocalAddr:              targetCtx.ParentLocalAddress,
				TargetableFromRangePtr: targetCtx.TargetableFromRangePtr,
			})
		}
	} else {
		// treat element targets as 1st class ones
		// if the list itself isn't targetable
		targets = elemTargets
	}

	return targets
}
