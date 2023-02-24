package decoder

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (set Set) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	if isEmptyExpression(set.expr) {
		label := "[ ]"
		triggerSuggest := false

		if set.cons.Elem != nil {
			label = fmt.Sprintf("[ %s ]", set.cons.Elem.FriendlyName())
			triggerSuggest = true
		}

		return []lang.Candidate{
			{
				Label:       label,
				Detail:      set.cons.FriendlyName(),
				Kind:        lang.SetCandidateKind,
				Description: set.cons.Description,
				TextEdit: lang.TextEdit{
					NewText: "[ ]",
					Snippet: "[ ${0} ]",
					Range: hcl.Range{
						Filename: set.expr.Range().Filename,
						Start:    pos,
						End:      pos,
					},
				},
				TriggerSuggest: triggerSuggest,
			},
		}
	}

	eType, ok := set.expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return []lang.Candidate{}
	}

	if set.cons.Elem == nil {
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
			return newExpression(set.pathCtx, expr, set.cons.Elem).CompletionAtPos(ctx, pos)
		}

		// TODO: depending on set.cons.Elem (Keyword, LiteralValue, Reference),
		// filter out declared elements to provide uniqueness as that is the nature of set

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
				return newExpression(set.pathCtx, elemExpr, set.cons.Elem).CompletionAtPos(ctx, pos)
			}
		}

		expr := newEmptyExpressionAtPos(eType.Range().Filename, pos)
		return newExpression(set.pathCtx, expr, set.cons.Elem).CompletionAtPos(ctx, pos)
	}

	return []lang.Candidate{}
}
