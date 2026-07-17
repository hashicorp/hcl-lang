// Copyright IBM Corp. 2020, 2026
// SPDX-License-Identifier: MPL-2.0

package schemahelper

import (
	"testing"

	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestMergeBlockBodySchemas_propagatesAnyAttributeFromDependentBody(t *testing.T) {
	cfg := `module "app" {
  source = "git::https://example.com/module.git"
}
`
	f, diags := hclsyntax.ParseConfig([]byte(cfg), "main.tf", hcl.InitialPos)
	if len(diags) > 0 {
		t.Fatal(diags)
	}
	body := f.Body.(*hclsyntax.Body)
	block := body.Blocks[0].AsHCLBlock()

	anyAttr := &schema.AttributeSchema{
		Constraint: schema.AnyExpression{OfType: cty.DynamicPseudoType},
	}

	blockSchema := &schema.BlockSchema{
		Labels: []*schema.LabelSchema{{Name: "name"}},
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"source": {
					Constraint: schema.LiteralType{Type: cty.String},
					IsRequired: true,
					IsDepKey:   true,
				},
			},
		},
		DependentBody: map[schema.SchemaKey]*schema.BodySchema{
			schema.NewSchemaKey(schema.DependencyKeys{
				Attributes: []schema.AttributeDependent{
					{
						Name: "source",
						Expr: schema.ExpressionValue{
							Static: cty.StringVal("git::https://example.com/module.git"),
						},
					},
				},
			}): {
				AnyAttribute: anyAttr,
			},
		},
	}

	merged, result := MergeBlockBodySchemas(block, blockSchema)
	if result != LookupSuccessful {
		t.Fatalf("expected successful dependent body lookup, got %v", result)
	}

	if merged.AnyAttribute == nil {
		t.Fatal("expected AnyAttribute to be propagated from dependent body, got nil")
	}
}
