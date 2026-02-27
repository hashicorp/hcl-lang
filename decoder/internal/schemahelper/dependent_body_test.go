// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package schemahelper

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestBodySchema_DependentBodySchema_label_basic(t *testing.T) {
	block := &hcl.Block{
		Labels: []string{"theircloud", "blah"},
		Body:   hcl.EmptyBody(),
	}

	bSchema := &schema.BlockSchema{
		Labels: []*schema.LabelSchema{
			{
				Name:     "type",
				IsDepKey: true,
			},
			{
				Name: "name",
			},
		},
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"alias": {
					Constraint: schema.LiteralType{Type: cty.String},
				},
			},
		},
		DependentBody: map[schema.SchemaKey]*schema.BodySchema{
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "theircloud"},
				},
			}): {
				Attributes: map[string]*schema.AttributeSchema{
					"bar": {Constraint: schema.LiteralType{Type: cty.Number}},
				},
			},
		},
	}

	bodySchema, _, result := NewBlockSchema(bSchema).DependentBodySchema(block)
	if result != LookupSuccessful {
		t.Fatal("expected to find body schema for 'theircloud' label")
	}
	expectedSchema := &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"bar": {Constraint: schema.LiteralType{Type: cty.Number}},
		},
	}
	if diff := cmp.Diff(expectedSchema, bodySchema, ctydebug.CmpOptions); diff != "" {
		t.Fatalf("unexpected body schema: %s", diff)
	}
}

func TestBodySchema_DependentBodySchema_mismatchingLabels(t *testing.T) {
	block := &hcl.Block{
		Labels: []string{"theircloud"},
		Body:   hcl.EmptyBody(),
	}

	bSchema := &schema.BlockSchema{
		Labels: []*schema.LabelSchema{
			{
				Name:     "type",
				IsDepKey: true,
			},
			{
				Name:     "name",
				IsDepKey: true,
			},
		},
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"alias": {
					Constraint: schema.LiteralType{Type: cty.String},
				},
			},
		},
		DependentBody: map[schema.SchemaKey]*schema.BodySchema{
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "theircloud"},
					{Index: 1, Value: "blah"},
				},
			}): {
				Attributes: map[string]*schema.AttributeSchema{
					"bar": {Constraint: schema.LiteralType{Type: cty.Number}},
				},
			},
		},
	}

	_, _, result := NewBlockSchema(bSchema).DependentBodySchema(block)
	if result != LookupFailed {
		t.Fatal("expected to not find body schema for mismatching label schema")
	}
}

func TestBodySchema_DependentBodySchema_dependentAttr(t *testing.T) {
	firstDepBody := &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"backend": {
				Constraint:             schema.LiteralType{Type: cty.String},
				IsDepKey:               true,
				SemanticTokenModifiers: lang.SemanticTokenModifiers{},
			},
		},
	}
	secondDepBody := &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"extra": {
				Constraint:             schema.LiteralType{Type: cty.Number},
				SemanticTokenModifiers: lang.SemanticTokenModifiers{},
			},
			"backend": {
				Constraint:             schema.LiteralType{Type: cty.String},
				IsDepKey:               true,
				SemanticTokenModifiers: lang.SemanticTokenModifiers{},
			},
		},
	}
	bSchema := &schema.BlockSchema{
		Labels: []*schema.LabelSchema{
			{
				Name:     "type",
				IsDepKey: true,
			},
			{
				Name: "name",
			},
		},
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"alias": {
					Constraint: schema.LiteralType{Type: cty.String},
				},
			},
		},
		DependentBody: map[schema.SchemaKey]*schema.BodySchema{
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "remote_state"},
				},
			}): firstDepBody,
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "remote_state"},
				},
				Attributes: []schema.AttributeDependent{
					{
						Name: "backend",
						Expr: schema.ExpressionValue{Static: cty.StringVal("special")},
					},
				},
			}): secondDepBody,
		},
	}

	bodySchema, _, result := NewBlockSchema(bSchema).DependentBodySchema(&hcl.Block{
		Labels: []string{"remote_state"},
	})
	if result != LookupSuccessful {
		t.Fatal("expected to find body schema for nested dependent schema")
	}
	if diff := cmp.Diff(firstDepBody, bodySchema, ctydebug.CmpOptions); diff != "" {
		t.Fatalf("mismatching body schema: %s", diff)
	}

	bodySchema, _, result = NewBlockSchema(bSchema).DependentBodySchema(&hcl.Block{
		Labels: []string{"remote_state"},
		Body: &hclsyntax.Body{
			Attributes: hclsyntax.Attributes{
				"backend": &hclsyntax.Attribute{
					Name: "backend",
					Expr: &hclsyntax.LiteralValueExpr{
						Val: cty.StringVal("special"),
					},
				},
			},
		},
	})
	if result != LookupSuccessful {
		t.Fatal("expected to find body schema for nested dependent schema")
	}
	if diff := cmp.Diff(secondDepBody, bodySchema, ctydebug.CmpOptions); diff != "" {
		t.Fatalf("mismatching body schema: %s", diff)
	}
}

