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

func (a Any) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	typ := a.cons.OfType

	if !a.cons.SkipLiteralComplexTypes && typ.IsListType() {
		expr, ok := a.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return a.completeNonComplexExprAtPos(ctx, pos)
		}

		cons := schema.List{
			Elem: schema.AnyExpression{
				OfType: typ.ElementType(),
			},
		}

		return newExpression(a.pathCtx, expr, cons).CompletionAtPos(ctx, pos)
	}

	if !a.cons.SkipLiteralComplexTypes && typ.IsSetType() {
		expr, ok := a.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return a.completeNonComplexExprAtPos(ctx, pos)
		}

		cons := schema.Set{
			Elem: schema.AnyExpression{
				OfType: typ.ElementType(),
			},
		}

		return newExpression(a.pathCtx, expr, cons).CompletionAtPos(ctx, pos)
	}

	if !a.cons.SkipLiteralComplexTypes && typ.IsTupleType() {
		expr, ok := a.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return a.completeNonComplexExprAtPos(ctx, pos)
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

		return newExpression(a.pathCtx, expr, cons).CompletionAtPos(ctx, pos)
	}

	if !a.cons.SkipLiteralComplexTypes && typ.IsMapType() {
		expr, ok := a.expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return a.completeNonComplexExprAtPos(ctx, pos)
		}

		cons := schema.Map{
			Elem: schema.AnyExpression{
				OfType: typ.ElementType(),
			},
		}
		return newExpression(a.pathCtx, expr, cons).CompletionAtPos(ctx, pos)
	}

	if !a.cons.SkipLiteralComplexTypes && typ.IsObjectType() {
		expr, ok := a.expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return a.completeNonComplexExprAtPos(ctx, pos)
		}

		cons := schema.Object{
			Attributes: ctyObjectToObjectAttributes(typ),
		}
		return newExpression(a.pathCtx, expr, cons).CompletionAtPos(ctx, pos)
	}

	return a.completeNonComplexExprAtPos(ctx, pos)
}

func (a Any) completeNonComplexExprAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	candidates := make([]lang.Candidate, 0)

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
	candidates = append(candidates, ref.CompletionAtPos(ctx, pos)...)

	fe := functionExpr{
		expr:       a.expr,
		returnType: a.cons.OfType,
		pathCtx:    a.pathCtx,
	}
	candidates = append(candidates, fe.CompletionAtPos(ctx, pos)...)

	lt := LiteralType{
		expr: a.expr,
		cons: schema.LiteralType{
			Type:             a.cons.OfType,
			SkipComplexTypes: a.cons.SkipLiteralComplexTypes,
		},
		pathCtx: a.pathCtx,
	}
	candidates = append(candidates, lt.CompletionAtPos(ctx, pos)...)

	return candidates
}
