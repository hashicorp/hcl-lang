package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
)

func (a Any) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	expr := OneOf{
		pathCtx: a.pathCtx,
		expr:    a.expr,
		cons: schema.OneOf{
			schema.Reference{OfType: a.cons.OfType},
			schema.LiteralType{Type: a.cons.OfType},
		},
	}

	return expr.ReferenceTargets(ctx, targetCtx)
}
