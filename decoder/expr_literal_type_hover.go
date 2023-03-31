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

func (lt LiteralType) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	typ := lt.cons.Type

	if typ == cty.DynamicPseudoType {
		val, diags := lt.expr.Value(nil)
		if !diags.HasErrors() {
			typ = val.Type()
		}
	}

	// string is a special case as it's always represented like a template
	// even if there's no templating involved
	if typ == cty.String {
		expr, ok := lt.expr.(*hclsyntax.TemplateExpr)
		if !ok {
			return nil
		}
		if expr.IsStringLiteral() || isMultilineStringLiteral(expr) {
			return &lang.HoverData{
				Content: lang.Markdown(fmt.Sprintf(`_%s_`, typ.FriendlyName())),
				Range:   expr.Range(),
			}
		}

		return nil
	}

	if typ.IsPrimitiveType() {
		expr, ok := lt.expr.(*hclsyntax.LiteralValueExpr)
		if !ok {
			return nil
		}

		if !lt.cons.Type.Equals(expr.Val.Type()) {
			return nil
		}

		return &lang.HoverData{
			Content: lang.Markdown(fmt.Sprintf(`_%s_`, typ.FriendlyName())),
			Range:   expr.Range(),
		}
	}

	if typ.IsListType() {
		expr, ok := lt.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return nil
		}

		cons := schema.List{
			Elem: schema.LiteralType{
				Type: typ.ElementType(),
			},
		}

		return newExpression(lt.pathCtx, expr, cons).HoverAtPos(ctx, pos)
	}

	if typ.IsSetType() {
		expr, ok := lt.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return nil
		}

		cons := schema.Set{
			Elem: schema.LiteralType{
				Type: typ.ElementType(),
			},
		}

		return newExpression(lt.pathCtx, expr, cons).HoverAtPos(ctx, pos)
	}

	if typ.IsTupleType() {
		expr, ok := lt.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return nil
		}

		elemTypes := typ.TupleElementTypes()
		cons := schema.Tuple{
			Elems: make([]schema.Constraint, len(elemTypes)),
		}
		for i, elemType := range elemTypes {
			cons.Elems[i] = schema.LiteralType{
				Type: elemType,
			}
		}

		return newExpression(lt.pathCtx, expr, cons).HoverAtPos(ctx, pos)
	}

	if typ.IsMapType() {
		expr, ok := lt.expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return nil
		}

		cons := schema.Map{
			Elem: schema.LiteralType{
				Type: typ.ElementType(),
			},
		}
		return newExpression(lt.pathCtx, expr, cons).HoverAtPos(ctx, pos)
	}

	if typ.IsObjectType() {
		expr, ok := lt.expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return nil
		}

		cons := schema.Object{
			Attributes: ctyObjectToObjectAttributes(typ),
		}
		return newExpression(lt.pathCtx, expr, cons).HoverAtPos(ctx, pos)
	}

	return nil
}

func isMultilineStringLiteral(tplExpr *hclsyntax.TemplateExpr) bool {
	if len(tplExpr.Parts) < 1 {
		return false
	}
	for _, part := range tplExpr.Parts {
		expr, ok := part.(*hclsyntax.LiteralValueExpr)
		if !ok {
			return false
		}
		if expr.Val.Type() != cty.String {
			return false
		}
	}
	return true
}
