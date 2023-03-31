// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reference

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
)

type originSigil struct{}

type Origin interface {
	isOriginImpl() originSigil
	Copy() Origin
	OriginRange() hcl.Range
}

type MatchableOrigin interface {
	Origin
	OriginConstraints() OriginConstraints
	AppendConstraints(OriginConstraints) MatchableOrigin
	Address() lang.Address
}
