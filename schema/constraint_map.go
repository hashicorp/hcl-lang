package schema

import (
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
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
		if m.Elem != nil && m.Elem.FriendlyName() != "" {
			return fmt.Sprintf("map of %s", m.Elem.FriendlyName())
		}
		return "map"
	}
	return m.Name
}

func (m Map) Copy() Constraint {
	var elem Constraint
	if m.Elem != nil {
		elem = m.Elem.Copy()
	}
	return Map{
		Elem:        elem,
		Name:        m.Name,
		Description: m.Description,
		MinItems:    m.MinItems,
		MaxItems:    m.MaxItems,
	}
}

func (m Map) EmptyCompletionData(nextPlaceholder int, nestingLevel int) CompletionData {
	// TODO
	return CompletionData{}
}

func (m Map) EmptyHoverData(nestingLevel int) *HoverData {
	// TODO
	return nil
}

func (m Map) ConstraintType() (cty.Type, bool) {
	// TODO
	return cty.NilType, false
}
