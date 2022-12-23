package schema

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
)

// LiteralValue represents a literal value, as defined by Value
// with additional metadata.
type LiteralValue struct {
	Value cty.Value

	// IsDeprecated defines whether the value is deprecated
	IsDeprecated bool

	// Description defines description of the value
	Description lang.MarkupContent
}

func (LiteralValue) isConstraintImpl() constraintSigil {
	return constraintSigil{}
}

func (lv LiteralValue) FriendlyName() string {
	return lv.Value.Type().FriendlyNameForConstraint()
}

func (lv LiteralValue) Copy() Constraint {
	return LiteralValue{
		Value:        lv.Value,
		IsDeprecated: lv.IsDeprecated,
		Description:  lv.Description,
	}
}

func (lv LiteralValue) EmptyCompletionData(nextPlaceholder int) CompletionData {
	// TODO
	return CompletionData{}
}
