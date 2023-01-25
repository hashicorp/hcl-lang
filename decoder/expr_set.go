package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Set struct {
	expr    hcl.Expression
	cons    schema.Set
	pathCtx *PathContext
}

func (s Set) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	// TODO
	return nil
}

func (s Set) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	// TODO
	return nil
}
