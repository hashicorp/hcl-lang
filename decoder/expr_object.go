package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Object struct {
	expr    hcl.Expression
	cons    schema.Object
	pathCtx *PathContext
}

func (obj Object) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	// TODO
	return nil
}

func (obj Object) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	// TODO
	return nil
}
