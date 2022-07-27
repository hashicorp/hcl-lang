package decoder

import (
	"context"
)

func (d *Decoder) ResolveCandidate(ctx context.Context, unresolvedCandidate UnresolvedCandidate) (*ResolvedCandidate, error) {
	if resolveFunc, ok := d.ctx.CompletionResolveHooks[unresolvedCandidate.ResolveHook.Name]; ok {
		return resolveFunc(ctx, unresolvedCandidate)
	}

	return nil, nil
}
