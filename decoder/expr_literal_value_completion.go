// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (lv LiteralValue) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	typ := lv.cons.Value.Type()

	if isEmptyExpression(lv.expr) {
		editRange := hcl.Range{
			Filename: lv.expr.Range().Filename,
			Start:    pos,
			End:      pos,
		}

		// We expect values to be always fully populated
		ctx = schema.WithPrefillRequiredFields(ctx, true)

		cd := lv.cons.EmptyCompletionData(ctx, 1, 0)

		return []lang.Candidate{
			{
				Label:        labelForLiteralValue(lv.cons.Value, false),
				Detail:       typ.FriendlyName(),
				Kind:         candidateKindForType(typ),
				IsDeprecated: lv.cons.IsDeprecated,
				Description:  lv.cons.Description,
				TextEdit: lang.TextEdit{
					Range:   editRange,
					NewText: cd.NewText,
					Snippet: cd.Snippet,
				},
				TriggerSuggest: cd.TriggerSuggest,
			},
		}
	}

	if typ == cty.Bool {
		return lv.completeBoolAtPos(ctx, pos)
	}

	editRange := lv.expr.Range()
	if editRange.End.Line != pos.Line {
		// account for quotes or brackets that are not closed
		editRange.End = pos
	}

	if !editRange.ContainsPos(pos) {
		// account for trailing character(s) which doesn't appear in AST
		// such as dot, opening bracket etc.
		editRange.End = pos
	}

	cd := lv.cons.EmptyCompletionData(ctx, 1, 0)
	return []lang.Candidate{
		{
			Label:        labelForLiteralValue(lv.cons.Value, false),
			Detail:       typ.FriendlyName(),
			Kind:         candidateKindForType(typ),
			IsDeprecated: lv.cons.IsDeprecated,
			Description:  lv.cons.Description,
			TextEdit: lang.TextEdit{
				Range:   editRange,
				NewText: cd.NewText,
				Snippet: cd.Snippet,
			},
			TriggerSuggest: cd.TriggerSuggest,
		},
	}

	// Avoid partial completion inside complex types for now
}

func (lv LiteralValue) completeBoolAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	switch eType := lv.expr.(type) {

	case *hclsyntax.ScopeTraversalExpr:
		prefixLen := pos.Byte - eType.Range().Start.Byte
		if prefixLen > len(eType.Traversal.RootName()) {
			// The user has probably typed an extra character, such as a
			// period, that is not (yet) part of the expression. This prefix
			// won't match anything, so we'll return early.
			return []lang.Candidate{}
		}
		prefix := eType.Traversal.RootName()[0:prefixLen]
		return lv.boolLiteralValueCandidates(prefix, eType.Range())

	case *hclsyntax.LiteralValueExpr:
		if eType.Val.Type() == cty.Bool {
			value := "false"
			if eType.Val.True() {
				value = "true"
			}
			prefixLen := pos.Byte - eType.Range().Start.Byte
			prefix := value[0:prefixLen]
			return lv.boolLiteralValueCandidates(prefix, eType.Range())
		}
	}

	return []lang.Candidate{}
}

func (lv LiteralValue) boolLiteralValueCandidates(prefix string, editRange hcl.Range) []lang.Candidate {
	candidates := make([]lang.Candidate, 0)

	if lv.cons.Value.False() && strings.HasPrefix("false", prefix) {
		candidates = append(candidates, lang.Candidate{
			Label:        "false",
			Detail:       cty.Bool.FriendlyNameForConstraint(),
			Kind:         lang.BoolCandidateKind,
			IsDeprecated: lv.cons.IsDeprecated,
			Description:  lv.cons.Description,
			TextEdit: lang.TextEdit{
				NewText: "false",
				Snippet: "false",
				Range:   editRange,
			},
		})
	}
	if lv.cons.Value.True() && strings.HasPrefix("true", prefix) {
		candidates = append(candidates, lang.Candidate{
			Label:        "true",
			Detail:       cty.Bool.FriendlyNameForConstraint(),
			Kind:         lang.BoolCandidateKind,
			IsDeprecated: lv.cons.IsDeprecated,
			Description:  lv.cons.Description,
			TextEdit: lang.TextEdit{
				NewText: "true",
				Snippet: "true",
				Range:   editRange,
			},
		})
	}

	return candidates
}
