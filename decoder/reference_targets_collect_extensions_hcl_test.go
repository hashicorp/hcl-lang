package decoder

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestCollectReferenceTargets_extension_hcl(t *testing.T) {
	testCases := []struct {
		name         string
		schema       *schema.BodySchema
		cfg          string
		expectedRefs reference.Targets
	}{
		{
			"self references",
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
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"static": {IsOptional: true, Expr: schema.LiteralTypeOnly(cty.String)},
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{
										Index: 0,
										Value: "aws_instance",
									},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"bar": {IsOptional: true, Expr: schema.LiteralTypeOnly(cty.Number)},
									"foo": {IsOptional: true, Expr: schema.LiteralTypeOnly(cty.String)},
								},
							},
						},
					},
				},
			},
			`resource "aws_instance" "blah" {
  static = "test"
  foo = "test"
  bar = 42
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_instance"},
						lang.AttrStep{Name: "blah"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   5,
							Column: 2,
							Byte:   78,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 31,
							Byte:   30,
						},
					},
					Type: cty.Object(map[string]cty.Type{
						"bar": cty.Number,
						"foo": cty.String,
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_instance"},
								lang.AttrStep{Name: "blah"},
								lang.AttrStep{Name: "bar"},
							},
							LocalAddr: lang.Address{
								lang.RootStep{Name: "self"},
								lang.AttrStep{Name: "bar"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   4,
									Column: 3,
									Byte:   68,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 11,
									Byte:   76,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   4,
									Column: 3,
									Byte:   68,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 6,
									Byte:   71,
								},
							},
							Type: cty.Number,
							TargetableFromRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 33,
									Byte:   32,
								},
								End: hcl.Pos{
									Line:   5,
									Column: 1,
									Byte:   77,
								},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_instance"},
								lang.AttrStep{Name: "blah"},
								lang.AttrStep{Name: "foo"},
							},
							LocalAddr: lang.Address{
								lang.RootStep{Name: "self"},
								lang.AttrStep{Name: "foo"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 3,
									Byte:   53,
								},
								End: hcl.Pos{
									Line:   3,
									Column: 15,
									Byte:   65,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 3,
									Byte:   53,
								},
								End: hcl.Pos{
									Line:   3,
									Column: 6,
									Byte:   56,
								},
							},
							Type: cty.String,
							TargetableFromRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 33,
									Byte:   32,
								},
								End: hcl.Pos{
									Line:   5,
									Column: 1,
									Byte:   77,
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
			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)

			d := testPathDecoder(t, &PathContext{
				Schema: tc.schema,
				Files: map[string]*hcl.File{
					"test.tf": f,
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
