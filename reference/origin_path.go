// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reference

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
)

// PathOrigin represents a resolved reference origin
// targeting an attribute or a block in a separate path
type PathOrigin struct {
	// Range represents a range of a local traversal or an attribute
	Range hcl.Range

	// TargetAddr describes the address of the targeted attribute or block
	TargetAddr lang.Address

	// TargetPath represents what Path does the origin target
	TargetPath lang.Path

	// Constraints represent any constraints to use when filtering
	// the targets within the destination Path
	Constraints OriginConstraints
}

func (po PathOrigin) Copy() Origin {
	return PathOrigin{
		Range:       po.Range,
		TargetAddr:  po.TargetAddr.Copy(),
		TargetPath:  po.TargetPath,
		Constraints: po.Constraints.Copy(),
	}
}

func (PathOrigin) isOriginImpl() originSigil {
	return originSigil{}
}

func (po PathOrigin) OriginRange() hcl.Range {
	return po.Range
}

func (po PathOrigin) OriginConstraints() OriginConstraints {
	return po.Constraints
}

func (po PathOrigin) AppendConstraints(oc OriginConstraints) MatchableOrigin {
	po.Constraints = append(po.Constraints, oc...)
	return po
}

func (po PathOrigin) Address() lang.Address {
	return po.TargetAddr
}
