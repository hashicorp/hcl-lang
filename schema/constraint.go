package schema

import (
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
	EmptyCompletionData(nextPlaceholder int) CompletionData
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

// Comparable represents a constraint which is type-aware,
// making it possible to compare a given type for conformity.
//
// This can affect completion hooks.
type Comparable interface {
	IsCompatible(typ cty.Type) bool
}

type CompletionData struct {
	NewText string

	// Snippet represents text to be inserted via text edits,
	// with snippet placeholder identifiers such as ${1} (if any) starting
	// from given nextPlaceholder (provided as arg to EmptyCompletionData).
	Snippet string

	TriggerSuggest  bool
	LastPlaceholder int
}

type HoverData struct {
	Content lang.MarkupContent
}
