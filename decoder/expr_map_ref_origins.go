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

func (m Map) ReferenceOrigins(ctx context.Context) reference.Origins {
	items, diags := hcl.ExprMap(m.expr)
	if diags.HasErrors() {
		return reference.Origins{}
	}

	if len(items) == 0 || m.cons.Elem == nil {
		return reference.Origins{}
	}

	origins := make(reference.Origins, 0)

	for _, item := range items {
		keyExpr, ok := item.Key.(*hclsyntax.ObjectConsKeyExpr)
		if ok {
			parensExpr, ok := keyExpr.Wrapped.(*hclsyntax.ParenthesesExpr)
			if ok {
				keyCons := schema.AnyExpression{
					OfType: cty.String,
				}
				kExpr := newExpression(m.pathCtx, parensExpr, keyCons)
				if expr, ok := kExpr.(ReferenceOriginsExpression); ok {
					origins = append(origins, expr.ReferenceOrigins(ctx)...)
				}
			}
		}

		valExpr := newExpression(m.pathCtx, item.Value, m.cons.Elem)
		if expr, ok := valExpr.(ReferenceOriginsExpression); ok {
			origins = append(origins, expr.ReferenceOrigins(ctx)...)
		}
	}

	return origins
}
