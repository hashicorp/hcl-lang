package decoder

import (
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
		triggerSuggest := false

		if len(tuple.cons.Elems) > 0 {
			elemLabel := ""
			for _, elemCons := range tuple.cons.Elems {
				elemLabel += fmt.Sprintf("%s, ", elemCons.FriendlyName())
			}
			if len(elemLabel) > 20 {
				elemLabel = fmt.Sprintf("%s…", elemLabel[0:19])
			}
			label = fmt.Sprintf("[ %s ]", elemLabel)
			triggerSuggest = true
		}

		return []lang.Candidate{
			{
				Label:       label,
				Detail:      tuple.cons.FriendlyName(),
				Kind:        lang.TupleCandidateKind,
				Description: tuple.cons.Description,
				TextEdit: lang.TextEdit{
					NewText: "[ ]",
					Snippet: "[ ${0} ]",
					Range: hcl.Range{
						Filename: tuple.expr.Range().Filename,
						Start:    pos,
						End:      pos,
					},
				},
				TriggerSuggest: triggerSuggest,
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

	fileBytes := tuple.pathCtx.Files[tuple.expr.Range().Filename].Bytes

	betweenBraces := hcl.Range{
		Filename: eType.Range().Filename,
		Start:    eType.OpenRange.End,
		End: hcl.Pos{
			// account for the trailing brace }
			Line:   eType.Range().End.Line,
			Column: eType.Range().End.Column - 1,
			Byte:   eType.Range().End.Byte - 1,
		},
	}

	if betweenBraces.ContainsPos(pos) {
		if len(eType.Exprs) == 0 {
			expr := newEmptyExpressionAtPos(eType.Range().Filename, pos)
			return newExpression(tuple.pathCtx, expr, tuple.cons.Elems[0]).CompletionAtPos(ctx, pos)
		}

		if len(eType.Exprs) <= len(tuple.cons.Elems) {
			var lastElemEndPos hcl.Pos
			// check for completion inside individual elements
			for i, elemExpr := range eType.Exprs {
				if elemExpr.Range().ContainsPos(pos) {
					return newExpression(tuple.pathCtx, elemExpr, tuple.cons.Elems[i]).CompletionAtPos(ctx, pos)
				}
				lastElemEndPos = elemExpr.Range().End
			}

			if pos.Byte > lastElemEndPos.Byte {
				if len(eType.Exprs) == len(tuple.cons.Elems) {
					// no more elements to complete, all declared
					return []lang.Candidate{}
				}

				rng := hcl.Range{
					Filename: eType.Range().Filename,
					Start:    lastElemEndPos,
					End:      pos,
				}
				// TODO: test with multi-line element expressions
				b := rng.SliceBytes(fileBytes)
				if strings.TrimSpace(string(b)) != "," {
					return []lang.Candidate{}
				}

				nextIdx := len(eType.Exprs)
				expr := newEmptyExpressionAtPos(eType.Range().Filename, pos)
				return newExpression(tuple.pathCtx, expr, tuple.cons.Elems[nextIdx]).CompletionAtPos(ctx, pos)
			}
		}
	}

	return []lang.Candidate{}
}
