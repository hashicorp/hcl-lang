// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (kw Keyword) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	if isEmptyExpression(kw.expr) {
		return []lang.Candidate{
			{
				Label:       kw.cons.Keyword,
				Detail:      kw.cons.FriendlyName(),
				Description: kw.cons.Description,
				Kind:        lang.KeywordCandidateKind,
				TextEdit: lang.TextEdit{
					NewText: kw.cons.Keyword,
					Snippet: kw.cons.Keyword,
					Range: hcl.Range{
						Filename: kw.expr.Range().Filename,
						Start:    pos,
						End:      pos,
					},
				},
			},
		}
	}

	eType, ok := kw.expr.(*hclsyntax.ScopeTraversalExpr)
	if !ok {
		return []lang.Candidate{}
	}

	if len(eType.Traversal) != 1 {
		return []lang.Candidate{}
	}

	prefixLen := pos.Byte - eType.Traversal.SourceRange().Start.Byte
	if prefixLen > len(eType.Traversal.RootName()) {
		// The user has probably typed an extra character, such as a
		// period, that is not (yet) part of the expression. This prefix
		// won't match anything, so we'll return early.
		return []lang.Candidate{}
	}
	prefix := eType.Traversal.RootName()[0:prefixLen]

	if strings.HasPrefix(kw.cons.Keyword, prefix) {
		return []lang.Candidate{
			{
				Label:       kw.cons.Keyword,
				Detail:      kw.cons.FriendlyName(),
				Description: kw.cons.Description,
				Kind:        lang.KeywordCandidateKind,
				TextEdit: lang.TextEdit{
					NewText: kw.cons.Keyword,
					Snippet: kw.cons.Keyword,
					Range:   eType.Range(),
				},
			},
		}
	}

	return []lang.Candidate{}
}
