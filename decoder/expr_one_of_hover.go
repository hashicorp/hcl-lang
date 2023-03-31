// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
)

func (oo OneOf) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	for _, con := range oo.cons {
		// since we cannot know which "constraint was typed"
		// we just pick the first which returns some data
		expr := newExpression(oo.pathCtx, oo.expr, con)
		hoverData := expr.HoverAtPos(ctx, pos)
		if hoverData != nil {
			return hoverData
		}
	}

	return nil
}
