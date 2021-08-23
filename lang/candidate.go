package lang

import (
	"github.com/hashicorp/hcl/v2"
)

const (
	NilCandidateKind CandidateKind = iota

	// structural kinds
	AttributeCandidateKind
	BlockCandidateKind
	LabelCandidateKind

	// expressions
	BoolCandidateKind
	KeywordCandidateKind
	ListCandidateKind
	MapCandidateKind
	NumberCandidateKind
	ObjectCandidateKind
	SetCandidateKind
	StringCandidateKind
	TupleCandidateKind
	TraversalCandidateKind
)

//go:generate stringer -type=CandidateKind -output=candidate_kind_string.go
type CandidateKind uint

// Candidate represents a completion candidate in the form of
// an attribute, block, or a label
type Candidate struct {
	Label               string
	Description         MarkupContent
	Detail              string
	IsDeprecated        bool
	TextEdit            TextEdit
	AdditionalTextEdits []TextEdit
	Kind                CandidateKind

	// TriggerSuggest allows server to instruct the client whether
	// to reopen candidate suggestion popup after insertion
	TriggerSuggest bool
}

// TextEdit represents a change (edit) of an HCL config file
// in the form of a Snippet *and* NewText to replace the given Range.
//
// Snippet and NewText are equivalent, but NewText is provided
// for backwards-compatible reasons.
// Snippet uses 1-indexed placeholders, such as name = ${1:value}.
type TextEdit struct {
	Range   hcl.Range
	NewText string
	Snippet string
}

// Candidates represents a list of candidates and indication
// whether the list is complete or if it needs further filtering
// because there is too many candidates.
//
// Decoder has an upper limit for the number of candidates it can return
// and when the limit is reached, the list is considered incomplete.
type Candidates struct {
	List       []Candidate
	IsComplete bool
}

func (ca Candidates) Len() int {
	return len(ca.List)
}

func (ca Candidates) Less(i, j int) bool {
	// TODO: sort by more metadata, such as IsRequired or IsDeprecated
	return ca.List[i].Label < ca.List[j].Label
}

func (ca Candidates) Swap(i, j int) {
	ca.List[i], ca.List[j] = ca.List[j], ca.List[i]
}

// NewCandidates creates a new (incomplete) list of candidates
// to be appended to.
func NewCandidates() Candidates {
	return Candidates{
		List:       make([]Candidate, 0),
		IsComplete: false,
	}
}

// ZeroCandidates returns a (complete) "list" of no candidates
func ZeroCandidates() Candidates {
	return Candidates{
		List:       make([]Candidate, 0),
		IsComplete: true,
	}
}

// CompleteCandidates creates a static (complete) list of candidates
//
// NewCandidates should be used at runtime instead, but CompleteCandidates
// provides a syntactic sugar useful in tests.
func CompleteCandidates(list []Candidate) Candidates {
	return Candidates{
		List:       list,
		IsComplete: true,
	}
}
