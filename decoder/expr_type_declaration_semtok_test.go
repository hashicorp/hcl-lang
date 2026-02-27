// Copyright IBM Corp. 2020, 2026
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
)

func TestSemanticTokens_exprTypeDeclaration(t *testing.T) {
	testCases := []struct {
		testName       string
		attrSchema     map[string]*schema.AttributeSchema
		cfg            string
		expectedTokens []lang.SemanticToken
	}{
		{
			"primitive type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = string`,
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
					Type:      lang.TokenTypePrimitive,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
					},
				},
			},
		},
		{
			"invalid primitive type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = foobar`,
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
			"single-argument complex type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = list(string)`,
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
					Type:      lang.TokenTypeComplex,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
					},
				},
				{
					Type:      lang.TokenTypePrimitive,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 13, Byte: 12},
						End:      hcl.Pos{Line: 1, Column: 19, Byte: 18},
					},
				},
			},
		},
		{
			"tuple",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = tuple([string, bool, number])`,
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
					Type:      lang.TokenTypeComplex,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
					},
				},
				{
					Type:      lang.TokenTypePrimitive,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
						End:      hcl.Pos{Line: 1, Column: 21, Byte: 20},
					},
				},
				{
					Type:      lang.TokenTypePrimitive,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
						End:      hcl.Pos{Line: 1, Column: 27, Byte: 26},
					},
				},
				{
					Type:      lang.TokenTypePrimitive,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 29, Byte: 28},
						End:      hcl.Pos{Line: 1, Column: 35, Byte: 34},
					},
				},
			},
		},
		{
			"object",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  foo = string
  bar = number
})`,
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
					Type:      lang.TokenTypeComplex,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
					},
				},
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 18},
						End:      hcl.Pos{Line: 2, Column: 6, Byte: 21},
					},
				},
				{
					Type:      lang.TokenTypePrimitive,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 9, Byte: 24},
						End:      hcl.Pos{Line: 2, Column: 15, Byte: 30},
					},
				},
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 33},
						End:      hcl.Pos{Line: 3, Column: 6, Byte: 36},
					},
				},
				{
					Type:      lang.TokenTypePrimitive,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 9, Byte: 39},
						End:      hcl.Pos{Line: 3, Column: 15, Byte: 45},
					},
				},
			},
		},
		{
			"object with complex types",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  foo = list(string)
  bar = tuple([bool, string])
  baz = object({
    paw = string
  })
})`,
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
					Type:      lang.TokenTypeComplex,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
					},
				},
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 18},
						End:      hcl.Pos{Line: 2, Column: 6, Byte: 21},
					},
				},
				{
					Type:      lang.TokenTypeComplex,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 9, Byte: 24},
						End:      hcl.Pos{Line: 2, Column: 13, Byte: 28},
					},
				},
				{
					Type:      lang.TokenTypePrimitive,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Start:    hcl.Pos{Line: 2, Column: 14, Byte: 29},
						End:      hcl.Pos{Line: 2, Column: 20, Byte: 35},
						Filename: "test.tf",
					},
				},
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 39},
						End:      hcl.Pos{Line: 3, Column: 6, Byte: 42},
						Filename: "test.tf",
					},
				},
				{
					Type:      lang.TokenTypeComplex,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Start:    hcl.Pos{Line: 3, Column: 9, Byte: 45},
						End:      hcl.Pos{Line: 3, Column: 14, Byte: 50},
						Filename: "test.tf",
					},
				},
				{
					Type:      lang.TokenTypePrimitive,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Start:    hcl.Pos{Line: 3, Column: 16, Byte: 52},
						End:      hcl.Pos{Line: 3, Column: 20, Byte: 56},
						Filename: "test.tf",
					},
				},
				{
					Type:      lang.TokenTypePrimitive,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 22, Byte: 58},
						End:      hcl.Pos{Line: 3, Column: 28, Byte: 64},
					},
				},
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 4, Column: 3, Byte: 69},
						End:      hcl.Pos{Line: 4, Column: 6, Byte: 72},
					},
				},
				{
					Type:      lang.TokenTypeComplex,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 4, Column: 9, Byte: 75},
						End:      hcl.Pos{Line: 4, Column: 15, Byte: 81},
					},
				},
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 5, Column: 5, Byte: 88},
						End:      hcl.Pos{Line: 5, Column: 8, Byte: 91},
					},
				},
				{
					Type:      lang.TokenTypePrimitive,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 5, Column: 11, Byte: 94},
						End:      hcl.Pos{Line: 5, Column: 17, Byte: 100},
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
			})

			ctx := context.Background()
			hoverData, err := d.SemanticTokensInFile(ctx, "test.tf")
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedTokens, hoverData); diff != "" {
				t.Fatalf("unexpected tokens: %s", diff)
			}
		})
	}
}
