package schema

import (
	"github.com/hashicorp/hcl-lang/lang"
)

// LabelSchema describes schema for a label on a particular position
type LabelSchema struct {
	Name        string
	Description lang.MarkupContent

	// IsDepKey describes whether to use this label as key
	// when looking up dependent schema
	IsDepKey bool
}

func (*LabelSchema) isSchemaImpl() schemaImplSigil {
	return schemaImplSigil{}
}
