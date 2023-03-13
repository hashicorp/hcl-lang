package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (set Set) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	eType, ok := set.expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return reference.Targets{}
	}

	if len(eType.Exprs) == 0 || set.cons.Elem == nil {
		return reference.Targets{}
	}

	targets := make(reference.Targets, 0)

	// TODO: collect parent target for the whole set
	// See https://github.com/hashicorp/hcl-lang/issues/228

	return targets
}
