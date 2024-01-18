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

func (obj Object) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	eType, ok := obj.expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return []lang.SemanticToken{}
	}

	if len(eType.Items) == 0 || len(obj.cons.Attributes) == 0 {
		return []lang.SemanticToken{}
	}

	tokens := make([]lang.SemanticToken, 0)

	for _, item := range eType.Items {
		attrName, _, isRawKey := rawObjectKey(item.KeyExpr)

		var aSchema *schema.AttributeSchema
		var isKnownAttr bool
		if isRawKey {
			aSchema, isKnownAttr = obj.cons.Attributes[attrName]
		}

		keyExpr, ok := item.KeyExpr.(*hclsyntax.ObjectConsKeyExpr)
		if ok && obj.cons.AllowInterpolatedAttrName {
			parensExpr, ok := keyExpr.Wrapped.(*hclsyntax.ParenthesesExpr)
			if ok {
				keyCons := schema.AnyExpression{
					OfType: cty.String,
				}
				kExpr := newExpression(obj.pathCtx, parensExpr, keyCons)
				tokens = append(tokens, kExpr.SemanticTokens(ctx)...)
			}
		}

		if isKnownAttr {
			tokens = append(tokens, lang.SemanticToken{
				Type:      lang.TokenObjectKey,
				Modifiers: lang.SemanticTokenModifiers{},
				// TODO: Consider not reporting the quotes?
				Range: item.KeyExpr.Range(),
			})
		}

		if isKnownAttr {
			expr := newExpression(obj.pathCtx, item.ValueExpr, aSchema.Constraint)
			tokens = append(tokens, expr.SemanticTokens(ctx)...)
			continue
		}
	}

	return tokens
}
