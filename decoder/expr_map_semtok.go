// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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
		_, _, ok := rawObjectKey(item.KeyExpr)
		if !ok {
			continue
		}
		tokens = append(tokens, lang.SemanticToken{
			Type:      lang.TokenMapKey,
			Modifiers: lang.SemanticTokenModifiers{},
			Range:     item.KeyExpr.Range(),
		})

		expr := newExpression(m.pathCtx, item.ValueExpr, m.cons.Elem)
		tokens = append(tokens, expr.SemanticTokens(ctx)...)
	}

	return tokens
}
