// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schemahelper

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/zclconf/go-cty/cty"
)

func CountAttributeSchema() *schema.AttributeSchema {
	return &schema.AttributeSchema{
		IsOptional: true,
		Constraint: schema.AnyExpression{OfType: cty.Number},
		Description: lang.Markdown("Total number of instances of this block.\n\n" +
			"**Note**: A given block cannot use both `count` and `for_each`."),
	}
}

func ForEachAttributeSchema() *schema.AttributeSchema {
	return &schema.AttributeSchema{
		IsOptional: true,
		Constraint: schema.OneOf{
			schema.AnyExpression{OfType: cty.Map(cty.DynamicPseudoType)},
			schema.AnyExpression{OfType: cty.Set(cty.String)},
		},
		Description: lang.Markdown("A meta-argument that accepts a map or a set of strings, and creates an instance for each item in that map or set.\n\n" +
			"**Note**: A given block cannot use both `count` and `for_each`."),
	}
}
