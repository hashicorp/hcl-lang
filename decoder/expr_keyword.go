package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Keyword struct {
	expr hcl.Expression
	cons schema.Keyword
}

func (kw Keyword) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	// TODO
	return nil
}

func (kw Keyword) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	// TODO
	return nil
}

func (kw Keyword) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	// TODO
	return nil
}

func (kw Keyword) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	// TODO
	return nil
}

func (kw Keyword) ReferenceTargets(ctx context.Context, attrAddr *schema.AttributeAddrSchema) reference.Targets {
	// TODO
	return nil
}
