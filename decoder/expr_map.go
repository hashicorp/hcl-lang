// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Map struct {
	expr    hcl.Expression
	cons    schema.Map
	pathCtx *PathContext
}

func (m Map) InferType() (cty.Type, bool) {
	if isEmptyExpression(m.expr) {
		return m.cons.ConstraintType()
	}

	elems, diags := hcl.ExprMap(m.expr)
	if diags.HasErrors() {
		return m.cons.ConstraintType()
	}

	if len(elems) == 0 {
		return m.cons.ConstraintType()
	}

	elemType, ok := m.inferExprMapElemType(elems)
	if !ok {
		return m.cons.ConstraintType()
	}

	return cty.Map(elemType), true
}

func (m Map) inferExprMapElemType(kvPairs []hcl.KeyValuePair) (cty.Type, bool) {
	var firstElemType cty.Type
	for _, kvPair := range kvPairs {
		elemExpr, ok := newExpression(m.pathCtx, kvPair.Value, m.cons.Elem).(CanInferTypeExpression)
		if !ok {
			return cty.NilType, false
		}
		elemType, ok := elemExpr.InferType()
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
