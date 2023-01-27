package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type OneOf struct {
	expr    hcl.Expression
	cons    schema.OneOf
	pathCtx *PathContext
}

func (oo OneOf) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	// TODO
	return nil
}
