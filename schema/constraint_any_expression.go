package schema

import (
	"github.com/zclconf/go-cty/cty"
)

// AnyExpression TODO
type AnyExpression struct {
	// OfType defines the type of a type-aware reference
	OfType cty.Type
}

func (AnyExpression) isConstraintImpl() constraintSigil {
	return constraintSigil{}
}

func (ae AnyExpression) FriendlyName() string {
	return ae.OfType.FriendlyNameForConstraint()
}

func (ae AnyExpression) Copy() Constraint {
	return AnyExpression{
		OfType: ae.OfType,
	}
}

func (ae AnyExpression) EmptyCompletionData(nextPlaceholder int) CompletionData {
	// TODO
	return CompletionData{}
}
