package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (set Set) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	eType, ok := set.expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return reference.Origins{}
	}

	if len(eType.Exprs) == 0 || set.cons.Elem == nil {
		return reference.Origins{}
	}

	origins := make(reference.Origins, 0)

	for _, elemExpr := range eType.Exprs {
		expr := newExpression(set.pathCtx, elemExpr, set.cons.Elem)
		if e, ok := expr.(ReferenceOriginsExpression); ok {
			origins = append(origins, e.ReferenceOrigins(ctx, allowSelfRefs)...)
		}
	}

	return origins
}
