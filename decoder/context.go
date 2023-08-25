// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
)

// DecoderContext represents global context relevant for all possible paths
// served by the Decoder
type DecoderContext struct {
	// UTM parameters for docs URLs
	// utm_source parameter, typically language server identification
	UtmSource string
	// utm_medium parameter, typically language client identification
	UtmMedium string
	// utm_content parameter, e.g. documentHover or documentLink
	UseUtmContent bool

	// CodeLenses represents a slice of executable lenses
	// which will be executed in the exact order they're declared
	CodeLenses []lang.CodeLensFunc

	// CompletionHooks represents a map of available hooks for completion.
	// One can register new hooks by adding an entry to this map. Inside the
	// attribute schema, one can refer to the hooks map key to enable the hook
	// execution on CompletionAtPos.
	CompletionHooks CompletionFuncMap

	// CompletionResolveHooks represents a map of available hooks for
	// completion candidate resolving. One can register new hooks by adding an
	// entry to this map. On completion candidate creation, one can specify a
	// resolve hook by using the map key string. If a completion candidate has
	// a resolve hook, ResolveCandidate will execute the hook and return
	// additional (resolved) data for the completion item.
	CompletionResolveHooks CompletionResolveFuncMap
}

func NewDecoderContext() DecoderContext {
	return DecoderContext{
		CompletionHooks:        make(CompletionFuncMap),
		CompletionResolveHooks: make(CompletionResolveFuncMap),
	}
}

func (d *Decoder) SetContext(ctx DecoderContext) {
	d.ctx = ctx
}

// CompletionFunc is the function signature for completion hooks.
//
// The completion func has access to path, filename, pos and maximum
// candidate count via context:
//
//	path, ok := decoder.PathFromContext(ctx)
//	filename, ok := decoder.FilenameFromContext(ctx)
//	pos, ok := decoder.PosFromContext(ctx)
//	maxCandidates, ok := decoder.MaxCandidatesFromContext(ctx)
type CompletionFunc func(ctx context.Context, value cty.Value) ([]Candidate, error)
type CompletionFuncMap map[string]CompletionFunc

// CompletionResolveFunc is the function signature for resolve hooks.
type CompletionResolveFunc func(ctx context.Context, unresolvedCandidate UnresolvedCandidate) (*ResolvedCandidate, error)
type CompletionResolveFuncMap map[string]CompletionResolveFunc

// Candidate represents a completion candidate created and returned from a
// completion hook.
type Candidate struct {
	// Label represents a human-readable name of the candidate
	// if one exists (otherwise Value is used)
	Label string

	// Detail represents a human-readable string with additional
	// information about this candidate, like symbol information.
	Detail string

	Kind lang.CandidateKind

	// Description represents human-readable description
	// of the candidate
	Description lang.MarkupContent

	// IsDeprecated indicates whether the candidate is deprecated
	IsDeprecated bool

	// RawInsertText represents the final text which is used to build the
	// TextEdit for completion. It should contain quotes when completing
	// strings.
	RawInsertText string

	// ResolveHook represents a resolve hook to call
	// and any arguments to pass to it
	ResolveHook *lang.ResolveHook

	// SortText is an optional string that will be used when comparing this
	// candidate with other candidates
	SortText string
}

// ExpressionCandidate is a simplified version of Candidate and the preferred
// way to create completion candidates from completion hooks for attributes
// values (expressions). One can use ExpressionCompletionCandidate to convert
// those into candidates.
type ExpressionCandidate struct {
	// Value represents the value to be inserted
	Value cty.Value

	// Detail represents a human-readable string with additional
	// information about this candidate, like symbol information.
	Detail string

	// Description represents human-readable description
	// of the candidate
	Description lang.MarkupContent

	// IsDeprecated indicates whether the candidate is deprecated
	IsDeprecated bool
}

// ExpressionCompletionCandidate converts a simplified ExpressionCandidate
// into a Candidate while taking care of populating fields and quoting strings
func ExpressionCompletionCandidate(c ExpressionCandidate) Candidate {
	// We're adding quotes to the string here, as we're always
	// replacing the whole edit range for attribute expressions
	text := fmt.Sprintf("%q", c.Value.AsString())

	return Candidate{
		Label:         text,
		Detail:        c.Detail,
		Kind:          candidateKindForType(c.Value.Type()),
		Description:   c.Description,
		IsDeprecated:  c.IsDeprecated,
		RawInsertText: text,
	}
}

// UnresolvedCandidate contains the information required to call a resolve
// hook for enriching a completion item with more information.
type UnresolvedCandidate struct {
	ResolveHook *lang.ResolveHook
}

// ResolvedCandidate is the result of a resolve hook and can enrich a
// completion item by adding additional content to any of the fields.
//
// A field should be empty if no update is intended.
type ResolvedCandidate struct {
	Description         lang.MarkupContent
	Detail              string
	AdditionalTextEdits []lang.TextEdit
}
