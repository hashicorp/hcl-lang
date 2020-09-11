package schema

import (
	"github.com/hashicorp/hcl-lang/lang"
)

// AttributeSchema describes schema for a block
// e.g. "resource" or "provider" in Terraform
type BlockSchema struct {
	Labels        []*LabelSchema
	Type          BlockType
	Body          *BodySchema
	DependentBody map[SchemaKey]*BodySchema

	Description  lang.MarkupContent
	IsDeprecated bool
	MinItems     uint64
	MaxItems     uint64
}

func (*BlockSchema) isSchemaImpl() schemaImplSigil {
	return schemaImplSigil{}
}
