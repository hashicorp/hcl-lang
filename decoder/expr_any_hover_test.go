// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestHoverAtPos_exprAny_functions(t *testing.T) {
	testCases := []struct {
		testName     string
		attrSchema   map[string]*schema.AttributeSchema
		cfg          string
		pos          hcl.Pos
		expectedData *lang.HoverData
	}{
		{
			"over unknown function",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = unknown()
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			nil,
		},
		{
			"over function name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = lower("FOO")
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			&lang.HoverData{
				Content: lang.MarkupContent{
					Value: "```terraform\nlower(str string) string\n```\n\n`lower` converts all cased letters in the given string to lowercase.",
					Kind:  lang.MarkdownKind,
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 20, Byte: 19},
				},
			},
		},
		{
			"over function parameter",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = join(",", ["a", "b"])
`,
			hcl.Pos{Line: 1, Column: 15, Byte: 14},
			&lang.HoverData{
				Content: lang.MarkupContent{
					Value: "_string_",
					Kind:  lang.MarkdownKind,
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 13, Byte: 12},
					End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
				},
			},
		},
		{
			"over complex variadic parameter",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = join(",", ["a", "b"])
`,
			hcl.Pos{Line: 1, Column: 21, Byte: 20},
			&lang.HoverData{
				Content: lang.MarkupContent{
					Value: "_string_",
					Kind:  lang.MarkdownKind,
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 19, Byte: 18},
					End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
				},
			},
		},
		{
			"over too many arguments",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = lower("FOO", "BAR")
`,
			hcl.Pos{Line: 1, Column: 24, Byte: 23},
			nil,
		},
		{
			"over mismatching argument",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = lower(["FOO"])
`,
			hcl.Pos{Line: 1, Column: 17, Byte: 16},
			nil,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%2d-%s", i, tc.testName), func(t *testing.T) {
			bodySchema := &schema.BodySchema{
				Attributes: tc.attrSchema,
			}

			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
				Functions: testFunctionSignatures(),
			})

			ctx := context.Background()
			data, err := d.HoverAtPos(ctx, "test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedData, data); diff != "" {
				t.Fatalf("unexpected data: %s", diff)
			}
		})
	}

}

func TestHoverAtPos_exprAny_literalTypes(t *testing.T) {
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
					Constraint: schema.AnyExpression{
						OfType: cty.DynamicPseudoType,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Number,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Number,
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
					Constraint: schema.AnyExpression{
						OfType: cty.String,
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
					Constraint: schema.AnyExpression{
						OfType: cty.String,
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
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = "foo${bar}"`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
					End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
				},
			},
		},

		// list
		{
			"list with any element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.DynamicPseudoType),
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
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.Number),
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
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Set(cty.DynamicPseudoType),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Set(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Set(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Set(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Set(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Set(cty.Number),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Set(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{}),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{}),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{
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
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{
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
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{
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
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{
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
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{
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
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{
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
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.DynamicPseudoType),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.DynamicPseudoType),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{}),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{}),
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
					Constraint: schema.AnyExpression{
						OfType: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
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
					Constraint: schema.AnyExpression{
						OfType: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
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
					Constraint: schema.AnyExpression{
						OfType: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
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
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
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
					Constraint: schema.AnyExpression{
						OfType: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
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
					Constraint: schema.AnyExpression{
						OfType: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
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
					Constraint: schema.AnyExpression{
						OfType: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
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
					Constraint: schema.AnyExpression{
						OfType: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
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
					Constraint: schema.AnyExpression{
						OfType: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
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

func TestHoverAtPos_exprAny_references(t *testing.T) {
	testCases := []struct {
		testName          string
		attrSchema        map[string]*schema.AttributeSchema
		refOrigins        reference.Origins
		refTargets        reference.Targets
		cfg               string
		pos               hcl.Pos
		expectedHoverData *lang.HoverData
	}{
		{
			"unknown origin",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "d"},
						lang.AttrStep{Name: "fx"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 13, Byte: 12},
						End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.String,
						},
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 13, Byte: 29},
					},
				},
			},
			`attr = l.ca+d.fx
foo = "noot"
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			nil,
		},
		{
			"matching origin no target",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.String,
						},
					},
				},
			},
			reference.Targets{},
			`attr = local.foo
foo = "noot"
`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			nil,
		},
		{
			"matching origin and target",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.String,
						},
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 13, Byte: 29},
					},
				},
			},
			`attr = local.foo
foo = "noot"
`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			&lang.HoverData{
				Content: lang.Markdown("`local.foo`\n_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
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
				ReferenceOrigins: tc.refOrigins,
				ReferenceTargets: tc.refTargets,
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

func TestHoverAtPos_exprAny_operators(t *testing.T) {
	testCases := []struct {
		testName          string
		attrSchema        map[string]*schema.AttributeSchema
		cfg               string
		pos               hcl.Pos
		expectedHoverData *lang.HoverData
	}{
		{
			"binary operator LHS",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Number,
					},
				},
			},
			`attr = 42 + 43
`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_number_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"binary operator RHS",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Number,
					},
				},
			},
			`attr = 42 + 43
`,
			hcl.Pos{Line: 1, Column: 14, Byte: 13},
			&lang.HoverData{
				Content: lang.Markdown("_number_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 13, Byte: 12},
					End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
				},
			},
		},
		{
			"binary operator mismatching constraint",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Number,
					},
				},
			},
			`attr = true || false
`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			nil,
		},
		{
			"unary operator",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
					},
				},
			},
			`attr = !true
`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("_bool_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
					End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
				},
			},
		},
		{
			"unary operator mismatching constraint",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Number,
					},
				},
			},
			`attr = !true
`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			nil,
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

