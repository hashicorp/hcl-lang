// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"sort"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

func (m Map) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	if isEmptyExpression(m.expr) && targetCtx != nil {
		return m.wholeMapReferenceTargets(targetCtx, nil)
	}

	items, diags := hcl.ExprMap(m.expr)
	if diags.HasErrors() {
		return reference.Targets{}
	}

	if m.cons.Elem == nil {
		return reference.Targets{}
	}

	elemTargets := make(reference.Targets, 0)

	for _, item := range items {
		keyName, _, ok := rawObjectKey(item.Key)
		if !ok {
			// avoid collecting item w/ invalid key
			continue
		}

		expr := newExpression(m.pathCtx, item.Value, m.cons.Elem)
		if e, ok := expr.(ReferenceTargetsExpression); ok {
			if targetCtx == nil {
				// collect any targets inside the expression
				// if attribute itself isn't targetable
				elemTargets = append(elemTargets, e.ReferenceTargets(ctx, nil)...)
				continue
			}

			elemCtx := targetCtx.Copy()

			elemCtx.ParentDefRangePtr = item.Key.Range().Ptr()
			elemCtx.ParentRangePtr = hcl.RangeBetween(item.Key.Range(), item.Value.Range()).Ptr()

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

	sort.Sort(elemTargets)

	if targetCtx == nil {
		// treat element targets as 1st class ones
		// if the map itself isn't targetable
		return elemTargets
	}

	return m.wholeMapReferenceTargets(targetCtx, elemTargets)
}

func (m Map) wholeMapReferenceTargets(targetCtx *TargetContext, nestedTargets reference.Targets) reference.Targets {
	// collect targets for the whole map
	targets := make(reference.Targets, 0)

	var rangePtr *hcl.Range
	if targetCtx.ParentRangePtr != nil {
		rangePtr = targetCtx.ParentRangePtr
	} else {
		rangePtr = m.expr.Range().Ptr()
	}

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
