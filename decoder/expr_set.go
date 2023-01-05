package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Set struct {
	expr hcl.Expression
	cons schema.Set
}

func (s Set) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	// TODO
	return nil
}

func (s Set) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	// TODO
	return nil
}

func (s Set) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	// TODO
	return nil
}

func (s Set) ReferenceOrigins(allowSelfRefs bool) reference.Origins {
	// TODO
	return nil
}

func (s Set) ReferenceTargets(attrAddr *schema.AttributeAddrSchema) reference.Targets {
	// TODO
	return nil
}
