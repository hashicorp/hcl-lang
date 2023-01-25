package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type List struct {
	expr    hcl.Expression
	cons    schema.List
	pathCtx *PathContext
}

func (l List) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	// TODO
	return nil
}

func (l List) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	// TODO
	return nil
}

func (l List) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	// TODO
	return nil
}
