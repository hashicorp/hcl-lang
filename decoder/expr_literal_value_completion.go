package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
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

		if typ.IsPrimitiveType() {
			if typ == cty.Bool {
				label := "false"
				if lv.cons.Value.True() {
					label = "true"
				}

				return []lang.Candidate{
					{
						Label:  label,
						Detail: typ.FriendlyName(),
						Kind:   lang.BoolCandidateKind,
						TextEdit: lang.TextEdit{
							NewText: label,
							Snippet: label,
							Range:   editRange,
						},
					},
				}
			}
			return []lang.Candidate{}
		}

		if typ == cty.DynamicPseudoType {
			return []lang.Candidate{}
		}

		return []lang.Candidate{}
	}

	if typ == cty.Bool {
		return []lang.Candidate{}
	}

	// TODO: delegate cty.Map to Map
	// TODO: delegate cty.Object to Object
	// TODO: delegate cty.Tuple to Tuple
	// TODO: delegate cty.Set to Set
	// TODO: delegate cty.List to List

	return nil
}
