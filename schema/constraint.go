// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
)

type constraintSigil struct{}

type Constraint interface {
	isConstraintImpl() constraintSigil
	FriendlyName() string
	Copy() Constraint

	// EmptyCompletionData provides completion data in context where
	// there is no corresponding configuration, such as when the Constraint
	// is part of another and it is desirable to complete
	// the parent constraint as whole.
	EmptyCompletionData(ctx context.Context, nextPlaceholder int, nestingLevel int) CompletionData
}

type ConstraintWithHoverData interface {
	// EmptyHoverData provides hover data in context where there is
	// no corresponding configuration, such as when the Constraint
	// is part of another and more detailed hover data is requested
	// for the parent.
	//
	// This enables e.g. rendering attributes under Object rather
	// than just "object".
	EmptyHoverData(nestingLevel int) *HoverData
}

type Validatable interface {
	Validate() error
}

// TypeAwareConstraint represents a constraint which may be type-aware.
// Most constraints which implement this are always type-aware,
// but for some this is runtime concern depending on the configuration.
//
// This makes it comparable to another type for conformity during completion
// and it enables collection of type-aware reference target, if the attribute
// itself is targetable as type-aware.
type TypeAwareConstraint interface {
	ConstraintType() (cty.Type, bool)
}

type CompletionData struct {
	NewText string

	// Snippet represents text to be inserted via text edits,
	// with snippet placeholder identifiers such as ${1} (if any) starting
	// from given nextPlaceholder (provided as arg to EmptyCompletionData).
	Snippet string

	TriggerSuggest  bool
	NextPlaceholder int
}

type HoverData struct {
	Content lang.MarkupContent
}
