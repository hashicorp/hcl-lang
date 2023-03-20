package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (set Set) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	eType, ok := set.expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return reference.Targets{}
	}

	if len(eType.Exprs) == 0 || set.cons.Elem == nil {
		return reference.Targets{}
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
				LocalAddr:              targetCtx.ParentLocalAddress,
				TargetableFromRangePtr: targetCtx.TargetableFromRangePtr,
			})
		}
	}

	return targets
}
