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
}

type ReferenceOriginsExpression interface {
	ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins
}

type ReferenceTargetsExpression interface {
	ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets
}

// TargetContext describes context for collecting reference targets
type TargetContext struct {
	// FriendlyName is (optional) human-readable name of the expression
	// interpreted as reference target.
	FriendlyName string

	// ScopeId defines scope of a reference to allow for more granular
	// filtering in completion and accurate matching, which is especially
	// important for type-less reference targets (i.e. AsReference: true).
	ScopeId lang.ScopeId

	// AsExprType defines whether the value of the attribute
	// is addressable as a matching literal type constraint included
	// in attribute Expr.
	//
	// cty.DynamicPseudoType (also known as "any type") will create
	// reference of the real type if value is present else cty.DynamicPseudoType.
	AsExprType bool

	// AsReference defines whether the attribute
	// is addressable as a type-less reference
	AsReference bool

	// AttributeAddress represents a resolved address for the attribute
	// to which the expression belongs.
	AttributeAddress lang.Address
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

// newEmptyExpressionAtPos returns a new "artificial" empty expression
// which can be used during completion inside of another expression
// in an empty space which isn't already represented by empty expression.
//
// For example, new argument after comma in function call,
// or new element in a list or set.
func newEmptyExpressionAtPos(filename string, pos hcl.Pos) hcl.Expression {
	return &hclsyntax.LiteralValueExpr{
		Val: cty.DynamicVal,
		SrcRange: hcl.Range{
			Filename: filename,
			Start:    pos,
			End:      pos,
		},
	}
}
