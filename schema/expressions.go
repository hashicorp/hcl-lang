package schema

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
)

type ExprConstraints []ExprConstraint

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
