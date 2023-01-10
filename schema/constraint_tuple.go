package schema

import (
	"github.com/hashicorp/hcl-lang/lang"
)

// Tuple represents a tuple, equivalent of hclsyntax.TupleConsExpr
// interpreted as tuple, i.e. collection of items where
// each one is of different type.
type Tuple struct {
	// Elems defines constraints to apply to each individual item
	// in the same order they would appear in the tuple
	Elems []Constraint

	// Description defines description of the whole tuple (affects hover)
	Description lang.MarkupContent
}

func (Tuple) isConstraintImpl() constraintSigil {
	return constraintSigil{}
}

func (t Tuple) FriendlyName() string {
	return "tuple"
}

func (t Tuple) Copy() Constraint {
	newTuple := Tuple{
		Description: t.Description,
	}
	if len(t.Elems) > 0 {
		newTuple.Elems = make([]Constraint, len(t.Elems))
		for i, elem := range t.Elems {
			newTuple.Elems[i] = elem.Copy()
		}
	}
	return newTuple
}

func (t Tuple) EmptyCompletionData(nextPlaceholder int) CompletionData {
	// TODO
	return CompletionData{}
}
