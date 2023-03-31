// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
)

func (oo OneOf) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	for _, con := range oo.cons {
		expr := newExpression(oo.pathCtx, oo.expr, con)
		e, ok := expr.(ReferenceTargetsExpression)
		if !ok {
			continue
		}
		targets := e.ReferenceTargets(ctx, targetCtx)
		if len(targets) > 0 {
			return targets
		}
	}

	return reference.Targets{}
}
