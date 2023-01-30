package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (m Map) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	eType, ok := m.expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return reference.Origins{}
	}

	if len(eType.Items) == 0 || m.cons.Elem == nil {
		return reference.Origins{}
	}

	origins := make(reference.Origins, 0)

	for _, item := range eType.Items {
		expr := newExpression(m.pathCtx, item.ValueExpr, m.cons.Elem)

		if elemExpr, ok := expr.(ReferenceOriginsExpression); ok {
			origins = append(origins, elemExpr.ReferenceOrigins(ctx, allowSelfRefs)...)
		}
	}

	return origins
}
