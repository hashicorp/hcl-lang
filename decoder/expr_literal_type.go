package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type LiteralType struct {
	expr hcl.Expression
	cons schema.LiteralType

	pathCtx *PathContext
}

func (lt LiteralType) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	// TODO
	return nil
}
