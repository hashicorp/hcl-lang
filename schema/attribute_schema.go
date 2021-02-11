package schema

import (
	"errors"

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

	return nil
}
