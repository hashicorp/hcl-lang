package schema

type constraintSigil struct{}

type Constraint interface {
	isConstraintImpl() constraintSigil
	FriendlyName() string
	Copy() Constraint
	// EmptyCompletionData provides completion data (to be used in text edits),
	// with snippet placeholder identifiers such as ${4} (if any) starting
	// from given nextPlaceholder.
	// 1 is the most appropriate placeholder, if none were yet assigned prior
	// to rendering completion data for this constraint (e.g. map key).
	EmptyCompletionData(nextPlaceholder int) CompletionData
}

type Validatable interface {
	Validate() error
}

type CompletionData struct {
	NewText         string
	Snippet         string
	TriggerSuggest  bool
	LastPlaceholder int
}
