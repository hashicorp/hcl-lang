package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Keyword struct {
	expr hcl.Expression
	cons schema.Keyword
}

func (kw Keyword) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	// TODO
	return nil
}
