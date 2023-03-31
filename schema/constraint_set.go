// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"
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

func (s Set) EmptyCompletionData(ctx context.Context, nextPlaceholder int, nestingLevel int) CompletionData {
	if s.Elem == nil {
		return CompletionData{
			NewText:         "[ ]",
			Snippet:         fmt.Sprintf("[ ${%d} ]", nextPlaceholder),
			NextPlaceholder: nextPlaceholder + 1,
		}
	}

	elemData := s.Elem.EmptyCompletionData(ctx, nextPlaceholder, nestingLevel)
	if elemData.NewText == "" || elemData.Snippet == "" {
		return CompletionData{
			NewText:         "[ ]",
			Snippet:         fmt.Sprintf("[ ${%d} ]", nextPlaceholder),
			TriggerSuggest:  elemData.TriggerSuggest,
			NextPlaceholder: nextPlaceholder + 1,
		}
	}

	return CompletionData{
		NewText:         fmt.Sprintf("[ %s ]", elemData.NewText),
		Snippet:         fmt.Sprintf("[ %s ]", elemData.Snippet),
		NextPlaceholder: elemData.NextPlaceholder,
	}
}

func (s Set) EmptyHoverData(nestingLevel int) *HoverData {
	elemCons, ok := s.Elem.(ConstraintWithHoverData)
	if !ok {
		return nil
	}

	hoverData := elemCons.EmptyHoverData(nestingLevel)
	if hoverData == nil {
		return nil
	}

	return &HoverData{
		Content: lang.Markdown(fmt.Sprintf(`set(%s)`, hoverData.Content.Value)),
	}
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
