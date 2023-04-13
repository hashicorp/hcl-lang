// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Object struct {
	expr    hcl.Expression
	cons    schema.Object
	pathCtx *PathContext
}

func (obj Object) InferType() (cty.Type, bool) {
	if isEmptyExpression(obj.expr) {
		return obj.cons.ConstraintType()
	}

	elems, diags := hcl.ExprMap(obj.expr)
	if diags.HasErrors() {
		return obj.cons.ConstraintType()
	}

	if len(elems) == 0 {
		return obj.cons.ConstraintType()
	}

	attrTypes, ok := obj.inferExprObjectAttrTypes(elems)
	if !ok {
		return obj.cons.ConstraintType()
	}

	return cty.Object(attrTypes), true
}

func (obj Object) inferExprObjectAttrTypes(kvPairs []hcl.KeyValuePair) (map[string]cty.Type, bool) {
	declaredAttributes := make(map[string]hcl.Expression)
	for _, kvPair := range kvPairs {
		keyName, _, ok := rawObjectKey(kvPair.Key)
		if !ok {
			// avoid collecting pair w/ invalid key
			continue
		}

		_, ok = obj.cons.Attributes[keyName]
		if !ok {
			// avoid collecting for unknown attribute
			continue
		}

		declaredAttributes[keyName] = kvPair.Value
	}

	attrTypes := make(map[string]cty.Type)
	for name, aSchema := range obj.cons.Attributes {
		valueExpr, ok := declaredAttributes[name]
		if !ok {
			valueExpr = newEmptyExpressionAtPos(obj.expr.Range().Filename, obj.expr.Range().Start)
		}

		expr, ok := newExpression(obj.pathCtx, valueExpr, aSchema.Constraint).(CanInferTypeExpression)
		if !ok {
			return map[string]cty.Type{}, false
		}
		attrType, ok := expr.InferType()
		if !ok {
			return map[string]cty.Type{}, false
		}
		attrTypes[name] = attrType
	}
	return attrTypes, true
}
