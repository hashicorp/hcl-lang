// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type LiteralType struct {
	expr hcl.Expression
	cons schema.LiteralType

	pathCtx *PathContext
}

func (lt LiteralType) InferType() (cty.Type, bool) {
	consType, ok := lt.cons.ConstraintType()
	if !ok {
		return consType, false
	}

	if consType == cty.DynamicPseudoType && !isEmptyExpression(lt.expr) {
		val, diags := lt.expr.Value(nil)
		if !diags.HasErrors() {
			consType = val.Type()
		}
	}

	return consType, true
}
