package schema

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
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

	Address *BlockAddrSchema
}

type BlockAddrSchema struct {
	Steps []AddrStep

	FriendlyName string
	ScopeId      lang.ScopeId

	// AsReference defines whether the block itself
	// is addressable as a type-less reference
	AsReference bool

	// BodyAsData defines whether the data in the block body
	// is addressable as cty.Object or cty.List(cty.Object),
	// cty.Set(cty.Object) etc. depending on block type
	BodyAsData bool
	// InferBody defines whether (static) Body's
	// blocks and attributes are also walked
	// and their addresses inferred as data
	InferBody bool

	// DependentBodyAsData defines whether the data in
	// the dependent block body is addressable as cty.Object
	// or cty.List(cty.Object), cty.Set(cty.Object) etc.
	// depending on block type
	DependentBodyAsData bool
	// InferDependentBody defines whether DependentBody's
	// blocks and attributes are also walked
	// and their addresses inferred as data
	InferDependentBody bool
}

func (as *BlockAddrSchema) Validate() error {
	for i, step := range as.Steps {
		if _, ok := step.(AttrNameStep); ok {
			return fmt.Errorf("Steps[%d]: AttrNameStep is not valid for attribute", i)
		}
	}

	if as.InferBody && !as.BodyAsData {
		return errors.New("InferBody requires BodyAsData")
	}

	if as.InferDependentBody && !as.DependentBodyAsData {
		return errors.New("InferDependentBody requires DependentBodyAsData")
	}

	return nil
}

func (*BlockSchema) isSchemaImpl() schemaImplSigil {
	return schemaImplSigil{}
}

func (bSchema *BlockSchema) Validate() error {
	var errs *multierror.Error

	if bSchema.Address != nil {
		err := bSchema.Address.Validate()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("Address: %w", err))
		}
	}

	if bSchema.Body != nil {
		err := bSchema.Body.Validate()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("Body: %w", err))
		}
	}

	if errs != nil && len(errs.Errors) == 1 {
		return errs.Errors[0]
	}

	return errs.ErrorOrNil()
}
