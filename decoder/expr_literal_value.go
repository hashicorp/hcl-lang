package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type LiteralValue struct {
	expr hcl.Expression
	cons schema.LiteralValue
}

func (lv LiteralValue) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	// TODO
	return nil
}

func (lv LiteralValue) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	// TODO
	return nil
}

func (lv LiteralValue) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	// TODO
	return nil
}
