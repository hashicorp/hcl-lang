package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (tuple Tuple) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	eType, ok := tuple.expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return reference.Origins{}
	}

	if len(eType.Exprs) == 0 || len(tuple.cons.Elems) == 0 {
		return reference.Origins{}
	}

	origins := make(reference.Origins, 0)

	for i, elemExpr := range eType.Exprs {
		if i+1 > len(tuple.cons.Elems) {
			break
		}

		expr := newExpression(tuple.pathCtx, elemExpr, tuple.cons.Elems[i])
		if e, ok := expr.(ReferenceOriginsExpression); ok {
			origins = append(origins, e.ReferenceOrigins(ctx, allowSelfRefs)...)
		}
	}

	return origins
}