func TestBodySchema_DependentBodySchema_missingDependentAttr(t *testing.T) {
	firstDepBody := &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"backend": {
				Constraint: schema.LiteralType{Type: cty.String},
				IsDepKey:   true,
			},
		},
	}
	secondDepBody := &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"extra": {Constraint: schema.LiteralType{Type: cty.Number}},
			"backend": {
				Constraint: schema.LiteralType{Type: cty.String},
				IsDepKey:   true,
			},
		},
	}
	bSchema := &schema.BlockSchema{
		Labels: []*schema.LabelSchema{
			{
				Name:     "type",
				IsDepKey: true,
			},
			{
				Name: "name",
			},
		},
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"alias": {
					Constraint: schema.LiteralType{Type: cty.String},
				},
			},
		},
		DependentBody: map[schema.SchemaKey]*schema.BodySchema{
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "remote_state"},
				},
			}): firstDepBody,
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "remote_state"},
				},
				Attributes: []schema.AttributeDependent{
					{
						Name: "backend",
						Expr: schema.ExpressionValue{Static: cty.StringVal("special")},
					},
				},
			}): secondDepBody,
		},
	}

	bodySchema, _, result := NewBlockSchema(bSchema).DependentBodySchema(&hcl.Block{
		Labels: []string{"remote_state"},
		Body: &hclsyntax.Body{
			Attributes: hclsyntax.Attributes{
				"backend": &hclsyntax.Attribute{
					Name: "backend",
					Expr: &hclsyntax.LiteralValueExpr{},
				},
			},
		},
	})
	if result != LookupPartiallySuccessful {
		t.Fatalf("expected to find first body schema for missing keys; reported: %q", result)
	}
	if diff := cmp.Diff(firstDepBody, bodySchema, ctydebug.CmpOptions); diff != "" {
		t.Fatalf("mismatching body schema: %s", diff)
	}
}

