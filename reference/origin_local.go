// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reference

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
)

// LocalOrigin represents a resolved reference origin (traversal)
// targeting a *local* attribute or a block within the same path
type LocalOrigin struct {
	// Addr describes the resolved address of the reference
	Addr lang.Address

	// Range represents the range of the traversal
	Range hcl.Range

	// Constraints represents any traversal expression constraints
	// for the attribute where the origin was found.
	//
	// Further matching against decoded reference targets is needed
	// for >1 constraints, which is done later at runtime as
	// targets and origins can be decoded at different times.
	Constraints OriginConstraints
}

func (lo LocalOrigin) Copy() Origin {
	return LocalOrigin{
		Addr:        lo.Addr.Copy(),
		Range:       lo.Range,
		Constraints: lo.Constraints.Copy(),
	}
}

func (LocalOrigin) isOriginImpl() originSigil {
	return originSigil{}
}

func (lo LocalOrigin) OriginRange() hcl.Range {
	return lo.Range
}

func (lo LocalOrigin) OriginConstraints() OriginConstraints {
	return lo.Constraints
}

func (lo LocalOrigin) AppendConstraints(oc OriginConstraints) MatchableOrigin {
	lo.Constraints = append(lo.Constraints, oc...)
	return lo
}

func (lo LocalOrigin) Address() lang.Address {
	return lo.Addr
}
