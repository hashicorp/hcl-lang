// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (a Any) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	typ := a.cons.OfType

	if typ.IsListType() {
		expr, ok := a.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return a.hoverNonComplexExprAtPos(ctx, pos)
		}

		cons := schema.List{
			Elem: schema.AnyExpression{
				OfType: typ.ElementType(),
			},
		}

		return newExpression(a.pathCtx, expr, cons).HoverAtPos(ctx, pos)
	}

	if typ.IsSetType() {
		expr, ok := a.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return a.hoverNonComplexExprAtPos(ctx, pos)
		}

		cons := schema.Set{
			Elem: schema.AnyExpression{
				OfType: typ.ElementType(),
			},
		}

		return newExpression(a.pathCtx, expr, cons).HoverAtPos(ctx, pos)
	}

	if typ.IsTupleType() {
		expr, ok := a.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return a.hoverNonComplexExprAtPos(ctx, pos)
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

		return newExpression(a.pathCtx, expr, cons).HoverAtPos(ctx, pos)
	}

	if typ.IsMapType() {
		expr, ok := a.expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return a.hoverNonComplexExprAtPos(ctx, pos)
		}

		cons := schema.Map{
			Elem: schema.AnyExpression{
				OfType: typ.ElementType(),
			},
		}
		return newExpression(a.pathCtx, expr, cons).HoverAtPos(ctx, pos)
	}

	if typ.IsObjectType() {
		expr, ok := a.expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return a.hoverNonComplexExprAtPos(ctx, pos)
		}

		cons := schema.Object{
			Attributes: ctyObjectToObjectAttributes(typ),
		}
		return newExpression(a.pathCtx, expr, cons).HoverAtPos(ctx, pos)
	}

	return a.hoverNonComplexExprAtPos(ctx, pos)
}

func (a Any) hoverNonComplexExprAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	// TODO: Support TemplateExpr https://github.com/hashicorp/terraform-ls/issues/522
	// TODO: Support splat expression https://github.com/hashicorp/terraform-ls/issues/526
	// TODO: Support for-in-if expression https://github.com/hashicorp/terraform-ls/issues/527
	// TODO: Support conditional expression https://github.com/hashicorp/terraform-ls/issues/528
	// TODO: Support operator expresssions https://github.com/hashicorp/terraform-ls/issues/529
	// TODO: Support complex index expressions https://github.com/hashicorp/terraform-ls/issues/531
	// TODO: Support relative traversals https://github.com/hashicorp/terraform-ls/issues/532

	ref := Reference{
		expr:    a.expr,
		cons:    schema.Reference{OfType: a.cons.OfType},
		pathCtx: a.pathCtx,
	}
	hoverData := ref.HoverAtPos(ctx, pos)
	if hoverData != nil {
		return hoverData
	}

	fe := functionExpr{
		expr:       a.expr,
		returnType: a.cons.OfType,
		pathCtx:    a.pathCtx,
	}
	hoverData = fe.HoverAtPos(ctx, pos)
	if hoverData != nil {
		return hoverData
	}

	lt := LiteralType{
		expr:    a.expr,
		cons:    schema.LiteralType{Type: a.cons.OfType},
		pathCtx: a.pathCtx,
	}
	return lt.HoverAtPos(ctx, pos)
}
