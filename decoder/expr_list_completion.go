package decoder

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (list List) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	if isEmptyExpression(list.expr) {
		label := "[ ]"
		triggerSuggest := false

		if list.cons.Elem != nil {
			label = fmt.Sprintf("[ %s ]", list.cons.Elem.FriendlyName())
			triggerSuggest = true
		}

		return []lang.Candidate{
			{
				Label:       label,
				Detail:      list.cons.FriendlyName(),
				Kind:        lang.ListCandidateKind,
				Description: list.cons.Description,
				TextEdit: lang.TextEdit{
					NewText: "[ ]",
					Snippet: "[ ${0} ]",
					Range: hcl.Range{
						Filename: list.expr.Range().Filename,
						Start:    pos,
						End:      pos,
					},
				},
				TriggerSuggest: triggerSuggest,
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

	fileBytes := list.pathCtx.Files[list.expr.Range().Filename].Bytes

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

		var lastElemEndPos hcl.Pos
		for _, elemExpr := range eType.Exprs {
			if elemExpr.Range().ContainsPos(pos) || elemExpr.Range().End.Byte == pos.Byte {
				return newExpression(list.pathCtx, elemExpr, list.cons.Elem).CompletionAtPos(ctx, pos)
			}
			lastElemEndPos = elemExpr.Range().End
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

		expr := newEmptyExpressionAtPos(eType.Range().Filename, pos)
		return newExpression(list.pathCtx, expr, list.cons.Elem).CompletionAtPos(ctx, pos)
	}

	return []lang.Candidate{}
}
