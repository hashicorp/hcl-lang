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

func TestSemanticTokens_exprAny_functions(t *testing.T) {
	testCases := []struct {
		testName               string
		attrSchema             map[string]*schema.AttributeSchema
		cfg                    string
		expectedSemanticTokens []lang.SemanticToken
	}{
		{
			"unknown function",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = unknown()
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
			},
		},
		{
			"function without arguments",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = lower()
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
					Type:      lang.TokenFunctionName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
					},
				},
			},
		},
		{
			"function with arguments",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = join(",", ["a", "b"])
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
					Type:      lang.TokenFunctionName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 13, Byte: 12},
						End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 19, Byte: 18},
						End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 24, Byte: 23},
						End:      hcl.Pos{Line: 1, Column: 27, Byte: 26},
					},
				},
			},
		},
		{
			"function with too many arguments",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = lower("FOO", "BAR")
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
					Type:      lang.TokenFunctionName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 14, Byte: 13},
						End:      hcl.Pos{Line: 1, Column: 19, Byte: 18},
					},
				},
			},
		},
		{
			"function with too mismatching arguments",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = lower(["FOO"])
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
					Type:      lang.TokenFunctionName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
					},
				},
			},
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

func TestSemanticTokens_exprAny_literalTypes(t *testing.T) {
	testCases := []struct {
		testName               string
		attrSchema             map[string]*schema.AttributeSchema
		cfg                    string
		expectedSemanticTokens []lang.SemanticToken
	}{
		{
			"any type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.DynamicPseudoType,
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

		// bool
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
					Constraint: schema.AnyExpression{
						OfType: cty.String,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Number,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Number,
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
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.DynamicPseudoType),
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
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
					},
				},
			},
			`attr = ["foobar"]`,
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
						End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
					},
				},
			},
		},
		{
			"single element multi-line list",
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
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
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
			"multi-element multi-line list with invalid element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
					},
				},
			},
			`attr = [
  "fooba",
  invalid,
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

		// set
		{
			"empty set with any element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Set(cty.DynamicPseudoType),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Set(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Set(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Set(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Set(cty.String),
					},
				},
			},
			`attr = [
  "fooba",
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
					Constraint: schema.AnyExpression{
						OfType: cty.Set(cty.String),
					},
				},
			},
			`attr = [
  "fooba",
  invalid,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{}),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{
							cty.String,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{
							cty.String,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{
							cty.String,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{
							cty.String,
							cty.Number,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{
							cty.String,
							cty.String,
							cty.Number,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.DynamicPseudoType),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
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

		// object
		{
			"undefined attributes",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{}),
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
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.String,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.String,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.String,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.String,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.String,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.String,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.Number,
							"bar": cty.String,
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
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.String,
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
			},
		},
		{
			"multi-line with nested object",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.Object(map[string]cty.Type{
								"noot": cty.String,
							}),
							"bar": cty.String,
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

func TestSemanticTokens_exprAny_references(t *testing.T) {
	testCases := []struct {
		testName               string
		attrSchema             map[string]*schema.AttributeSchema
		refOrigins             reference.Origins
		refTargets             reference.Targets
		cfg                    string
		expectedSemanticTokens []lang.SemanticToken
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
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "foo"},
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
			`attr = local.foox
foo = "noot"
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
			},
		},
		{
			"matching origin with no target",
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
					Type:      lang.TokenTraversalStep,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
					},
				},
				{
					Type:      lang.TokenTraversalStep,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 14, Byte: 13},
						End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
					},
				},
			},
		},
		{
			"matching reference with numerical index",
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
						lang.IndexStep{Key: cty.NumberIntVal(42)},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 21, Byte: 20},
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
						lang.IndexStep{Key: cty.NumberIntVal(42)},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 21},
						End:      hcl.Pos{Line: 2, Column: 13, Byte: 33},
					},
				},
			},
			`attr = local.foo[42]
foo = "noot"
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
					Type:      lang.TokenTraversalStep,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
					},
				},
				{
					Type:      lang.TokenTraversalStep,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 14, Byte: 13},
						End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
					},
				},
				{
					Type:      lang.TokenNumber,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 18, Byte: 17},
						End:      hcl.Pos{Line: 1, Column: 20, Byte: 19},
					},
				},
			},
		},
		{
			"matching reference with string index",
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
						lang.IndexStep{Key: cty.StringVal("bar")},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
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
						lang.IndexStep{Key: cty.StringVal("bar")},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 24},
						End:      hcl.Pos{Line: 2, Column: 13, Byte: 36},
					},
				},
			},
			`attr = local.foo["bar"]
foo = "noot"
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
					Type:      lang.TokenTraversalStep,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
					},
				},
				{
					Type:      lang.TokenTraversalStep,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 14, Byte: 13},
						End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 18, Byte: 17},
						End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
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
