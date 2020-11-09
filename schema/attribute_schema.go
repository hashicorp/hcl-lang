package schema

import (
	"errors"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
)

// AttributeSchema describes schema for an attribute
type AttributeSchema struct {
	Description  lang.MarkupContent
	IsRequired   bool
	IsOptional   bool
	IsDeprecated bool
	IsComputed   bool

	ValueType  cty.Type
	ValueTypes ValueTypes

	// IsDepKey describes whether to use this attribute (and its value)
	// as key when looking up dependent schema
	IsDepKey bool
}

type ValueTypes []cty.Type

func (vt ValueTypes) FriendlyNames() []string {
	names := make([]string, len(vt))
	for i, t := range vt {
		names[i] = t.FriendlyName()
	}
	return names
}

func (*AttributeSchema) isSchemaImpl() schemaImplSigil {
	return schemaImplSigil{}
}

func (as *AttributeSchema) Validate() error {
	if len(as.ValueTypes) == 0 && as.ValueType == cty.NilType {
		return errors.New("one of ValueType or ValueTypes must be specified")
	}
	if len(as.ValueTypes) > 0 && as.ValueType != cty.NilType {
		return errors.New("ValueType or ValueTypes must be specified, not both")
	}

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
