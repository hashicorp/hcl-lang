// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

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
				Label:  labelForLiteralValue(lv.cons.Value, false),
				Detail: typ.FriendlyName(),
				Kind:   candidateKindForType(typ),
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
			Label:  labelForLiteralValue(lv.cons.Value, false),
			Detail: typ.FriendlyName(),
			Kind:   candidateKindForType(typ),
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

func (lt LiteralValue) completeBoolAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	switch eType := lt.expr.(type) {

	case *hclsyntax.ScopeTraversalExpr:
		prefixLen := pos.Byte - eType.Range().Start.Byte
		prefix := eType.Traversal.RootName()[0:prefixLen]
		return boolLiteralCandidates(prefix, eType.Range())

	case *hclsyntax.LiteralValueExpr:
		if eType.Val.Type() == cty.Bool {
			value := "false"
			if eType.Val.True() {
				value = "true"
			}
			prefixLen := pos.Byte - eType.Range().Start.Byte
			prefix := value[0:prefixLen]
			return boolLiteralCandidates(prefix, eType.Range())
		}
	}

	return []lang.Candidate{}
}
