// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (ref Reference) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	eType, ok := ref.expr.(*hclsyntax.ScopeTraversalExpr)
	if !ok {
		return []lang.SemanticToken{}
	}

	pos := ref.expr.Range().Start
	origins, ok := ref.pathCtx.ReferenceOrigins.AtPos(eType.Range().Filename, pos)
	if !ok {
		return []lang.SemanticToken{}
	}

	for _, origin := range origins {
		matchableOrigin, ok := origin.(reference.MatchableOrigin)
		if !ok {
			continue
		}
		_, ok = ref.pathCtx.ReferenceTargets.Match(matchableOrigin)
		if !ok {
			// target not found
			continue
		}

		return semanticTokensForTraversal(eType.Traversal)
	}

	return []lang.SemanticToken{}
}

func semanticTokensForTraversal(traversal hcl.Traversal) []lang.SemanticToken {
	tokens := make([]lang.SemanticToken, 0)

	for _, t := range traversal {
		// TODO: Add meaning to each step/token?
		// This would require declaring the meaning in schema.AddrStep
		// and exposing it via lang.AddressStep
		// See https://github.com/hashicorp/vscode-terraform/issues/574

		switch ts := t.(type) {
		case hcl.TraverseRoot:
			tokens = append(tokens, lang.SemanticToken{
				Type:      lang.TokenTraversalStep,
				Modifiers: []lang.SemanticTokenModifier{},
				Range:     t.SourceRange(),
			})
		case hcl.TraverseAttr:
			rng := t.SourceRange()
			tokens = append(tokens, lang.SemanticToken{
				Type:      lang.TokenTraversalStep,
				Modifiers: []lang.SemanticTokenModifier{},
				Range: hcl.Range{
					Filename: rng.Filename,
					// omit the initial '.'
					Start: hcl.Pos{
						Line:   rng.Start.Line,
						Column: rng.Start.Column + 1,
						Byte:   rng.Start.Byte + 1,
					},
					End: rng.End,
				},
			})
		case hcl.TraverseIndex:
			// for index steps we only report
			// what's inside brackets
			rng := t.SourceRange()
			idxRange := hcl.Range{
				Filename: rng.Filename,
				Start: hcl.Pos{
					Line:   rng.Start.Line,
					Column: rng.Start.Column + 1,
					Byte:   rng.Start.Byte + 1,
				},
				End: hcl.Pos{
					Line:   rng.End.Line,
					Column: rng.End.Column - 1,
					Byte:   rng.End.Byte - 1,
				},
			}

			if ts.Key.Type() == cty.String {
				tokens = append(tokens, lang.SemanticToken{
					Type:      lang.TokenMapKey,
					Modifiers: []lang.SemanticTokenModifier{},
					Range:     idxRange,
				})
			}
			if ts.Key.Type() == cty.Number {
				tokens = append(tokens, lang.SemanticToken{
					Type:      lang.TokenNumber,
					Modifiers: []lang.SemanticTokenModifier{},
					Range:     idxRange,
				})
			}
		}
	}

	return tokens
}
