package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

// Expression represents an expression capable of providing
// various LSP features for given hcl.Expression and schema.Constraint.
type Expression interface {
	CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate
	HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData
	SemanticTokens(ctx context.Context) []lang.SemanticToken
	ReferenceOrigins(allowSelfRefs bool) reference.Origins
	ReferenceTargets(attrAddr *schema.AttributeAddrSchema) reference.Targets
}

func NewExpression(expr hcl.Expression, cons schema.Constraint) Expression {
	// TODO
	return nil
}
