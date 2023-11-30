// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
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

	// TODO: Support splat expression https://github.com/hashicorp/terraform-ls/issues/526
	// TODO: Support for-in-if expression https://github.com/hashicorp/terraform-ls/issues/527
	// TODO: Support complex index expressions https://github.com/hashicorp/terraform-ls/issues/531
	// TODO: Support relative traversals https://github.com/hashicorp/terraform-ls/issues/532

	opCandidates, ok := a.completeOperatorExprAtPos(ctx, pos)
	if !ok {
		return candidates
	}
	candidates = append(candidates, opCandidates...)

	templateCandidates, ok := a.completeTemplateExprAtPos(ctx, pos)
	if !ok {
		return candidates
	}
	candidates = append(candidates, templateCandidates...)

	condCandidates, ok := a.completeConditionalExprAtPos(ctx, pos)
	if !ok {
		return candidates
	}
	candidates = append(candidates, condCandidates...)

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

func (a Any) completeOperatorExprAtPos(ctx context.Context, pos hcl.Pos) ([]lang.Candidate, bool) {
	candidates := make([]lang.Candidate, 0)

	switch eType := a.expr.(type) {
	case *hclsyntax.BinaryOpExpr:
		opReturnType := eType.Op.Type

		// Check if such an operation is even allowed within the constraint
		if _, err := convert.Convert(cty.UnknownVal(opReturnType), a.cons.OfType); err != nil {
			// This could illustrate a situation such as `list_attr = 42 +`
			// which is invalid syntax as add (+) op will never produce a list
			return candidates, false
		}

		opFuncParams := eType.Op.Impl.Params()
		if len(opFuncParams) != 2 {
			// This should never happen if HCL implementation is correct
			return candidates, false
		}

		if eType.LHS.Range().ContainsPos(pos) {
			cons := schema.AnyExpression{
				OfType: opFuncParams[0].Type,
			}
			return newExpression(a.pathCtx, eType.LHS, cons).CompletionAtPos(ctx, pos), true
		}
		if eType.RHS.Range().ContainsPos(pos) || eType.RHS.Range().End.Byte == pos.Byte {
			cons := schema.AnyExpression{
				OfType: opFuncParams[1].Type,
			}
			return newExpression(a.pathCtx, eType.RHS, cons).CompletionAtPos(ctx, pos), true
		}

		return candidates, false

	case *hclsyntax.UnaryOpExpr:
		opReturnType := eType.Op.Type

		// Check if such an operation is even allowed within the constraint
		if _, err := convert.Convert(cty.UnknownVal(opReturnType), a.cons.OfType); err != nil {
			// This could illustrate a situation such as `list_attr = !`
			// which is invalid syntax as negation (!) op will never produce a list
			return candidates, false
		}

		opFuncParams := eType.Op.Impl.Params()
		if len(opFuncParams) != 1 {
			// This should never happen if HCL implementation is correct
			return candidates, false
		}

		if eType.Val.Range().ContainsPos(pos) || eType.Val.Range().End.Byte == pos.Byte {
			cons := schema.AnyExpression{
				OfType: opFuncParams[0].Type,
			}
			return newExpression(a.pathCtx, eType.Val, cons).CompletionAtPos(ctx, pos), true
		}

		// Trailing dot may be ignored by the parser so we attempt to recover it
		if pos.Byte-eType.Val.Range().End.Byte == 1 {
			fileBytes := a.pathCtx.Files[eType.Range().Filename].Bytes
			trailingRune := fileBytes[eType.Val.Range().End.Byte:pos.Byte][0]

			if trailingRune == '.' {
				cons := schema.AnyExpression{
					OfType: opFuncParams[0].Type,
				}
				return newExpression(a.pathCtx, eType.Val, cons).CompletionAtPos(ctx, pos), true
			}
		}

		return candidates, false
	}

	return candidates, true
}

func (a Any) completeTemplateExprAtPos(ctx context.Context, pos hcl.Pos) ([]lang.Candidate, bool) {
	candidates := make([]lang.Candidate, 0)

	switch eType := a.expr.(type) {
	case *hclsyntax.TemplateExpr:
		if eType.IsStringLiteral() {
			return candidates, false
		}

		for _, partExpr := range eType.Parts {
			// We overshot the position and stop
			if partExpr.Range().Start.Byte > pos.Byte {
				break
			}

			// We're not checking the end byte position here, because we don't
			// allow completion after the }
			if partExpr.Range().ContainsPos(pos) || partExpr.Range().End.Byte == pos.Byte {
				cons := schema.AnyExpression{
					OfType: cty.String,
				}
				return newExpression(a.pathCtx, partExpr, cons).CompletionAtPos(ctx, pos), true
			}

			// Trailing dot may be ignored by the parser so we attempt to recover it
			if pos.Byte-partExpr.Range().End.Byte == 1 {
				fileBytes := a.pathCtx.Files[partExpr.Range().Filename].Bytes
				trailingRune := fileBytes[partExpr.Range().End.Byte:pos.Byte][0]

				if trailingRune == '.' {
					cons := schema.AnyExpression{
						OfType: cty.String,
					}
					return newExpression(a.pathCtx, partExpr, cons).CompletionAtPos(ctx, pos), true
				}
			}
		}

		return candidates, false
	case *hclsyntax.TemplateWrapExpr:
		if eType.Wrapped.Range().ContainsPos(pos) {
			cons := schema.AnyExpression{
				OfType: cty.String,
			}
			return newExpression(a.pathCtx, eType.Wrapped, cons).CompletionAtPos(ctx, pos), true
		}

		return candidates, false
	case *hclsyntax.TemplateJoinExpr:
		// TODO: implement when support for expressions https://github.com/hashicorp/terraform-ls/issues/527 lands
	}

	return candidates, true
}
