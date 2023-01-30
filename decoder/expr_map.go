package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Map struct {
	expr    hcl.Expression
	cons    schema.Map
	pathCtx *PathContext
}

func (m Map) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	// TODO
	return nil
}

func (m Map) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	// TODO
	return nil
}
