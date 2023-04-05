// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

func (set Set) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	if isEmptyExpression(set.expr) && targetCtx != nil {
		return set.wholeSetReferenceTargets(targetCtx, nil)
	}

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

	if targetCtx == nil {
		// treat element targets as 1st class ones
		// if the list itself isn't targetable
		return elemTargets
	}

	return set.wholeSetReferenceTargets(targetCtx, elemTargets)
}

func (set Set) wholeSetReferenceTargets(targetCtx *TargetContext, nestedTargets reference.Targets) reference.Targets {
	targets := make(reference.Targets, 0)

	// collect target for the whole set

	var rangePtr *hcl.Range
	if targetCtx.ParentRangePtr != nil {
		rangePtr = targetCtx.ParentRangePtr
	} else {
		rangePtr = set.expr.Range().Ptr()
	}

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
