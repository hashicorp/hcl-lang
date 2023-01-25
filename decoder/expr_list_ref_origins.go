package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (list List) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	eType, ok := list.expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return reference.Origins{}
	}

	if len(eType.Exprs) == 0 || list.cons.Elem == nil {
		return reference.Origins{}
	}

	origins := make(reference.Origins, 0)

	for _, elemExpr := range eType.Exprs {
		expr := newExpression(list.pathCtx, elemExpr, list.cons.Elem)
		if e, ok := expr.(ReferenceOriginsExpression); ok {
			origins = append(origins, e.ReferenceOrigins(ctx, allowSelfRefs)...)
		}
	}

	return origins
}
