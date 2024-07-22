// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (obj Object) ReferenceOrigins(ctx context.Context) reference.Origins {
	items, diags := hcl.ExprMap(obj.expr)
	if diags.HasErrors() {
		return reference.Origins{}
	}

	if len(items) == 0 || len(obj.cons.Attributes) == 0 {
		return reference.Origins{}
	}

	origins := make(reference.Origins, 0)

	for _, item := range items {
		attrName, _, isRawKey := rawObjectKey(item.Key)

		var aSchema *schema.AttributeSchema
		var isKnownAttr bool
		if isRawKey {
			aSchema, isKnownAttr = obj.cons.Attributes[attrName]
		}

		keyExpr, ok := item.Key.(*hclsyntax.ObjectConsKeyExpr)
		if ok {
			parensExpr, ok := keyExpr.Wrapped.(*hclsyntax.ParenthesesExpr)
			if ok {
				keyCons := schema.AnyExpression{
					OfType: cty.String,
				}
				kExpr := newExpression(obj.pathCtx, parensExpr, keyCons)
				if expr, ok := kExpr.(ReferenceOriginsExpression); ok {
					origins = append(origins, expr.ReferenceOrigins(ctx)...)
				}
			}
		}

		if isKnownAttr {
			expr := newExpression(obj.pathCtx, item.Value, aSchema.Constraint)
			if elemExpr, ok := expr.(ReferenceOriginsExpression); ok {
				origins = append(origins, elemExpr.ReferenceOrigins(ctx)...)
			}
		}
	}

	return origins
}
