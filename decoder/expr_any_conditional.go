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
)

func (a Any) completeConditionalExprAtPos(ctx context.Context, pos hcl.Pos) ([]lang.Candidate, bool) {
	candidates := make([]lang.Candidate, 0)

	switch eType := a.expr.(type) {
	case *hclsyntax.ConditionalExpr:
		if eType.Condition.Range().ContainsPos(pos) || eType.Condition.Range().End.Byte == pos.Byte {
			cons := schema.AnyExpression{
				OfType: cty.Bool,
			}
			return newExpression(a.pathCtx, eType.Condition, cons).CompletionAtPos(ctx, pos), true
		}
		if eType.TrueResult.Range().ContainsPos(pos) || eType.TrueResult.Range().End.Byte == pos.Byte {
			cons := schema.AnyExpression{
				OfType: cty.DynamicPseudoType,
			}
			return newExpression(a.pathCtx, eType.TrueResult, cons).CompletionAtPos(ctx, pos), true
		}
		if eType.FalseResult.Range().ContainsPos(pos) || eType.FalseResult.Range().End.Byte == pos.Byte {
			cons := schema.AnyExpression{
				OfType: cty.DynamicPseudoType,
			}
			return newExpression(a.pathCtx, eType.FalseResult, cons).CompletionAtPos(ctx, pos), true
		}

		return candidates, false
	}

	return candidates, true
}

func (a Any) hoverConditionalExprAtPos(ctx context.Context, pos hcl.Pos) (*lang.HoverData, bool) {
	switch eType := a.expr.(type) {
	case *hclsyntax.ConditionalExpr:
		if eType.Condition.Range().ContainsPos(pos) {
			cons := schema.AnyExpression{
				OfType: cty.Bool,
			}
			return newExpression(a.pathCtx, eType.Condition, cons).HoverAtPos(ctx, pos), true
		}
		if eType.TrueResult.Range().ContainsPos(pos) {
			cons := schema.AnyExpression{
				OfType: cty.DynamicPseudoType,
			}
			return newExpression(a.pathCtx, eType.TrueResult, cons).HoverAtPos(ctx, pos), true
		}
		if eType.FalseResult.Range().ContainsPos(pos) {
			cons := schema.AnyExpression{
				OfType: cty.DynamicPseudoType,
			}
			return newExpression(a.pathCtx, eType.FalseResult, cons).HoverAtPos(ctx, pos), true
		}
	}

	return nil, false
}

func (a Any) refOriginsForConditionalExpr(ctx context.Context) (reference.Origins, bool) {
	origins := make(reference.Origins, 0)

	// There is currently no way of decoding conditional expressions in JSON
	// so we just collect them using the fallback logic assuming "any"
	// constraint and focus on collecting expressions in HCL with more
	// accurate constraints below.

	switch eType := a.expr.(type) {
	case *hclsyntax.ConditionalExpr:
		condExpr := newExpression(a.pathCtx, eType.Condition, schema.AnyExpression{
			OfType: cty.Bool,
		})
		if expr, ok := condExpr.(ReferenceOriginsExpression); ok {
			origins = append(origins, expr.ReferenceOrigins(ctx)...)
		}

		trueExpr := newExpression(a.pathCtx, eType.TrueResult, schema.AnyExpression{
			OfType:                  a.cons.OfType,
			SkipLiteralComplexTypes: a.cons.SkipLiteralComplexTypes,
		})
		if expr, ok := trueExpr.(ReferenceOriginsExpression); ok {
			origins = append(origins, expr.ReferenceOrigins(ctx)...)
		}

		falseExpr := newExpression(a.pathCtx, eType.FalseResult, schema.AnyExpression{
			OfType:                  a.cons.OfType,
			SkipLiteralComplexTypes: a.cons.SkipLiteralComplexTypes,
		})
		if expr, ok := falseExpr.(ReferenceOriginsExpression); ok {
			origins = append(origins, expr.ReferenceOrigins(ctx)...)
		}

		return origins, true
	}

	return origins, false
}

func (a Any) semanticTokensForConditionalExpr(ctx context.Context) ([]lang.SemanticToken, bool) {
	tokens := make([]lang.SemanticToken, 0)

	switch eType := a.expr.(type) {
	case *hclsyntax.ConditionalExpr:
		tokens = append(tokens, newExpression(a.pathCtx, eType.Condition, schema.AnyExpression{
			OfType: cty.Bool,
		}).SemanticTokens(ctx)...)
		tokens = append(tokens, newExpression(a.pathCtx, eType.TrueResult, schema.AnyExpression{
			OfType:                  a.cons.OfType,
			SkipLiteralComplexTypes: a.cons.SkipLiteralComplexTypes,
		}).SemanticTokens(ctx)...)
		tokens = append(tokens, newExpression(a.pathCtx, eType.FalseResult, schema.AnyExpression{
			OfType:                  a.cons.OfType,
			SkipLiteralComplexTypes: a.cons.SkipLiteralComplexTypes,
		}).SemanticTokens(ctx)...)

		return tokens, true
	}

	return tokens, false
}
