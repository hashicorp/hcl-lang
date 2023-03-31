// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (set Set) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	eType, ok := set.expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return []lang.SemanticToken{}
	}

	if len(eType.Exprs) == 0 || set.cons.Elem == nil {
		return []lang.SemanticToken{}
	}

	tokens := make([]lang.SemanticToken, 0)

	for _, elemExpr := range eType.Exprs {
		expr := newExpression(set.pathCtx, elemExpr, set.cons.Elem)
		tokens = append(tokens, expr.SemanticTokens(ctx)...)
	}

	return tokens
}
