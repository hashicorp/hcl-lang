// Copyright IBM Corp. 2020, 2026
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

func TestCollectRefTargets_exprLiteralType_hcl(t *testing.T) {
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
					Constraint: schema.LiteralType{Type: cty.String},
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
			"bool",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{Type: cty.Bool},
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
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.Bool,
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
			},
		},
		{
			"string",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{Type: cty.String},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						AsExprType: true,
					},
				},
			},
			`attr = "foobar"`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
			},
		},
		{
			"number",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{Type: cty.Number},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						AsExprType: true,
					},
				},
			},
			`attr = 42`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.Number,
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
				},
			},
		},
		{
			"list constraint mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.Bool),
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
			"list empty type-aware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.String),
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
					Type:          cty.List(cty.String),
					NestedTargets: reference.Targets{},
				},
			},
		},
		{
			"list type-aware with invalid element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.String),
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
			`attr = ["one", foo, "two"]`,
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
					Type: cty.List(cty.String),
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
							Type: cty.String,
						},
					},
				},
			},
		},
		{
			"list type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.String),
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
			"list type-aware nested",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.List(cty.String)),
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
					Type: cty.List(cty.List(cty.String)),
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
							Type: cty.List(cty.String),
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
							Type: cty.List(cty.String),
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
			"set constraint mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Set(cty.Bool),
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
			"set empty type-aware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Set(cty.String),
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
					Type:          cty.Set(cty.String),
					NestedTargets: reference.Targets{},
				},
			},
		},
		{
			"set type-aware with invalid element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Set(cty.String),
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
			`attr = ["one", foo, "two"]`,
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
					Type:          cty.Set(cty.String),
					NestedTargets: reference.Targets{},
				},
			},
		},
		{
			"set type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Set(cty.String),
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
			"set type-aware nested",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Set(cty.Set(cty.String)),
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
					Type:          cty.Set(cty.Set(cty.String)),
					NestedTargets: reference.Targets{},
				},
			},
		},
		{
			"tuple constraint mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{cty.Bool}),
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
			"tuple empty type-aware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{cty.String}),
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
			"tuple type-aware with invalid element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{cty.String, cty.String, cty.Number}),
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
			"tuple type-aware with extra element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{cty.String, cty.Number}),
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
			"tuple type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{cty.String, cty.String}),
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
			"tuple type-aware nested",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{
							cty.Tuple([]cty.Type{
								cty.String,
							}),
							cty.Tuple([]cty.Type{
								cty.String,
							}),
						}),
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
			"object constraint mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
						}, []string{"foo"}),
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
			"object empty type-aware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Number,
						}, []string{"foo"}),
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
			"object type-aware with invalid key type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Number,
						}, []string{"foo"}),
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
			"object type-aware with invalid attribute name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Number,
						}, []string{"foo"}),
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
			"object type-aware with invalid value type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Number,
						}, []string{"foo"}),
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
			"object type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Number,
						}, []string{"foo"}),
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
			"object nested type-unaware",
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
			"map constraint mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.String),
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
			"map empty type-aware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.String),
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
					Type: cty.Map(cty.String),
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
					NestedTargets: reference.Targets{},
				},
			},
		},
		{
			"map type-aware with invalid key type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.Number),
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
					Type: cty.Map(cty.Number),
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
								lang.IndexStep{Key: cty.StringVal("bar")},
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
			"map type-aware with multiple items",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.Number),
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
  fox = 12345
  bar = 42
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.Map(cty.Number),
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
								lang.IndexStep{Key: cty.StringVal("bar")},
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
								lang.IndexStep{Key: cty.StringVal("fox")},
							},
							Type: cty.Number,
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
			"map type-aware with invalid value type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.Number),
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
  bar = 42
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.Map(cty.Number),
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
								lang.IndexStep{Key: cty.StringVal("bar")},
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
			"map type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.Number),
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
  foo = 12345
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
			"map nested type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.Map(cty.String)),
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
  foo = {   }
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
					Type: cty.Map(cty.Map(cty.String)),
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
								lang.IndexStep{Key: cty.StringVal("bar")},
							},
							Type: cty.Map(cty.String),
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
										lang.IndexStep{Key: cty.StringVal("bar")},
										lang.IndexStep{Key: cty.StringVal("baz")},
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
								lang.IndexStep{Key: cty.StringVal("foo")},
							},
							Type: cty.Map(cty.String),
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
							NestedTargets: reference.Targets{},
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