func TestBodySchema_DependentBodySchema_attributes(t *testing.T) {
	testCases := []struct {
		name           string
		attributes     hclsyntax.Attributes
		schema         blockSchema
		expectedSchema *schema.BodySchema
	}{
		{
			"single string attribute",
			map[string]*hclsyntax.Attribute{
				"depattr": {
					Name: "depattr",
					Expr: &hclsyntax.LiteralValueExpr{
						Val: cty.StringVal("dep-val"),
					},
				},
			},
			testSchemaWithAttributes,
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"depval_attr": {Constraint: schema.LiteralType{Type: cty.String}},
				},
			},
		},
		{
			"single numeric attribute",
			map[string]*hclsyntax.Attribute{
				"depnum": {
					Name: "depnum",
					Expr: &hclsyntax.LiteralValueExpr{
						Val: cty.NumberIntVal(42),
					},
				},
			},
			testSchemaWithAttributes,
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"number_found": {Constraint: schema.LiteralType{Type: cty.Number}},
				},
			},
		},
		{
			"reference attribute",
			map[string]*hclsyntax.Attribute{
				"depref": {
					Name: "depref",
					Expr: &hclsyntax.ScopeTraversalExpr{
						Traversal: hcl.Traversal{
							hcl.TraverseRoot{Name: "myroot"},
							hcl.TraverseAttr{Name: "attrstep"},
						},
					},
				},
			},
			testSchemaWithAttributes,
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"refbar": {Constraint: schema.LiteralType{Type: cty.Number}},
				},
			},
		},
		{
			"two attributes sorted",
			map[string]*hclsyntax.Attribute{
				"depattr": {
					Name: "depattr",
					Expr: &hclsyntax.LiteralValueExpr{
						Val: cty.StringVal("pumpkin"),
					},
				},
				"depnum": {
					Name: "depnum",
					Expr: &hclsyntax.LiteralValueExpr{
						Val: cty.NumberIntVal(55),
					},
				},
			},
			testSchemaWithAttributes,
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"sortedattr": {Constraint: schema.LiteralType{Type: cty.String}},
				},
			},
		},
		{
			"two attributes unsorted",
			map[string]*hclsyntax.Attribute{
				"depattr": {
					Name: "depattr",
					Expr: &hclsyntax.LiteralValueExpr{
						Val: cty.StringVal("pumpkin"),
					},
				},
				"depnum": {
					Name: "depnum",
					Expr: &hclsyntax.LiteralValueExpr{
						Val: cty.NumberIntVal(2),
					},
				},
			},
			testSchemaWithAttributes,
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"unsortedattr": {Constraint: schema.LiteralType{Type: cty.String}},
				},
			},
		},
		{
			"attribute with default value only",
			map[string]*hclsyntax.Attribute{},
			testSchemaWithAttributesWithDefaultValue,
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"defaultattr": {Constraint: schema.LiteralType{Type: cty.String}},
				},
			},
		},
		{
			"attribute with default value and explicit value",
			map[string]*hclsyntax.Attribute{
				"depattr": {
					Name: "depattr",
					Expr: &hclsyntax.LiteralValueExpr{
						Val: cty.StringVal("pumpkin"),
					},
				},
			},
			testSchemaWithAttributesWithDefaultValue,
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"combinedattr": {Constraint: schema.LiteralType{Type: cty.String}},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			block := &hcl.Block{
				Body: &hclsyntax.Body{
					Attributes: tc.attributes,
				},
			}
			bodySchema, _, result := tc.schema.DependentBodySchema(block)
			if result != LookupSuccessful {
				t.Fatalf("expected to find body schema for given block with %d attributes",
					len(tc.attributes))
			}
			if diff := cmp.Diff(tc.expectedSchema, bodySchema, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("unexpected body schema: %s", diff)
			}
		})
	}
}

func TestBodySchema_DependentBodySchema_partialMergeFailure(t *testing.T) {
	testSchema := NewBlockSchema(&schema.BlockSchema{
		Labels: []*schema.LabelSchema{
			{
				Name:     "type",
				IsDepKey: true,
			},
		},
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"count": {
					Constraint: schema.AnyExpression{OfType: cty.Number},
				},
			},
		},
		DependentBody: map[schema.SchemaKey]*schema.BodySchema{
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{
						Index: 0,
						Value: "terraform_remote_state",
					},
				},
			}): {
				Attributes: map[string]*schema.AttributeSchema{
					"first": {
						Constraint: schema.LiteralType{Type: cty.String},
					},
					"backend": {
						Constraint: schema.AnyExpression{OfType: cty.String},
						IsDepKey:   true,
					},
				},
			},
			schema.NewSchemaKey(schema.DependencyKeys{
				Attributes: []schema.AttributeDependent{
					{
						Name: "backend",
						Expr: schema.ExpressionValue{
							Static: cty.StringVal("remote"),
						},
					},
				},
			}): {
				Attributes: map[string]*schema.AttributeSchema{
					"second": {Constraint: schema.LiteralType{Type: cty.String}},
				},
			},
		},
	})

	block := &hcl.Block{
		Labels: []string{"terraform_remote_state"},
		Body: &hclsyntax.Body{
			Attributes: map[string]*hclsyntax.Attribute{
				"backend": {
					Name: "backend",
					Expr: &hclsyntax.ScopeTraversalExpr{
						Traversal: hcl.Traversal{
							hcl.TraverseRoot{Name: "referencestep"},
						},
					},
				},
			},
		},
	}

	_, result := MergeBlockBodySchemas(block, testSchema.BlockSchema)
	if result != LookupPartiallySuccessful {
		t.Fatal("expected partially failed dependent body lookup to fail")
	}
}

