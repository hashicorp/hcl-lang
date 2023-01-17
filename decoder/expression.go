package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
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

func (d *PathDecoder) newExpression(expr hcl.Expression, cons schema.Constraint) Expression {
	return newExpression(d.pathCtx, expr, cons)
}

func newExpression(pathContext *PathContext, expr hcl.Expression, cons schema.Constraint) Expression {
	switch c := cons.(type) {
	case schema.AnyExpression:
		return Any{
			expr:    expr,
			cons:    c,
			pathCtx: pathContext,
		}
	case schema.LiteralType:
		return LiteralType{
			expr: expr,
			cons: c,
		}
	case schema.LiteralValue:
		return LiteralValue{
			expr: expr,
			cons: c,
		}
	case schema.TypeDeclaration:
		return TypeDeclaration{
			expr: expr,
			cons: c,
		}
	case schema.Keyword:
		return Keyword{
			expr: expr,
			cons: c,
		}
	case schema.Reference:
		return Reference{
			expr: expr,
			cons: c,
		}
	case schema.List:
		return List{
			expr:    expr,
			cons:    c,
			pathCtx: pathContext,
		}
	case schema.Set:
		return Set{
			expr:    expr,
			cons:    c,
			pathCtx: pathContext,
		}
	case schema.Tuple:
		return Tuple{
			expr:    expr,
			cons:    c,
			pathCtx: pathContext,
		}
	case schema.Object:
		return Object{
			expr:    expr,
			cons:    c,
			pathCtx: pathContext,
		}
	case schema.Map:
		return Map{
			expr:    expr,
			cons:    c,
			pathCtx: pathContext,
		}
	case schema.OneOf:
		return OneOf{
			expr:    expr,
			cons:    c,
			pathCtx: pathContext,
		}

	}

	return unknownExpression{}
}

// isEmptyExpression returns true if given expression is suspected
// to be empty, e.g. newline after equal sign.
//
// Because upstream HCL parser doesn't always handle incomplete
// configuration gracefully, this may not cover all cases.
func isEmptyExpression(expr hcl.Expression) bool {
	l, ok := expr.(*hclsyntax.LiteralValueExpr)
	if !ok {
		return false
	}
	if l.Val != cty.DynamicVal {
		return false
	}

	return true
}
