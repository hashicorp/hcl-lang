// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type List struct {
	expr    hcl.Expression
	cons    schema.List
	pathCtx *PathContext
}

func (list List) InferType() (cty.Type, bool) {
	if isEmptyExpression(list.expr) {
		return list.cons.ConstraintType()
	}

	elems, diags := hcl.ExprList(list.expr)
	if diags.HasErrors() {
		return list.cons.ConstraintType()
	}

	if len(elems) == 0 {
		return list.cons.ConstraintType()
	}

	elemType, ok := list.inferExprListElemType(elems)
	if !ok {
		return list.cons.ConstraintType()
	}

	return cty.List(elemType), true
}

func (list List) inferExprListElemType(elems []hcl.Expression) (cty.Type, bool) {
	var firstElemType cty.Type

	// Try to infer element type from declared element expressions
	for _, elemExpr := range elems {
		elem, ok := newExpression(list.pathCtx, elemExpr, list.cons.Elem).(CanInferTypeExpression)
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
