// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
)

func (oo OneOf) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	for _, con := range oo.cons {
		// since we cannot know which "constraint was typed"
		// we just pick the first which returns some tokens
		expr := newExpression(oo.pathCtx, oo.expr, con)
		tokens := expr.SemanticTokens(ctx)
		if len(tokens) > 0 {
			return tokens
		}
	}

	return []lang.SemanticToken{}
}
