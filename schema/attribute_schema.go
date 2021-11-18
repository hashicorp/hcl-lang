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

	// IsDepKey describes whether to use this attribute (and its value)
	// as key when looking up dependent schema
	IsDepKey bool

	// Address describes whether and how the attribute itself is targetable
	Address *AttributeAddrSchema

	// OriginForTarget describes whether the attribute is treated
	// as an origin for another target (e.g. module inputs,
	// or tfvars entires in Terraform)
	OriginForTarget *PathTarget
}

type AttributeAddrSchema struct {
	Steps Address

	FriendlyName string
	ScopeId      lang.ScopeId

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

	return as.Expr.Validate()
}

func (as *AttributeSchema) Copy() *AttributeSchema {
	if as == nil {
		return nil
	}

	newAs := &AttributeSchema{
		IsRequired:      as.IsRequired,
		IsOptional:      as.IsOptional,
		IsDeprecated:    as.IsDeprecated,
		IsComputed:      as.IsComputed,
		IsSensitive:     as.IsSensitive,
		IsDepKey:        as.IsDepKey,
		Description:     as.Description,
		Expr:            as.Expr.Copy(),
		Address:         as.Address.Copy(),
		OriginForTarget: as.OriginForTarget.Copy(),
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