func TestBodySchema_DependentBodySchema_label_notFound(t *testing.T) {
	block := &hcl.Block{
		Labels: []string{"test", "mycloud"},
		Body:   hcl.EmptyBody(),
	}
	_, _, result := testSchemaWithLabels.DependentBodySchema(block)
	if result != LookupFailed {
		t.Fatal("expected not to find body schema for 'mycloud' 2nd label")
	}
}

func TestBodySchema_DependentBodySchema_label_storedUnsorted(t *testing.T) {
	block := &hcl.Block{
		Labels: []string{"complexcloud", "pumpkin"},
		Body:   hcl.EmptyBody(),
	}
	bodySchema, _, result := testSchemaWithLabels.DependentBodySchema(block)
	if result != LookupSuccessful {
		t.Fatal("expected to find body schema stored with unsorted keys")
	}
	expectedSchema := &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"event": {Constraint: schema.LiteralType{Type: cty.String}},
		},
	}
	if diff := cmp.Diff(expectedSchema, bodySchema, ctydebug.CmpOptions); diff != "" {
		t.Fatalf("unexpected body schema: %s", diff)
	}
}

func TestBodySchema_DependentBodySchema_label_lookupUnsorted(t *testing.T) {
	block := &hcl.Block{
		Labels: []string{"apple", "crazycloud"},
		Body:   hcl.EmptyBody(),
	}
	_, _, result := testSchemaWithLabels.DependentBodySchema(block)
	if result != LookupFailed {
		t.Fatal("expected to not find body schema based on wrongly sorted labels")
	}
}

func TestBodySchema_DependentBodySchema_allows_overriding(t *testing.T) {
	testSchema := NewBlockSchema(&schema.BlockSchema{
		Labels: []*schema.LabelSchema{
			{
				Name:     "type",
				IsDepKey: true,
			},
		},
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"value": {
					Constraint: schema.AnyExpression{},
				},
			},
			Blocks: map[string]*schema.BlockSchema{
				"config": {
					Description: lang.Markdown("Provider configuration"),
					MaxItems:    1,
				},
			},
		},
		DependentBody: map[schema.SchemaKey]*schema.BodySchema{
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{
						Index: 0,
						Value: "specific",
					},
				},
			}): {
				Attributes: map[string]*schema.AttributeSchema{
					"value": {
						Constraint: schema.LiteralType{Type: cty.String},
					},
				},
				Blocks: map[string]*schema.BlockSchema{
					"config": {
						Description: lang.Markdown("Provider configuration"),
						MaxItems:    1,
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"extra": {
									Constraint: schema.LiteralType{Type: cty.String},
								},
							},
						},
					},
				},
			},
		},
	})

	block := &hcl.Block{
		Labels: []string{"specific"},
		Body: &hclsyntax.Body{
			Attributes: map[string]*hclsyntax.Attribute{
				"value": {
					Name: "value",
					Expr: &hclsyntax.LiteralValueExpr{
						Val: cty.StringVal("hello"),
					},
				},
			},
			Blocks: []*hclsyntax.Block{
				{
					Body: &hclsyntax.Body{
						Attributes: map[string]*hclsyntax.Attribute{
							"extra": {
								Name: "extra",
								Expr: &hclsyntax.LiteralValueExpr{
									Val: cty.StringVal("world"),
								},
							},
						},
					},
				},
			},
		},
	}

	merged, result := MergeBlockBodySchemas(block, testSchema.BlockSchema)
	if result != LookupSuccessful {
		t.Fatal("expected lookup result to be successful")
	}
	if merged.Blocks["config"].Body == nil {
		t.Fatal("expected to find overridden attribute in merged schema for blocks")
	}
	if _, ok := merged.Attributes["value"].Constraint.(schema.AnyExpression); ok {
		t.Fatal("expected to find overridden attribute in merged schema for attributes")
	}
}

