package schema

import (
	"github.com/hashicorp/hcl-lang/lang"
)

// BodySchema describes schema of a body comprised of blocks or attributes
// (if any), where body can be root or body of any block in the hierarchy.
type BodySchema struct {
	Blocks       map[string]*BlockSchema
	Attributes   map[string]*AttributeSchema
	AnyAttribute *AttributeSchema
	IsDeprecated bool
	Detail       string
	Description  lang.MarkupContent

	// TODO: validate conflict between Attributes and AnyAttribute
	// TODO: Functions
}

func (*BodySchema) isSchemaImpl() schemaImplSigil {
	return schemaImplSigil{}
}

// NewBodySchema creates a new BodySchema instance
func NewBodySchema() *BodySchema {
	return &BodySchema{
		Blocks:     make(map[string]*BlockSchema, 0),
		Attributes: make(map[string]*AttributeSchema, 0),
	}
}
