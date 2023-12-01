package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
)

func (d *Decoder) CodeActionsForRange(ctx context.Context, path lang.Path, rng hcl.Range) []lang.CodeAction {
	actions := make([]lang.CodeAction, 0)

	pathCtx, err := d.pathReader.PathContext(path)
	if err == nil {
		ctx = withPathContext(ctx, pathCtx)
	}

	for _, ca := range d.ctx.CodeActions {
		actions = append(actions, ca.CodeActions(ctx, path, rng)...)
	}

	return actions
}
