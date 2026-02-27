// Copyright IBM Corp. 2020, 2026
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestSemanticTokens_exprLiteralValue(t *testing.T) {

	testCases := []struct {
		testName               string
		attrSchema             map[string]*schema.AttributeSchema
		cfg                    string
		expectedSemanticTokens []lang.SemanticToken
	}{
		{
			"boolean",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.BoolVal(false),
					},
				},
			},
			`attr = false`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenBool,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
					},
				},
			},
		},

		// string
		{
			"single-line string",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.StringVal("foobar"),
					},
				},
			},
			`attr = "foobar"`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
					},
				},
			},
		},
		{
			"multi-line string",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.StringVal("foo\nbar\n"),
					},
				},
			},
			`attr = <<TEXT
foo
bar
TEXT
`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 4, Column: 5, Byte: 26},
					},
				},
			},
		},

		// number
		{
			"number whole",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.NumberIntVal(4223),
					},
				},
			},
			`attr = 4223`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenNumber,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
					},
				},
			},
		},
		{
			"number fractional",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.NumberFloatVal(4.222),
					},
				},
			},
			`attr = 4.222`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenNumber,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
					},
				},
			},
		},

		// list
		{
			"empty list with any element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ListValEmpty(cty.DynamicPseudoType),
					},
				},
			},
			`attr = []`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
			},
		},
		{
			"empty list with element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ListVal([]cty.Value{
							cty.StringVal("foo"),
							cty.StringVal("bar"),
						}),
					},
				},
			},
			`attr = []`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
			},
		},
		{
			"single element list",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ListVal([]cty.Value{
							cty.StringVal("fooba"),
						}),
					},
				},
			},
			`attr = ["fooba"]`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
					},
				},
			},
		},
		{
			"single element multi-line list",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ListVal([]cty.Value{
							cty.StringVal("foobar"),
							cty.StringVal("barbar"),
						}),
					},
				},
			},
			`attr = [
  "foobar",
]`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 11, Byte: 19},
					},
				},
			},
		},
		{
			"multi-element multi-line list",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ListVal([]cty.Value{
							cty.StringVal("foobar"),
							cty.StringVal("barfoo"),
						}),
					},
				},
			},
			`attr = [
  "foobar",
  "barfoo",
]`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 11, Byte: 19},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 23},
						End:      hcl.Pos{Line: 3, Column: 11, Byte: 31},
					},
				},
			},
		},
		{
			"multi-element multi-line list without last token",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ListVal([]cty.Value{
							cty.StringVal("foobar"),
							cty.StringVal("barbar"),
						}),
					},
				},
			},
			`attr = [
  "foobar",
  "barfoo",
]`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 11, Byte: 19},
					},
				},
			},
		},
		{
			"multi-element multi-line list with invalid element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ListVal([]cty.Value{
							cty.StringVal("fooba"),
							cty.StringVal("foobar"),
							cty.StringVal("barbar"),
						}),
					},
				},
			},
			`attr = [
  "fooba",
  invalid,
  "barbar",
]`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 10, Byte: 18},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 4, Column: 3, Byte: 33},
						End:      hcl.Pos{Line: 4, Column: 11, Byte: 41},
					},
				},
			},
		},

		// set
		{
			"empty set with any element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.SetVal([]cty.Value{
							cty.DynamicVal,
						}),
					},
				},
			},
			`attr = []`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
			},
		},
		{
			"empty set with element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.SetVal([]cty.Value{
							cty.StringVal("foo"),
						}),
					},
				},
			},
			`attr = []`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
			},
		},
		{
			"single element set",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.SetVal([]cty.Value{
							cty.StringVal("fooba"),
						}),
					},
				},
			},
			`attr = ["fooba"]`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
					},
				},
			},
		},
		{
			"single element multi-line set",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.SetVal([]cty.Value{
							cty.StringVal("fooba"),
						}),
					},
				},
			},
			`attr = [
  "fooba",
]`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 10, Byte: 18},
					},
				},
			},
		},
		{
			"multi-element multi-line set",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.SetVal([]cty.Value{
							cty.StringVal("fooba"),
							cty.StringVal("barba"),
						}),
					},
				},
			},
			`attr = [
  "fooba",
  "barba",
]`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 10, Byte: 18},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 22},
						End:      hcl.Pos{Line: 3, Column: 10, Byte: 29},
					},
				},
			},
		},
		{
			"multi-element multi-line set with invalid element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.SetVal([]cty.Value{
							cty.StringVal("fooba"),
							cty.StringVal("barba"),
							cty.StringVal("waaba"),
						}),
					},
				},
			},
			`attr = [
  "fooba",
  invalid,
  "waaba",
]`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 10, Byte: 18},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 4, Column: 3, Byte: 33},
						End:      hcl.Pos{Line: 4, Column: 10, Byte: 40},
					},
				},
			},
		},

		// tuple
		{
			"empty tuple without element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.TupleVal([]cty.Value{}),
					},
				},
			},
			`attr = []`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
			},
		},
		{
			"empty tuple with element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.TupleVal([]cty.Value{
							cty.StringVal("foo"),
						}),
					},
				},
			},
			`attr = []`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
			},
		},
		{
			"single element tuple",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.TupleVal([]cty.Value{
							cty.StringVal("fooba"),
						}),
					},
				},
			},
			`attr = ["fooba"]`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
					},
				},
			},
		},
		{
			"single element multi-line tuple",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.TupleVal([]cty.Value{
							cty.StringVal("fooba"),
						}),
					},
				},
			},
			`attr = [
  "fooba",
]`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 10, Byte: 18},
					},
				},
			},
		},
		{
			"multi-element multi-line tuple",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.TupleVal([]cty.Value{
							cty.StringVal("fooba"),
							cty.NumberVal(big.NewFloat(1234567)),
						}),
					},
				},
			},
			`attr = [
  "fooba",
  1234567,
]`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 10, Byte: 18},
					},
				},
				{
					Type:      lang.TokenNumber,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 22},
						End:      hcl.Pos{Line: 3, Column: 10, Byte: 29},
					},
				},
			},
		},
		{
			"multi-element multi-line tuple with invalid element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.TupleVal([]cty.Value{
							cty.StringVal("fooba"),
							cty.StringVal("foobadddd"),
							cty.NumberVal(big.NewFloat(1234567)),
						}),
					},
				},
			},
			`attr = [
  "fooba",
  invalid,
  1234567,
]`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 10, Byte: 18},
					},
				},
				{
					Type:      lang.TokenNumber,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 4, Column: 3, Byte: 33},
						End:      hcl.Pos{Line: 4, Column: 10, Byte: 40},
					},
				},
			},
		},

		// map
		{
			"any element constraint",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.MapVal(map[string]cty.Value{
							"foo": cty.StringVal("foobar"),
						}),
					},
				},
			},
			`attr = {}`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
			},
		},
		{
			"single-line with mismatching expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.MapVal(map[string]cty.Value{
							"foobar": cty.StringVal("foobar"),
						}),
					},
				},
			},
			`attr = [ foobar ]`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
			},
		},
		{
			"single-line with mismatching key type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.MapVal(map[string]cty.Value{
							"422": cty.StringVal("foobar"),
						}),
					},
				},
			},
			`attr = { 422 = foobar }`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
			},
		},
		{
			"single-line with valid item",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.MapVal(map[string]cty.Value{
							"foo": cty.StringVal("noot"),
						}),
					},
				},
			},
			`attr = { foo = "noot" }`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
						End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 16, Byte: 15},
						End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
					},
				},
			},
		},
		{
			"single-line with valid multiple items",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.MapVal(map[string]cty.Value{
							"foo": cty.StringVal("noot"),
							"bar": cty.StringVal("noot"),
						}),
					},
				},
			},
			`attr = { foo = "noot", bar = "noot" }`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
						End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 16, Byte: 15},
						End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 24, Byte: 23},
						End:      hcl.Pos{Line: 1, Column: 27, Byte: 26},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 30, Byte: 29},
						End:      hcl.Pos{Line: 1, Column: 36, Byte: 35},
					},
				},
			},
		},
		{
			"single-line with valid multiple items with quoted keys",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.MapVal(map[string]cty.Value{
							"foo": cty.StringVal("noot"),
							"bar": cty.StringVal("noot"),
						}),
					},
				},
			},
			`attr = { "foo" = "noot", "bar" = "noot" }`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
						End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 18, Byte: 17},
						End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 26, Byte: 25},
						End:      hcl.Pos{Line: 1, Column: 31, Byte: 30},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 34, Byte: 33},
						End:      hcl.Pos{Line: 1, Column: 40, Byte: 39},
					},
				},
			},
		},
		{
			"single-line with multiple items and one mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.MapVal(map[string]cty.Value{
							"foo": cty.StringVal("bar"),
							"bar": cty.StringVal("noot"),
						}),
					},
				},
			},
			`attr = { foo = bar, bar = "noot" }`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
						End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 21, Byte: 20},
						End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 27, Byte: 26},
						End:      hcl.Pos{Line: 1, Column: 33, Byte: 32},
					},
				},
			},
		},
		{
			"multi-line with valid item",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.MapVal(map[string]cty.Value{
							"foo": cty.StringVal("noot"),
						}),
					},
				},
			},
			`attr = {
  foo = "noot"
}`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 6, Byte: 14},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 9, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 15, Byte: 23},
					},
				},
			},
		},
		{
			"multi-line with valid multiple items",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.MapVal(map[string]cty.Value{
							"foo": cty.StringVal("noot"),
							"bar": cty.StringVal("toot"),
						}),
					},
				},
			},
			`attr = {
  foo = "noot"
  bar = "toot"
}`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 6, Byte: 14},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 9, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 15, Byte: 23},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 26},
						End:      hcl.Pos{Line: 3, Column: 6, Byte: 29},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 9, Byte: 32},
						End:      hcl.Pos{Line: 3, Column: 15, Byte: 38},
					},
				},
			},
		},
		{
			"multi-line with multiple items and one mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.MapVal(map[string]cty.Value{
							"foo": cty.StringVal("bar"),
							"bar": cty.StringVal("noot"),
						}),
					},
				},
			},
			`attr = {
  foo = bar
  bar = "noot"
}`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 6, Byte: 14},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 23},
						End:      hcl.Pos{Line: 3, Column: 6, Byte: 26},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 9, Byte: 29},
						End:      hcl.Pos{Line: 3, Column: 15, Byte: 35},
					},
				},
			},
		},

		// 		// object
		{
			"undefined attributes",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ObjectVal(map[string]cty.Value{}),
					},
				},
			},
			`attr = {}`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
			},
		},
		{
			"single-line with mismatching expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ObjectVal(map[string]cty.Value{
							"foo": cty.StringVal("foobar"),
						}),
					},
				},
			},
			`attr = [ foobar ]`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
			},
		},
		{
			"single-line with mismatching key type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ObjectVal(map[string]cty.Value{
							"422": cty.StringVal("noot"),
						}),
					},
				},
			},
			`attr = { 422 = "noot" }`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
			},
		},
		{
			"single-line with valid item",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ObjectVal(map[string]cty.Value{
							"foo": cty.StringVal("noot"),
						}),
					},
				},
			},
			`attr = { foo = "noot" }`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenObjectKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
						End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 16, Byte: 15},
						End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
					},
				},
			},
		},
		{
			"single-line with valid multiple items",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ObjectVal(map[string]cty.Value{
							"foo": cty.StringVal("noot"),
							"bar": cty.StringVal("toot"),
						}),
					},
				},
			},
			`attr = { foo = "noot", bar = "toot" }`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenObjectKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
						End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 16, Byte: 15},
						End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
					},
				},
				{
					Type:      lang.TokenObjectKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 24, Byte: 23},
						End:      hcl.Pos{Line: 1, Column: 27, Byte: 26},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 30, Byte: 29},
						End:      hcl.Pos{Line: 1, Column: 36, Byte: 35},
					},
				},
			},
		},
		{
			"single-line with valid multiple items with valid quoted attributes",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ObjectVal(map[string]cty.Value{
							"foo": cty.StringVal("noot"),
							"bar": cty.StringVal("toot"),
						}),
					},
				},
			},
			`attr = { "foo" = "noot", "bar" = "toot" }`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenObjectKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
						End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 18, Byte: 17},
						End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
					},
				},
				{
					Type:      lang.TokenObjectKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 26, Byte: 25},
						End:      hcl.Pos{Line: 1, Column: 31, Byte: 30},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 34, Byte: 33},
						End:      hcl.Pos{Line: 1, Column: 40, Byte: 39},
					},
				},
			},
		},
		{
			"single-line with multiple items with invalid quoted attribute",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ObjectVal(map[string]cty.Value{
							"foo": cty.StringVal("noot"),
							"bar": cty.StringVal("toot"),
						}),
					},
				},
			},
			`attr = { "foo" = "noot", "baz" = "noot" }`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenObjectKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
						End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 18, Byte: 17},
						End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
					},
				},
			},
		},
		{
			"single-line with multiple items and one value mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ObjectVal(map[string]cty.Value{
							"foo": cty.StringVal("bar"),
							"bar": cty.StringVal("noot"),
						}),
					},
				},
			},
			`attr = { foo = bar, bar = "noot" }`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenObjectKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
						End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
					},
				},
				{
					Type:      lang.TokenObjectKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 21, Byte: 20},
						End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 27, Byte: 26},
						End:      hcl.Pos{Line: 1, Column: 33, Byte: 32},
					},
				},
			},
		},
		{
			"multi-line with valid item",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ObjectVal(map[string]cty.Value{
							"foo": cty.StringVal("noot"),
						}),
					},
				},
			},
			`attr = {
  foo = "noot"
}`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenObjectKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 6, Byte: 14},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 9, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 15, Byte: 23},
					},
				},
			},
		},
		{
			"multi-line with valid multiple items",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ObjectVal(map[string]cty.Value{
							"foo": cty.StringVal("noot"),
							"bar": cty.StringVal("toot"),
						}),
					},
				},
			},
			`attr = {
  foo = "noot"
  bar = "toot"
}`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenObjectKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 6, Byte: 14},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 9, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 15, Byte: 23},
					},
				},
				{
					Type:      lang.TokenObjectKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 26},
						End:      hcl.Pos{Line: 3, Column: 6, Byte: 29},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 9, Byte: 32},
						End:      hcl.Pos{Line: 3, Column: 15, Byte: 38},
					},
				},
			},
		},
		{
			"multi-line with multiple items and one value mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ObjectVal(map[string]cty.Value{
							"foo": cty.StringVal("bar"),
							"bar": cty.StringVal("noot"),
						}),
					},
				},
			},
			`attr = {
  foo = bar
  bar = "noot"
}`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenObjectKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 6, Byte: 14},
					},
				},
				{
					Type:      lang.TokenObjectKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 23},
						End:      hcl.Pos{Line: 3, Column: 6, Byte: 26},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 9, Byte: 29},
						End:      hcl.Pos{Line: 3, Column: 15, Byte: 35},
					},
				},
			},
		},
		{
			"multi-line with multiple items and one attribute mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ObjectVal(map[string]cty.Value{
							"foo": cty.StringVal("x"),
							"baz": cty.StringVal("x"),
						}),
					},
				},
			},
			`attr = {
  foo = "x"
  baz = "x"
}`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenObjectKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 6, Byte: 14},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 9, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 12, Byte: 20},
					},
				},
				{
					Type:      lang.TokenObjectKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 23},
						End:      hcl.Pos{Line: 3, Column: 6, Byte: 26},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 9, Byte: 29},
						End:      hcl.Pos{Line: 3, Column: 12, Byte: 32},
					},
				},
			},
		},
		{
			"multi-line with nested object",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ObjectVal(map[string]cty.Value{
							"foo": cty.ObjectVal(map[string]cty.Value{
								"noot": cty.StringVal("to"),
							}),
							"bar": cty.StringVal("noot"),
						}),
					},
				},
			},
			`attr = {
  foo = {
    noot = "to"
  }
  bar = "noot"
}`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenObjectKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 6, Byte: 14},
					},
				},
				{
					Type:      lang.TokenObjectKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 5, Byte: 23},
						End:      hcl.Pos{Line: 3, Column: 9, Byte: 27},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 12, Byte: 30},
						End:      hcl.Pos{Line: 3, Column: 16, Byte: 34},
					},
				},
				{
					Type:      lang.TokenObjectKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 5, Column: 3, Byte: 41},
						End:      hcl.Pos{Line: 5, Column: 6, Byte: 44},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 5, Column: 9, Byte: 47},
						End:      hcl.Pos{Line: 5, Column: 15, Byte: 53},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%02d-%s", i, tc.testName), func(t *testing.T) {
			bodySchema := &schema.BodySchema{
				Attributes: tc.attrSchema,
			}

			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			})

			ctx := context.Background()
			tokens, err := d.SemanticTokensInFile(ctx, "test.tf")
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedSemanticTokens, tokens); diff != "" {
				t.Fatalf("unexpected tokens: %s", diff)
			}
		})
	}
}
