package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (list List) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	eType, ok := list.expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return []lang.SemanticToken{}
	}

	if len(eType.Exprs) == 0 || list.cons.Elem == nil {
		return []lang.SemanticToken{}
	}

	tokens := make([]lang.SemanticToken, 0)

	for _, elemExpr := range eType.Exprs {
		expr := newExpression(list.pathCtx, elemExpr, list.cons.Elem)
		tokens = append(tokens, expr.SemanticTokens(ctx)...)
	}

	return tokens
}