var testSchemaWithLabels = NewBlockSchema(&schema.BlockSchema{
	Labels: []*schema.LabelSchema{
		{
			Name:     "type",
			IsDepKey: true,
		},
		{
			Name:     "name",
			IsDepKey: true,
		},
	},
	Body: &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"alias": {
				Constraint: schema.LiteralType{Type: cty.String},
			},
		},
	},
	DependentBody: map[schema.SchemaKey]*schema.BodySchema{
		schema.NewSchemaKey(schema.DependencyKeys{
			Labels: []schema.LabelDependent{
				{Index: 0, Value: "mycloud"},
			},
		}): {
			Attributes: map[string]*schema.AttributeSchema{
				"special_attr": {Constraint: schema.LiteralType{Type: cty.String}},
			},
		},
		schema.NewSchemaKey(schema.DependencyKeys{
			Labels: []schema.LabelDependent{
				{Index: 1, Value: "theircloud"},
			},
		}): {
			Attributes: map[string]*schema.AttributeSchema{
				"foo": {Constraint: schema.LiteralType{Type: cty.Number}},
			},
		},
		schema.NewSchemaKey(schema.DependencyKeys{
			Labels: []schema.LabelDependent{
				{Index: 0, Value: "theircloud"},
			},
		}): {
			Attributes: map[string]*schema.AttributeSchema{
				"bar": {Constraint: schema.LiteralType{Type: cty.Number}},
			},
		},
		schema.NewSchemaKey(schema.DependencyKeys{
			Labels: []schema.LabelDependent{
				{Index: 1, Value: "pumpkin"},
				{Index: 0, Value: "complexcloud"},
			},
		}): {
			Attributes: map[string]*schema.AttributeSchema{
				"event": {Constraint: schema.LiteralType{Type: cty.String}},
			},
		},
		schema.NewSchemaKey(schema.DependencyKeys{
			Labels: []schema.LabelDependent{
				{Index: 0, Value: "crazycloud"},
				{Index: 1, Value: "apple"},
			},
		}): {
			Attributes: map[string]*schema.AttributeSchema{
				"another": {Constraint: schema.LiteralType{Type: cty.String}},
			},
		},
	},
})

