// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
)

// Keyword represents a keyword, represented as hcl.Traversal
// of a single segment.
type Keyword struct {
	// Keyword defines the literal keyword
	Keyword string

	// Name overrides friendly name of the constraint
	Name string

	// Description defines description of the keyword
	Description lang.MarkupContent
}

func (Keyword) isConstraintImpl() constraintSigil {
	return constraintSigil{}
}

func (k Keyword) FriendlyName() string {
	if k.Name == "" {
		return "keyword"
	}
	return k.Name
}

func (k Keyword) Copy() Constraint {
	return Keyword{
		Keyword:     k.Keyword,
		Name:        k.Name,
		Description: k.Description,
	}
}

func (k Keyword) EmptyCompletionData(ctx context.Context, nextPlaceholder int, nestingLevel int) CompletionData {
	return CompletionData{
		TriggerSuggest:  true,
		NextPlaceholder: nextPlaceholder,
	}
}
