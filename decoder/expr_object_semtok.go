// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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
		attrName, _, ok := rawObjectKey(item.KeyExpr)
		if !ok {
			// invalid expression
			continue
		}

		aSchema, ok := obj.cons.Attributes[attrName]
		if !ok {
			// skip unknown attribute
			continue
		}

		tokens = append(tokens, lang.SemanticToken{
			Type:      lang.TokenObjectKey,
			Modifiers: lang.SemanticTokenModifiers{},
			// TODO: Consider not reporting the quotes?
			Range: item.KeyExpr.Range(),
		})

		expr := newExpression(obj.pathCtx, item.ValueExpr, aSchema.Constraint)
		tokens = append(tokens, expr.SemanticTokens(ctx)...)
	}

	return tokens
}
