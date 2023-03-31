// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (tuple Tuple) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	if isEmptyExpression(tuple.expr) {
		label := "[ ]"

		if len(tuple.cons.Elems) > 0 {
			names := make([]string, 0, len(tuple.cons.Elems))
			for _, elemCons := range tuple.cons.Elems {
				names = append(names, elemCons.FriendlyName())
			}

			elemLabel := strings.Join(names, ", ")
			if len(elemLabel) > 20 {
				elemLabel = fmt.Sprintf("%s…", elemLabel[0:19])
			}
			label = fmt.Sprintf("[ %s ]", elemLabel)
		}

		d := tuple.cons.EmptyCompletionData(ctx, 1, 0)

		return []lang.Candidate{
			{
				Label:       label,
				Detail:      tuple.cons.FriendlyName(),
				Kind:        lang.TupleCandidateKind,
				Description: tuple.cons.Description,
				TextEdit: lang.TextEdit{
					NewText: d.NewText,
					Snippet: d.Snippet,
					Range: hcl.Range{
						Filename: tuple.expr.Range().Filename,
						Start:    pos,
						End:      pos,
					},
				},
				TriggerSuggest: d.TriggerSuggest,
			},
		}
	}

	eType, ok := tuple.expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return []lang.Candidate{}
	}

	if len(tuple.cons.Elems) == 0 {
		return []lang.Candidate{}
	}

	betweenBraces := hcl.Range{
		Filename: eType.Range().Filename,
		Start:    eType.OpenRange.End,
		End:      eType.Range().End,
	}

	if !betweenBraces.ContainsPos(pos) {
		return []lang.Candidate{}
	}

	if len(eType.Exprs) == 0 {
		expr := newEmptyExpressionAtPos(eType.Range().Filename, pos)
		return newExpression(tuple.pathCtx, expr, tuple.cons.Elems[0]).CompletionAtPos(ctx, pos)
	}

	if len(eType.Exprs) > len(tuple.cons.Elems) {
		return []lang.Candidate{}
	}

	lastElemEndPos := eType.OpenRange.Start
	lastElemIdx := 0
	// check for completion inside individual elements
	for i, elemExpr := range eType.Exprs {
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
			return newExpression(tuple.pathCtx, elemExpr, tuple.cons.Elems[i]).CompletionAtPos(ctx, pos)
		}
		lastElemEndPos = elemExpr.Range().End
		lastElemIdx = i
	}

	if pos.Byte <= lastElemEndPos.Byte {
		return []lang.Candidate{}
	}

	if len(eType.Exprs) == len(tuple.cons.Elems) {
		// no more elements to complete, all declared
		return []lang.Candidate{}
	}

	fileBytes := tuple.pathCtx.Files[eType.Range().Filename].Bytes
	recoveredBytes := recoverLeftBytes(fileBytes, pos, func(byteOffset int, r rune) bool {
		return (r == '[' || r == ',') && byteOffset > lastElemEndPos.Byte
	})
	trimmedBytes := bytes.TrimRight(recoveredBytes, " \t\n")

	if len(trimmedBytes) == 0 {
		return []lang.Candidate{}
	}

	nextIdx := len(eType.Exprs)
	if string(trimmedBytes) == "[" {
		// We're at the beginning of a tuple and want the
		// to complete the first element
		nextIdx = 0
	}
	if string(trimmedBytes) == "," {
		// We're likely within an empty expression and
		// want to provide completion for the current element
		// instead of the next one
		nextIdx = lastElemIdx + 1
	}

	expr := newEmptyExpressionAtPos(eType.Range().Filename, pos)
	return newExpression(tuple.pathCtx, expr, tuple.cons.Elems[nextIdx]).CompletionAtPos(ctx, pos)
}
