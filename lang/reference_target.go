package lang

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type ScopeId string

type ReferenceTarget struct {
	Addr    Address
	ScopeId ScopeId

	// RangePtr represents range of the whole attribute or block
	// or nil if the target is not addressable.
	RangePtr *hcl.Range

	// DefRangePtr represents a definition range, i.e. block header,
	// or an attribute name or nil if the target is not addressable
	// or when it represents multiple list, set or map blocks.
	//
	// This is useful in situation where a representative single-line
	// range is needed - e.g. to render a contextual UI element in
	// the editor near the middle of this range.
	DefRangePtr *hcl.Range

	Type        cty.Type
	Name        string
	Description MarkupContent

	NestedTargets ReferenceTargets
}

type ReferenceTargets []ReferenceTarget

func (refs ReferenceTargets) Copy() ReferenceTargets {
	if refs == nil {
		return nil
	}

	newRefs := make(ReferenceTargets, len(refs))
	for i, ref := range refs {
		newRefs[i] = ref.Copy()
	}

	return newRefs
}

func (ref ReferenceTarget) Copy() ReferenceTarget {
	return ReferenceTarget{
		Addr:          ref.Addr,
		ScopeId:       ref.ScopeId,
		RangePtr:      copyHclRangePtr(ref.RangePtr),
		DefRangePtr:   copyHclRangePtr(ref.DefRangePtr),
		Type:          ref.Type, // cty.Type is immutable by design
		Name:          ref.Name,
		Description:   ref.Description,
		NestedTargets: ref.NestedTargets.Copy(),
	}
}

func copyHclRangePtr(rng *hcl.Range) *hcl.Range {
	if rng == nil {
		return nil
	}
	return rng.Ptr()
}

func (r ReferenceTargets) Len() int {
	return len(r)
}

func (r ReferenceTargets) Less(i, j int) bool {
	return r[i].Addr.String() < r[j].Addr.String()
}

func (r ReferenceTargets) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r ReferenceTarget) Address() Address {
	return r.Addr
}

func (r ReferenceTarget) FriendlyName() string {
	if r.Name != "" {
		return r.Name
	}

	if r.Type != cty.NilType {
		return r.Type.FriendlyName()
	}

	return "reference"
}

func (r ReferenceTarget) TargetRange() (hcl.Range, bool) {
	if r.RangePtr == nil {
		return hcl.Range{}, false
	}

	return *r.RangePtr, true
}
