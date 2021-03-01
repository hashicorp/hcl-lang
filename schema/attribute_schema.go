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

	Expr ExprConstraints

	// IsDepKey describes whether to use this attribute (and its value)
	// as key when looking up dependent schema
	IsDepKey bool

	Address *AttributeAddrSchema
}

type AttributeAddrSchema struct {
	Steps []AddrStep

	FriendlyName string
	ScopeId      lang.ScopeId

	// AsData defines whether the value of the attribute
	// is addressable as a matching literal type
	// included in Expr
	AsData bool

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
		if !as.Address.AsData && !as.Address.AsReference {
			return fmt.Errorf("Address: at least one of AsData or AsReference must be set")
		}

		for i, step := range as.Address.Steps {
			if _, ok := step.(LabelStep); ok {
				return fmt.Errorf("Address[%d]: LabelStep is not valid for attribute", i)
			}
			if _, ok := step.(AttrValueStep); ok {
				return fmt.Errorf("Address[%d]: AttrValueStep is not implemented for attribute", i)
			}
		}
	}

	return as.Expr.Validate()
}
