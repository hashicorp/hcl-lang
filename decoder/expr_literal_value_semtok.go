// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (lv LiteralValue) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	typ := lv.cons.Value.Type()

	if typ == cty.DynamicPseudoType {
		val, diags := lv.expr.Value(nil)
		if !diags.HasErrors() {
			typ = val.Type()
		}
	}

	// string is a special case as it's always represented like a template
	// even if there's no templating involved
	if typ == cty.String {
		expr, ok := lv.expr.(*hclsyntax.TemplateExpr)
		if !ok {
			return []lang.SemanticToken{}
		}

		val, diags := expr.Value(nil)
		if diags.HasErrors() {
			return []lang.SemanticToken{}
		}
		if !lv.cons.Value.RawEquals(val) {
			return nil
		}

		if expr.IsStringLiteral() || isMultilineStringLiteral(expr) {
			return []lang.SemanticToken{
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range:     expr.Range(),
				},
			}
		}

		// TODO: consider reporting multiline/HEREDOC notation as a different token

		return []lang.SemanticToken{}
	}

	if typ.IsPrimitiveType() {
		expr, ok := lv.expr.(*hclsyntax.LiteralValueExpr)
		if !ok {
			return []lang.SemanticToken{}
		}

		val, diags := expr.Value(nil)
		if diags.HasErrors() {
			return []lang.SemanticToken{}
		}
		if !lv.cons.Value.RawEquals(val) {
			return nil
		}

		if typ == cty.Bool {
			return []lang.SemanticToken{
				{
					Type:      lang.TokenBool,
					Modifiers: lang.SemanticTokenModifiers{},
					Range:     expr.Range(),
				},
			}
		}

		if typ == cty.Number {
			return []lang.SemanticToken{
				{
					Type:      lang.TokenNumber,
					Modifiers: lang.SemanticTokenModifiers{},
					Range:     expr.Range(),
				},
			}
		}

		return []lang.SemanticToken{}
	}

	if typ.IsListType() {
		expr, ok := lv.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return []lang.SemanticToken{}
		}

		tokens := make([]lang.SemanticToken, 0)
		values := lv.cons.Value.AsValueSlice()
		for i, elemExpr := range expr.Exprs {
			if len(values) < i+1 {
				break
			}

			val, diags := elemExpr.Value(nil)
			if diags.HasErrors() {
				continue
			}

			elemCons := schema.LiteralValue{
				Value: values[i],
			}
			if !elemCons.Value.RawEquals(val) {
				continue
			}

			expr := newExpression(lv.pathCtx, elemExpr, elemCons)
			tokens = append(tokens, expr.SemanticTokens(ctx)...)
		}

		return tokens
	}

	if typ.IsSetType() {
		expr, ok := lv.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return []lang.SemanticToken{}
		}

		tokens := make([]lang.SemanticToken, 0)
		values := lv.cons.Value.AsValueSet()
		for i, elemExpr := range expr.Exprs {
			if values.Length() < i+1 {
				break
			}

			val, diags := elemExpr.Value(nil)
			if diags.HasErrors() {
				continue
			}

			if !values.ElementType().Equals(val.Type()) {
				continue
			}

			if !values.Has(val) {
				continue
			}

			elemCons := schema.LiteralValue{
				Value: val,
			}

			expr := newExpression(lv.pathCtx, elemExpr, elemCons)
			tokens = append(tokens, expr.SemanticTokens(ctx)...)
		}

		return tokens
	}

	if typ.IsTupleType() {
		expr, ok := lv.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return []lang.SemanticToken{}
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

		return newExpression(lv.pathCtx, expr, cons).SemanticTokens(ctx)
	}

	if typ.IsMapType() {
		expr, ok := lv.expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return []lang.SemanticToken{}
		}

		tokens := make([]lang.SemanticToken, 0)
		values := lv.cons.Value.AsValueMap()
		for _, item := range expr.Items {
			keyStr, _, ok := rawObjectKey(item.KeyExpr)
			if !ok {
				continue
			}

			if _, ok := values[keyStr]; !ok {
				continue
			}

			tokens = append(tokens, lang.SemanticToken{
				Type:      lang.TokenMapKey,
				Modifiers: lang.SemanticTokenModifiers{},
				Range:     item.KeyExpr.Range(),
			})

			val, diags := item.ValueExpr.Value(nil)
			if diags.HasErrors() {
				continue
			}

			if !values[keyStr].RawEquals(val) {
				continue
			}

			elemCons := schema.LiteralValue{
				Value: val,
			}
			expr := newExpression(lv.pathCtx, item.ValueExpr, elemCons)
			tokens = append(tokens, expr.SemanticTokens(ctx)...)
		}

		return tokens
	}

	if typ.IsObjectType() {
		expr, ok := lv.expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return []lang.SemanticToken{}
		}

		cons := schema.Object{
			Attributes: ctyObjectToObjectAttributes(typ),
		}
		return newExpression(lv.pathCtx, expr, cons).SemanticTokens(ctx)
	}

	return []lang.SemanticToken{}
}
