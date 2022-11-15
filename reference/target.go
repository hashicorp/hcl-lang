package reference

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Target struct {
	// Addr represents the address of the target, as available
	// elsewhere in the configuration
	Addr lang.Address

	// LocalAddr represents the address of the target
	// as available *locally* (e.g. self.attr_name)
	LocalAddr lang.Address

	// TargetableFromRangePtr defines where the target is targetable from.
	// This is considered when matching the target against origin.
	//
	// e.g. count.index is only available within the body of the block
	// where count is declared (and extension enabled)
	TargetableFromRangePtr *hcl.Range

	// ScopeId provides scope for matching/filtering
	// (in addition to Type & Addr/LocalAddr).
	//
	// There should never be two targets with the same Type & address,
	// but there are contexts (e.g. completion) where we don't filter
	// by address and may not have type either (e.g. because targets
	// are type-unaware).
	ScopeId lang.ScopeId

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
	Description lang.MarkupContent

	NestedTargets Targets
}

// rangeOverlaps is a copy of hcl.Range.Overlaps
// https://github.com/hashicorp/hcl/blob/v2.14.1/pos.go#L195-L212
// which accounts for empty ranges that are common in the context of LS
func rangeOverlaps(one, other hcl.Range) bool {
	switch {
	case one.Filename != other.Filename:
		// If the ranges are in different files then they can't possibly overlap
		return false
	case one.Empty() && other.Empty():
		// Empty ranges can never overlap
		return false
	case one.ContainsOffset(other.Start.Byte) || one.ContainsOffset(other.End.Byte):
		return true
	case other.ContainsOffset(one.Start.Byte) || other.ContainsOffset(one.End.Byte):
		return true
	default:
		return false
	}
}

func (ref Target) Copy() Target {
	return Target{
		Addr:                   ref.Addr,
		LocalAddr:              ref.LocalAddr,
		TargetableFromRangePtr: copyHclRangePtr(ref.TargetableFromRangePtr),
		ScopeId:                ref.ScopeId,
		RangePtr:               copyHclRangePtr(ref.RangePtr),
		DefRangePtr:            copyHclRangePtr(ref.DefRangePtr),
		Type:                   ref.Type, // cty.Type is immutable by design
		Name:                   ref.Name,
		Description:            ref.Description,
		NestedTargets:          ref.NestedTargets.Copy(),
	}
}

func copyHclRangePtr(rng *hcl.Range) *hcl.Range {
	if rng == nil {
		return nil
	}
	return rng.Ptr()
}

// Address returns any of the two non-empty addresses
//
// TODO: Return address based on context when we have both
func (r Target) Address() lang.Address {
	addr := r.Addr
	if len(r.LocalAddr) > 0 {
		addr = r.LocalAddr
	}

	return addr
}

func (r Target) FriendlyName() string {
	if r.Name != "" {
		return r.Name
	}

	if r.Type != cty.NilType {
		return r.Type.FriendlyName()
	}

	return "reference"
}

func (r Target) TargetRange() (hcl.Range, bool) {
	if r.RangePtr == nil {
		return hcl.Range{}, false
	}

	return *r.RangePtr, true
}

func (ref Target) MatchesConstraint(te schema.TraversalExpr) bool {
	return ref.MatchesScopeId(te.OfScopeId) && ref.ConformsToType(te.OfType)
}

func (ref Target) MatchesScopeId(scopeId lang.ScopeId) bool {
	return scopeId == "" || ref.ScopeId == scopeId
}

func (ref Target) ConformsToType(typ cty.Type) bool {
	conformsToType := false
	if typ != cty.NilType && ref.Type != cty.NilType {
		if ref.Type == cty.DynamicPseudoType {
			// anything conforms with dynamic
			conformsToType = true
		}
		if errs := ref.Type.TestConformance(typ); len(errs) == 0 {
			conformsToType = true
		}
	}

	return conformsToType || (typ == cty.NilType && ref.Type == cty.NilType)
}

func (target Target) Matches(origin MatchableOrigin) bool {
	originAddr, localOriginAddr := origin.Address(), origin.Address()

	matchesCons := false

	// Unconstrained origins should be uncommon, but they match any target
	if len(origin.OriginConstraints()) == 0 {
		// As long as the target is type-aware. Type-unaware targets
		// generally don't have Type, so we avoid false positive here.
		if target.Type != cty.NilType {
			matchesCons = true
		}
	}

	for _, cons := range origin.OriginConstraints() {
		if !target.MatchesScopeId(cons.OfScopeId) {
			continue
		}

		if target.Type == cty.DynamicPseudoType {
			// Account for the case where the origin address points to a nested
			// segment, which the target address doesn't explicitly contain
			// but implies.
			// e.g. If self.foo target is of "any type" (cty.DynamicPseudoType),
			// then we assume it is a match for self.foo.anything
			// by ignoring the last "anything" segment.
			if len(target.Addr) < len(origin.Address()) {
				originAddr = origin.Address().FirstSteps(uint(len(target.Addr)))
			}
			if len(target.LocalAddr) < len(origin.Address()) {
				localOriginAddr = origin.Address().FirstSteps(uint(len(target.LocalAddr)))
			}
			matchesCons = true
			continue
		}
		if cons.OfType != cty.NilType && target.ConformsToType(cons.OfType) {
			matchesCons = true
		}
		if cons.OfType == cty.NilType && target.Type == cty.NilType {
			// This just simplifies testing
			matchesCons = true
		}
	}

	// Reject origin if it's outside the targetable range
	if target.TargetableFromRangePtr != nil && !rangeOverlaps(*target.TargetableFromRangePtr, origin.OriginRange()) {
		return false
	}

	return (target.LocalAddr.Equals(localOriginAddr) || target.Addr.Equals(originAddr)) && matchesCons
}
