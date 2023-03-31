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

func (lt LiteralType) SemanticTokens(ctx context.Context) []lang.SemanticToken {
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
			return []lang.SemanticToken{}
		}
		if expr.IsStringLiteral() {
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
		expr, ok := lt.expr.(*hclsyntax.LiteralValueExpr)
		if !ok {
			return []lang.SemanticToken{}
		}

		if !lt.cons.Type.Equals(expr.Val.Type()) {
			return []lang.SemanticToken{}
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
		expr, ok := lt.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return []lang.SemanticToken{}
		}

		cons := schema.List{
			Elem: schema.LiteralType{
				Type: typ.ElementType(),
			},
		}

		return newExpression(lt.pathCtx, expr, cons).SemanticTokens(ctx)
	}

	if typ.IsSetType() {
		expr, ok := lt.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return []lang.SemanticToken{}
		}

		cons := schema.Set{
			Elem: schema.LiteralType{
				Type: typ.ElementType(),
			},
		}

		return newExpression(lt.pathCtx, expr, cons).SemanticTokens(ctx)
	}

	if typ.IsTupleType() {
		expr, ok := lt.expr.(*hclsyntax.TupleConsExpr)
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

		return newExpression(lt.pathCtx, expr, cons).SemanticTokens(ctx)
	}

	if typ.IsMapType() {
		expr, ok := lt.expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return []lang.SemanticToken{}
		}

		cons := schema.Map{
			Elem: schema.LiteralType{
				Type: typ.ElementType(),
			},
		}
		return newExpression(lt.pathCtx, expr, cons).SemanticTokens(ctx)
	}

	if typ.IsObjectType() {
		expr, ok := lt.expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return []lang.SemanticToken{}
		}

		cons := schema.Object{
			Attributes: ctyObjectToObjectAttributes(typ),
		}
		return newExpression(lt.pathCtx, expr, cons).SemanticTokens(ctx)
	}

	return []lang.SemanticToken{}
}
