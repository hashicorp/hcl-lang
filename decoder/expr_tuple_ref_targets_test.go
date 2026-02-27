// Copyright IBM Corp. 2026
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

func TestCollectRefTargets_exprTuple_hcl(t *testing.T) {
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
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.LiteralType{Type: cty.Bool},
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
			`attr = true`,
			reference.Targets{},
		},
		{
			"tuple of keyword",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{Keyword: "foo"},
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
			`attr = [foo]`,
			reference.Targets{},
		},
		{
			"tuple of addressable reference",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Reference{
								Address: &schema.ReferenceAddrSchema{
									ScopeId: lang.ScopeId("test"),
								},
							},
						},
					},
					IsOptional: true,
				},
			},
			`attr = [foo]`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					ScopeId: lang.ScopeId("test"),
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
					},
				},
			},
		},
		{
			"empty type-aware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.LiteralType{
								Type: cty.String,
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
			`attr = []`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
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
					Type: cty.Tuple([]cty.Type{cty.String}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.IndexStep{Key: cty.NumberIntVal(0)},
							},
							RangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
								End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
							},
							Type: cty.String,
						},
					},
				},
			},
		},
		{
			"type-aware with invalid element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.LiteralType{
								Type: cty.String,
							},
							schema.LiteralType{
								Type: cty.String,
							},
							schema.LiteralType{
								Type: cty.Number,
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
			`attr = ["one", foo, 42224]`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 27, Byte: 26},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
					Type: cty.Tuple([]cty.Type{cty.String, cty.String, cty.Number}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.IndexStep{Key: cty.NumberIntVal(0)},
							},
							RangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
								End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
							},
							Type: cty.String,
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.IndexStep{Key: cty.NumberIntVal(2)},
							},
							RangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 1, Column: 21, Byte: 20},
								End:      hcl.Pos{Line: 1, Column: 26, Byte: 25},
							},
							Type: cty.Number,
						},
					},
				},
			},
		},
		{
			"type-aware with extra element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.LiteralType{
								Type: cty.String,
							},
							schema.LiteralType{
								Type: cty.Number,
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
			`attr = ["one", 422, "two"]`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 27, Byte: 26},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
					Type: cty.Tuple([]cty.Type{cty.String, cty.Number}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.IndexStep{Key: cty.NumberIntVal(0)},
							},
							RangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
								End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
							},
							Type: cty.String,
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.IndexStep{Key: cty.NumberIntVal(1)},
							},
							RangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 1, Column: 16, Byte: 15},
								End:      hcl.Pos{Line: 1, Column: 19, Byte: 18},
							},
							Type: cty.Number,
						},
					},
				},
			},
		},
		{
			"type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.LiteralType{
								Type: cty.String,
							},
							schema.LiteralType{
								Type: cty.String,
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
			`attr = ["one", "two"]`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
					ScopeId:       lang.ScopeId("test"),
					NestedTargets: reference.Targets{},
				},
			},
		},
		{
			"type-aware nested",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Tuple{
								Elems: []schema.Constraint{
									schema.LiteralType{
										Type: cty.String,
									},
								},
							},
							schema.Tuple{
								Elems: []schema.Constraint{
									schema.LiteralType{
										Type: cty.String,
									},
								},
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
			`attr = [
  ["one"],
  ["two"],
]
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 4, Column: 2, Byte: 32},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
					Type: cty.Tuple([]cty.Type{
						cty.Tuple([]cty.Type{cty.String}),
						cty.Tuple([]cty.Type{cty.String}),
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.IndexStep{Key: cty.NumberIntVal(0)},
							},
							RangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
								End:      hcl.Pos{Line: 2, Column: 10, Byte: 18},
							},
							Type: cty.Tuple([]cty.Type{cty.String}),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "attr"},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
									},
									Type: cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.hcl",
										Start:    hcl.Pos{Line: 2, Column: 4, Byte: 12},
										End:      hcl.Pos{Line: 2, Column: 9, Byte: 17},
									},
								},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.IndexStep{Key: cty.NumberIntVal(1)},
							},
							RangePtr: &hcl.Range{
								Filename: "test.hcl",
								Start:    hcl.Pos{Line: 3, Column: 3, Byte: 22},
								End:      hcl.Pos{Line: 3, Column: 10, Byte: 29},
							},
							Type: cty.Tuple([]cty.Type{cty.String}),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "attr"},
										lang.IndexStep{Key: cty.NumberIntVal(1)},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
									},
									Type: cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.hcl",
										Start:    hcl.Pos{Line: 3, Column: 4, Byte: 23},
										End:      hcl.Pos{Line: 3, Column: 9, Byte: 28},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"tuple from for-expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.LiteralType{
								Type: cty.String,
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
			`attr = [for sa in ["one", "two"]: sa]`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 38, Byte: 37},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
					ScopeId: lang.ScopeId("test"),
					Type:    cty.DynamicPseudoType,
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

func TestCollectRefTargets_exprTuple_implied_hcl(t *testing.T) {
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
									Constraint: schema.Tuple{
										Elems: []schema.Constraint{
											schema.LiteralType{Type: cty.Bool},
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
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.Tuple([]cty.Type{cty.Bool}),
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
							Type: cty.Tuple([]cty.Type{cty.Bool}),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "blk"},
										lang.AttrStep{Name: "attr"},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
									},
									RangePtr: &hcl.Range{
										Filename: "test.hcl",
										Start:    hcl.Pos{Line: 1, Column: 5, Byte: 4},
										End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
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
									Constraint: schema.Tuple{
										Elems: []schema.Constraint{
											schema.LiteralType{Type: cty.Bool},
										},
									},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`blk { attr = [] }`,
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
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.Tuple([]cty.Type{cty.Bool}),
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
							Type: cty.Tuple([]cty.Type{cty.Bool}),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "blk"},
										lang.AttrStep{Name: "attr"},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
									},
									RangePtr: &hcl.Range{
										Filename: "test.hcl",
										Start:    hcl.Pos{Line: 1, Column: 14, Byte: 13},
										End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
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
									Constraint: schema.Tuple{
										Elems: []schema.Constraint{
											schema.LiteralType{Type: cty.Bool},
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

func TestCollectRefTargets_exprTuple_json(t *testing.T) {
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
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.LiteralType{Type: cty.Bool},
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
			"tuple of keyword",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{Keyword: "foo"},
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
			`{"attr": ["foo"]}`,
			reference.Targets{},
		},
		{
			"tuple of addressable reference",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Reference{
								Address: &schema.ReferenceAddrSchema{
									ScopeId: lang.ScopeId("test"),
								},
							},
						},
					},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
					},
				},
			},
			`{"attr": ["foo"]}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					ScopeId: lang.ScopeId("test"),
					RangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 12, Byte: 11},
						End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
					},
				},
			},
		},
		{
			"empty type-aware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.LiteralType{
								Type: cty.String,
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
			`{"attr": []}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
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
					Type: cty.Tuple([]cty.Type{cty.String}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.IndexStep{Key: cty.NumberIntVal(0)},
							},
							RangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
								End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
							},
							Type: cty.String,
						},
					},
				},
			},
		},
		{
			"type-aware with invalid element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.LiteralType{
								Type: cty.String,
							},
							schema.LiteralType{
								Type: cty.String,
							},
							schema.LiteralType{
								Type: cty.Number,
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
			`{"attr": ["one", 422, 42223]}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 1, Column: 29, Byte: 28},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
					},
					Type: cty.Tuple([]cty.Type{cty.String, cty.String, cty.Number}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.IndexStep{Key: cty.NumberIntVal(0)},
							},
							RangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
								End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
							},
							Type: cty.String,
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.IndexStep{Key: cty.NumberIntVal(2)},
							},
							RangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
								End:      hcl.Pos{Line: 1, Column: 28, Byte: 27},
							},
							Type: cty.Number,
						},
					},
				},
			},
		},
		{
			"type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.LiteralType{
								Type: cty.String,
							},
							schema.LiteralType{
								Type: cty.String,
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
			`{"attr": ["one", "two"]}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
					},
					ScopeId:       lang.ScopeId("test"),
					NestedTargets: reference.Targets{},
				},
			},
		},
		{
			"type-aware nested",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Tuple{
								Elems: []schema.Constraint{
									schema.LiteralType{
										Type: cty.String,
									},
								},
							},
							schema.Tuple{
								Elems: []schema.Constraint{
									schema.LiteralType{
										Type: cty.String,
									},
								},
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
			`{"attr": [
  ["one"],
  ["two"]
]}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 4, Column: 2, Byte: 33},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
					},
					Type: cty.Tuple([]cty.Type{
						cty.Tuple([]cty.Type{cty.String}),
						cty.Tuple([]cty.Type{cty.String}),
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.IndexStep{Key: cty.NumberIntVal(0)},
							},
							RangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 2, Column: 3, Byte: 13},
								End:      hcl.Pos{Line: 2, Column: 10, Byte: 20},
							},
							Type: cty.Tuple([]cty.Type{cty.String}),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "attr"},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
									},
									Type: cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.hcl.json",
										Start:    hcl.Pos{Line: 2, Column: 4, Byte: 14},
										End:      hcl.Pos{Line: 2, Column: 9, Byte: 19},
									},
								},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "attr"},
								lang.IndexStep{Key: cty.NumberIntVal(1)},
							},
							RangePtr: &hcl.Range{
								Filename: "test.hcl.json",
								Start:    hcl.Pos{Line: 3, Column: 3, Byte: 24},
								End:      hcl.Pos{Line: 3, Column: 10, Byte: 31},
							},
							Type: cty.Tuple([]cty.Type{cty.String}),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "attr"},
										lang.IndexStep{Key: cty.NumberIntVal(1)},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
									},
									Type: cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.hcl.json",
										Start:    hcl.Pos{Line: 3, Column: 4, Byte: 25},
										End:      hcl.Pos{Line: 3, Column: 9, Byte: 30},
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

