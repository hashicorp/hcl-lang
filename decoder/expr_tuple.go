package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Tuple struct {
	expr    hcl.Expression
	cons    schema.Tuple
	pathCtx *PathContext
}

func (t Tuple) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	// TODO
	return nil
}
