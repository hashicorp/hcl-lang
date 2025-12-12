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
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestCollectRefTargets_exprObject_hcl(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefTargets reference.Targets
	}{
		{
			"constraint mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.Keyword{
									Keyword: "keyword",
								},
								IsOptional: true,
							},
						},
					},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						AsExprType: true,
					},
				},
			},
			`attr = keyword`,
			reference.Targets{},
		},
		{
			"no collectable constraint",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.Keyword{
									Keyword: "keyword",
								},
								IsOptional: true,
							},
						},
					},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						AsExprType: true,
					},
				},
			},
			`attr = { foo = keyword }`,
			reference.Targets{},
		},
		{
			"addressable reference only",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.Reference{
									Address: &schema.ReferenceAddrSchema{
										ScopeId: lang.ScopeId("test"),
									},
								},
								IsOptional: true,
							},
						},
					},
					IsOptional: true,
				},
			},
			`attr = {
  foo = foo
}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					ScopeId: lang.ScopeId("test"),
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 2, Column: 9, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 12, Byte: 20},
					},
				},
			},
		},
		{
			"empty type-aware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
								IsOptional: true,
							},
							"bar": {
								Constraint: schema.LiteralType{
									Type: cty.Number,
								},
								IsRequired: true,
							},
						},
					},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						AsExprType: true,
					},
				},
			},
			`attr = {}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"foo": cty.String,
						"bar": cty.Number,
					}, []string{"foo"}),
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.AttrStep{Name: "bar"},
							},
							Type: cty.Number,
							RangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
								End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.AttrStep{Name: "foo"},
							},
							Type: cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
								End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
							},
						},
					},
				},
			},
		},
		{
			"type-aware with invalid key type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
								IsOptional: true,
							},
							"bar": {
								Constraint: schema.LiteralType{
									Type: cty.Number,
								},
								IsRequired: true,
							},
						},
					},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						AsExprType: true,
					},
				},
			},
			`attr = {
  422 = "foo"
  bar = 42
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"foo": cty.String,
						"bar": cty.Number,
					}, []string{"foo"}),
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 4, Column: 2, Byte: 35},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.AttrStep{Name: "bar"},
							},
							Type: cty.Number,
							RangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 3, Column: 3, Byte: 25},
								End:      hcl.Pos{Line: 3, Column: 11, Byte: 33},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 3, Column: 3, Byte: 25},
								End:      hcl.Pos{Line: 3, Column: 6, Byte: 28},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.AttrStep{Name: "foo"},
							},
							Type: cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
								End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
							},
						},
					},
				},
			},
		},
		{
			"type-aware with invalid attribute name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
								IsOptional: true,
							},
							"bar": {
								Constraint: schema.LiteralType{
									Type: cty.Number,
								},
								IsRequired: true,
							},
						},
					},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						AsExprType: true,
					},
				},
			},
			`attr = {
  fox = "foo"
  bar = 42
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"foo": cty.String,
						"bar": cty.Number,
					}, []string{"foo"}),
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 4, Column: 2, Byte: 35},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.AttrStep{Name: "bar"},
							},
							Type: cty.Number,
							RangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 3, Column: 3, Byte: 25},
								End:      hcl.Pos{Line: 3, Column: 11, Byte: 33},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 3, Column: 3, Byte: 25},
								End:      hcl.Pos{Line: 3, Column: 6, Byte: 28},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.AttrStep{Name: "foo"},
							},
							Type: cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
								End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
							},
						},
					},
				},
			},
		},
		{
			"type-aware with invalid value type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
								IsOptional: true,
							},
							"bar": {
								Constraint: schema.LiteralType{
									Type: cty.Number,
								},
								IsRequired: true,
							},
						},
					},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						AsExprType: true,
					},
				},
			},
			`attr = {
  foo = 12345
  bar = 42
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"foo": cty.String,
						"bar": cty.Number,
					}, []string{"foo"}),
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 4, Column: 2, Byte: 35},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.AttrStep{Name: "bar"},
							},
							Type: cty.Number,
							RangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 3, Column: 3, Byte: 25},
								End:      hcl.Pos{Line: 3, Column: 11, Byte: 33},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 3, Column: 3, Byte: 25},
								End:      hcl.Pos{Line: 3, Column: 6, Byte: 28},
							},
						},
					},
				},
			},
		},
		{
			"type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
								IsOptional: true,
							},
							"bar": {
								Constraint: schema.LiteralType{
									Type: cty.Number,
								},
								IsRequired: true,
							},
						},
					},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						ScopeId:     lang.ScopeId("test"),
						AsReference: true,
					},
				},
			},
			`attr = {
  foo = "foo"
  bar = 42
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					ScopeId: lang.ScopeId("test"),
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 4, Column: 2, Byte: 35},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
					NestedTargets: reference.Targets{},
				},
			},
		},
		{
			"nested type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
								IsOptional: true,
							},
							"bar": {
								Constraint: schema.Object{
									Attributes: schema.ObjectAttributes{
										"baz": {
											Constraint: schema.LiteralType{
												Type: cty.String,
											},
											IsRequired: true,
										},
									},
								},
								IsRequired: true,
							},
						},
					},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						AsExprType: true,
					},
				},
			},
			`attr = {
  foo = "foo"
  bar = {
    baz = "noot"
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"foo": cty.String,
						"bar": cty.Object(map[string]cty.Type{
							"baz": cty.String,
						}),
					}, []string{"foo"}),
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 6, Column: 2, Byte: 55},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.AttrStep{Name: "bar"},
							},
							Type: cty.Object(map[string]cty.Type{
								"baz": cty.String,
							}),
							RangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 3, Column: 3, Byte: 25},
								End:      hcl.Pos{Line: 5, Column: 4, Byte: 53},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 3, Column: 3, Byte: 25},
								End:      hcl.Pos{Line: 3, Column: 6, Byte: 28},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "attr"},
										lang.AttrStep{Name: "bar"},
										lang.AttrStep{Name: "baz"},
									},
									Type: cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.hcl",
										Start:    hcl.Pos{Line: 4, Column: 5, Byte: 37},
										End:      hcl.Pos{Line: 4, Column: 17, Byte: 49},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.hcl",
										Start:    hcl.Pos{Line: 4, Column: 5, Byte: 37},
										End:      hcl.Pos{Line: 4, Column: 8, Byte: 40},
									},
								},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.AttrStep{Name: "foo"},
							},
							Type: cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
								End:      hcl.Pos{Line: 2, Column: 14, Byte: 22},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
								End:      hcl.Pos{Line: 2, Column: 6, Byte: 14},
							},
						},
					},
				},
			},
		},
		{
			"object from for-expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
								IsOptional: true,
							},
						},
					},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						ScopeId:    lang.ScopeId("test"),
						AsExprType: true,
					},
				},
			},
			`attr = { for s in ["a", "b"] : s => "s${s}" }`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					ScopeId: lang.ScopeId("test"),
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 46, Byte: 45},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
					Type: cty.DynamicPseudoType,
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			bodySchema := &schema.BodySchema{
				Attributes: tc.attrSchema,
			}

			f, diags := hclsyntax.ParseConfig([]byte(tc.cfg), "test.hcl", hcl.InitialPos)
			if len(diags) > 0 {
				t.Error(diags)
			}
			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.hcl": f,
				},
			})

			targets, err := d.CollectReferenceTargets()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedRefTargets, targets, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("unexpected targets: %s", diff)
			}
		})
	}
}

func TestCollectRefTargets_exprObject_implied_hcl(t *testing.T) {
	testCases := []struct {
		testName           string
		bodySchema         *schema.BodySchema
		cfg                string
		expectedRefTargets reference.Targets
	}{
		{
			"undeclared implied as type",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"blk": {
						Address: &schema.BlockAddrSchema{
							Steps: schema.Address{
								schema.StaticStep{Name: "blk"},
							},
							BodyAsData: true,
							InferBody:  true,
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Constraint: schema.Object{
										Attributes: map[string]*schema.AttributeSchema{
											"foo": {
												Constraint: schema.LiteralType{Type: cty.Bool},
											},
										},
									},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`blk {}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "blk"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 7, Byte: 6},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
					RootBlockRange: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.Object(map[string]cty.Type{
							"foo": cty.Bool,
						}),
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "blk"},
								lang.AttrStep{Name: "attr"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
							},
							RootBlockRange: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
								End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
							},
							Type: cty.Object(map[string]cty.Type{
								"foo": cty.Bool,
							}),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "blk"},
										lang.AttrStep{Name: "attr"},
										lang.AttrStep{Name: "foo"},
									},
									RangePtr: &hcl.Range{
										Filename: "test.hcl",
										Start:    hcl.Pos{Line: 1, Column: 5, Byte: 4},
										End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
									},
									RootBlockRange: &hcl.Range{
										Filename: "test.hcl",
										Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
										End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
									},
									Type: cty.Bool,
								},
							},
						},
					},
				},
			},
		},
		{
			"declared as type",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"blk": {
						Address: &schema.BlockAddrSchema{
							Steps: schema.Address{
								schema.StaticStep{Name: "blk"},
							},
							BodyAsData: true,
							InferBody:  true,
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Constraint: schema.Object{
										Attributes: schema.ObjectAttributes{
											"foo": {Constraint: schema.LiteralType{Type: cty.Bool}},
										},
									},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`blk { attr = {} }`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "blk"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 18, Byte: 17},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
					RootBlockRange: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.Object(map[string]cty.Type{
							"foo": cty.Bool,
						}),
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "blk"},
								lang.AttrStep{Name: "attr"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
							},
							RootBlockRange: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
								End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
							},
							Type: cty.Object(map[string]cty.Type{
								"foo": cty.Bool,
							}),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "blk"},
										lang.AttrStep{Name: "attr"},
										lang.AttrStep{Name: "foo"},
									},
									RangePtr: &hcl.Range{
										Filename: "test.hcl",
										Start:    hcl.Pos{Line: 1, Column: 14, Byte: 13},
										End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
									},
									RootBlockRange: &hcl.Range{
										Filename: "test.hcl",
										Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
										End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
									},
									Type: cty.Bool,
								},
							},
						},
					},
				},
			},
		},
		{
			"undeclared as reference",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"blk": {
						Address: &schema.BlockAddrSchema{
							Steps: schema.Address{
								schema.StaticStep{Name: "blk"},
							},
							AsReference: true,
							ScopeId:     lang.ScopeId("foo"),
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Constraint: schema.Object{
										Attributes: map[string]*schema.AttributeSchema{
											"foo": {
												Constraint: schema.LiteralType{Type: cty.Bool},
											},
										},
									},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`blk {}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "blk"},
					},
					ScopeId: lang.ScopeId("foo"),
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 7, Byte: 6},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
					RootBlockRange: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			bodySchema := tc.bodySchema

			f, diags := hclsyntax.ParseConfig([]byte(tc.cfg), "test.hcl", hcl.InitialPos)
			if len(diags) > 0 {
				t.Log(diags)
			}
			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.hcl": f,
				},
			})

			targets, err := d.CollectReferenceTargets()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedRefTargets, targets, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("unexpected targets: %s", diff)
			}
		})
	}
}

func TestCollectRefTargets_exprObject_json(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefTargets reference.Targets
	}{
		{
			"constraint mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.Keyword{
									Keyword: "keyword",
								},
								IsOptional: true,
							},
						},
					},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						AsExprType: true,
					},
				},
			},
			`{"attr": true}`,
			reference.Targets{},
		},
		{
			"no collectable constraint",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.Keyword{
									Keyword: "keyword",
								},
								IsOptional: true,
							},
						},
					},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						AsExprType: true,
					},
				},
			},
			`{"attr": { "foo": "keyword" }}`,
			reference.Targets{},
		},
		{
			"addressable reference only",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.Reference{
									Address: &schema.ReferenceAddrSchema{
										ScopeId: lang.ScopeId("test"),
									},
								},
								IsOptional: true,
							},
						},
					},
					IsOptional: true,
				},
			},
			`{"attr": {
  "foo": "foo"
}}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					ScopeId: lang.ScopeId("test"),
					RangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 2, Column: 11, Byte: 21},
						End:      hcl.Pos{Line: 2, Column: 14, Byte: 24},
					},
				},
			},
		},
		{
			"empty type-aware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
								IsOptional: true,
							},
							"bar": {
								Constraint: schema.LiteralType{
									Type: cty.Number,
								},
								IsRequired: true,
							},
						},
					},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						AsExprType: true,
					},
				},
			},
			`{"attr": {}}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"foo": cty.String,
						"bar": cty.Number,
					}, []string{"foo"}),
					RangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.AttrStep{Name: "bar"},
							},
							Type: cty.Number,
							RangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
								End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.AttrStep{Name: "foo"},
							},
							Type: cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
								End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
							},
						},
					},
				},
			},
		},
		{
			"type-aware with invalid key type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
								IsOptional: true,
							},
							"bar": {
								Constraint: schema.LiteralType{
									Type: cty.Number,
								},
								IsRequired: true,
							},
						},
					},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						AsExprType: true,
					},
				},
			},
			`{"attr": {
  "422": "foo",
  "bar": 42
}}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"foo": cty.String,
						"bar": cty.Number,
					}, []string{"foo"}),
					RangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 4, Column: 2, Byte: 40},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.AttrStep{Name: "bar"},
							},
							Type: cty.Number,
							RangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 3, Column: 3, Byte: 29},
								End:      hcl.Pos{Line: 3, Column: 12, Byte: 38},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 3, Column: 3, Byte: 29},
								End:      hcl.Pos{Line: 3, Column: 8, Byte: 34},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.AttrStep{Name: "foo"},
							},
							Type: cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
								End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
							},
						},
					},
				},
			},
		},
		{
			"type-aware with invalid attribute name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
								IsOptional: true,
							},
							"bar": {
								Constraint: schema.LiteralType{
									Type: cty.Number,
								},
								IsRequired: true,
							},
						},
					},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						AsExprType: true,
					},
				},
			},
			`{"attr": {
  "fox": "foo",
  "bar": 42
}}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"foo": cty.String,
						"bar": cty.Number,
					}, []string{"foo"}),
					RangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 4, Column: 2, Byte: 40},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.AttrStep{Name: "bar"},
							},
							Type: cty.Number,
							RangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 3, Column: 3, Byte: 29},
								End:      hcl.Pos{Line: 3, Column: 12, Byte: 38},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 3, Column: 3, Byte: 29},
								End:      hcl.Pos{Line: 3, Column: 8, Byte: 34},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.AttrStep{Name: "foo"},
							},
							Type: cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
								End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
							},
						},
					},
				},
			},
		},
		{
			"type-aware with invalid value type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
								IsOptional: true,
							},
							"bar": {
								Constraint: schema.LiteralType{
									Type: cty.Number,
								},
								IsRequired: true,
							},
						},
					},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						AsExprType: true,
					},
				},
			},
			`{"attr": {
  "foo": 12345,
  "bar": 42
}}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"foo": cty.String,
						"bar": cty.Number,
					}, []string{"foo"}),
					RangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 4, Column: 2, Byte: 40},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.AttrStep{Name: "bar"},
							},
							Type: cty.Number,
							RangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 3, Column: 3, Byte: 29},
								End:      hcl.Pos{Line: 3, Column: 12, Byte: 38},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 3, Column: 3, Byte: 29},
								End:      hcl.Pos{Line: 3, Column: 8, Byte: 34},
							},
						},
					},
				},
			},
		},
		{
			"type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
								IsOptional: true,
							},
							"bar": {
								Constraint: schema.LiteralType{
									Type: cty.Number,
								},
								IsRequired: true,
							},
						},
					},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						ScopeId:     lang.ScopeId("test"),
						AsReference: true,
					},
				},
			},
			`{"attr": {
  "foo": "foo",
  "bar": 42
}}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					ScopeId: lang.ScopeId("test"),
					RangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 4, Column: 2, Byte: 40},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
					},
					NestedTargets: reference.Targets{},
				},
			},
		},
		{
			"nested type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
								IsOptional: true,
							},
							"bar": {
								Constraint: schema.Object{
									Attributes: schema.ObjectAttributes{
										"baz": {
											Constraint: schema.LiteralType{
												Type: cty.String,
											},
											IsRequired: true,
										},
									},
								},
								IsRequired: true,
							},
						},
					},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						AsExprType: true,
					},
				},
			},
			`{"attr": {
  "foo": "foo",
  "bar": {
    "baz": "noot"
  }
}}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"foo": cty.String,
						"bar": cty.Object(map[string]cty.Type{
							"baz": cty.String,
						}),
					}, []string{"foo"}),
					RangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 6, Column: 2, Byte: 61},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.AttrStep{Name: "bar"},
							},
							Type: cty.Object(map[string]cty.Type{
								"baz": cty.String,
							}),
							RangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 3, Column: 3, Byte: 29},
								End:      hcl.Pos{Line: 5, Column: 4, Byte: 59},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 3, Column: 3, Byte: 29},
								End:      hcl.Pos{Line: 3, Column: 8, Byte: 34},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "attr"},
										lang.AttrStep{Name: "bar"},
										lang.AttrStep{Name: "baz"},
									},
									Type: cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.hcl.json",
										Start:    hcl.Pos{Line: 4, Column: 5, Byte: 42},
										End:      hcl.Pos{Line: 4, Column: 18, Byte: 55},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.hcl.json",
										Start:    hcl.Pos{Line: 4, Column: 5, Byte: 42},
										End:      hcl.Pos{Line: 4, Column: 10, Byte: 47},
									},
								},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.AttrStep{Name: "foo"},
							},
							Type: cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 2, Column: 3, Byte: 13},
								End:      hcl.Pos{Line: 2, Column: 15, Byte: 25},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 2, Column: 3, Byte: 13},
								End:      hcl.Pos{Line: 2, Column: 8, Byte: 18},
							},
						},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			bodySchema := &schema.BodySchema{
				Attributes: tc.attrSchema,
			}

			f, diags := json.ParseWithStartPos([]byte(tc.cfg), "test.hcl.json", hcl.InitialPos)
			if len(diags) > 0 {
				t.Error(diags)
			}
			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.hcl.json": f,
				},
			})

			targets, err := d.CollectReferenceTargets()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedRefTargets, targets, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("unexpected targets: %s", diff)
			}
		})
	}
}