func TestCollectRefTargets_exprTuple_implied_json(t *testing.T) {
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
									Constraint: schema.Tuple{
										Elems: []schema.Constraint{
											schema.LiteralType{Type: cty.Bool},
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
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.Tuple([]cty.Type{cty.Bool}),
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
							Type: cty.Tuple([]cty.Type{cty.Bool}),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "blk"},
										lang.AttrStep{Name: "attr"},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
									},
									RangePtr: &hcl.Range{
										Filename: "test.hcl.json",
										Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
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
									Constraint: schema.Tuple{
										Elems: []schema.Constraint{
											schema.LiteralType{Type: cty.Bool},
										},
									},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`{"blk": {"attr": []}}`,
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
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.Tuple([]cty.Type{cty.Bool}),
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
							Type: cty.Tuple([]cty.Type{cty.Bool}),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "blk"},
										lang.AttrStep{Name: "attr"},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
									},
									RangePtr: &hcl.Range{
										Filename: "test.hcl.json",
										Start:    hcl.Pos{Line: 1, Column: 18, Byte: 17},
										End:      hcl.Pos{Line: 1, Column: 18, Byte: 17},
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
									Constraint: schema.Tuple{
										Elems: []schema.Constraint{
											schema.LiteralType{Type: cty.Bool},
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
