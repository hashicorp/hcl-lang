package schema

import (
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
)

// Map represents a map, equivalent of hclsyntax.ObjectConsExpr
// interpreted as map, i.e. with items of unknown keys
// and same value types.
type Map struct {
	// Elem defines constraint to apply to each item of the map
	Elem Constraint

	// Name overrides friendly name of the constraint
	Name string

	// Description defines description of the whole map (affects hover)
	Description lang.MarkupContent

	// MinItems defines minimum number of items (affects completion)
	MinItems uint64

	// MaxItems defines maximum number of items (affects completion)
	MaxItems uint64
}

func (Map) isConstraintImpl() constraintSigil {
	return constraintSigil{}
}

func (m Map) FriendlyName() string {
	if m.Name == "" {
		elemName := m.Elem.FriendlyName()
		if elemName != "" {
			return fmt.Sprintf("map of %s", elemName)
		}
		return "map"
	}
	return m.Name
}

func (m Map) Copy() Constraint {
	return Map{
		Elem:        m.Elem.Copy(),
		Name:        m.Name,
		Description: m.Description,
		MinItems:    m.MinItems,
		MaxItems:    m.MaxItems,
	}
}
