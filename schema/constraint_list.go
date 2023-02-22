package schema

import (
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
)

// List represents a list, equivalent of hclsyntax.TupleConsExpr
// interpreted as list, i.e. ordering of item (which are all
// of the same type) matters.
type List struct {
	// Elem defines constraint to apply to each item
	Elem Constraint

	// Description defines description of the whole list (affects hover)
	Description lang.MarkupContent

	// MinItems defines minimum number of items (affects completion)
	MinItems uint64

	// MaxItems defines maximum number of items (affects completion)
	MaxItems uint64
}

func (List) isConstraintImpl() constraintSigil {
	return constraintSigil{}
}

func (l List) FriendlyName() string {
	if l.Elem != nil && l.Elem.FriendlyName() != "" {
		return fmt.Sprintf("list of %s", l.Elem.FriendlyName())
	}
	return "list"
}

func (l List) Copy() Constraint {
	var elem Constraint
	if l.Elem != nil {
		elem = l.Elem.Copy()
	}
	return List{
		Elem:        elem,
		Description: l.Description,
		MinItems:    l.MinItems,
		MaxItems:    l.MaxItems,
	}
}

func (l List) EmptyCompletionData(nextPlaceholder int) CompletionData {
	// TODO
	return CompletionData{}
}

func (l List) EmptyHoverData(nestingLevel int) *HoverData {
	// TODO
	return nil
}

func (l List) ConstraintType() (cty.Type, bool) {
	if l.Elem == nil {
		return cty.NilType, false
	}

	elemCons, ok := l.Elem.(TypeAwareConstraint)
	if !ok {
		return cty.NilType, false
	}

	elemType, ok := elemCons.ConstraintType()
	if !ok {
		return cty.NilType, false
	}

	return cty.List(elemType), true
}
