// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
)

// AttributeSchema describes schema for an attribute
type AttributeSchema struct {
	Description  lang.MarkupContent
	IsRequired   bool
	IsOptional   bool
	IsDeprecated bool
	IsComputed   bool
	IsSensitive  bool

	// Expr represents expression constraints e.g. what types of
	// expressions are expected for the attribute
	Expr ExprConstraints

	// Constraint represents expression constraint e.g. what types of
	// expressions are expected for the attribute
	Constraint Constraint

	// IsDepKey describes whether to use this attribute (and its value)
	// as key when looking up dependent schema
	IsDepKey bool

	// Address describes whether and how the attribute itself is targetable
	Address *AttributeAddrSchema

	// OriginForTarget describes whether the attribute is treated
	// as an origin for another target (e.g. module inputs,
	// or tfvars entires in Terraform)
	OriginForTarget *PathTarget

	// SemanticTokenModifiers represents the semantic token modifiers
	// to report for the attribute name
	// (in addition to any modifiers of any parent blocks)
	SemanticTokenModifiers lang.SemanticTokenModifiers

	// CompletionHooks represent any hooks which provide
	// additional completion candidates for the attribute.
	// These are typically candidates which cannot be provided
	// via schema and come from external APIs or other sources.
	CompletionHooks lang.CompletionHooks
}

type AttributeAddrSchema struct {
	// Steps describes address steps used to describe the attribute as whole.
	// The last step would typically be AttrNameStep{}.
	Steps Address

	// FriendlyName is (optional) human-readable name of the *outermost*
	// expression interpreted as reference target.
	//
	// The name is used in completion item and in hover data.
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
}

func (*AttributeSchema) isSchemaImpl() schemaImplSigil {
	return schemaImplSigil{}
}

func (as *AttributeSchema) Validate() error {
	if as.IsOptional && as.IsRequired {
		return errors.New("IsOptional or IsRequired must be set, not both")
	}

	if as.IsRequired && as.IsComputed {
		return errors.New("cannot be both IsRequired and IsComputed")
	}

	if !as.IsRequired && !as.IsOptional && !as.IsComputed {
		return errors.New("one of IsRequired, IsOptional, or IsComputed must be set")
	}

	if as.Address != nil {
		if !as.Address.AsExprType && !as.Address.AsReference {
			return fmt.Errorf("Address: at least one of AsExprType or AsReference must be set")
		}

		if err := as.Address.Steps.AttributeValidate(); err != nil {
			return err
		}
	}

	if as.OriginForTarget != nil {
		if err := as.OriginForTarget.Address.AttributeValidate(); err != nil {
			return err
		}
	}

	if (as.Constraint == nil && len(as.Expr) == 0) ||
		(as.Constraint != nil && len(as.Expr) > 0) {
		return errors.New("expected one of Constraint or Expr")
	}

	if as.Constraint != nil {
		if con, ok := as.Constraint.(Validatable); ok {
			err := con.Validate()
			if err != nil {
				return fmt.Errorf("Constraint: %T: %s", as.Constraint, err)
			}
		}
	} else {
		return as.Expr.Validate()
	}

	return nil
}

func (as *AttributeSchema) Copy() *AttributeSchema {
	if as == nil {
		return nil
	}

	newAs := &AttributeSchema{
		IsRequired:             as.IsRequired,
		IsOptional:             as.IsOptional,
		IsDeprecated:           as.IsDeprecated,
		IsComputed:             as.IsComputed,
		IsSensitive:            as.IsSensitive,
		IsDepKey:               as.IsDepKey,
		Description:            as.Description,
		Address:                as.Address.Copy(),
		OriginForTarget:        as.OriginForTarget.Copy(),
		SemanticTokenModifiers: as.SemanticTokenModifiers.Copy(),
		CompletionHooks:        as.CompletionHooks.Copy(),
	}

	if as.Constraint != nil {
		newAs.Constraint = as.Constraint.Copy()
	} else {
		newAs.Expr = as.Expr.Copy()
	}

	return newAs
}

func (aas *AttributeAddrSchema) Copy() *AttributeAddrSchema {
	if aas == nil {
		return nil
	}

	newAas := &AttributeAddrSchema{
		FriendlyName: aas.FriendlyName,
		ScopeId:      aas.ScopeId,
		AsExprType:   aas.AsExprType,
		AsReference:  aas.AsReference,
		Steps:        aas.Steps.Copy(),
	}

	return newAas
}
