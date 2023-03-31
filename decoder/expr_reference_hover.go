// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (ref Reference) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	eType, ok := ref.expr.(*hclsyntax.ScopeTraversalExpr)
	if !ok {
		return nil
	}

	origins, ok := ref.pathCtx.ReferenceOrigins.AtPos(eType.Range().Filename, pos)
	if !ok {
		return nil
	}

	for _, origin := range origins {
		matchableOrigin, ok := origin.(reference.MatchableOrigin)
		if !ok {
			continue
		}
		targets, ok := ref.pathCtx.ReferenceTargets.Match(matchableOrigin)
		if !ok {
			// target not found
			continue
		}

		// TODO: Reflect additional found targets here?

		content, err := hoverContentForReferenceTarget(ctx, targets[0], pos)
		if err == nil {
			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   eType.Range(),
			}
		}
	}

	return nil
}
