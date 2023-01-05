package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type List struct {
	expr hcl.Expression
	cons schema.List
}

func (l List) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	// TODO
	return nil
}

func (l List) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	// TODO
	return nil
}

func (l List) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	// TODO
	return nil
}

func (l List) ReferenceOrigins(allowSelfRefs bool) reference.Origins {
	// TODO
	return nil
}

func (l List) ReferenceTargets(attrAddr *schema.AttributeAddrSchema) reference.Targets {
	// TODO
	return nil
}