func TestCollectRefTargets_exprLiteralType_json(t *testing.T) {
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
					Constraint: schema.LiteralType{Type: cty.String},
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
			"bool",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{Type: cty.Bool},
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
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.Bool,
					RangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
					},
				},
			},
		},
		{
			"string",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{Type: cty.String},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						AsExprType: true,
					},
				},
			},
			`{"attr": "foobar"}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 1, Column: 18, Byte: 17},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 2, Byte: 1},
						End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
					},
				},
			},
		},
		{
			"number",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{Type: cty.Number},
					IsOptional: true,
					Address: &schema.AttributeAddrSchema{
						Steps: schema.Address{
							schema.AttrNameStep{},
						},
						AsExprType: true,
					},
				},
			},
			`{"attr": 42}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.Number,
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
				},
			},
		},
		{
			"list constraint mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.Bool),
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
			"list empty type-aware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.String),
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
					Type:          cty.List(cty.String),
					NestedTargets: reference.Targets{},
				},
			},
		},
		{
			"list type-aware with invalid element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.String),
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
			`{"attr": ["one", 422, "two"]}`,
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
					Type: cty.List(cty.String),
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
							Type: cty.String,
						},
					},
				},
			},
		},
		{
			"list type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.String),
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
			"list type-aware nested",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.List(cty.String)),
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
					Type: cty.List(cty.List(cty.String)),
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
							Type: cty.List(cty.String),
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
							Type: cty.List(cty.String),
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
		{
			"set constraint mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Set(cty.Bool),
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
			"set empty type-aware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Set(cty.String),
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
					Type:          cty.Set(cty.String),
					NestedTargets: reference.Targets{},
				},
			},
		},
		{
			"set type-aware with invalid element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Set(cty.String),
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
			`{"attr": ["one", 422, "two"]}`,
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
					Type:          cty.Set(cty.String),
					NestedTargets: reference.Targets{},
				},
			},
		},
		{
			"set type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Set(cty.String),
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
			"set type-aware nested",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Set(cty.Set(cty.String)),
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
					Type:          cty.Set(cty.Set(cty.String)),
					NestedTargets: reference.Targets{},
				},
			},
		},
		{
			"tuple constraint mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{cty.Bool}),
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
			"tuple empty type-aware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{cty.String}),
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
			"tuple type-aware with invalid element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{cty.String, cty.String, cty.Number}),
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
			"tuple type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{cty.String, cty.String}),
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
			"tuple type-aware nested",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{
							cty.Tuple([]cty.Type{cty.String}),
							cty.Tuple([]cty.Type{cty.String}),
						}),
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
		{
			"object constraint mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
						}, []string{"foo"}),
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
			"object empty type-aware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Number,
						}, []string{"foo"}),
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
			"object type-aware with invalid key type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Number,
						}, []string{"foo"}),
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
			"object type-aware with invalid attribute name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Number,
						}, []string{"foo"}),
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
			"object type-aware with invalid value type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Number,
						}, []string{"foo"}),
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
			"object type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Number,
						}, []string{"foo"}),
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
			"object nested type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Object(map[string]cty.Type{
								"baz": cty.String,
							}),
						}, []string{"foo"}),
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
		{
			"map constraint mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.String),
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
			"map empty type-aware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.String),
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
					Type: cty.Map(cty.String),
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
					NestedTargets: reference.Targets{},
				},
			},
		},
		{
			"map type-aware with invalid key type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.Number),
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
					Type: cty.Map(cty.Number),
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
								lang.IndexStep{Key: cty.StringVal("bar")},
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
			"map type-aware with multiple items",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.Number),
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
  "fox": 12345,
  "bar": 42
}}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.Map(cty.Number),
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
								lang.IndexStep{Key: cty.StringVal("bar")},
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
								lang.IndexStep{Key: cty.StringVal("fox")},
							},
							Type: cty.Number,
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
		{
			"map type-aware with invalid value type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.Number),
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
  "bar": 42
}}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.Map(cty.Number),
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
								lang.IndexStep{Key: cty.StringVal("bar")},
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
			"map type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.Number),
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
  "foo": 12345,
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
			"map nested type-unaware",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.Map(cty.String)),
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
  "foo": {   },
  "bar": {
    "baz": "noot"
  }
}}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "attr"},
					},
					Type: cty.Map(cty.Map(cty.String)),
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
								lang.IndexStep{Key: cty.StringVal("bar")},
							},
							Type: cty.Map(cty.String),
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
										lang.IndexStep{Key: cty.StringVal("bar")},
										lang.IndexStep{Key: cty.StringVal("baz")},
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
								lang.IndexStep{Key: cty.StringVal("foo")},
							},
							Type: cty.Map(cty.String),
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
							NestedTargets: reference.Targets{},
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
