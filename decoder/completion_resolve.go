package decoder

import (
	"context"
	"errors"
)

// ResolveCandidate tries to gather more information for a candidate,
// by checking for a resolve hook and executing it.
func (d *Decoder) ResolveCandidate(ctx context.Context, unresolvedCandidate UnresolvedCandidate) (*ResolvedCandidate, error) {
	if unresolvedCandidate.ResolveHook == nil {
		return nil, errors.New("missing resolve hook")
	}

	if resolveFunc, ok := d.ctx.CompletionResolveHooks[unresolvedCandidate.ResolveHook.Name]; ok {
		return resolveFunc(ctx, unresolvedCandidate)
	}

	return nil, nil
}
