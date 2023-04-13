// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Set struct {
	expr    hcl.Expression
	cons    schema.Set
	pathCtx *PathContext
}

func (set Set) InferType() (cty.Type, bool) {
	if isEmptyExpression(set.expr) {
		return set.cons.ConstraintType()
	}

	elems, diags := hcl.ExprList(set.expr)
	if diags.HasErrors() {
		return set.cons.ConstraintType()
	}

	if len(elems) == 0 {
		return set.cons.ConstraintType()
	}

	elemType, ok := set.inferExprSetElemType(elems)
	if !ok {
		return set.cons.ConstraintType()
	}

	return cty.Set(elemType), true
}

func (set Set) inferExprSetElemType(elems []hcl.Expression) (cty.Type, bool) {
	var firstElemType cty.Type

	// Try to infer element type from declared element expressions
	for _, elemExpr := range elems {
		elem, ok := newExpression(set.pathCtx, elemExpr, set.cons.Elem).(CanInferTypeExpression)
		if !ok {
			return cty.NilType, false
		}
		elemType, ok := elem.InferType()
		if !ok {
			return cty.NilType, false
		}
		if firstElemType == cty.NilType {
			firstElemType = elemType
			continue
		}
		if !firstElemType.Equals(elemType) {
			// elements of mismatching type
			return cty.NilType, false
		}
	}

	return firstElemType, true
}
