// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

func (tuple Tuple) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	if isEmptyExpression(tuple.expr) && targetCtx != nil {
		return tuple.wholeTupleReferenceTargets(targetCtx, tuple.collectTupleElemTargets(ctx, targetCtx, []hcl.Expression{}))
	}

	elems, diags := hcl.ExprList(tuple.expr)
	if diags.HasErrors() {
		return reference.Targets{}
	}

	if len(tuple.cons.Elems) == 0 {
		return reference.Targets{}
	}

	elemTargets := tuple.collectTupleElemTargets(ctx, targetCtx, elems)

	if targetCtx == nil {
		// treat element targets as 1st class ones
		// if the tuple itself isn't targetable
		return elemTargets
	}

	return tuple.wholeTupleReferenceTargets(targetCtx, elemTargets)
}

func (tuple Tuple) collectTupleElemTargets(ctx context.Context, targetCtx *TargetContext, declaredElems []hcl.Expression) reference.Targets {
	elemTargets := make(reference.Targets, 0)

	for i, elemCons := range tuple.cons.Elems {
		var elemExpr hcl.Expression
		if len(declaredElems) >= i+1 {
			elemExpr = declaredElems[i]
		} else {
			elemExpr = newEmptyExpressionAtPos(tuple.expr.Range().Filename, tuple.expr.Range().Start)
		}

		expr := newExpression(tuple.pathCtx, elemExpr, elemCons)
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

	return elemTargets
}

func (tuple Tuple) wholeTupleReferenceTargets(targetCtx *TargetContext, nestedTargets reference.Targets) reference.Targets {
	targets := make(reference.Targets, 0)

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