func TestCollectRefTargets_exprObject_implied_json(t *testing.T) {
	testCases := []struct {
		testName           string
		bodySchema         *schema.BodySchema
		cfg                string
		expectedRefTargets reference.Targets
	}{
		{
			"undeclared implied as type",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"blk": {
						Address: &schema.BlockAddrSchema{
							Steps: schema.Address{
								schema.StaticStep{Name: "blk"},
							},
							BodyAsData: true,
							InferBody:  true,
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Constraint: schema.Object{
										Attributes: map[string]*schema.AttributeSchema{
											"foo": {
												Constraint: schema.LiteralType{Type: cty.Bool},
											},
										},
									},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`{"blk": {}}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "blk"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
					},
					RootBlockRange: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
					},
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.Object(map[string]cty.Type{
							"foo": cty.Bool,
						}),
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "blk"},
								lang.AttrStep{Name: "attr"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
								End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
							},
							RootBlockRange: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
								End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
							},
							Type: cty.Object(map[string]cty.Type{
								"foo": cty.Bool,
							}),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "blk"},
										lang.AttrStep{Name: "attr"},
										lang.AttrStep{Name: "foo"},
									},
									RangePtr: &hcl.Range{
										Filename: "test.hcl.json",
										Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
										End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
									},
									RootBlockRange: &hcl.Range{
										Filename: "test.hcl.json",
										Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
										End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
									},
									Type: cty.Bool,
								},
							},
						},
					},
				},
			},
		},
		{
			"declared as type",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"blk": {
						Address: &schema.BlockAddrSchema{
							Steps: schema.Address{
								schema.StaticStep{Name: "blk"},
							},
							BodyAsData: true,
							InferBody:  true,
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Constraint: schema.Object{
										Attributes: map[string]*schema.AttributeSchema{
											"foo": {
												Constraint: schema.LiteralType{Type: cty.Bool},
											},
										},
									},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`{"blk": {"attr": {}}}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "blk"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 21, Byte: 20},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
					},
					RootBlockRange: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
					},
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.Object(map[string]cty.Type{
							"foo": cty.Bool,
						}),
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "blk"},
								lang.AttrStep{Name: "attr"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
								End:      hcl.Pos{Line: 1, Column: 20, Byte: 19},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
								End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
							},
							RootBlockRange: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
								End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
							},
							Type: cty.Object(map[string]cty.Type{
								"foo": cty.Bool,
							}),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "blk"},
										lang.AttrStep{Name: "attr"},
										lang.AttrStep{Name: "foo"},
									},
									RangePtr: &hcl.Range{
										Filename: "test.hcl.json",
										Start:    hcl.Pos{Line: 1, Column: 18, Byte: 17},
										End:      hcl.Pos{Line: 1, Column: 18, Byte: 17},
									},
									RootBlockRange: &hcl.Range{
										Filename: "test.hcl.json",
										Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
										End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
									},
									Type: cty.Bool,
								},
							},
						},
					},
				},
			},
		},
		{
			"undeclared as reference",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"blk": {
						Address: &schema.BlockAddrSchema{
							Steps: schema.Address{
								schema.StaticStep{Name: "blk"},
							},
							AsReference: true,
							ScopeId:     lang.ScopeId("foo"),
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Constraint: schema.Object{
										Attributes: map[string]*schema.AttributeSchema{
											"foo": {
												Constraint: schema.LiteralType{Type: cty.Bool},
											},
										},
									},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`{"blk": {}}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "blk"},
					},
					ScopeId: lang.ScopeId("foo"),
					RangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
					},
					RootBlockRange: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			bodySchema := tc.bodySchema

			f, diags := json.ParseWithStartPos([]byte(tc.cfg), "test.hcl.json", hcl.InitialPos)
			if len(diags) > 0 {
				t.Error(diags)
			}
			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.hcl.json": f,
				},
			})

			targets, err := d.CollectReferenceTargets()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedRefTargets, targets, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("unexpected targets: %s", diff)
			}
		})
	}
}
