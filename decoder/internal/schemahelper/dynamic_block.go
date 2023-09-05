// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schemahelper

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/zclconf/go-cty/cty"
)

func buildDynamicBlockSchema(inputSchema *schema.BodySchema) *schema.BlockSchema {
	dependentBody := make(map[schema.SchemaKey]*schema.BodySchema)
	for blockName, block := range inputSchema.Blocks {
		dependentBody[schema.NewSchemaKey(schema.DependencyKeys{
			Labels: []schema.LabelDependent{
				{Index: 0, Value: blockName},
			},
		})] = &schema.BodySchema{
			Blocks: map[string]*schema.BlockSchema{
				"content": {
					Description: lang.PlainText("The body of each generated block"),
					MinItems:    1,
					MaxItems:    1,
					Body:        block.Body.Copy(),
				},
			},
		}
	}

	return &schema.BlockSchema{
		Description: lang.Markdown("A dynamic block to produce blocks dynamically by iterating over a given complex value"),
		Labels: []*schema.LabelSchema{
			{
				Name:        "name",
				Completable: true,
				IsDepKey:    true,
			},
		},
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"for_each": {
					Constraint: schema.OneOf{
						schema.AnyExpression{OfType: cty.Map(cty.DynamicPseudoType)},
						schema.AnyExpression{OfType: cty.List(cty.DynamicPseudoType)},
						schema.AnyExpression{OfType: cty.Set(cty.String)},
					},
					IsRequired:  true,
					Description: lang.Markdown("A meta-argument that accepts a list, map or a set of strings, and creates an instance for each item in that list, map or set."),
				},
				"iterator": {
					Constraint: schema.LiteralType{Type: cty.String},
					IsOptional: true,
					Description: lang.Markdown("The name of a temporary variable that represents the current " +
						"element of the complex value. Defaults to the label of the dynamic block."),
				},
				"labels": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
					},
					IsOptional: true,
					Description: lang.Markdown("A list of strings that specifies the block labels, " +
						"in order, to use for each generated block."),
				},
			},
		},
		DependentBody: dependentBody,
	}
}
