package schema

import (
	"strings"

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

type MapExpr struct {
	Elem        ExprConstraints
	Name        string
	Description lang.MarkupContent
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

type ObjectExprAttributes map[string]ObjectAttribute

func (ObjectExprAttributes) isExprConstraintImpl() exprConstrSigil {
	return exprConstrSigil{}
}

func (oe ObjectExprAttributes) FriendlyName() string {
	return "attributes"
}

type ObjectAttribute struct {
	Expr        ExprConstraints
	Description lang.MarkupContent
}

func (oa ObjectAttribute) FriendlyName() string {
	return oa.Expr.FriendlyName()
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

func LiteralTypeOnly(t cty.Type) ExprConstraints {
	return ExprConstraints{
		LiteralTypeExpr{Type: t},
	}
}
