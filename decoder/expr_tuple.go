// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Tuple struct {
	expr    hcl.Expression
	cons    schema.Tuple
	pathCtx *PathContext
}

func (tuple Tuple) InferType() (cty.Type, bool) {
	if isEmptyExpression(tuple.expr) {
		return tuple.cons.ConstraintType()
	}

	elems, diags := hcl.ExprList(tuple.expr)
	if diags.HasErrors() {
		return tuple.cons.ConstraintType()
	}

	if len(elems) == 0 {
		return tuple.cons.ConstraintType()
	}

	elemTypes, ok := tuple.inferExprTupleElemTypes(elems)
	if !ok {
		return tuple.cons.ConstraintType()
	}

	return cty.Tuple(elemTypes), true
}

func (tuple Tuple) inferExprTupleElemTypes(elems []hcl.Expression) ([]cty.Type, bool) {
	elemTypes := make([]cty.Type, len(tuple.cons.Elems))

	for i, elemCons := range tuple.cons.Elems {
		if len(elems) < i+1 {
			return []cty.Type{}, false
		}

		elemExpr := elems[i]
		elem, ok := newExpression(tuple.pathCtx, elemExpr, elemCons).(CanInferTypeExpression)
		if !ok {
			return []cty.Type{}, false
		}

		elemType, ok := elem.InferType()
		if !ok {
			return []cty.Type{}, false
		}
		elemTypes[i] = elemType
	}

	return elemTypes, true
}
