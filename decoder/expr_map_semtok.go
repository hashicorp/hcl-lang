// Copyright IBM Corp. 2020, 2025
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (m Map) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	eType, ok := m.expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return []lang.SemanticToken{}
	}

	if len(eType.Items) == 0 || m.cons.Elem == nil {
		return []lang.SemanticToken{}
	}

	tokens := make([]lang.SemanticToken, 0)

	for _, item := range eType.Items {
		_, _, isRawKey := rawObjectKey(item.KeyExpr)
		if isRawKey {
			tokens = append(tokens, lang.SemanticToken{
				Type:      lang.TokenMapKey,
				Modifiers: lang.SemanticTokenModifiers{},
				Range:     item.KeyExpr.Range(),
			})

			vExpr := newExpression(m.pathCtx, item.ValueExpr, m.cons.Elem)
			tokens = append(tokens, vExpr.SemanticTokens(ctx)...)
			continue
		}

		keyExpr, ok := item.KeyExpr.(*hclsyntax.ObjectConsKeyExpr)
		if ok && m.cons.AllowInterpolatedKeys {
			parensExpr, ok := keyExpr.Wrapped.(*hclsyntax.ParenthesesExpr)
			if ok {
				keyCons := schema.AnyExpression{
					OfType: cty.String,
				}
				kExpr := newExpression(m.pathCtx, parensExpr, keyCons)
				tokens = append(tokens, kExpr.SemanticTokens(ctx)...)

				vExpr := newExpression(m.pathCtx, item.ValueExpr, m.cons.Elem)
				tokens = append(tokens, vExpr.SemanticTokens(ctx)...)
			}
		}
	}

	return tokens
}
