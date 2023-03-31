// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
)

func (oo OneOf) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	candidates := make([]lang.Candidate, 0)

	for _, con := range oo.cons {
		expr := newExpression(oo.pathCtx, oo.expr, con)
		candidates = append(candidates, expr.CompletionAtPos(ctx, pos)...)
	}

	return candidates
}
