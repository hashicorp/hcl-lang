package lang

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type ReferenceOrigin struct {
	Addr  Address
	Range hcl.Range

	// Constraints represents any traversal expression constraints
	// for the attribute where the origin was found.
	//
	// Further matching against decoded reference targets is needed
	// for >1 constraints, which is done later at runtime as
	// targets and origins can be decoded at different times.
	Constraints ReferenceOriginConstraints
}

type ReferenceOrigins []ReferenceOrigin

func (ro ReferenceOrigins) Copy() ReferenceOrigins {
	if ro == nil {
		return nil
	}

	newOrigins := make(ReferenceOrigins, len(ro))
	for i, origin := range ro {
		newOrigins[i] = origin.Copy()
	}

	return newOrigins
}

func (ro ReferenceOrigin) Copy() ReferenceOrigin {
	return ReferenceOrigin{
		Addr:        ro.Addr,
		Range:       ro.Range,
		Constraints: ro.Constraints.Copy(),
	}
}

type ReferenceOriginConstraint struct {
	OfScopeId ScopeId
	OfType    cty.Type
}

type ReferenceOriginConstraints []ReferenceOriginConstraint

func (roc ReferenceOriginConstraints) Copy() ReferenceOriginConstraints {
	if roc == nil {
		return nil
	}

	cons := make(ReferenceOriginConstraints, 0)
	for _, oc := range roc {
		cons = append(cons, oc)
	}

	return cons
}
