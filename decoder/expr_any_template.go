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
		if eType.Wrapped.Range().ContainsPos(pos) || eType.Wrapped.Range().End.Byte == pos.Byte {
			cons := schema.AnyExpression{
				OfType: cty.String,
			}
			return newExpression(a.pathCtx, eType.Wrapped, cons).CompletionAtPos(ctx, pos), true
		}

		// Trailing dot may be ignored by the parser so we attempt to recover it
		if pos.Byte-eType.Wrapped.Range().End.Byte == 1 {
			fileBytes := a.pathCtx.Files[eType.Wrapped.Range().Filename].Bytes
			trailingRune := fileBytes[eType.Wrapped.Range().End.Byte:pos.Byte][0]

			if trailingRune == '.' {
				cons := schema.AnyExpression{
					OfType: cty.String,
				}
				return newExpression(a.pathCtx, eType.Wrapped, cons).CompletionAtPos(ctx, pos), true
			}
		}

		return candidates, false
	case *hclsyntax.TemplateJoinExpr:
		// TODO: implement when support for expressions https://github.com/hashicorp/terraform-ls/issues/527 lands
	}

	return candidates, true
}

func (a Any) hoverTemplateExprAtPos(ctx context.Context, pos hcl.Pos) (*lang.HoverData, bool) {
	switch eType := a.expr.(type) {
	case *hclsyntax.TemplateExpr:
		if eType.IsStringLiteral() {
			return nil, false
		}

		for _, partExpr := range eType.Parts {
			if partExpr.Range().ContainsPos(pos) {
				cons := schema.AnyExpression{
					OfType: cty.String,
				}
				return newExpression(a.pathCtx, partExpr, cons).HoverAtPos(ctx, pos), true
			}
		}

		return nil, true
	case *hclsyntax.TemplateWrapExpr:
		if eType.Wrapped.Range().ContainsPos(pos) {
			cons := schema.AnyExpression{
				OfType: cty.String,
			}
			return newExpression(a.pathCtx, eType.Wrapped, cons).HoverAtPos(ctx, pos), true
		}

		return nil, true
	}

	return nil, false
}

func (a Any) refOriginsForTemplateExpr(ctx context.Context) (reference.Origins, bool) {
	origins := make(reference.Origins, 0)

	switch eType := a.expr.(type) {
	case *hclsyntax.TemplateExpr:
		if eType.IsStringLiteral() {
			return nil, false
		}

		for _, partExpr := range eType.Parts {
			cons := schema.AnyExpression{
				OfType: cty.String,
			}
			expr := newExpression(a.pathCtx, partExpr, cons)

			if e, ok := expr.(ReferenceOriginsExpression); ok {
				origins = append(origins, e.ReferenceOrigins(ctx)...)
			}
		}

		return origins, true
	case *hclsyntax.TemplateWrapExpr:
		cons := schema.AnyExpression{
			OfType: cty.String,
		}
		expr := newExpression(a.pathCtx, eType.Wrapped, cons)

		if e, ok := expr.(ReferenceOriginsExpression); ok {
			origins = append(origins, e.ReferenceOrigins(ctx)...)
		}

		return origins, true
	}

	return origins, false
}

func (a Any) semanticTokensForTemplateExpr(ctx context.Context) ([]lang.SemanticToken, bool) {
	tokens := make([]lang.SemanticToken, 0)

	switch eType := a.expr.(type) {
	case *hclsyntax.TemplateExpr:
		if eType.IsStringLiteral() {
			cons := schema.LiteralType{
				Type: cty.String,
			}
			expr := newExpression(a.pathCtx, eType, cons)
			tokens = append(tokens, expr.SemanticTokens(ctx)...)
			return tokens, true
		}

		for _, partExpr := range eType.Parts {
			cons := schema.AnyExpression{
				OfType: cty.String,
			}
			expr := newExpression(a.pathCtx, partExpr, cons)
			tokens = append(tokens, expr.SemanticTokens(ctx)...)
		}

		return tokens, true
	case *hclsyntax.TemplateWrapExpr:
		cons := schema.AnyExpression{
			OfType: cty.String,
		}
		expr := newExpression(a.pathCtx, eType.Wrapped, cons)
		tokens = append(tokens, expr.SemanticTokens(ctx)...)

		return tokens, true
	}

	return tokens, false
}
