package decoder

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
					Expr: schema.LiteralTypeOnly(cty.String),
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
					"bar": {Expr: schema.LiteralTypeOnly(cty.Number)},
				},
			},
		},
	}

	bodySchema, _, ok := NewBlockSchema(bSchema).DependentBodySchema(block)
	if !ok {
		t.Fatal("expected to find body schema for 'theircloud' label")
	}
	expectedSchema := &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"bar": {Expr: schema.LiteralTypeOnly(cty.Number)},
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
					Expr: schema.LiteralTypeOnly(cty.String),
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
					"bar": {Expr: schema.LiteralTypeOnly(cty.Number)},
				},
			},
		},
	}

	_, _, ok := NewBlockSchema(bSchema).DependentBodySchema(block)
	if ok {
		t.Fatal("expected to not find body schema for mismatching label schema")
	}
}

func TestBodySchema_DependentBodySchema_dependentAttr(t *testing.T) {
	firstDepBody := &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"backend": {
				Expr:     schema.LiteralTypeOnly(cty.String),
				IsDepKey: true,
			},
		},
	}
	secondDepBody := &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"extra": {Expr: schema.LiteralTypeOnly(cty.Number)},
			"backend": {
				Expr:     schema.LiteralTypeOnly(cty.String),
				IsDepKey: true,
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
					Expr: schema.LiteralTypeOnly(cty.String),
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

	bodySchema, _, ok := NewBlockSchema(bSchema).DependentBodySchema(&hcl.Block{
		Labels: []string{"remote_state"},
	})
	if !ok {
		t.Fatal("expected to find body schema for nested dependent schema")
	}
	if diff := cmp.Diff(firstDepBody, bodySchema, ctydebug.CmpOptions); diff != "" {
		t.Fatalf("mismatching body schema: %s", diff)
	}

	bodySchema, _, ok = NewBlockSchema(bSchema).DependentBodySchema(&hcl.Block{
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
	if !ok {
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
				Expr:     schema.LiteralTypeOnly(cty.String),
				IsDepKey: true,
			},
		},
	}
	secondDepBody := &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"extra": {Expr: schema.LiteralTypeOnly(cty.Number)},
			"backend": {
				Expr:     schema.LiteralTypeOnly(cty.String),
				IsDepKey: true,
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
					Expr: schema.LiteralTypeOnly(cty.String),
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

	bodySchema, _, ok := NewBlockSchema(bSchema).DependentBodySchema(&hcl.Block{
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
	if !ok {
		t.Fatal("expected to find first body schema for missing keys")
	}
	if diff := cmp.Diff(firstDepBody, bodySchema, ctydebug.CmpOptions); diff != "" {
		t.Fatalf("mismatching body schema: %s", diff)
	}
}

func TestBodySchema_DependentBodySchema_attributes(t *testing.T) {
	testCases := []struct {
		name           string
		attributes     hclsyntax.Attributes
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
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"depval_attr": {Expr: schema.LiteralTypeOnly(cty.String)},
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
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"number_found": {Expr: schema.LiteralTypeOnly(cty.Number)},
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
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"refbar": {Expr: schema.LiteralTypeOnly(cty.Number)},
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
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"sortedattr": {Expr: schema.LiteralTypeOnly(cty.String)},
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
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"unsortedattr": {Expr: schema.LiteralTypeOnly(cty.String)},
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
			bodySchema, _, ok := testSchemaWithAttributes.DependentBodySchema(block)
			if !ok {
				t.Fatal("expected to find body schema for 'depattr' attribute")
			}
			if diff := cmp.Diff(tc.expectedSchema, bodySchema, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("unexpected body schema: %s", diff)
			}
		})
	}
}

func TestBodySchema_DependentBodySchema_label_notFound(t *testing.T) {
	block := &hcl.Block{
		Labels: []string{"test", "mycloud"},
		Body:   hcl.EmptyBody(),
	}
	_, _, ok := testSchemaWithLabels.DependentBodySchema(block)
	if ok {
		t.Fatal("expected not to find body schema for 'mycloud' 2nd label")
	}
}

func TestBodySchema_DependentBodySchema_label_storedUnsorted(t *testing.T) {
	block := &hcl.Block{
		Labels: []string{"complexcloud", "pumpkin"},
		Body:   hcl.EmptyBody(),
	}
	bodySchema, _, ok := testSchemaWithLabels.DependentBodySchema(block)
	if !ok {
		t.Fatal("expected to find body schema stored with unsorted keys")
	}
	expectedSchema := &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"event": {Expr: schema.LiteralTypeOnly(cty.String)},
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
	_, _, ok := testSchemaWithLabels.DependentBodySchema(block)
	if ok {
		t.Fatal("expected to not find body schema based on wrongly sorted labels")
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
				Expr: schema.LiteralTypeOnly(cty.String),
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
				"special_attr": {Expr: schema.LiteralTypeOnly(cty.String)},
			},
		},
		schema.NewSchemaKey(schema.DependencyKeys{
			Labels: []schema.LabelDependent{
				{Index: 1, Value: "theircloud"},
			},
		}): {
			Attributes: map[string]*schema.AttributeSchema{
				"foo": {Expr: schema.LiteralTypeOnly(cty.Number)},
			},
		},
		schema.NewSchemaKey(schema.DependencyKeys{
			Labels: []schema.LabelDependent{
				{Index: 0, Value: "theircloud"},
			},
		}): {
			Attributes: map[string]*schema.AttributeSchema{
				"bar": {Expr: schema.LiteralTypeOnly(cty.Number)},
			},
		},
		schema.NewSchemaKey(schema.DependencyKeys{
			Labels: []schema.LabelDependent{
				{Index: 1, Value: "pumpkin"},
				{Index: 0, Value: "complexcloud"},
			},
		}): {
			Attributes: map[string]*schema.AttributeSchema{
				"event": {Expr: schema.LiteralTypeOnly(cty.String)},
			},
		},
		schema.NewSchemaKey(schema.DependencyKeys{
			Labels: []schema.LabelDependent{
				{Index: 0, Value: "crazycloud"},
				{Index: 1, Value: "apple"},
			},
		}): {
			Attributes: map[string]*schema.AttributeSchema{
				"another": {Expr: schema.LiteralTypeOnly(cty.String)},
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
				Expr:     schema.LiteralTypeOnly(cty.String),
				IsDepKey: true,
			},
			"depnum": {
				Expr:     schema.LiteralTypeOnly(cty.Number),
				IsDepKey: true,
			},
			"depref": {
				Expr:     schema.LiteralTypeOnly(cty.DynamicPseudoType),
				IsDepKey: true,
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
				"depval_attr": {Expr: schema.LiteralTypeOnly(cty.String)},
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
				"number_found": {Expr: schema.LiteralTypeOnly(cty.Number)},
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
				"refbar": {Expr: schema.LiteralTypeOnly(cty.Number)},
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
				"sortedattr": {Expr: schema.LiteralTypeOnly(cty.String)},
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
				"unsortedattr": {Expr: schema.LiteralTypeOnly(cty.String)},
			},
		},
	},
})
