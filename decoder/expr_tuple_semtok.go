// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (tuple Tuple) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	eType, ok := tuple.expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return []lang.SemanticToken{}
	}

	if len(eType.Exprs) == 0 || len(tuple.cons.Elems) == 0 {
		return []lang.SemanticToken{}
	}

	tokens := make([]lang.SemanticToken, 0)

	for i, elemExpr := range eType.Exprs {
		if i+1 > len(tuple.cons.Elems) {
			break
		}

		expr := newExpression(tuple.pathCtx, elemExpr, tuple.cons.Elems[i])
		tokens = append(tokens, expr.SemanticTokens(ctx)...)
	}

	return tokens
}
