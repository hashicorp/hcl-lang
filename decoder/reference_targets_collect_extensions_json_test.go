// Copyright (c) HashiCorp, Inc.
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

func TestCollectReferenceTargets_extension_json(t *testing.T) {
	testCases := []struct {
		name         string
		schema       *schema.BodySchema
		cfg          string
		expectedRefs reference.Targets
	}{
		{
			"self references collection - list block",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Address: &schema.BlockAddrSchema{
							Steps: schema.Address{
								schema.LabelStep{Index: 0},
								schema.LabelStep{Index: 1},
							},
							DependentBodyAsData:  true,
							InferDependentBody:   true,
							DependentBodySelfRef: true,
						},
						Labels: []*schema.LabelSchema{
							{
								Name:     "type",
								IsDepKey: true,
							},
							{
								Name: "name",
							},
						},
						Body: schema.NewBodySchema(),
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{
										Index: 0,
										Value: "aws_instance",
									},
								},
							}): {
								Blocks: map[string]*schema.BlockSchema{
									"foo": {
										Type: schema.BlockTypeList,
										Body: &schema.BodySchema{
											Attributes: map[string]*schema.AttributeSchema{
												"bar": {IsOptional: true, Constraint: schema.LiteralType{Type: cty.Number}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			`{
  "resource": {
    "aws_instance": {
      "blah": {
        "foo": {
          "bar": 42
        }
      }
    }
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_instance"},
						lang.AttrStep{Name: "blah"},
					},
					LocalAddr: lang.Address{
						lang.RootStep{Name: "self"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   4,
							Column: 15,
							Byte:   54,
						},
						End: hcl.Pos{
							Line:   8,
							Column: 8,
							Byte:   110,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   4,
							Column: 15,
							Byte:   54,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 16,
							Byte:   55,
						},
					},
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   4,
							Column: 15,
							Byte:   54,
						},
						End: hcl.Pos{
							Line:   8,
							Column: 8,
							Byte:   110,
						},
					},
					Type: cty.Object(map[string]cty.Type{
						"foo": cty.List(cty.Object(map[string]cty.Type{
							"bar": cty.Number,
						})),
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_instance"},
								lang.AttrStep{Name: "blah"},
								lang.AttrStep{Name: "foo"},
							},
							LocalAddr: lang.Address{}, // no LocalAddr for JSON files
							RangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   5,
									Column: 16,
									Byte:   71,
								},
								End: hcl.Pos{
									Line:   7,
									Column: 10,
									Byte:   102,
								},
							},
							Type: cty.List(cty.Object(map[string]cty.Type{
								"bar": cty.Number,
							})),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws_instance"},
										lang.AttrStep{Name: "blah"},
										lang.AttrStep{Name: "foo"},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
									},
									LocalAddr: lang.Address{}, // no LocalAddr for JSON files
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start: hcl.Pos{
											Line:   5,
											Column: 16,
											Byte:   71,
										},
										End: hcl.Pos{
											Line:   7,
											Column: 10,
											Byte:   102,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start: hcl.Pos{
											Line:   5,
											Column: 16,
											Byte:   71,
										},
										End: hcl.Pos{
											Line:   5,
											Column: 17,
											Byte:   72,
										},
									},
									Type: cty.Object(map[string]cty.Type{
										"bar": cty.Number,
									}),
									NestedTargets: reference.Targets{
										{
											Addr: lang.Address{
												lang.RootStep{Name: "aws_instance"},
												lang.AttrStep{Name: "blah"},
												lang.AttrStep{Name: "foo"},
												lang.IndexStep{Key: cty.NumberIntVal(0)},
												lang.AttrStep{Name: "bar"},
											},
											LocalAddr: nil, // no LocalAddr for JSON files
											RangePtr: &hcl.Range{
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   6,
													Column: 11,
													Byte:   83,
												},
												End: hcl.Pos{
													Line:   6,
													Column: 20,
													Byte:   92,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   6,
													Column: 11,
													Byte:   83,
												},
												End: hcl.Pos{
													Line:   6,
													Column: 16,
													Byte:   88,
												},
											},
											Type: cty.Number,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
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

			refs, err := d.CollectReferenceTargets()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedRefs, refs, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of references: %s", diff)
			}
		})
	}
}
