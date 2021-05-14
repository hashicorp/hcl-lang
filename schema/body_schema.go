package schema

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
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

	DocsLink *DocsLink

	// TODO: Functions
}

type DocsLink struct {
	URL     string
	Tooltip string
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

func (bs *BodySchema) Validate() error {
	if len(bs.Attributes) > 0 && bs.AnyAttribute != nil {
		return fmt.Errorf("one of Attributes or AnyAttribute must be set, not both")
	}

	var result *multierror.Error
	for name, attr := range bs.Attributes {
		err := attr.Validate()
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("%s: %w", name, err))
		}
	}

	for bType, block := range bs.Blocks {
		err := block.Validate()
		if err != nil {
			if me, ok := err.(*multierror.Error); ok {
				for _, err := range me.Errors {
					result = multierror.Append(result, fmt.Errorf("%s: %w", bType, err))
				}
			} else {
				result = multierror.Append(result, fmt.Errorf("%s: %w", bType, err))
			}
		}
	}

	return result.ErrorOrNil()
}
