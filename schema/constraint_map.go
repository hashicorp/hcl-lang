package schema

import (
	"fmt"
	"strings"

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
	if m.Elem == nil {
		return CompletionData{
			NewText:         "{}",
			Snippet:         fmt.Sprintf("{ ${%d} }", nextPlaceholder),
			LastPlaceholder: nextPlaceholder + 1,
		}
	}

	elemData := m.Elem.EmptyCompletionData(nextPlaceholder+1, nestingLevel+1)
	if elemData.NewText == "" || elemData.Snippet == "" {
		return CompletionData{
			NewText:         "{}",
			Snippet:         fmt.Sprintf("{ ${%d} }", nextPlaceholder),
			LastPlaceholder: nextPlaceholder + 1,
			TriggerSuggest:  elemData.TriggerSuggest,
		}
	}

	nesting := strings.Repeat("  ", nestingLevel)

	return CompletionData{
		NewText:         fmt.Sprintf("{\n%s\"name\" = %s\n}", nesting, elemData.NewText),
		Snippet:         fmt.Sprintf("{\n%s\"${%d:name}\" = %s\n}", nesting, nextPlaceholder, elemData.Snippet),
		LastPlaceholder: elemData.LastPlaceholder,
		TriggerSuggest:  elemData.TriggerSuggest,
	}
}

func (m Map) EmptyHoverData(nestingLevel int) *HoverData {
	elemCons, ok := m.Elem.(ConstraintWithHoverData)
	if !ok {
		return nil
	}

	hoverData := elemCons.EmptyHoverData(nestingLevel)
	if hoverData == nil {
		return nil
	}

	return &HoverData{
		Content: lang.Markdown(fmt.Sprintf(`map(%s)`, hoverData.Content.Value)),
	}
}

func (m Map) ConstraintType() (cty.Type, bool) {
	if m.Elem == nil {
		return cty.NilType, false
	}

	elemCons, ok := m.Elem.(TypeAwareConstraint)
	if !ok {
		return cty.NilType, false
	}

	elemType, ok := elemCons.ConstraintType()
	if !ok {
		return cty.NilType, false
	}

	return cty.Map(elemType), true
}
