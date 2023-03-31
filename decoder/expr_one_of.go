// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type OneOf struct {
	expr    hcl.Expression
	cons    schema.OneOf
	pathCtx *PathContext
}

func (oo OneOf) InferType() (cty.Type, bool) {
	consType, ok := oo.cons.ConstraintType()
	if !ok {
		return consType, false
	}

	if consType == cty.DynamicPseudoType && !isEmptyExpression(oo.expr) {
		for _, cons := range oo.cons {
			c, ok := cons.(CanInferTypeExpression)
			if !ok {
				continue
			}
			typ, ok := c.InferType()
			if !ok {
				continue
			}

			// Picking first type-aware constraint may not always be
			// appropriate since we cannot match it against configuration,
			// but it is mostly a pragmatic choice to mimic existing behaviours
			// based on common schema, such as OneOf{Reference{}, LiteralType{}}.
			// TODO: Revisit when AnyExpression{} is implemented & rolled out
			return typ, true
		}
	}

	return consType, true
}
