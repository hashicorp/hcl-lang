package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
)

func (oo OneOf) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	origins := make(reference.Targets, 0)

	for _, con := range oo.cons {
		expr := newExpression(oo.pathCtx, oo.expr, con)
		// TODO should we only consider the first non-empty constraint, rather than collect all?
		if e, ok := expr.(ReferenceTargetsExpression); ok {
			origins = append(origins, e.ReferenceTargets(ctx, targetCtx)...)
		}
	}

	return origins
}
