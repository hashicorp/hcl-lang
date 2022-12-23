package schema

import (
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
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
	elemName := l.Elem.FriendlyName()
	if elemName != "" {
		return fmt.Sprintf("list of %s", elemName)
	}
	return "list"
}

func (l List) Copy() Constraint {
	return List{
		Elem:        l.Elem.Copy(),
		Description: l.Description,
		MinItems:    l.MinItems,
		MaxItems:    l.MaxItems,
	}
}
