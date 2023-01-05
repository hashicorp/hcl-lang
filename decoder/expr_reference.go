package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Reference struct {
	expr hcl.Expression
	cons schema.Reference
}

func (ref Reference) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	// TODO
	return nil
}

func (ref Reference) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	// TODO
	return nil
}

func (ref Reference) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	// TODO
	return nil
}

func (ref Reference) ReferenceOrigins(allowSelfRefs bool) reference.Origins {
	// TODO
	return nil
}

func (ref Reference) ReferenceTargets(attrAddr *schema.AttributeAddrSchema) reference.Targets {
	// TODO
	return nil
}
