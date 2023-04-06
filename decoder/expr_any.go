// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Any struct {
	expr    hcl.Expression
	cons    schema.AnyExpression
	pathCtx *PathContext
}

func (a Any) InferType() (cty.Type, bool) {
	consType, ok := a.cons.ConstraintType()
	if !ok {
		return consType, false
	}

	if consType == cty.DynamicPseudoType && !isEmptyExpression(a.expr) {
		val, diags := a.expr.Value(nil)
		if !diags.HasErrors() {
			consType = val.Type()
		}
	}

	return consType, true
}
