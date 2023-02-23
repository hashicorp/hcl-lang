package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
)

func (oo OneOf) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	for _, con := range oo.cons {
		expr := newExpression(oo.pathCtx, oo.expr, con)
		e, ok := expr.(ReferenceOriginsExpression)
		if !ok {
			continue
		}
		origins := e.ReferenceOrigins(ctx, allowSelfRefs)
		if len(origins) > 0 {
			return origins
		}
	}

	return reference.Origins{}
}
