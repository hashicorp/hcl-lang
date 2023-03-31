// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"unicode/utf8"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
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

type CanInferTypeExpression interface {
	InferType() (cty.Type, bool)
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

	// ParentAddress represents a resolved "parent" absolute address,
	// such as data.aws_instance.foo.attr_name.
	// This may be address of the attribute, or implied element/item address
	// for complex-type expressions such as object, list, map etc.
	ParentAddress lang.Address

	// ParentLocalAddress represents a resolved "parent" local address,
	// such as self.attr_name.
	// This may be address of the attribute, or implied element/item address
	// for complex-type expressions such as object, list, map etc.
	ParentLocalAddress lang.Address

	// TargetableFromRangePtr defines where the target is locally targetable
	// from via the ParentLocalAddress.
	TargetableFromRangePtr *hcl.Range

	// ParentRangePtr represents the range of the parent target being collected
	// e.g. whole object/map item
	ParentRangePtr *hcl.Range

	// ParentDefRangePtr represents the range of the parent target's definition
	// e.g. object attribute name or map key
	ParentDefRangePtr *hcl.Range
}

func (tctx *TargetContext) Copy() *TargetContext {
	if tctx == nil {
		return nil
	}

	newCtx := &TargetContext{
		FriendlyName:  tctx.FriendlyName,
		ScopeId:       tctx.ScopeId,
		AsExprType:    tctx.AsExprType,
		AsReference:   tctx.AsReference,
		ParentAddress: tctx.ParentAddress.Copy(),
	}

	if tctx.ParentLocalAddress != nil {
		newCtx.ParentLocalAddress = tctx.ParentLocalAddress.Copy()
	}
	if tctx.TargetableFromRangePtr != nil {
		newCtx.TargetableFromRangePtr = tctx.TargetableFromRangePtr.Ptr()
	}

	return newCtx
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
			expr:    expr,
			cons:    c,
			pathCtx: pathContext,
		}
	case schema.LiteralValue:
		return LiteralValue{
			expr: expr,
			cons: c,
		}
	case schema.TypeDeclaration:
		return TypeDeclaration{
			expr:    expr,
			cons:    c,
			pathCtx: pathContext,
		}
	case schema.Keyword:
		return Keyword{
			expr: expr,
			cons: c,
		}
	case schema.Reference:
		return Reference{
			expr:    expr,
			cons:    c,
			pathCtx: pathContext,
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

// recoverLeftBytes seeks left from given pos in given slice of bytes
// and recovers all bytes up until f matches, including that match.
// This allows recovery of incomplete configuration which is not
// present in the parsed AST during completion.
//
// Zero bytes is returned if no match was found.
func recoverLeftBytes(b []byte, pos hcl.Pos, f func(byteOffset int, r rune) bool) []byte {
	firstRune, size := utf8.DecodeLastRune(b[:pos.Byte])
	offset := pos.Byte - size

	// check for early match
	if f(pos.Byte, firstRune) {
		return b[offset:pos.Byte]
	}

	for offset > 0 {
		nextRune, size := utf8.DecodeLastRune(b[:offset])
		if f(offset, nextRune) {
			// record the matched offset
			// and include the matched last rune
			startByte := offset - size
			return b[startByte:pos.Byte]
		}
		offset -= size
	}

	return []byte{}
}

// isObjectItemTerminatingRune returns true if the given rune
// is considered a left terminating character for an item
// in hclsyntax.ObjectConsExpr.
func isObjectItemTerminatingRune(r rune) bool {
	return r == '\n' || r == ',' || r == '{'
}

// rawObjectKey extracts raw key (as string) from KeyExpr of
// any hclsyntax.ObjectConsExpr along with the corresponding range
// and boolean indicating whether the extraction was successful.
//
// This accounts for the two common key representations (quoted and unquoted)
// and enables validation, filtering of object attributes and accurate
// calculation of edit range.
//
// It does *not* account for interpolation inside the key,
// such as { (var.key_name) = "foo" }.
func rawObjectKey(expr hcl.Expression) (string, *hcl.Range, bool) {
	if json.IsJSONExpression(expr) {
		val, diags := expr.Value(&hcl.EvalContext{})
		if diags.HasErrors() {
			return "", nil, false
		}
		if val.Type() != cty.String {
			return "", nil, false
		}

		return val.AsString(), expr.Range().Ptr(), true
	}

	// regardless of what expression it is always wrapped
	keyExpr, ok := expr.(*hclsyntax.ObjectConsKeyExpr)
	if !ok {
		return "", nil, false
	}

	switch eType := keyExpr.Wrapped.(type) {
	// most common "naked" keys
	case *hclsyntax.ScopeTraversalExpr:
		if len(eType.Traversal) != 1 {
			return "", nil, false
		}
		return eType.Traversal.RootName(), eType.Range().Ptr(), true

	// less common quoted keys
	case *hclsyntax.TemplateExpr:
		if !eType.IsStringLiteral() {
			return "", nil, false
		}

		// string literals imply exactly 1 part
		lvExpr, ok := eType.Parts[0].(*hclsyntax.LiteralValueExpr)
		if !ok {
			return "", nil, false
		}

		if lvExpr.Val.Type() != cty.String {
			return "", nil, false
		}
		return lvExpr.Val.AsString(), lvExpr.Range().Ptr(), true
	}

	return "", nil, false
}
