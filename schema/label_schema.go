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

	// In cases where label's IsDepKey=true any DependentKey label values
	// within Blocks's DependentBody can be used for completion
	// This enables such behaviour.
	Completable bool
}

func (*LabelSchema) isSchemaImpl() schemaImplSigil {
	return schemaImplSigil{}
}

func (ls *LabelSchema) Copy() *LabelSchema {
	if ls == nil {
		return nil
	}

	return &LabelSchema{
		Name:        ls.Name,
		Completable: ls.Completable,
		Description: ls.Description,
		IsDepKey:    ls.IsDepKey,
	}
}
