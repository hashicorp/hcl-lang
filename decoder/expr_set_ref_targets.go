package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

func (set Set) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	elems, diags := hcl.ExprList(set.expr)
	if diags.HasErrors() {
		return reference.Targets{}
	}

	if set.cons.Elem == nil {
		return reference.Targets{}
	}

	elemTargets := make(reference.Targets, 0)

	for _, elemExpr := range elems {
		expr := newExpression(set.pathCtx, elemExpr, set.cons.Elem)
		if e, ok := expr.(ReferenceTargetsExpression); ok {
			if targetCtx == nil {
				// collect any targets inside the expression
				// as set elements aren't addressable by themselves
				elemTargets = append(elemTargets, e.ReferenceTargets(ctx, nil)...)
				continue
			}
		}
	}

	targets := make(reference.Targets, 0)

	if targetCtx != nil {
		// collect target for the whole set

		// type-aware
		elemCons, ok := set.cons.Elem.(schema.TypeAwareConstraint)
		if targetCtx.AsExprType && ok {
			elemType, ok := elemCons.ConstraintType()
			if ok {
				targets = append(targets, reference.Target{
					Addr:                   targetCtx.ParentAddress,
					Name:                   targetCtx.FriendlyName,
					Type:                   cty.Set(elemType),
					ScopeId:                targetCtx.ScopeId,
					RangePtr:               set.expr.Range().Ptr(),
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
				RangePtr:               set.expr.Range().Ptr(),
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
