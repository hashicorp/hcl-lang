package schema

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
)

type ExprConstraints []ExprConstraint

func (ec ExprConstraints) FriendlyName() string {
	names := make([]string, 0)
	for _, constraint := range ec {
		if name := constraint.FriendlyName(); name != "" &&
			!namesContain(names, name) {
			names = append(names, name)
		}
	}
	if len(names) > 0 {
		return strings.Join(names, " or ")
	}
	return ""
}

func (ec ExprConstraints) Validate() error {
	if len(ec) == 0 {
		return nil
	}

	type validatable interface {
		Validate() error
	}
	var errs *multierror.Error

	for i, constraint := range ec {
		if c, ok := constraint.(validatable); ok {
			err := c.Validate()
			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("(%d: %T) %w", i, constraint, err))
			}
		}
	}

	if errs != nil && len(errs.Errors) == 1 {
		return errs.Errors[0]
	}

	return errs.ErrorOrNil()
}

func namesContain(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

type exprConstrSigil struct{}

type ExprConstraint interface {
	isExprConstraintImpl() exprConstrSigil
	FriendlyName() string
}

type LiteralTypeExpr struct {
	Type cty.Type
}

func (LiteralTypeExpr) isExprConstraintImpl() exprConstrSigil {
	return exprConstrSigil{}
}

func (lt LiteralTypeExpr) FriendlyName() string {
	return lt.Type.FriendlyNameForConstraint()
}

type LiteralValue struct {
	Val         cty.Value
	Description lang.MarkupContent
}

func (LiteralValue) isExprConstraintImpl() exprConstrSigil {
	return exprConstrSigil{}
}

func (lv LiteralValue) FriendlyName() string {
	return lv.Val.Type().FriendlyNameForConstraint()
}

// TODO: Consider removing TupleConsExpr
// in favour of ListExpr, SetExpr and TupleExpr
type TupleConsExpr struct {
	AnyElem     ExprConstraints
	Name        string
	Description lang.MarkupContent
}

func (TupleConsExpr) isExprConstraintImpl() exprConstrSigil {
	return exprConstrSigil{}
}

func (tc TupleConsExpr) FriendlyName() string {
	if tc.Name == "" {
		return "tuple"
	}
	return tc.Name
}

type ListExpr struct {
	Elem        ExprConstraints
	Description lang.MarkupContent
	MinItems    uint64
	MaxItems    uint64
}

func (ListExpr) isExprConstraintImpl() exprConstrSigil {
	return exprConstrSigil{}
}

func (ListExpr) FriendlyName() string {
	return "list"
}

type SetExpr struct {
	Elem        ExprConstraints
	Description lang.MarkupContent
	MinItems    uint64
	MaxItems    uint64
}

func (SetExpr) isExprConstraintImpl() exprConstrSigil {
	return exprConstrSigil{}
}

func (SetExpr) FriendlyName() string {
	return "set"
}

type TupleExpr struct {
	Elems       []ExprConstraints
	Description lang.MarkupContent
}

func (TupleExpr) isExprConstraintImpl() exprConstrSigil {
	return exprConstrSigil{}
}

func (le TupleExpr) FriendlyName() string {
	return "tuple"
}

type MapExpr struct {
	Elem        ExprConstraints
	Name        string
	Description lang.MarkupContent
	MinItems    uint64
	MaxItems    uint64
}

func (MapExpr) isExprConstraintImpl() exprConstrSigil {
	return exprConstrSigil{}
}

func (me MapExpr) FriendlyName() string {
	if me.Name == "" {
		return "map"
	}
	return me.Name
}

type ObjectExpr struct {
	Attributes  ObjectExprAttributes
	Name        string
	Description lang.MarkupContent
}

func (ObjectExpr) isExprConstraintImpl() exprConstrSigil {
	return exprConstrSigil{}
}

func (oe ObjectExpr) FriendlyName() string {
	if oe.Name == "" {
		return "object"
	}
	return oe.Name
}

type ObjectExprAttributes map[string]*AttributeSchema

func (ObjectExprAttributes) isExprConstraintImpl() exprConstrSigil {
	return exprConstrSigil{}
}

func (oe ObjectExprAttributes) FriendlyName() string {
	return "attributes"
}

type KeywordExpr struct {
	Keyword     string
	Name        string
	Description lang.MarkupContent
}

func (KeywordExpr) isExprConstraintImpl() exprConstrSigil {
	return exprConstrSigil{}
}

func (ke KeywordExpr) FriendlyName() string {
	if ke.Name == "" {
		return "keyword"
	}
	return ke.Name
}

type TraversalExpr struct {
	OfScopeId lang.ScopeId
	OfType    cty.Type
	Name      string

	// Address (if not nil) makes the expression
	// itself addressable and provides scope
	// for the decoded reference
	// Only one of Address or OfScopeId/OfType can be declared
	Address *TraversalAddrSchema
}

type TraversalAddrSchema struct {
	ScopeId lang.ScopeId
}

func (TraversalExpr) isExprConstraintImpl() exprConstrSigil {
	return exprConstrSigil{}
}

func (te TraversalExpr) FriendlyName() string {
	if te.Name != "" {
		return te.Name
	}
	if te.OfType != cty.NilType {
		return te.OfType.FriendlyNameForConstraint()
	}

	return "reference"
}

func (te TraversalExpr) Validate() error {
	if te.Address != nil && (te.OfType != cty.NilType || te.OfScopeId != "") {
		return errors.New("cannot be have both Address and OfType/OfScopeId set")
	}
	if te.Address != nil && te.Address.ScopeId == "" {
		return errors.New("Address requires non-emmpty ScopeId")
	}
	return nil
}

func LiteralTypeOnly(t cty.Type) ExprConstraints {
	return ExprConstraints{
		LiteralTypeExpr{Type: t},
	}
}
