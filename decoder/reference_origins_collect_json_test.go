// Copyright IBM Corp. 2020, 2025
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestCollectReferenceOrigins_json(t *testing.T) {
	testCases := []struct {
		name            string
		schema          *schema.BodySchema
		cfg             string
		expectedOrigins reference.Origins
	}{
		{
			"no origins",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attribute": {
						Constraint: schema.LiteralType{Type: cty.String},
					},
				},
			},
			`{"attribute": "foo-bar"}`,
			reference.Origins{},
		},
		{
			"root attribute single step",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Constraint: schema.Reference{},
					},
				},
			},
			`{"attr": "${onestep}"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   1,
							Column: 13,
							Byte:   12,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 20,
							Byte:   19,
						},
					},
					Constraints: reference.OriginConstraints{},
				},
			},
		},
		{
			"multiple root attributes single step",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr1": {
						Constraint: schema.Reference{},
					},
					"attr2": {
						Constraint: schema.Reference{},
					},
					"attr3": {
						Constraint: schema.Reference{},
					},
				},
			},
			`{
  "attr1": "${onestep}",
  "attr2": "${anotherstep}",
  "attr3": "${onestep}"
}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   2,
							Column: 15,
							Byte:   16,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 22,
							Byte:   23,
						},
					},
					Constraints: reference.OriginConstraints{},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "anotherstep"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 15,
							Byte:   41,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 26,
							Byte:   52,
						},
					},
					Constraints: reference.OriginConstraints{},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   4,
							Column: 15,
							Byte:   70,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 22,
							Byte:   77,
						},
					},
					Constraints: reference.OriginConstraints{},
				},
			},
		},
		{
			"root attribute multiple origins",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr1": {
						Constraint: schema.AnyExpression{
							OfType: cty.DynamicPseudoType,
						},
					},
				},
			},
			`{"attr1": "${onestep}-${onestep}-${another.foo.bar}"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   1,
							Column: 14,
							Byte:   13,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 21,
							Byte:   20,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.DynamicPseudoType},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   1,
							Column: 25,
							Byte:   24,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 32,
							Byte:   31,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.DynamicPseudoType},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "another"},
						lang.AttrStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   1,
							Column: 36,
							Byte:   35,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 51,
							Byte:   50,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.DynamicPseudoType},
					},
				},
			},
		},
		{
			"root attribute multi-step",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Constraint: schema.Reference{
							OfType: cty.Number,
						},
					},
				},
			},
			`{"attr": "one.two[\"key\"].attr[0]"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "one"},
						lang.AttrStep{Name: "two"},
						lang.IndexStep{Key: cty.StringVal("key")},
						lang.AttrStep{Name: "attr"},
						lang.IndexStep{Key: cty.NumberIntVal(0)},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   1,
							Column: 11,
							Byte:   10,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 33,
							Byte:   32,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.Number},
					},
				},
			},
		},
		{
			"attribute in block",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"myblock": {
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Constraint: schema.Reference{},
								},
							},
						},
					},
				},
			},
			`{
  "myblock": {
    "attr": "${onestep}"
  }
}
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 16,
							Byte:   32,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 23,
							Byte:   39,
						},
					},
					Constraints: reference.OriginConstraints{},
				},
			},
		},
		{
			"any attribute in block",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"myblock": {
						Body: &schema.BodySchema{
							AnyAttribute: &schema.AttributeSchema{
								Constraint: schema.Reference{},
							},
						},
					},
				},
			},
			`{
  "myblock": {
    "attr": "${onestep}"
  }
}
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 16,
							Byte:   32,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 23,
							Byte:   39,
						},
					},
					Constraints: reference.OriginConstraints{},
				},
			},
		},
		{
			"origins within block with matching dependent body",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"myblock": {
						Labels: []*schema.LabelSchema{
							{Name: "type", IsDepKey: true},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"static": {
									Constraint: schema.Reference{},
								},
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "special"},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"dep_attr": {
										Constraint: schema.Reference{},
									},
								},
							},
						},
					},
				},
			},
			`{
  "myblock": {
  	"special": {
      "static": "${var.first}",
      "dep_attr": "${var.second}"
    }
  }
}
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   4,
							Column: 20,
							Byte:   52,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 29,
							Byte:   61,
						},
					},
					Constraints: reference.OriginConstraints{},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   5,
							Column: 22,
							Byte:   86,
						},
						End: hcl.Pos{
							Line:   5,
							Column: 32,
							Byte:   96,
						},
					},
					Constraints: reference.OriginConstraints{},
				},
			},
		},
		{
			"origins within block with mismatching dependent body",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"myblock": {
						Labels: []*schema.LabelSchema{
							{Name: "type", IsDepKey: true},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"static": {
									Constraint: schema.Reference{},
								},
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "special"},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"dep_attr": {
										Constraint: schema.Reference{},
									},
								},
							},
						},
					},
				},
			},
			`{
  "myblock": {
  	"different": {
      "static": "${var.first}",
      "dep_attr": "${var.second}"
    }
  }
}
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   4,
							Column: 20,
							Byte:   54,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 29,
							Byte:   63,
						},
					},
					Constraints: reference.OriginConstraints{},
				},
			},
		},
		{
			"origin inside collection expressions",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"list": {
						Constraint: schema.List{
							Elem: schema.Reference{
								OfScopeId: lang.ScopeId("test"),
							},
						},
					},
					"set": {
						Constraint: schema.Set{
							Elem: schema.Reference{
								OfScopeId: lang.ScopeId("test"),
							},
						},
					},
					"tuple": {
						Constraint: schema.Tuple{
							Elems: []schema.Constraint{
								schema.Reference{
									OfScopeId: lang.ScopeId("test"),
								},
							},
						},
					},
				},
			},
			`{
  "list": [ "${var.first}" ],
  "set": [ "${var.second}" ],
  "tuple": [ "${var.third}" ]
}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   2,
							Column: 16,
							Byte:   17,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 25,
							Byte:   26,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfScopeId: lang.ScopeId("test")},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 15,
							Byte:   46,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 25,
							Byte:   56,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfScopeId: lang.ScopeId("test")},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "third"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   4,
							Column: 17,
							Byte:   78,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 26,
							Byte:   87,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfScopeId: lang.ScopeId("test")},
					},
				},
			},
		},
		{
			"origin inside object expression",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"obj": {
						Constraint: schema.Object{
							Attributes: schema.ObjectAttributes{
								"attr": &schema.AttributeSchema{
									Constraint: schema.Reference{
										OfScopeId: lang.ScopeId("test"),
									},
								},
							},
						},
					},
				},
			},
			`{
  "obj": {
    "attr": "${var.first}"
  }
}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 16,
							Byte:   28,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 25,
							Byte:   37,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfScopeId: lang.ScopeId("test")},
					},
				},
			},
		},
		{
			"origin inside map expression",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"map": {
						Constraint: schema.Map{
							Elem: schema.Reference{
								OfScopeId: lang.ScopeId("test"),
							},
						},
					},
				},
			},
			`{
  "map": {
    "key": "${var.first}"
  }
}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 15,
							Byte:   27,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 24,
							Byte:   36,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfScopeId: lang.ScopeId("test")},
					},
				},
			},
		},
		{
			"origin inside object and map expression with multiple matches",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"map": {
						Constraint: schema.OneOf{
							schema.Reference{OfType: cty.Map(cty.String)},
							schema.Map{
								Elem: schema.Reference{OfType: cty.String},
							},
						},
					},
					"obj": {
						Constraint: schema.OneOf{
							schema.Reference{OfType: cty.Object(map[string]cty.Type{
								"foo": cty.String,
							})},
							schema.Object{
								Attributes: schema.ObjectAttributes{
									"foo": &schema.AttributeSchema{
										Constraint: schema.Reference{OfType: cty.String},
									},
								},
							},
						},
					},
				},
			},
			`{
  "map": {
    "bar": "${var.one}"
  },
  "obj": {
    "foo": "${var.two}"
  }
}
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "one"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 15,
							Byte:   27,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 22,
							Byte:   34,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "two"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   6,
							Column: 15,
							Byte:   67,
						},
						End: hcl.Pos{
							Line:   6,
							Column: 22,
							Byte:   74,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
			},
		},
		{
			"origin inside list, set and tuple expression with multiple matches",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"list": {
						Constraint: schema.OneOf{
							schema.Reference{OfType: cty.List(cty.String)},
							schema.List{
								Elem: schema.Reference{OfType: cty.String},
							},
						},
					},
					"set": {
						Constraint: schema.OneOf{
							schema.Reference{OfType: cty.Set(cty.String)},
							schema.Set{
								Elem: schema.Reference{OfType: cty.String},
							},
						},
					},
					"tup": {
						Constraint: schema.OneOf{
							schema.Reference{OfType: cty.Tuple([]cty.Type{cty.String})},
							schema.Tuple{
								Elems: []schema.Constraint{
									schema.Reference{OfType: cty.String},
								},
							},
						},
					},
				},
			},
			`{
  "list": [ "${var.one}" ],
  "set": [ "${var.two}" ],
  "tup": [ "${var.three}" ]
}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "one"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   2,
							Column: 16,
							Byte:   17,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 23,
							Byte:   24,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "two"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 15,
							Byte:   44,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 22,
							Byte:   51,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "three"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   4,
							Column: 15,
							Byte:   71,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 24,
							Byte:   80,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%2d-%s", i, tc.name), func(t *testing.T) {
			f, diags := json.Parse([]byte(tc.cfg), "test.tf.json")
			if len(diags) > 0 {
				t.Fatal(diags)
			}

			d := testPathDecoder(t, &PathContext{
				Schema: tc.schema,
				Files: map[string]*hcl.File{
					"test.tf.json": f,
				},
			})

			origins, err := d.CollectReferenceOrigins()
			if err != nil {
				t.Fatal(err)
			}

			opts := cmp.Options{
				ctydebug.CmpOptions,
			}

			if diff := cmp.Diff(tc.expectedOrigins, origins, opts...); diff != "" {
				t.Fatalf("mismatched reference origins: %s", diff)
			}
		})
	}
}
