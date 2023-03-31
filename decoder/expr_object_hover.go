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
)

func (obj Object) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	eType, ok := obj.expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return nil
	}

	for _, item := range eType.Items {
		attrName, _, ok := rawObjectKey(item.KeyExpr)
		if !ok {
			continue
		}

		aSchema, ok := obj.cons.Attributes[attrName]
		if !ok {
			// unknown attribute
			continue
		}

		if item.KeyExpr.Range().ContainsPos(pos) {
			itemRng := hcl.RangeBetween(item.KeyExpr.Range(), item.ValueExpr.Range())
			content := hoverContentForAttribute(attrName, aSchema)

			return &lang.HoverData{
				Content: content,
				Range:   itemRng,
			}
		}

		if item.ValueExpr.Range().ContainsPos(pos) {
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
