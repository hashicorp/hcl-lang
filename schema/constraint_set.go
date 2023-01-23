package schema

import (
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
)

// Set represents a set, equivalent of hclsyntax.TupleConsExpr
// interpreted as set, i.e. ordering of items (which are all
// of the same type) does not matter.
type Set struct {
	// Elem defines constraint to apply to each item
	Elem Constraint

	// Description defines description of the whole list (affects hover)
	Description lang.MarkupContent

	// MinItems defines minimum number of items (affects completion)
	MinItems uint64

	// MaxItems defines maximum number of items (affects completion)
	MaxItems uint64
}

func (Set) isConstraintImpl() constraintSigil {
	return constraintSigil{}
}

func (s Set) FriendlyName() string {
	if s.Elem != nil && s.Elem.FriendlyName() != "" {
		return fmt.Sprintf("set of %s", s.Elem.FriendlyName())
	}
	return "set"
}

func (s Set) Copy() Constraint {
	var elem Constraint
	if s.Elem != nil {
		elem = s.Elem.Copy()
	}
	return Set{
		Elem:        elem,
		Description: s.Description,
		MinItems:    s.MinItems,
		MaxItems:    s.MaxItems,
	}
}

func (s Set) EmptyCompletionData(nextPlaceholder int) CompletionData {
	// TODO
	return CompletionData{}
}
