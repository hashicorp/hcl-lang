package schema

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
)

// AttributeSchema describes schema for an attribute
type AttributeSchema struct {
	Description  lang.MarkupContent
	IsRequired   bool
	IsDeprecated bool
	IsReadOnly   bool
	ValueType    cty.Type

	// IsDepKey describes whether to use this attribute (and its value)
	// as key when looking up dependent schema
	IsDepKey bool
}

func (*AttributeSchema) isSchemaImpl() schemaImplSigil {
	return schemaImplSigil{}
}