var testSchemaWithAttributes = NewBlockSchema(&schema.BlockSchema{
	Labels: []*schema.LabelSchema{
		{
			Name: "type",
		},
		{
			Name: "name",
		},
	},
	Body: &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"depattr": {
				Constraint: schema.LiteralType{Type: cty.String},
				IsDepKey:   true,
			},
			"depnum": {
				Constraint: schema.LiteralType{Type: cty.Number},
				IsDepKey:   true,
			},
			"depref": {
				Constraint: schema.LiteralType{Type: cty.DynamicPseudoType},
				IsDepKey:   true,
			},
		},
	},
	DependentBody: map[schema.SchemaKey]*schema.BodySchema{
		schema.NewSchemaKey(schema.DependencyKeys{
			Attributes: []schema.AttributeDependent{
				{
					Name: "depattr",
					Expr: schema.ExpressionValue{
						Static: cty.StringVal("dep-val"),
					},
				},
			},
		}): {
			Attributes: map[string]*schema.AttributeSchema{
				"depval_attr": {Constraint: schema.LiteralType{Type: cty.String}},
			},
		},
		schema.NewSchemaKey(schema.DependencyKeys{
			Attributes: []schema.AttributeDependent{
				{
					Name: "depnum",
					Expr: schema.ExpressionValue{
						Static: cty.NumberIntVal(42),
					},
				},
			},
		}): {
			Attributes: map[string]*schema.AttributeSchema{
				"number_found": {Constraint: schema.LiteralType{Type: cty.Number}},
			},
		},
		schema.NewSchemaKey(schema.DependencyKeys{
			Attributes: []schema.AttributeDependent{
				{
					Name: "depref",
					Expr: schema.ExpressionValue{
						Address: lang.Address{
							lang.RootStep{Name: "myroot"},
							lang.AttrStep{Name: "attrstep"},
						},
					},
				},
			},
		}): {
			Attributes: map[string]*schema.AttributeSchema{
				"refbar": {Constraint: schema.LiteralType{Type: cty.Number}},
			},
		},
		schema.NewSchemaKey(schema.DependencyKeys{
			Attributes: []schema.AttributeDependent{
				{
					Name: "depattr",
					Expr: schema.ExpressionValue{
						Static: cty.StringVal("pumpkin"),
					},
				},
				{
					Name: "depnum",
					Expr: schema.ExpressionValue{
						Static: cty.NumberIntVal(55),
					},
				},
			},
		}): {
			Attributes: map[string]*schema.AttributeSchema{
				"sortedattr": {Constraint: schema.LiteralType{Type: cty.String}},
			},
		},
		schema.NewSchemaKey(schema.DependencyKeys{
			Attributes: []schema.AttributeDependent{
				{
					Name: "depnum",
					Expr: schema.ExpressionValue{
						Static: cty.NumberIntVal(2),
					},
				},
				{
					Name: "depattr",
					Expr: schema.ExpressionValue{
						Static: cty.StringVal("pumpkin"),
					},
				},
			},
		}): {
			Attributes: map[string]*schema.AttributeSchema{
				"unsortedattr": {Constraint: schema.LiteralType{Type: cty.String}},
			},
		},
	},
})

var testSchemaWithAttributesWithDefaultValue = NewBlockSchema(&schema.BlockSchema{
	Labels: []*schema.LabelSchema{
		{
			Name: "type",
		},
		{
			Name: "name",
		},
	},
	Body: &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"depattr": {
				Constraint: schema.LiteralType{Type: cty.String},
				IsDepKey:   true,
			},
			"depdefault": {
				Constraint:   schema.LiteralType{Type: cty.String},
				IsDepKey:     true,
				DefaultValue: schema.DefaultValue{Value: cty.StringVal("foobar")},
			},
		},
	},
	DependentBody: map[schema.SchemaKey]*schema.BodySchema{
		schema.NewSchemaKey(schema.DependencyKeys{
			Attributes: []schema.AttributeDependent{
				{
					Name: "depattr",
					Expr: schema.ExpressionValue{
						Static: cty.StringVal("dep-val"),
					},
				},
			},
		}): {
			Attributes: map[string]*schema.AttributeSchema{
				"depval_attr": {Constraint: schema.LiteralType{Type: cty.String}},
			},
		},
		schema.NewSchemaKey(schema.DependencyKeys{
			Attributes: []schema.AttributeDependent{
				{
					Name: "depdefault",
					Expr: schema.ExpressionValue{
						Static: cty.StringVal("foobar"),
					},
				},
			},
		}): {
			Attributes: map[string]*schema.AttributeSchema{
				"defaultattr": {Constraint: schema.LiteralType{Type: cty.String}},
			},
		},
		schema.NewSchemaKey(schema.DependencyKeys{
			Attributes: []schema.AttributeDependent{
				{
					Name: "depattr",
					Expr: schema.ExpressionValue{
						Static: cty.StringVal("pumpkin"),
					},
				},
				{
					Name: "depdefault",
					Expr: schema.ExpressionValue{
						Static: cty.StringVal("foobar"),
					},
				},
			},
		}): {
			Attributes: map[string]*schema.AttributeSchema{
				"combinedattr": {Constraint: schema.LiteralType{Type: cty.String}},
			},
		},
	},
})
