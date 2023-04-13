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

func (lv LiteralValue) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	typ := lv.cons.Value.Type()

	// string is a special case as it's always represented like a template
	// even if there's no templating involved
	if typ == cty.String {
		expr, ok := lv.expr.(*hclsyntax.TemplateExpr)
		if !ok {
			return nil
		}

		val, diags := expr.Value(nil)
		if diags.HasErrors() {
			return nil
		}
		if !lv.cons.Value.RawEquals(val) {
			return nil
		}

		if expr.IsStringLiteral() || isMultilineStringLiteral(expr) {
			content := fmt.Sprintf(`_%s_`, typ.FriendlyName())
			if lv.cons.Description.Value != "" {
				content += "\n\n" + lv.cons.Description.Value
			}

			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   expr.Range(),
			}
		}

		return nil
	}

	if typ.IsPrimitiveType() {
		expr, ok := lv.expr.(*hclsyntax.LiteralValueExpr)
		if !ok {
			return nil
		}

		val, diags := expr.Value(nil)
		if diags.HasErrors() {
			return nil
		}
		if !lv.cons.Value.RawEquals(val) {
			return nil
		}

		content := fmt.Sprintf(`_%s_`, typ.FriendlyName())
		if lv.cons.Description.Value != "" {
			content += "\n\n" + lv.cons.Description.Value
		}

		return &lang.HoverData{
			Content: lang.Markdown(content),
			Range:   expr.Range(),
		}
	}

	if typ.IsListType() {
		expr, ok := lv.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return nil
		}

		values := lv.cons.Value.AsValueSlice()
		for i, elemExpr := range expr.Exprs {
			if len(values) < i+1 {
				return nil
			}

			if elemExpr.Range().ContainsPos(pos) {
				elemCons := schema.LiteralValue{
					Value: values[i],
				}

				val, diags := elemExpr.Value(nil)
				if diags.HasErrors() {
					continue
				}
				if !elemCons.Value.RawEquals(val) {
					return nil
				}

				expr := newExpression(lv.pathCtx, elemExpr, elemCons)
				return expr.HoverAtPos(ctx, pos)
			}
		}

		cons := schema.List{
			Elem: schema.LiteralType{
				Type: typ.ElementType(),
			},
			Description: lv.cons.Description,
		}

		return newExpression(lv.pathCtx, expr, cons).HoverAtPos(ctx, pos)
	}

	if typ.IsSetType() {
		expr, ok := lv.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return nil
		}

		values := lv.cons.Value.AsValueSet()
		for i, elemExpr := range expr.Exprs {
			if values.Length() < i+1 {
				return nil
			}

			if elemExpr.Range().ContainsPos(pos) {
				val, diags := elemExpr.Value(nil)
				if diags.HasErrors() {
					continue
				}

				if !values.ElementType().Equals(val.Type()) {
					return nil
				}

				if !values.Has(val) {
					return nil
				}

				elemCons := schema.LiteralValue{
					Value: val,
				}

				expr := newExpression(lv.pathCtx, elemExpr, elemCons)
				return expr.HoverAtPos(ctx, pos)
			}
		}

		cons := schema.Set{
			Elem: schema.LiteralType{
				Type: typ.ElementType(),
			},
			Description: lv.cons.Description,
		}

		return newExpression(lv.pathCtx, expr, cons).HoverAtPos(ctx, pos)
	}

	if typ.IsTupleType() {
		expr, ok := lv.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return nil
		}

		elemTypes := typ.TupleElementTypes()
		cons := schema.Tuple{
			Elems:       make([]schema.Constraint, len(elemTypes)),
			Description: lv.cons.Description,
		}
		for i, elemType := range elemTypes {
			cons.Elems[i] = schema.LiteralType{
				Type: elemType,
			}
		}

		return newExpression(lv.pathCtx, expr, cons).HoverAtPos(ctx, pos)
	}

	if typ.IsMapType() {
		expr, ok := lv.expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return nil
		}

		values := lv.cons.Value.AsValueMap()

		for _, item := range expr.Items {
			keyStr, _, ok := rawObjectKey(item.KeyExpr)
			if !ok {
				return nil
			}

			if _, ok := values[keyStr]; !ok {
				return nil
			}

			if item.ValueExpr.Range().ContainsPos(pos) {
				val, diags := item.ValueExpr.Value(nil)
				if diags.HasErrors() {
					continue
				}

				if !values[keyStr].RawEquals(val) {
					return nil
				}

				elemCons := schema.LiteralValue{
					Value: val,
				}

				expr := newExpression(lv.pathCtx, item.ValueExpr, elemCons)
				return expr.HoverAtPos(ctx, pos)
			}
		}

		cons := schema.Map{
			Elem: schema.LiteralType{
				Type: typ.ElementType(),
			},
			Description: lv.cons.Description,
		}
		return newExpression(lv.pathCtx, expr, cons).HoverAtPos(ctx, pos)
	}

	if typ.IsObjectType() {
		expr, ok := lv.expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return nil
		}

		cons := schema.Object{
			Attributes:  ctyObjectToObjectAttributes(typ),
			Description: lv.cons.Description,
		}
		return newExpression(lv.pathCtx, expr, cons).HoverAtPos(ctx, pos)
	}

	return nil
}
