package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
)

func (oo OneOf) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	origins := make(reference.Origins, 0)

	for _, con := range oo.cons {
		expr := newExpression(oo.pathCtx, oo.expr, con)
		if e, ok := expr.(ReferenceOriginsExpression); ok {
			origins = append(origins, e.ReferenceOrigins(ctx, allowSelfRefs)...)
		}
	}

	return origins
}
