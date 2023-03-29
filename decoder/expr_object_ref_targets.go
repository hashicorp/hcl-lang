package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (obj Object) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	items, diags := hcl.ExprMap(obj.expr)
	if diags.HasErrors() {
		return reference.Targets{}
	}

	if len(obj.cons.Attributes) == 0 {
		return reference.Targets{}
	}

	attrTargets := make(reference.Targets, 0)

	declaredAttributes := make(map[string]hcl.KeyValuePair, 0)
	for _, item := range items {
		keyName, _, ok := rawObjectKey(item.Key)
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
		item, attrDeclared := declaredAttributes[name]
		if attrDeclared {
			valueExpr = item.Value
		} else {
			valueExpr = newEmptyExpressionAtPos(obj.expr.Range().Filename, obj.expr.Range().Start)
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

			if attrDeclared {
				elemCtx.ParentDefRangePtr = item.Key.Range().Ptr()
				elemCtx.ParentRangePtr = hcl.RangeBetween(item.Key.Range(), item.Value.Range()).Ptr()
			}

			if hclsyntax.ValidIdentifier(name) {
				// Prefer simpler syntax - e.g. myobj.attribute if possible
				elemCtx.ParentAddress = append(elemCtx.ParentAddress, lang.AttrStep{
					Name: name,
				})
				if elemCtx.ParentLocalAddress != nil {
					elemCtx.ParentLocalAddress = append(elemCtx.ParentLocalAddress, lang.AttrStep{
						Name: name,
					})
				}
			} else {
				// Fall back to indexing syntax - e.g. myobj["attr-foo"]
				elemCtx.ParentAddress = append(elemCtx.ParentAddress, lang.IndexStep{
					Key: cty.StringVal(name),
				})
				if elemCtx.ParentLocalAddress != nil {
					elemCtx.ParentLocalAddress = append(elemCtx.ParentLocalAddress, lang.IndexStep{
						Key: cty.StringVal(name),
					})
				}
			}

			attrTargets = append(attrTargets, e.ReferenceTargets(ctx, elemCtx)...)
		}
	}

	targets := make(reference.Targets, 0)

	if targetCtx != nil {
		// collect target for the whole object

		var rangePtr *hcl.Range
		if targetCtx.ParentRangePtr != nil {
			rangePtr = targetCtx.ParentRangePtr
		} else {
			rangePtr = obj.expr.Range().Ptr()
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
				DefRangePtr:            targetCtx.ParentDefRangePtr,
				RangePtr:               rangePtr,
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
