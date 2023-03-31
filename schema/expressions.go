// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

func (ec ExprConstraints) Copy() ExprConstraints {
	if ec == nil {
		return make(ExprConstraints, 0)
	}

	newEc := make(ExprConstraints, len(ec))
	for i, c := range ec {
		newEc[i] = c.Copy()
	}

	return newEc
}

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

type exprConstrSigil struct{}

type ExprConstraint interface {
	isExprConstraintImpl() exprConstrSigil
	FriendlyName() string
	Copy() ExprConstraint
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

func (lt LiteralTypeExpr) Copy() ExprConstraint {
	return LiteralTypeExpr{
		// cty.Type is immutable by design
		Type: lt.Type,
	}
}

type LegacyLiteralValue struct {
	Val          cty.Value
	IsDeprecated bool
	Description  lang.MarkupContent
}

func (LegacyLiteralValue) isExprConstraintImpl() exprConstrSigil {
	return exprConstrSigil{}
}

func (lv LegacyLiteralValue) FriendlyName() string {
	return lv.Val.Type().FriendlyNameForConstraint()
}

func (lv LegacyLiteralValue) Copy() ExprConstraint {
	return LegacyLiteralValue{
		// cty.Value is immutable by design
		Val:          lv.Val,
		IsDeprecated: lv.IsDeprecated,
		Description:  lv.Description,
	}
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

func (le ListExpr) FriendlyName() string {
	elemName := le.Elem.FriendlyName()
	if elemName != "" {
		return fmt.Sprintf("list of %s", elemName)
	}
	return "list"
}

func (le ListExpr) Copy() ExprConstraint {
	return ListExpr{
		Elem:        le.Elem.Copy(),
		Description: le.Description,
		MinItems:    le.MinItems,
		MaxItems:    le.MaxItems,
	}
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

func (se SetExpr) FriendlyName() string {
	elemName := se.Elem.FriendlyName()
	if elemName != "" {
		return fmt.Sprintf("set of %s", elemName)
	}
	return "set"
}

func (se SetExpr) Copy() ExprConstraint {
	return SetExpr{
		Elem:        se.Elem.Copy(),
		Description: se.Description,
		MinItems:    se.MinItems,
		MaxItems:    se.MaxItems,
	}
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

func (te TupleExpr) Copy() ExprConstraint {
	newTe := TupleExpr{
		Description: te.Description,
	}
	if len(te.Elems) > 0 {
		newTe.Elems = make([]ExprConstraints, len(te.Elems))
		for i, elem := range te.Elems {
			newTe.Elems[i] = elem.Copy()
		}
	}
	return newTe
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
		elemName := me.Elem.FriendlyName()
		if elemName != "" {
			return fmt.Sprintf("map of %s", elemName)
		}
		return "map"
	}
	return me.Name
}

func (me MapExpr) Copy() ExprConstraint {
	return MapExpr{
		Elem:        me.Elem.Copy(),
		Name:        me.Name,
		Description: me.Description,
		MinItems:    me.MinItems,
		MaxItems:    me.MaxItems,
	}
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

func (oe ObjectExpr) Copy() ExprConstraint {
	return ObjectExpr{
		Attributes:  oe.Attributes.Copy().(ObjectExprAttributes),
		Name:        oe.Name,
		Description: oe.Description,
	}
}

type ObjectExprAttributes map[string]*AttributeSchema

func (ObjectExprAttributes) isExprConstraintImpl() exprConstrSigil {
	return exprConstrSigil{}
}

func (oe ObjectExprAttributes) FriendlyName() string {
	return "attributes"
}

func (oe ObjectExprAttributes) Copy() ExprConstraint {
	m := make(ObjectExprAttributes, 0)
	for name, aSchema := range oe {
		m[name] = aSchema.Copy()
	}
	return m
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

func (ke KeywordExpr) Copy() ExprConstraint {
	return KeywordExpr{
		Keyword:     ke.Keyword,
		Name:        ke.Name,
		Description: ke.Description,
	}
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

type TraversalExprs []TraversalExpr

func (tes TraversalExprs) AsConstraints() ExprConstraints {
	if tes == nil {
		return nil
	}
	ec := make(ExprConstraints, 0)
	for _, te := range tes {
		ec = append(ec, te)
	}
	return ec
}

type TraversalAddrSchema struct {
	ScopeId lang.ScopeId
}

func (tas *TraversalAddrSchema) Copy() *TraversalAddrSchema {
	if tas == nil {
		return nil
	}
	return &TraversalAddrSchema{
		ScopeId: tas.ScopeId,
	}
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

func (te TraversalExpr) Copy() ExprConstraint {
	return TraversalExpr{
		OfScopeId: te.OfScopeId,
		OfType:    te.OfType,
		Name:      te.Name,
		Address:   te.Address.Copy(),
	}
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

type TypeDeclarationExpr struct{}

func (TypeDeclarationExpr) isExprConstraintImpl() exprConstrSigil {
	return exprConstrSigil{}
}

func (td TypeDeclarationExpr) FriendlyName() string {
	return "type"
}

func (td TypeDeclarationExpr) Copy() ExprConstraint {
	return TypeDeclarationExpr{}
}
