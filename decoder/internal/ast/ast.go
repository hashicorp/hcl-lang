// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ast

import (
	"github.com/hashicorp/hcl/v2"
)

// blockContent represents HCL or JSON block content
type BlockContent struct {
	*hcl.Block

	// Range represents range of the block in HCL syntax
	// or closest available representative range in JSON
	Range hcl.Range
}

// bodyContent represents an HCL or JSON body content
type BodyContent struct {
	Attributes hcl.Attributes
	Blocks     []*BlockContent
	RangePtr   *hcl.Range
}
