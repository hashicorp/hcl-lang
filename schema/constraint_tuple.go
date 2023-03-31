// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
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

func (t Tuple) EmptyCompletionData(ctx context.Context, nextPlaceholder int, nestingLevel int) CompletionData {
	if len(t.Elems) == 0 {
		return CompletionData{
			NewText:         "[ ]",
			Snippet:         fmt.Sprintf("[ ${%d} ]", nextPlaceholder),
			NextPlaceholder: nextPlaceholder + 1,
		}
	}

	elemNewText := make([]string, len(t.Elems))
	elemSnippets := make([]string, len(t.Elems))
	lastPlaceholder := nextPlaceholder

	for i, elem := range t.Elems {
		cData := elem.EmptyCompletionData(ctx, lastPlaceholder, nestingLevel)
		if cData.NewText == "" || cData.Snippet == "" {
			return CompletionData{
				NewText:         "[ ]",
				Snippet:         fmt.Sprintf("[ ${%d} ]", nextPlaceholder),
				TriggerSuggest:  cData.TriggerSuggest,
				NextPlaceholder: nextPlaceholder + 1,
			}
		}
		elemNewText[i] = cData.NewText
		elemSnippets[i] = cData.Snippet
		lastPlaceholder = cData.NextPlaceholder
	}

	return CompletionData{
		NewText:         fmt.Sprintf("[ %s ]", strings.Join(elemNewText, ", ")),
		Snippet:         fmt.Sprintf("[ %s ]", strings.Join(elemSnippets, ", ")),
		NextPlaceholder: lastPlaceholder,
	}
}

func (t Tuple) EmptyHoverData(nestingLevel int) *HoverData {
	elems := make([]string, len(t.Elems))
	for i, elem := range t.Elems {
		elemCons, ok := elem.(ConstraintWithHoverData)
		if !ok {
			return nil
		}

		hoverData := elemCons.EmptyHoverData(nestingLevel)
		if hoverData == nil {
			return nil
		}

		elems[i] = hoverData.Content.Value
	}

	return &HoverData{
		Content: lang.Markdown(fmt.Sprintf(`tuple([%s])`, strings.Join(elems, ", "))),
	}
}

func (t Tuple) ConstraintType() (cty.Type, bool) {
	elemCons := make([]cty.Type, 0)

	for _, elem := range t.Elems {
		cons, ok := elem.(TypeAwareConstraint)
		if !ok {
			return cty.NilType, false
		}

		elemType, ok := cons.ConstraintType()
		if !ok {
			return cty.NilType, false
		}
		elemCons = append(elemCons, elemType)
	}

	return cty.Tuple(elemCons), true
}
