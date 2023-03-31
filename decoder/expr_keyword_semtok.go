// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (kw Keyword) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	eType, ok := kw.expr.(*hclsyntax.ScopeTraversalExpr)
	if !ok {
		return []lang.SemanticToken{}
	}

	if len(eType.Traversal) != 1 {
		return []lang.SemanticToken{}
	}

	if eType.Traversal.RootName() == kw.cons.Keyword {
		return []lang.SemanticToken{
			{
				Type:      lang.TokenKeyword,
				Modifiers: []lang.SemanticTokenModifier{},
				Range:     eType.Range(),
			},
		}
	}

	return []lang.SemanticToken{}
}
