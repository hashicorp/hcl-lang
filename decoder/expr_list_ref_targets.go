// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

func (list List) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	if isEmptyExpression(list.expr) && targetCtx != nil {
		return list.wholeListReferenceTargets(targetCtx, nil)
	}

	elems, diags := hcl.ExprList(list.expr)
	if diags.HasErrors() {
		return reference.Targets{}
	}

	if list.cons.Elem == nil {
		return reference.Targets{}
	}

	elemTargets := make(reference.Targets, 0)

	for i, elemExpr := range elems {
		expr := newExpression(list.pathCtx, elemExpr, list.cons.Elem)
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

	if targetCtx == nil {
		// treat element targets as 1st class ones
		// if the list itself isn't targetable
		return elemTargets
	}

	return list.wholeListReferenceTargets(targetCtx, elemTargets)
}

func (list List) wholeListReferenceTargets(targetCtx *TargetContext, nestedTargets reference.Targets) reference.Targets {
	targets := make(reference.Targets, 0)

	// collect target for the whole list
	var rangePtr *hcl.Range
	if targetCtx.ParentRangePtr != nil {
		rangePtr = targetCtx.ParentRangePtr
	} else {
		rangePtr = list.expr.Range().Ptr()
	}

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
				RangePtr:               rangePtr,
				DefRangePtr:            targetCtx.ParentDefRangePtr,
				NestedTargets:          nestedTargets,
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
			NestedTargets:          nestedTargets,
			LocalAddr:              targetCtx.ParentLocalAddress,
			TargetableFromRangePtr: targetCtx.TargetableFromRangePtr,
		})
	}

	return targets
}
