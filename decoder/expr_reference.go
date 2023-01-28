package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Reference struct {
	expr    hcl.Expression
	cons    schema.Reference
	pathCtx *PathContext
}

func (ref Reference) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	// TODO
	return nil
}
