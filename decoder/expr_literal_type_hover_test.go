// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestHoverAtPos_exprLiteralType(t *testing.T) {
	testCases := []struct {
		testName          string
		attrSchema        map[string]*schema.AttributeSchema
		cfg               string
		pos               hcl.Pos
		expectedHoverData *lang.HoverData
	}{
		{
			"any type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.DynamicPseudoType,
					},
					IsOptional: true,
				},
			},
			`attr = "foobar"`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			&lang.HoverData{
				Content: lang.Markdown("_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
				},
			},
		},

		// primitive types
		{
			"boolean",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Bool,
					},
				},
			},
			`attr = false`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_bool_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
				},
			},
		},
		{
			"number whole",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Number,
					},
				},
			},
			`attr = 4222"`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_number_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
				},
			},
		},
		{
			"number fractional",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Number,
					},
				},
			},
			`attr = 4.222"`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_number_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
				},
			},
		},
		{
			"string single-line",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.String,
					},
				},
			},
			`attr = "foo"`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
				},
			},
		},
		{
			"string multi-line",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.String,
					},
				},
			},
			`attr = <<TEXT
foo
bar
TEXT
`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 4, Column: 5, Byte: 26},
				},
			},
		},
		{
			"string template",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.String,
					},
				},
			},
			`attr = "foo${bar}"`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			nil,
		},

		// list
		{
			"list with any element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.DynamicPseudoType),
					},
				},
			},
			`attr = []`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_list of any type_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"empty single-line list with element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.String),
					},
				},
			},
			`attr = []`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_list of string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"empty multi-line list with element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.String),
					},
				},
			},
			`attr = [
  
]`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("_list of string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 3, Column: 2, Byte: 13},
				},
			},
		},
		{
			"single element single-line list on element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.String),
					},
				},
			},
			`attr = ["fooba"]`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			&lang.HoverData{
				Content: lang.Markdown("_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
					End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
				},
			},
		},
		{
			"multi-element single-line list on list",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.String),
					},
				},
			},
			`attr = [ "one", "two" ]`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			&lang.HoverData{
				Content: lang.Markdown("_list of string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 8,
						Byte:   7,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 24,
						Byte:   23,
					},
				},
			},
		},
		{
			"single element multi-line list on element with custom data",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.String),
					},
				},
			},
			`attr = [
  "foobar",
]`,
			hcl.Pos{Line: 2, Column: 6, Byte: 14},
			&lang.HoverData{
				Content: lang.Markdown("_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
					End:      hcl.Pos{Line: 2, Column: 11, Byte: 19},
				},
			},
		},
		{
			"multi-element multi-line list on invalid element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.Number),
					},
				},
			},
			`attr = [
  true,
  12345678,
]`,
			hcl.Pos{Line: 2, Column: 6, Byte: 14},
			nil,
		},
		{
			"multi-element multi-line list on second element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.String),
					},
				},
			},
			`attr = [
  "foo",
  "bar",
]`,
			hcl.Pos{Line: 3, Column: 4, Byte: 22},
			&lang.HoverData{
				Content: lang.Markdown("_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 3, Column: 3, Byte: 20},
					End:      hcl.Pos{Line: 3, Column: 8, Byte: 25},
				},
			},
		},

		// set
		{
			"set with any element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Set(cty.DynamicPseudoType),
					},
				},
			},
			`attr = []`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_set of any type_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"empty single-line set with element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Set(cty.String),
					},
				},
			},
			`attr = []`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_set of string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"empty multi-line set with element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Set(cty.String),
					},
				},
			},
			`attr = [
  
]`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("_set of string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 3, Column: 2, Byte: 13},
				},
			},
		},
		{
			"single element single-line set on element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Set(cty.String),
					},
				},
			},
			`attr = ["fooba"]`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			&lang.HoverData{
				Content: lang.Markdown("_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
					End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
				},
			},
		},
		{
			"multi-element single-line set on set",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Set(cty.String),
					},
				},
			},
			`attr = [ "one", "two" ]`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			&lang.HoverData{
				Content: lang.Markdown("_set of string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 8,
						Byte:   7,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 24,
						Byte:   23,
					},
				},
			},
		},
		{
			"multi-element multi-line set on invalid element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Set(cty.Number),
					},
				},
			},
			`attr = [
  false,
  12345678,
]`,
			hcl.Pos{Line: 2, Column: 6, Byte: 14},
			nil,
		},
		{
			"multi-element multi-line set on second element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Set(cty.String),
					},
				},
			},
			`attr = [
  "foo",
  "bar",
]`,
			hcl.Pos{Line: 3, Column: 5, Byte: 22},
			&lang.HoverData{
				Content: lang.Markdown("_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 3, Column: 3, Byte: 20},
					End:      hcl.Pos{Line: 3, Column: 8, Byte: 25},
				},
			},
		},

		// tuple
		{
			"empty single-line tuple without element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{}),
					},
				},
			},
			`attr = []`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_tuple_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"empty multi-line tuple without element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{}),
					},
				},
			},
			`attr = [
  
]`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("_tuple_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 3, Column: 2, Byte: 13},
				},
			},
		},
		{
			"empty single-line tuple with one element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{
							cty.String,
							cty.Number,
						}),
					},
				},
			},
			`attr = []`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_tuple_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"empty multi-line tuple with one element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{
							cty.String,
						}),
					},
				},
			},
			`attr = [
  
]`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("_tuple_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 3, Column: 2, Byte: 13},
				},
			},
		},
		{
			"single element single-line tuple on element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{
							cty.String,
							cty.Bool,
						}),
					},
				},
			},
			`attr = ["fooba"]`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			&lang.HoverData{
				Content: lang.Markdown("_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
					End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
				},
			},
		},
		{
			"multi-element single-line tuple on tuple",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{
							cty.String,
							cty.Number,
						}),
					},
				},
			},
			`attr = [ "one", 42234 ]`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			&lang.HoverData{
				Content: lang.Markdown("_tuple_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 8,
						Byte:   7,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 24,
						Byte:   23,
					},
				},
			},
		},
		{
			"multi-element multi-line tuple on invalid element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{
							cty.Number,
							cty.String,
						}),
					},
				},
			},
			`attr = [
  "foo",
  42223,
]`,
			hcl.Pos{Line: 2, Column: 6, Byte: 14},
			nil,
		},
		{
			"multi-element multi-line tuple on second element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{
							cty.String,
							cty.String,
						}),
					},
				},
			},
			`attr = [
  "foobar",
  "barfoo",
]`,
			hcl.Pos{Line: 3, Column: 6, Byte: 26},
			&lang.HoverData{
				Content: lang.Markdown("_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 3, Column: 3, Byte: 23},
					End:      hcl.Pos{Line: 3, Column: 11, Byte: 31},
				},
			},
		},

		// map
		{
			"empty single-line map with any element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.DynamicPseudoType),
					},
				},
			},
			`attr = {}`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_map of any type_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"empty multi-line map with any element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.DynamicPseudoType),
					},
				},
			},
			`attr = {
  
}`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("_map of any type_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 3, Column: 2, Byte: 13},
				},
			},
		},
		{
			"empty single-line map with element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.String),
					},
				},
			},
			`attr = {}`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_map of string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"single item map on key name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.String),
					},
				},
			},
			`attr = {
  foo = "bar"
}`,
			hcl.Pos{Line: 2, Column: 5, Byte: 13},
			nil,
		},
		{
			"single item map on invalid key type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.String),
					},
				},
			},
			`attr = {
  422 = "bar"
}`,
			hcl.Pos{Line: 2, Column: 5, Byte: 13},
			nil,
		},
		{
			"multi item map on valid key type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.String),
					},
				},
			},
			`attr = {
  422 = "foo"
  bar = "bar"
  432 = "baz"
}`,
			hcl.Pos{Line: 3, Column: 5, Byte: 27},
			nil,
		},
		{
			"multi item map on matching value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.String),
					},
				},
			},
			`attr = {
  foo = invalid
  bar = "foobar"
  baz = "foobaz"
}`,
			hcl.Pos{Line: 3, Column: 13, Byte: 37},
			&lang.HoverData{
				Content: lang.Markdown("_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 3, Column: 9, Byte: 33},
					End:      hcl.Pos{Line: 3, Column: 17, Byte: 41},
				},
			},
		},
		{
			"multi item map on mismatching value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.String),
					},
				},
			},
			`attr = {
  foo = invalid
  bar = "foobar"
  baz = "foobaz"
}`,
			hcl.Pos{Line: 2, Column: 13, Byte: 21},
			nil,
		},

		// object
		{
			"empty single-line object without attributes",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Object(map[string]cty.Type{}),
					},
				},
			},
			`attr = {}`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_object_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"empty multi-line object without attributes",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Object(map[string]cty.Type{}),
					},
				},
			},
			`attr = {
  
}`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("_object_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 3, Column: 2, Byte: 13},
				},
			},
		},
		{
			"empty single-line object with attributes",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
						}, []string{"foo"}),
					},
				},
			},
			`attr = {}`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("```\n{\n  foo = string # optional\n}\n```\n_object_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"empty multi-line object with attributes",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
						}, []string{"foo"}),
					},
				},
			},
			`attr = {
  
}`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("```\n{\n  foo = string # optional\n}\n```\n_object_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 3, Column: 2, Byte: 13},
				},
			},
		},
		{
			"single item object on valid attribute name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
						}, []string{"foo"}),
					},
				},
			},
			`attr = {
  foo = "fooba"
}`,
			hcl.Pos{Line: 2, Column: 5, Byte: 13},
			&lang.HoverData{
				Content: lang.Markdown("**foo** _optional, string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
					End:      hcl.Pos{Line: 2, Column: 16, Byte: 24},
				},
			},
		},
		{
			"single item object on invalid attribute name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Object(map[string]cty.Type{
							"foo": cty.String,
						}),
					},
				},
			},
			`attr = {
  bar = "fooba"
}`,
			hcl.Pos{Line: 2, Column: 5, Byte: 13},
			&lang.HoverData{
				Content: lang.Markdown("```" + `
{
  foo = string
}
` + "```\n" + `_object_`),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 3, Column: 2, Byte: 26},
				},
			},
		},
		{
			"multi item object on valid attribute name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.String,
							"baz": cty.String,
						}, []string{"foo", "baz"}),
					},
				},
			},
			`attr = {
  foo = "foo"
  bar = "bar"
  baz = "baz"
}`,
			hcl.Pos{Line: 3, Column: 5, Byte: 27},
			&lang.HoverData{
				Content: lang.Markdown("**bar** _required, string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 3, Column: 3, Byte: 25},
					End:      hcl.Pos{Line: 3, Column: 14, Byte: 36},
				},
			},
		},
		{
			"multi item object on matching value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.String,
							"baz": cty.String,
						}, []string{"foo", "bar", "baz"}),
					},
				},
			},
			`attr = {
  foo = invalid
  bar = "barbar"
  baz = "bazbaz"
}`,
			hcl.Pos{Line: 3, Column: 16, Byte: 40},
			&lang.HoverData{
				Content: lang.Markdown("_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 3, Column: 9, Byte: 33},
					End:      hcl.Pos{Line: 3, Column: 17, Byte: 41},
				},
			},
		},
		{
			"multi item object on mismatching value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.String,
							"baz": cty.String,
						}, []string{"foo", "bar", "baz"}),
					},
				},
			},
			`attr = {
  foo = invalid
  bar = "barbar"
  baz = "bazbaz"
}`,
			hcl.Pos{Line: 2, Column: 13, Byte: 21},
			nil,
		},
		{
			"multi item object in empty space",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.Number,
							"bar": cty.String,
							"baz": cty.String,
						}, []string{"foo", "bar", "baz"}),
					},
				},
			},
			`attr = {
  bar = "bar"
  baz = "baz"
}`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("```" + `
{
  bar = string # optional
  baz = string # optional
  foo = number # optional
}
` + "```\n_object_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 4, Column: 2, Byte: 38},
				},
			},
		},
		{
			"multi item nested object",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
							"foo": cty.Number,
							"bar": cty.ObjectWithOptionalAttrs(map[string]cty.Type{
								"noot":   cty.Bool,
								"animal": cty.String,
							}, []string{"animal"}),
							"baz": cty.String,
						}, []string{"foo", "bar", "baz"}),
					},
				},
			},
			`attr = {
  bar = {}
  baz = "baz"
}`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("```" + `
{
  bar = {
    animal = string # optional
    noot = bool
  } # optional
  baz = string # optional
  foo = number # optional
}
` + "```\n_object_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 4, Column: 2, Byte: 35},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
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
			hoverData, err := d.HoverAtPos(ctx, "test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedHoverData, hoverData); diff != "" {
				t.Fatalf("unexpected hover data: %s", diff)
			}
		})
	}
}
