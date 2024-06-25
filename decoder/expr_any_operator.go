// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

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

	case *hclsyntax.ParenthesesExpr:
		if eType.Expression.Range().ContainsPos(pos) || eType.Expression.Range().End.Byte == pos.Byte {
			return newExpression(a.pathCtx, eType.Expression, a.cons).CompletionAtPos(ctx, pos), true
		}
	}

	return candidates, true
}

func (a Any) hoverOperatorExprAtPos(ctx context.Context, pos hcl.Pos) (*lang.HoverData, bool) {
	switch eType := a.expr.(type) {
	case *hclsyntax.BinaryOpExpr:
		opReturnType := eType.Op.Type

		// Check if such an operation is even allowed within the constraint
		if _, err := convert.Convert(cty.UnknownVal(opReturnType), a.cons.OfType); err != nil {
			return nil, true
		}

		opFuncParams := eType.Op.Impl.Params()
		if len(opFuncParams) != 2 {
			// This should never happen if HCL implementation is correct
			return nil, true
		}

		if eType.LHS.Range().ContainsPos(pos) {
			cons := schema.AnyExpression{
				OfType: opFuncParams[0].Type,
			}
			return newExpression(a.pathCtx, eType.LHS, cons).HoverAtPos(ctx, pos), true
		}
		if eType.RHS.Range().ContainsPos(pos) {
			cons := schema.AnyExpression{
				OfType: opFuncParams[1].Type,
			}
			return newExpression(a.pathCtx, eType.RHS, cons).HoverAtPos(ctx, pos), true
		}

		return nil, true

	case *hclsyntax.UnaryOpExpr:
		opReturnType := eType.Op.Type

		// Check if such an operation is even allowed within the constraint
		if _, err := convert.Convert(cty.UnknownVal(opReturnType), a.cons.OfType); err != nil {
			return nil, true
		}

		opFuncParams := eType.Op.Impl.Params()
		if len(opFuncParams) != 1 {
			// This should never happen if HCL implementation is correct
			return nil, true
		}

		if eType.Val.Range().ContainsPos(pos) {
			cons := schema.AnyExpression{
				OfType: opFuncParams[0].Type,
			}
			return newExpression(a.pathCtx, eType.Val, cons).HoverAtPos(ctx, pos), true
		}

		return nil, true
	case *hclsyntax.ParenthesesExpr:
		if eType.Expression.Range().ContainsPos(pos) {
			return newExpression(a.pathCtx, eType.Expression, a.cons).HoverAtPos(ctx, pos), true
		}
	}

	return nil, false
}

func (a Any) refOriginsForOperatorExpr(ctx context.Context) (reference.Origins, bool) {
	origins := make(reference.Origins, 0)

	// There is currently no way of decoding operator expressions in JSON
	// so we just collect them using the fallback logic assuming "any"
	// constraint and focus on collecting expressions in HCL with more
	// accurate constraints below.

	switch eType := a.expr.(type) {
	case *hclsyntax.BinaryOpExpr:
		opReturnType := eType.Op.Type

		// Check if such an operation is even allowed within the constraint
		if _, err := convert.Convert(cty.UnknownVal(opReturnType), a.cons.OfType); err != nil {
			return origins, true
		}

		opFuncParams := eType.Op.Impl.Params()
		if len(opFuncParams) != 2 {
			// This should never happen if HCL implementation is correct
			return origins, true
		}

		leftExpr := newExpression(a.pathCtx, eType.LHS, schema.AnyExpression{
			OfType: opFuncParams[0].Type,
		})
		if expr, ok := leftExpr.(ReferenceOriginsExpression); ok {
			origins = append(origins, expr.ReferenceOrigins(ctx)...)
		}

		rightExpr := newExpression(a.pathCtx, eType.RHS, schema.AnyExpression{
			OfType: opFuncParams[1].Type,
		})
		if expr, ok := rightExpr.(ReferenceOriginsExpression); ok {
			origins = append(origins, expr.ReferenceOrigins(ctx)...)
		}

		return origins, true

	case *hclsyntax.UnaryOpExpr:
		opReturnType := eType.Op.Type

		// Check if such an operation is even allowed within the constraint
		if _, err := convert.Convert(cty.UnknownVal(opReturnType), a.cons.OfType); err != nil {
			return origins, true
		}

		opFuncParams := eType.Op.Impl.Params()
		if len(opFuncParams) != 1 {
			// This should never happen if HCL implementation is correct
			return origins, true
		}

		expr := newExpression(a.pathCtx, eType.Val, schema.AnyExpression{
			OfType: opFuncParams[0].Type,
		})
		if expr, ok := expr.(ReferenceOriginsExpression); ok {
			origins = append(origins, expr.ReferenceOrigins(ctx)...)
		}

		return origins, true
	case *hclsyntax.ParenthesesExpr:
		expr := newExpression(a.pathCtx, eType.Expression, a.cons)
		if expr, ok := expr.(ReferenceOriginsExpression); ok {
			origins = append(origins, expr.ReferenceOrigins(ctx)...)
		}

		return origins, true
	}

	return origins, false
}

func (a Any) semanticTokensForOperatorExpr(ctx context.Context) ([]lang.SemanticToken, bool) {
	tokens := make([]lang.SemanticToken, 0)

	switch eType := a.expr.(type) {
	case *hclsyntax.BinaryOpExpr:
		opReturnType := eType.Op.Type

		// Check if such an operation is even allowed within the constraint
		if _, err := convert.Convert(cty.UnknownVal(opReturnType), a.cons.OfType); err != nil {
			return tokens, true
		}

		opFuncParams := eType.Op.Impl.Params()
		if len(opFuncParams) != 2 {
			// This should never happen if HCL implementation is correct
			return tokens, true
		}

		tokens = append(tokens, newExpression(a.pathCtx, eType.LHS, schema.AnyExpression{
			OfType: opFuncParams[0].Type,
		}).SemanticTokens(ctx)...)

		tokens = append(tokens, newExpression(a.pathCtx, eType.RHS, schema.AnyExpression{
			OfType: opFuncParams[1].Type,
		}).SemanticTokens(ctx)...)

		return tokens, true

	case *hclsyntax.UnaryOpExpr:
		opReturnType := eType.Op.Type

		// Check if such an operation is even allowed within the constraint
		if _, err := convert.Convert(cty.UnknownVal(opReturnType), a.cons.OfType); err != nil {
			return tokens, true
		}

		opFuncParams := eType.Op.Impl.Params()
		if len(opFuncParams) != 1 {
			// This should never happen if HCL implementation is correct
			return tokens, true
		}

		tokens = append(tokens, newExpression(a.pathCtx, eType.Val, schema.AnyExpression{
			OfType: opFuncParams[0].Type,
		}).SemanticTokens(ctx)...)

		return tokens, true

	case *hclsyntax.ParenthesesExpr:
		tokens = append(tokens, newExpression(a.pathCtx, eType.Expression, a.cons).SemanticTokens(ctx)...)
		return tokens, true
	}

	return tokens, false
}
