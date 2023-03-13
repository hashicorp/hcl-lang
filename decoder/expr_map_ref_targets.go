package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (m Map) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	eType, ok := m.expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return reference.Targets{}
	}

	if len(eType.Items) == 0 || m.cons.Elem == nil {
		return reference.Targets{}
	}

	elemTargets := make(reference.Targets, 0)

	for _, item := range eType.Items {
		keyName, _, ok := rawObjectKey(item.KeyExpr)
		if !ok {
			// avoid collecting item w/ invalid key
			continue
		}

		expr := newExpression(m.pathCtx, item.ValueExpr, m.cons.Elem)
		if e, ok := expr.(ReferenceTargetsExpression); ok {
			if targetCtx == nil {
				// collect any targets inside the expression
				// if attribute itself isn't targetable
				elemTargets = append(elemTargets, e.ReferenceTargets(ctx, nil)...)
				continue
			}

			elemCtx := targetCtx.Copy()
			elemCtx.ParentAddress = append(elemCtx.ParentAddress, lang.IndexStep{
				Key: cty.StringVal(keyName),
			})
			if elemCtx.ParentLocalAddress != nil {
				elemCtx.ParentLocalAddress = append(elemCtx.ParentLocalAddress, lang.IndexStep{
					Key: cty.StringVal(keyName),
				})
			}

			elemTargets = append(elemTargets, e.ReferenceTargets(ctx, elemCtx)...)
		}
	}

	targets := make(reference.Targets, 0)

	if targetCtx != nil {
		// collect target for the whole map

		// type-aware
		elemCons, ok := m.cons.Elem.(schema.TypeAwareConstraint)
		if targetCtx.AsExprType && ok {
			elemType, ok := elemCons.ConstraintType()
			if ok {
				targets = append(targets, reference.Target{
					Addr:                   targetCtx.ParentAddress,
					Name:                   targetCtx.FriendlyName,
					Type:                   cty.Map(elemType),
					ScopeId:                targetCtx.ScopeId,
					RangePtr:               m.expr.Range().Ptr(),
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
				RangePtr:               m.expr.Range().Ptr(),
				NestedTargets:          elemTargets,
				LocalAddr:              targetCtx.ParentLocalAddress,
				TargetableFromRangePtr: targetCtx.TargetableFromRangePtr,
			})
		}
	} else {
		// treat element targets as 1st class ones
		// if the map itself isn't targetable
		targets = elemTargets
	}

	return targets
}
