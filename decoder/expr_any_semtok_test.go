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
