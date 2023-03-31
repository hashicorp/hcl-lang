// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (list List) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	if isEmptyExpression(list.expr) {
		label := "[ ]"

		if list.cons.Elem != nil {
			label = fmt.Sprintf("[ %s ]", list.cons.Elem.FriendlyName())
		}

		d := list.cons.EmptyCompletionData(ctx, 1, 0)

		return []lang.Candidate{
			{
				Label:       label,
				Detail:      list.cons.FriendlyName(),
				Kind:        lang.ListCandidateKind,
				Description: list.cons.Description,
				TextEdit: lang.TextEdit{
					NewText: d.NewText,
					Snippet: d.Snippet,
					Range: hcl.Range{
						Filename: list.expr.Range().Filename,
						Start:    pos,
						End:      pos,
					},
				},
				TriggerSuggest: d.TriggerSuggest,
			},
		}
	}

	eType, ok := list.expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return []lang.Candidate{}
	}

	if list.cons.Elem == nil {
		return []lang.Candidate{}
	}

	betweenBraces := hcl.Range{
		Filename: eType.Range().Filename,
		Start:    eType.OpenRange.End,
		End:      eType.Range().End,
	}

	if betweenBraces.ContainsPos(pos) {
		if len(eType.Exprs) == 0 {
			expr := newEmptyExpressionAtPos(eType.Range().Filename, pos)
			return newExpression(list.pathCtx, expr, list.cons.Elem).CompletionAtPos(ctx, pos)
		}

		for _, elemExpr := range eType.Exprs {
			// We cannot trust ranges of empty expressions, so we imply
			// that invalid configuration follows and we stop here
			// e.g. for completion between commas [keyword, ,keyword]
			if isEmptyExpression(elemExpr) {
				break
			}
			// We overshot the position and stop
			if elemExpr.Range().Start.Byte > pos.Byte {
				break
			}
			if elemExpr.Range().ContainsPos(pos) || elemExpr.Range().End.Byte == pos.Byte {
				return newExpression(list.pathCtx, elemExpr, list.cons.Elem).CompletionAtPos(ctx, pos)
			}
		}

		expr := newEmptyExpressionAtPos(eType.Range().Filename, pos)
		return newExpression(list.pathCtx, expr, list.cons.Elem).CompletionAtPos(ctx, pos)
	}

	return []lang.Candidate{}
}
