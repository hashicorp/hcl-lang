package decoder

import (
	"context"

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

	CompletionHooks        CompletionFuncMap
	CompletionResolveHooks CompletionResolveFuncMap
}

func (d *Decoder) SetContext(ctx DecoderContext) {
	d.ctx = ctx
}

type CompletionFunc func(ctx context.Context, value cty.Value) ([]Candidate, error)
type CompletionFuncMap map[string]CompletionFunc

type CompletionResolveFuncMap map[string]CompletionResolveFunc
type CompletionResolveFunc func(ctx context.Context, unresolvedCandidate UnresolvedCandidate) (*ResolvedCandidate, error)

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

	RawInsertText string

	// ResolveHook represents a resolve hook to call
	// and any arguments to pass to it
	ResolveHook *lang.ResolveHook
}

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

func ExpressionCompletionCandidate(c ExpressionCandidate) Candidate {
	return Candidate{
		Label:         c.Value.AsString(),
		Detail:        c.Detail,
		Kind:          candidateKindForType(c.Value.Type()),
		Description:   c.Description,
		IsDeprecated:  c.IsDeprecated,
		RawInsertText: c.Value.AsString(),
	}
}

type UnresolvedCandidate struct {
	ResolveHook *lang.ResolveHook
}

type ResolvedCandidate struct {
	Description         lang.MarkupContent
	Detail              string
	AdditionalTextEdits []lang.TextEdit
}
