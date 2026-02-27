// Copyright IBM Corp. 2026
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

func (obj Object) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	eType, ok := obj.expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return nil
	}

	for _, item := range eType.Items {
		attrName, _, isRawKey := rawObjectKey(item.KeyExpr)

		var aSchema *schema.AttributeSchema
		var isKnownAttr bool
		if isRawKey {
			aSchema, isKnownAttr = obj.cons.Attributes[attrName]
		}

		if item.KeyExpr.Range().ContainsPos(pos) {
			// handle any interpolation if it is allowed
			keyExpr, ok := item.KeyExpr.(*hclsyntax.ObjectConsKeyExpr)
			if ok && obj.cons.AllowInterpolatedKeys {
				parensExpr, ok := keyExpr.Wrapped.(*hclsyntax.ParenthesesExpr)
				if ok {
					keyCons := schema.AnyExpression{
						OfType: cty.String,
					}
					return newExpression(obj.pathCtx, parensExpr, keyCons).HoverAtPos(ctx, pos)
				}
			}

			if isKnownAttr {
				itemRng := hcl.RangeBetween(item.KeyExpr.Range(), item.ValueExpr.Range())
				content := hoverContentForAttribute(attrName, aSchema)

				return &lang.HoverData{
					Content: content,
					Range:   itemRng,
				}
			}
		}

		if isKnownAttr && item.ValueExpr.Range().ContainsPos(pos) {
			expr := newExpression(obj.pathCtx, item.ValueExpr, aSchema.Constraint)
			return expr.HoverAtPos(ctx, pos)
		}
	}

	content := ""
	hoverData := obj.cons.EmptyHoverData(0)
	if hoverData != nil {
		content = hoverData.Content.Value
	}
	content += fmt.Sprintf("_%s_", obj.cons.FriendlyName())
	if obj.cons.Description.Value != "" {
		content += "\n\n" + obj.cons.Description.Value
	}

	return &lang.HoverData{
		Content: lang.Markdown(content),
		Range:   eType.Range(),
	}
}

func hoverContentForAttribute(name string, aSchema *schema.AttributeSchema) lang.MarkupContent {
	value := fmt.Sprintf("**%s** _%s_", name, detailForAttribute(aSchema))
	if aSchema.Description.Value != "" {
		value += fmt.Sprintf("\n\n%s", aSchema.Description.Value)
	}
	return lang.MarkupContent{
		Kind:  lang.MarkdownKind,
		Value: value,
	}
}
