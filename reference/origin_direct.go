// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reference

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
)

// DirectOrigin represents an origin which directly targets a file
// and doesn't need a matching target
type DirectOrigin struct {
	// Range represents a range of a local traversal, attribute, or an expression
	Range hcl.Range

	// TargetPath represents what (directory) Path does the origin targets
	TargetPath lang.Path

	// TargetRange represents which file and line the origin targets
	TargetRange hcl.Range
}

func (do DirectOrigin) Copy() Origin {
	return DirectOrigin{
		Range:       do.Range,
		TargetPath:  do.TargetPath,
		TargetRange: do.TargetRange,
	}
}

func (DirectOrigin) isOriginImpl() originSigil {
	return originSigil{}
}

func (do DirectOrigin) OriginRange() hcl.Range {
	return do.Range
}
