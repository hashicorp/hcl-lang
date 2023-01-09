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
	ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins
	ReferenceTargets(ctx context.Context, attrAddr *schema.AttributeAddrSchema) reference.Targets
}

func NewExpression(expr hcl.Expression, cons schema.Constraint) Expression {
	switch c := cons.(type) {
	case schema.LiteralType:
		return LiteralType{expr: expr, cons: c}
	case schema.Reference:
		return Reference{expr: expr, cons: c}
	case schema.TypeDeclaration:
		return TypeDeclaration{expr: expr, cons: c}
	case schema.Keyword:
		return Keyword{expr: expr, cons: c}
	case schema.List:
		return List{expr: expr, cons: c}
	case schema.Set:
		return Set{expr: expr, cons: c}
	case schema.Tuple:
		return Tuple{expr: expr, cons: c}
	case schema.Object:
		return Object{expr: expr, cons: c}
	case schema.Map:
		return Map{expr: expr, cons: c}
	case schema.OneOf:
		return OneOf{expr: expr, cons: c}
	case schema.LiteralValue:
		return LiteralValue{expr: expr, cons: c}
	}

	return unknownExpression{}
}
