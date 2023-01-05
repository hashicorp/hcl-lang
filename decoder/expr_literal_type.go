package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type LiteralType struct {
	expr hcl.Expression
	cons schema.LiteralType
}

func (lt LiteralType) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	// TODO
	return nil
}

func (lt LiteralType) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	// TODO
	return nil
}

func (lt LiteralType) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	// TODO
	return nil
}

func (lt LiteralType) ReferenceOrigins(allowSelfRefs bool) reference.Origins {
	// TODO
	return nil
}

func (lt LiteralType) ReferenceTargets(attrAddr *schema.AttributeAddrSchema) reference.Targets {
	// TODO
	return nil
}
