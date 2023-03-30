package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

func (tuple Tuple) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	elems, diags := hcl.ExprList(tuple.expr)
	if diags.HasErrors() {
		return reference.Targets{}
	}

	if len(tuple.cons.Elems) == 0 {
		return reference.Targets{}
	}

	elemTargets := make(reference.Targets, 0)

	for i, elemExpr := range elems {
		if i+1 > len(tuple.cons.Elems) {
			break
		}

		expr := newExpression(tuple.pathCtx, elemExpr, tuple.cons.Elems[i])
		if e, ok := expr.(ReferenceTargetsExpression); ok {
			if targetCtx == nil {
				// collect any targets inside the expression
				// if attribute itself isn't targetable
				elemTargets = append(elemTargets, e.ReferenceTargets(ctx, nil)...)
				continue
			}

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
		// collect target for the whole tuple

		var rangePtr *hcl.Range
		if targetCtx.ParentRangePtr != nil {
			rangePtr = targetCtx.ParentRangePtr
		} else {
			rangePtr = tuple.expr.Range().Ptr()
		}

		// type-aware
		elemType, ok := tuple.cons.ConstraintType()
		if targetCtx.AsExprType && ok {
			if ok {
				targets = append(targets, reference.Target{
					Addr:                   targetCtx.ParentAddress,
					Name:                   targetCtx.FriendlyName,
					Type:                   elemType,
					ScopeId:                targetCtx.ScopeId,
					RangePtr:               rangePtr,
					DefRangePtr:            targetCtx.ParentDefRangePtr,
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
				RangePtr:               rangePtr,
				DefRangePtr:            targetCtx.ParentDefRangePtr,
				NestedTargets:          elemTargets,
				LocalAddr:              targetCtx.ParentLocalAddress,
				TargetableFromRangePtr: targetCtx.TargetableFromRangePtr,
			})
		}
	} else {
		// treat element targets as 1st class ones
		// if the tuple itself isn't targetable
		targets = elemTargets
	}

	return targets
}
