package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type unknownExpression struct{}

func (oo unknownExpression) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	return []lang.Candidate{}
}

func (oo unknownExpression) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	return nil
}

func (oo unknownExpression) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	return []lang.SemanticToken{}
}

func (oo unknownExpression) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	return reference.Origins{}
}

func (oo unknownExpression) ReferenceTargets(ctx context.Context, attrAddr *schema.AttributeAddrSchema) reference.Targets {
	return reference.Targets{}
}
