package schema

import (
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
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

func (s Set) EmptyCompletionData(nextPlaceholder int, nestingLevel int) CompletionData {
	// TODO
	return CompletionData{}
}

func (s Set) EmptyHoverData(nestingLevel int) *HoverData {
	// TODO
	return nil
}

func (s Set) ConstraintType() (cty.Type, bool) {
	if s.Elem == nil {
		return cty.NilType, false
	}

	elemCons, ok := s.Elem.(TypeAwareConstraint)
	if !ok {
		return cty.NilType, false
	}

	elemType, ok := elemCons.ConstraintType()
	if !ok {
		return cty.NilType, false
	}

	return cty.Set(elemType), true
}
