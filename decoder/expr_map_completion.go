// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (m Map) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	if isEmptyExpression(m.expr) {
		label := `{ "key" = value }`

		if m.cons.Elem != nil {
			label = fmt.Sprintf(`{ "key" = %s }`, m.cons.Elem.FriendlyName())
		}

		cData := m.cons.EmptyCompletionData(ctx, 1, 0)

		return []lang.Candidate{
			{
				Label:       label,
				Detail:      m.cons.FriendlyName(),
				Kind:        lang.MapCandidateKind,
				Description: m.cons.Description,
				TextEdit: lang.TextEdit{
					NewText: cData.NewText,
					Snippet: cData.Snippet,
					Range: hcl.Range{
						Filename: m.expr.Range().Filename,
						Start:    pos,
						End:      pos,
					},
				},
				TriggerSuggest: cData.TriggerSuggest,
			},
		}
	}

	eType, ok := m.expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return []lang.Candidate{}
	}

	betweenBraces := hcl.Range{
		Filename: eType.Range().Filename,
		Start:    eType.OpenRange.End,
		End: hcl.Pos{
			// exclude the trailing brace } from range
			// to make byte slice comparison easier
			Line:   eType.Range().End.Line,
			Column: eType.Range().End.Column - 1,
			Byte:   eType.Range().End.Byte - 1,
		},
	}

	if betweenBraces.ContainsPos(pos) {
		if m.cons.Elem == nil {
			return []lang.Candidate{}
		}

		cData := m.cons.Elem.EmptyCompletionData(ctx, 2, 0)
		kind := lang.AttributeCandidateKind
		// TODO: replace "attribute" kind w/ Elem type

		editRange := hcl.Range{
			Filename: eType.Range().Filename,
			Start:    pos,
			End:      pos,
		}

		mapItemCandidate := lang.Candidate{
			Label:  fmt.Sprintf("\"key\" = %s", m.cons.Elem.FriendlyName()),
			Detail: m.cons.Elem.FriendlyName(),
			Kind:   kind,
			TextEdit: lang.TextEdit{
				NewText: fmt.Sprintf("\"key\" = %s", cData.NewText),
				Snippet: fmt.Sprintf("\"${1:key}\" = %s", cData.Snippet),
				Range:   editRange,
			},
		}

		if len(eType.Items) == 0 {
			// check for incomplete configuration between {}
			betweenBraces := hcl.Range{
				Filename: eType.Range().Filename,
				Start:    eType.OpenRange.End,
				End: hcl.Pos{
					Line:   eType.SrcRange.End.Line,
					Column: eType.SrcRange.End.Column - 1,
					Byte:   eType.SrcRange.End.Byte - 1,
				},
			}
			fileBytes := m.pathCtx.Files[eType.Range().Filename].Bytes
			remainingBytes := bytes.TrimSpace(betweenBraces.SliceBytes(fileBytes))

			if len(remainingBytes) == 0 {
				return []lang.Candidate{
					mapItemCandidate,
				}
			}

			// if last byte is =, then it's incomplete attribute
			if remainingBytes[len(remainingBytes)-1] == '=' {
				emptyExpr := newEmptyExpressionAtPos(eType.Range().Filename, pos)
				cons := newExpression(m.pathCtx, emptyExpr, m.cons.Elem)
				return cons.CompletionAtPos(ctx, pos)
			}
		}

		recoveryPos := eType.OpenRange.Start
		for _, item := range eType.Items {
			emptyRange := hcl.Range{
				Filename: eType.Range().Filename,
				Start:    item.KeyExpr.Range().End,
				End:      item.ValueExpr.Range().Start,
			}
			if emptyRange.ContainsPos(pos) {
				// exit early if we're in empty space between key and value
				return []lang.Candidate{}
			}

			// check if we've just missed the position
			if pos.Byte < item.KeyExpr.Range().Start.Byte {
				// enable recovery between last item's end and position
				break
			}

			recoveryPos = item.ValueExpr.Range().End

			if item.KeyExpr.Range().ContainsPos(pos) {
				return []lang.Candidate{}
			}
			if item.ValueExpr.Range().ContainsPos(pos) || item.ValueExpr.Range().End.Byte == pos.Byte {
				cons := newExpression(m.pathCtx, item.ValueExpr, m.cons.Elem)
				return cons.CompletionAtPos(ctx, pos)
			}
		}

		// check any incomplete configuration up to a terminating character
		fileBytes := m.pathCtx.Files[eType.Range().Filename].Bytes
		recoveredBytes := recoverLeftBytes(fileBytes, pos, func(offset int, r rune) bool {
			return isObjectItemTerminatingRune(r) && offset > recoveryPos.Byte
		})
		trimmedBytes := bytes.TrimRight(recoveredBytes, " \t")

		if len(trimmedBytes) == 0 {
			return []lang.Candidate{
				mapItemCandidate,
			}
		}

		if len(trimmedBytes) == 1 && isObjectItemTerminatingRune(rune(trimmedBytes[0])) {
			return []lang.Candidate{
				mapItemCandidate,
			}
		}

		// if last byte is =, then it's incomplete attribute
		if trimmedBytes[len(trimmedBytes)-1] == '=' {
			emptyExpr := newEmptyExpressionAtPos(eType.Range().Filename, pos)
			cons := newExpression(m.pathCtx, emptyExpr, m.cons.Elem)
			return cons.CompletionAtPos(ctx, pos)
		}

		return []lang.Candidate{}
	}
	return []lang.Candidate{}
}
