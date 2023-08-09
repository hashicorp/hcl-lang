// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

func countIndexReferenceTarget(attr *hcl.Attribute, bodyRange hcl.Range) reference.Target {
	return reference.Target{
		LocalAddr: lang.Address{
			lang.RootStep{Name: "count"},
			lang.AttrStep{Name: "index"},
		},
		TargetableFromRangePtr: bodyRange.Ptr(),
		Type:                   cty.Number,
		Description:            lang.Markdown("The distinct index number (starting with 0) corresponding to the instance"),
		RangePtr:               attr.Range.Ptr(),
		DefRangePtr:            attr.NameRange.Ptr(),
	}
}

func forEachReferenceTargets(attr *hcl.Attribute, bodyRange hcl.Range) reference.Targets {
	return reference.Targets{
		{
			LocalAddr: lang.Address{
				lang.RootStep{Name: "each"},
				lang.AttrStep{Name: "key"},
			},
			TargetableFromRangePtr: bodyRange.Ptr(),
			Type:                   cty.String,
			Description:            lang.Markdown("The map key (or set member) corresponding to this instance"),
			RangePtr:               attr.Range.Ptr(),
			DefRangePtr:            attr.NameRange.Ptr(),
		},
		{
			LocalAddr: lang.Address{
				lang.RootStep{Name: "each"},
				lang.AttrStep{Name: "value"},
			},
			TargetableFromRangePtr: bodyRange.Ptr(),
			Type:                   cty.DynamicPseudoType,
			Description:            lang.Markdown("The map value corresponding to this instance. (If a set was provided, this is the same as `each.key`.)"),
			RangePtr:               attr.Range.Ptr(),
			DefRangePtr:            attr.NameRange.Ptr(),
		},
	}
}
