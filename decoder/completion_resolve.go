// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
)

// ResolveCandidate gathers more information for a completion candidate
// by checking for a resolve hook and executing it.
// This would be called as part of `completionItem/resolve` LSP method.
func (d *Decoder) ResolveCandidate(ctx context.Context, unresolvedCandidate UnresolvedCandidate) (*ResolvedCandidate, error) {
	if unresolvedCandidate.ResolveHook == nil {
		return nil, nil
	}

	if resolveFunc, ok := d.ctx.CompletionResolveHooks[unresolvedCandidate.ResolveHook.Name]; ok {
		return resolveFunc(ctx, unresolvedCandidate)
	}

	return nil, nil
}
