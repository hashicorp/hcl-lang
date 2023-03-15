package schema

import (
	"context"

	"github.com/zclconf/go-cty/cty"
)

// AnyExpression represents any expression type convertible
// to the given data type (OfType).
//
// For example function call returning cty.String complies with
// AnyExpression{OfType: cty.String}.
type AnyExpression struct {
	// OfType defines the type which the outermost expression is constrained to
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

func (ae AnyExpression) EmptyCompletionData(ctx context.Context, nextPlaceholder int, nestingLevel int) CompletionData {
	// TODO
	return CompletionData{}
}

func (ae AnyExpression) ConstraintType() (cty.Type, bool) {
	return ae.OfType, true
}