func TestHoverAtPos_exprAny_parenthesis(t *testing.T) {
	testCases := []struct {
		testName          string
		attrSchema        map[string]*schema.AttributeSchema
		cfg               string
		pos               hcl.Pos
		expectedHoverData *lang.HoverData
	}{
		{
			"number in parenthesis",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Number,
					},
				},
			},
			`attr = (42+3)*2
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			&lang.HoverData{
				Content: lang.Markdown("_number_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
					End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
				},
			},
		},
		{
			"bool in parenthesis",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
					},
				},
			},
			`attr = (true || false) && true
`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("_bool_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
					End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
				},
			},
		},
		{
			"mismatched constraints",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Number,
					},
				},
			},
			`attr = (true || false) && true
`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			nil,
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

func TestHoverAtPos_exprAny_templates(t *testing.T) {
	testCases := []struct {
		testName          string
		attrSchema        map[string]*schema.AttributeSchema
		refOrigins        reference.Origins
		refTargets        reference.Targets
		cfg               string
		pos               hcl.Pos
		expectedHoverData *lang.HoverData
	}{
		{
			"expression string",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{},
			reference.Targets{},
			`attr = "foo-${"bar"}-bar"
`,
			hcl.Pos{Line: 1, Column: 17, Byte: 16},
			&lang.HoverData{
				Content: lang.Markdown("_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
					End:      hcl.Pos{Line: 1, Column: 20, Byte: 19},
				},
			},
		},
		{
			"expression with reference",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
						End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.String,
						},
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 13, Byte: 29},
					},
				},
			},
			`attr = "foo-${local.foo}-bar"
`,
			hcl.Pos{Line: 1, Column: 17, Byte: 16},
			&lang.HoverData{
				Content: lang.Markdown("`local.foo`\n_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
					End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
				},
			},
		},
		{
			"heredoc with reference",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 19},
						End:      hcl.Pos{Line: 3, Column: 12, Byte: 28},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.String,
						},
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 5, Column: 1, Byte: 34},
						End:      hcl.Pos{Line: 5, Column: 13, Byte: 45},
					},
				},
			},
			`attr = <<EOT
foo
${local.foo}
bar
EOT
`,
			hcl.Pos{Line: 2, Column: 2, Byte: 14},
			&lang.HoverData{
				Content: lang.Markdown("_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 1, Byte: 13},
					End:      hcl.Pos{Line: 3, Column: 1, Byte: 17},
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
				ReferenceOrigins: tc.refOrigins,
				ReferenceTargets: tc.refTargets,
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

func TestHoverAtPos_exprAny_conditional(t *testing.T) {
	testCases := []struct {
		testName          string
		attrSchema        map[string]*schema.AttributeSchema
		refOrigins        reference.Origins
		refTargets        reference.Targets
		cfg               string
		pos               hcl.Pos
		expectedHoverData *lang.HoverData
	}{
		{
			"condition",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{},
			reference.Targets{},
			`attr = true ? "bar" : "baz"
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			&lang.HoverData{
				Content: lang.Markdown("_bool_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
				},
			},
		},
		{
			"true",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{},
			reference.Targets{},
			`attr = true ? 4223 : "baz"
`,
			hcl.Pos{Line: 1, Column: 17, Byte: 16},
			&lang.HoverData{
				Content: lang.Markdown("_number_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
					End:      hcl.Pos{Line: 1, Column: 19, Byte: 18},
				},
			},
		},
		{
			"false",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{},
			reference.Targets{},
			`attr = true ? 4223 : 5713
`,
			hcl.Pos{Line: 1, Column: 24, Byte: 23},
			&lang.HoverData{
				Content: lang.Markdown("_number_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 22, Byte: 21},
					End:      hcl.Pos{Line: 1, Column: 26, Byte: 25},
				},
			},
		},
		{
			"condition in template",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{},
			reference.Targets{},
			`attr = "x${true ? 4223 : 5713}"
`,
			hcl.Pos{Line: 1, Column: 14, Byte: 13},
			&lang.HoverData{
				Content: lang.Markdown("_bool_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 12, Byte: 11},
					End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
				},
			},
		},
		{
			"condition as directive in template",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{},
			reference.Targets{},
			`attr = "x%{if true}4223%{else}5713%{endif}"
`,
			hcl.Pos{Line: 1, Column: 17, Byte: 16},
			&lang.HoverData{
				Content: lang.Markdown("_bool_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
					End:      hcl.Pos{Line: 1, Column: 19, Byte: 18},
				},
			},
		},
		{
			"condition as directive in template inside true",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{},
			reference.Targets{},
			`attr = "x%{if true}4223%{else}5713%{endif}"
`,
			hcl.Pos{Line: 1, Column: 22, Byte: 21},
			&lang.HoverData{
				Content: lang.Markdown("_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 20, Byte: 19},
					End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
				},
			},
		},
		{
			"condition as directive in template inside false",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{},
			reference.Targets{},
			`attr = "x%{if true}4223%{else}5713%{endif}"
`,
			hcl.Pos{Line: 1, Column: 33, Byte: 32},
			&lang.HoverData{
				Content: lang.Markdown("_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 31, Byte: 30},
					End:      hcl.Pos{Line: 1, Column: 35, Byte: 34},
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
				ReferenceOrigins: tc.refOrigins,
				ReferenceTargets: tc.refTargets,
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
