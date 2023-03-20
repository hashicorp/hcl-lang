package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/zclconf/go-cty/cty"
)

func (obj Object) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	if json.IsJSONExpression(obj.expr) {
		targets := make(reference.Targets, 0)

		if targetCtx != nil {
			// collect target for the whole object
			rangePtr := obj.expr.Range().Ptr()
			if targetCtx.ParentRangePtr != nil {
				rangePtr = targetCtx.ParentRangePtr
			}

			// type-aware
			if targetCtx.AsExprType {
				objType, ok := obj.cons.ConstraintType()
				if ok {
					targets = append(targets, reference.Target{
						Addr:                   targetCtx.ParentAddress,
						Name:                   targetCtx.FriendlyName,
						Type:                   objType,
						ScopeId:                targetCtx.ScopeId,
						DefRangePtr:            targetCtx.ParentDefRangePtr,
						RangePtr:               rangePtr,
						NestedTargets:          reference.Targets{},
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
					DefRangePtr:            targetCtx.ParentDefRangePtr,
					RangePtr:               rangePtr,
					NestedTargets:          reference.Targets{},
					LocalAddr:              targetCtx.ParentLocalAddress,
					TargetableFromRangePtr: targetCtx.TargetableFromRangePtr,
				})
			}
		}

		return targets
	}

	eType, ok := obj.expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return reference.Targets{}
	}

	if len(obj.cons.Attributes) == 0 {
		return reference.Targets{}
	}

	attrTargets := make(reference.Targets, 0)

	declaredAttributes := make(map[string]hclsyntax.ObjectConsItem, 0)
	for _, item := range eType.Items {
		keyName, _, ok := rawObjectKey(item.KeyExpr)
		if !ok {
			// avoid collecting item w/ invalid key
			continue
		}

		_, ok = obj.cons.Attributes[keyName]
		if !ok {
			// avoid collecting for unknown attribute
			continue
		}

		declaredAttributes[keyName] = item
	}

	attrNames := sortedAttributeNames(obj.cons.Attributes)
	for _, name := range attrNames {
		var valueExpr hcl.Expression
		item, ok := declaredAttributes[name]
		if ok {
			valueExpr = item.ValueExpr
		} else {
			valueExpr = newEmptyExpressionAtPos(eType.Range().Filename, eType.Range().Start)
		}

		aSchema := obj.cons.Attributes[name]
		expr := newExpression(obj.pathCtx, valueExpr, aSchema.Constraint)
		if e, ok := expr.(ReferenceTargetsExpression); ok {
			if targetCtx == nil {
				// collect any targets inside the expression
				// if attribute itself isn't targetable
				attrTargets = append(attrTargets, e.ReferenceTargets(ctx, nil)...)
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

			attrTargets = append(attrTargets, e.ReferenceTargets(ctx, elemCtx)...)
		}
	}

	targets := make(reference.Targets, 0)

	if targetCtx != nil {
		// collect target for the whole object

		// type-aware
		if targetCtx.AsExprType {
			objType, ok := obj.cons.ConstraintType()
			if ok {
				targets = append(targets, reference.Target{
					Addr:                   targetCtx.ParentAddress,
					Name:                   targetCtx.FriendlyName,
					Type:                   objType,
					ScopeId:                targetCtx.ScopeId,
					RangePtr:               obj.expr.Range().Ptr(),
					NestedTargets:          attrTargets,
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
				RangePtr:               obj.expr.Range().Ptr(),
				NestedTargets:          attrTargets,
				LocalAddr:              targetCtx.ParentLocalAddress,
				TargetableFromRangePtr: targetCtx.TargetableFromRangePtr,
			})
		}
	} else {
		// treat element targets as 1st class ones
		// if the object itself isn't targetable
		targets = attrTargets
	}

	return targets
}
