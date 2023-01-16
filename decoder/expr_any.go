package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Any struct {
	expr    hcl.Expression
	cons    schema.AnyExpression
	pathCtx *PathContext
}

func (a Any) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	// TODO
	return nil
}

func (a Any) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	// TODO
	return nil
}

func (a Any) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	// TODO
	return nil
}

func (a Any) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	// TODO
	return nil
}

func (a Any) ReferenceTargets(ctx context.Context, attrAddr *schema.AttributeAddrSchema) reference.Targets {
	// TODO
	return nil
}
