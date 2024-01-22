// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (m Map) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	eType, ok := m.expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return nil
	}

	for _, item := range eType.Items {
		if item.KeyExpr.Range().ContainsPos(pos) {
			keyExpr, ok := item.KeyExpr.(*hclsyntax.ObjectConsKeyExpr)
			if ok && m.cons.AllowInterpolatedKeys {
				parensExpr, ok := keyExpr.Wrapped.(*hclsyntax.ParenthesesExpr)
				if ok {
					keyCons := schema.AnyExpression{
						OfType: cty.String,
					}
					expr := newExpression(m.pathCtx, parensExpr, keyCons)
					return expr.HoverAtPos(ctx, pos)
				}
			}
			return nil
		}

		if item.ValueExpr.Range().ContainsPos(pos) {
			expr := newExpression(m.pathCtx, item.ValueExpr, m.cons.Elem)
			return expr.HoverAtPos(ctx, pos)
		}
	}

	content := fmt.Sprintf("_%s_", m.cons.FriendlyName())
	if m.cons.Description.Value != "" {
		content += "\n\n" + m.cons.Description.Value
	}

	return &lang.HoverData{
		Content: lang.Markdown(content),
		Range:   eType.Range(),
	}
}
