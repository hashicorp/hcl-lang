package reference

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Target struct {
	Addr    lang.Address
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

func (ref Target) Copy() Target {
	return Target{
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

func (r Target) Address() lang.Address {
	return r.Addr
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

func (target Target) Matches(addr lang.Address, cons OriginConstraints) bool {
	if len(target.Addr) > len(addr) {
		return false
	}

	originAddr := addr

	matchesCons := false

	if len(cons) == 0 && target.Type != cty.NilType {
		matchesCons = true
	}

	for _, cons := range cons {
		if !target.MatchesScopeId(cons.OfScopeId) {
			continue
		}

		if target.Type == cty.DynamicPseudoType {
			originAddr = addr.FirstSteps(uint(len(target.Addr)))
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

	return target.Addr.Equals(originAddr) && matchesCons
}
