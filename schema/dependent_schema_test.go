package schema

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestBodySchema_FindSchemaDependingOn_label_basic(t *testing.T) {
	dks := DependencyKeys{
		Labels: []LabelDependent{
			{Index: 0, Value: "theircloud"},
		},
	}
	bodySchema, ok := testSchemaWithLabels.DependentBodySchema(dks)
	if !ok {
		t.Fatal("expected to find body schema for 'theircloud' label")
	}
	expectedSchema := &BodySchema{
		Attributes: map[string]*AttributeSchema{
			"bar": {Expr: LiteralTypeOnly(cty.Number)},
		},
	}
	if diff := cmp.Diff(expectedSchema, bodySchema, ctydebug.CmpOptions); diff != "" {
		t.Fatalf("unexpected body schema: %s", diff)
	}
}

func TestBodySchema_FindSchemaDependingOn_attributes(t *testing.T) {
	testCases := []struct {
		name           string
		attrKeys       []AttributeDependent
		expectedSchema *BodySchema
	}{
		{
			"single string attribute",
			[]AttributeDependent{
				{
					Name: "depattr",
					Expr: ExpressionValue{
						Static: cty.StringVal("dep-val"),
					},
				},
			},
			&BodySchema{
				Attributes: map[string]*AttributeSchema{
					"depval_attr": {Expr: LiteralTypeOnly(cty.String)},
				},
			},
		},
		{
			"single numeric attribute",
			[]AttributeDependent{
				{
					Name: "depnum",
					Expr: ExpressionValue{
						Static: cty.NumberIntVal(42),
					},
				},
			},
			&BodySchema{
				Attributes: map[string]*AttributeSchema{
					"number_found": {Expr: LiteralTypeOnly(cty.Number)},
				},
			},
		},
		{
			"reference attribute",
			[]AttributeDependent{
				{
					Name: "depref",
					Expr: ExpressionValue{
						Reference: lang.Reference{
							lang.RootStep{Name: "myroot"},
							lang.AttrStep{Name: "attrstep"},
						},
					},
				},
			},
			&BodySchema{
				Attributes: map[string]*AttributeSchema{
					"refbar": {Expr: LiteralTypeOnly(cty.Number)},
				},
			},
		},
		{
			"two attributes sorted",
			[]AttributeDependent{
				{
					Name: "depattr",
					Expr: ExpressionValue{
						Static: cty.StringVal("pumpkin"),
					},
				},
				{
					Name: "depnum",
					Expr: ExpressionValue{
						Static: cty.NumberIntVal(55),
					},
				},
			},
			&BodySchema{
				Attributes: map[string]*AttributeSchema{
					"sortedattr": {Expr: LiteralTypeOnly(cty.String)},
				},
			},
		},
		{
			"two attributes unsorted",
			[]AttributeDependent{
				{
					Name: "depattr",
					Expr: ExpressionValue{
						Static: cty.StringVal("pumpkin"),
					},
				},
				{
					Name: "depnum",
					Expr: ExpressionValue{
						Static: cty.NumberIntVal(2),
					},
				},
			},
			&BodySchema{
				Attributes: map[string]*AttributeSchema{
					"unsortedattr": {Expr: LiteralTypeOnly(cty.String)},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			dks := DependencyKeys{
				Attributes: tc.attrKeys,
			}
			bodySchema, ok := testSchemaWithAttributes.DependentBodySchema(dks)
			if !ok {
				t.Fatal("expected to find body schema for 'depattr' attribute")
			}
			if diff := cmp.Diff(tc.expectedSchema, bodySchema, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("unexpected body schema: %s", diff)
			}
		})
	}
}

func TestBodySchema_FindSchemaDependingOn_label_notFound(t *testing.T) {
	dks := DependencyKeys{
		Labels: []LabelDependent{
			{Index: 1, Value: "mycloud"},
		},
	}
	_, ok := testSchemaWithLabels.DependentBodySchema(dks)
	if ok {
		t.Fatal("expected not to find body schema for 'mycloud' 2nd label")
	}
}

func TestBodySchema_FindSchemaDependingOn_label_storedUnsorted(t *testing.T) {
	dks := DependencyKeys{
		Labels: []LabelDependent{
			{Index: 0, Value: "complexcloud"},
			{Index: 1, Value: "pumpkin"},
		},
	}
	bodySchema, ok := testSchemaWithLabels.DependentBodySchema(dks)
	if !ok {
		t.Fatal("expected to find body schema stored with unsorted keys")
	}
	expectedSchema := &BodySchema{
		Attributes: map[string]*AttributeSchema{
			"event": {Expr: LiteralTypeOnly(cty.String)},
		},
	}
	if diff := cmp.Diff(expectedSchema, bodySchema, ctydebug.CmpOptions); diff != "" {
		t.Fatalf("unexpected body schema: %s", diff)
	}
}

func TestBodySchema_FindSchemaDependingOn_label_lookupUnsorted(t *testing.T) {
	dks := DependencyKeys{
		Labels: []LabelDependent{
			{Index: 1, Value: "apple"},
			{Index: 0, Value: "crazycloud"},
		},
	}
	bodySchema, ok := testSchemaWithLabels.DependentBodySchema(dks)
	if !ok {
		t.Fatal("expected to find body schema based on unsorted keys")
	}
	expectedSchema := &BodySchema{
		Attributes: map[string]*AttributeSchema{
			"another": {Expr: LiteralTypeOnly(cty.String)},
		},
	}
	if diff := cmp.Diff(expectedSchema, bodySchema, ctydebug.CmpOptions); diff != "" {
		t.Fatalf("unexpected body schema: %s", diff)
	}
}

var testSchemaWithLabels = &BlockSchema{
	Labels: []*LabelSchema{
		{
			Name:     "type",
			IsDepKey: true,
		},
		{
			Name:     "name",
			IsDepKey: true,
		},
	},
	Body: &BodySchema{
		Attributes: map[string]*AttributeSchema{
			"alias": {
				Expr: LiteralTypeOnly(cty.String),
			},
		},
	},
	DependentBody: map[SchemaKey]*BodySchema{
		NewSchemaKey(DependencyKeys{
			Labels: []LabelDependent{
				{Index: 0, Value: "mycloud"},
			},
		}): {
			Attributes: map[string]*AttributeSchema{
				"special_attr": {Expr: LiteralTypeOnly(cty.String)},
			},
		},
		NewSchemaKey(DependencyKeys{
			Labels: []LabelDependent{
				{Index: 1, Value: "theircloud"},
			},
		}): {
			Attributes: map[string]*AttributeSchema{
				"foo": {Expr: LiteralTypeOnly(cty.Number)},
			},
		},
		NewSchemaKey(DependencyKeys{
			Labels: []LabelDependent{
				{Index: 0, Value: "theircloud"},
			},
		}): {
			Attributes: map[string]*AttributeSchema{
				"bar": {Expr: LiteralTypeOnly(cty.Number)},
			},
		},
		NewSchemaKey(DependencyKeys{
			Labels: []LabelDependent{
				{Index: 1, Value: "pumpkin"},
				{Index: 0, Value: "complexcloud"},
			},
		}): {
			Attributes: map[string]*AttributeSchema{
				"event": {Expr: LiteralTypeOnly(cty.String)},
			},
		},
		NewSchemaKey(DependencyKeys{
			Labels: []LabelDependent{
				{Index: 0, Value: "crazycloud"},
				{Index: 1, Value: "apple"},
			},
		}): {
			Attributes: map[string]*AttributeSchema{
				"another": {Expr: LiteralTypeOnly(cty.String)},
			},
		},
	},
}

var testSchemaWithAttributes = &BlockSchema{
	Labels: []*LabelSchema{
		{
			Name: "type",
		},
		{
			Name: "name",
		},
	},
	Body: &BodySchema{
		Attributes: map[string]*AttributeSchema{
			"depattr": {
				Expr:     LiteralTypeOnly(cty.String),
				IsDepKey: true,
			},
			"depnum": {
				Expr:     LiteralTypeOnly(cty.Number),
				IsDepKey: true,
			},
			"depref": {
				Expr:     LiteralTypeOnly(cty.DynamicPseudoType),
				IsDepKey: true,
			},
		},
	},
	DependentBody: map[SchemaKey]*BodySchema{
		NewSchemaKey(DependencyKeys{
			Attributes: []AttributeDependent{
				{
					Name: "depattr",
					Expr: ExpressionValue{
						Static: cty.StringVal("dep-val"),
					},
				},
			},
		}): {
			Attributes: map[string]*AttributeSchema{
				"depval_attr": {Expr: LiteralTypeOnly(cty.String)},
			},
		},
		NewSchemaKey(DependencyKeys{
			Attributes: []AttributeDependent{
				{
					Name: "depnum",
					Expr: ExpressionValue{
						Static: cty.NumberIntVal(42),
					},
				},
			},
		}): {
			Attributes: map[string]*AttributeSchema{
				"number_found": {Expr: LiteralTypeOnly(cty.Number)},
			},
		},
		NewSchemaKey(DependencyKeys{
			Attributes: []AttributeDependent{
				{
					Name: "depref",
					Expr: ExpressionValue{
						Reference: lang.Reference{
							lang.RootStep{Name: "myroot"},
							lang.AttrStep{Name: "attrstep"},
						},
					},
				},
			},
		}): {
			Attributes: map[string]*AttributeSchema{
				"refbar": {Expr: LiteralTypeOnly(cty.Number)},
			},
		},
		NewSchemaKey(DependencyKeys{
			Attributes: []AttributeDependent{
				{
					Name: "depattr",
					Expr: ExpressionValue{
						Static: cty.StringVal("pumpkin"),
					},
				},
				{
					Name: "depnum",
					Expr: ExpressionValue{
						Static: cty.NumberIntVal(55),
					},
				},
			},
		}): {
			Attributes: map[string]*AttributeSchema{
				"sortedattr": {Expr: LiteralTypeOnly(cty.String)},
			},
		},
		NewSchemaKey(DependencyKeys{
			Attributes: []AttributeDependent{
				{
					Name: "depnum",
					Expr: ExpressionValue{
						Static: cty.NumberIntVal(2),
					},
				},
				{
					Name: "depattr",
					Expr: ExpressionValue{
						Static: cty.StringVal("pumpkin"),
					},
				},
			},
		}): {
			Attributes: map[string]*AttributeSchema{
				"unsortedattr": {Expr: LiteralTypeOnly(cty.String)},
			},
		},
	},
}
