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
)

func TestSemanticTokens_exprOneOf(t *testing.T) {
	testCases := []struct {
		testName       string
		attrSchema     map[string]*schema.AttributeSchema
		cfg            string
		expectedTokens []lang.SemanticToken
	}{
		{
			"matching first expr",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.OneOf{
						schema.Keyword{
							Keyword: "cpu",
						},
						schema.Keyword{
							Keyword: "memory",
						},
						schema.Keyword{
							Keyword: "storage",
						},
					},
				},
			},
			`attr = cpu`,
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
					Type:      lang.TokenKeyword,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
					},
				},
			},
		},
		{
			"matching second expr",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.OneOf{
						schema.Keyword{
							Keyword: "cpu",
						},
						schema.Keyword{
							Keyword: "memory",
						},
						schema.Keyword{
							Keyword: "storage",
						},
					},
				},
			},
			`attr = memory`,
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
					Type:      lang.TokenKeyword,
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
			"no matching expr",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.OneOf{
						schema.Keyword{
							Keyword: "cpu",
						},
						schema.Keyword{
							Keyword: "memory",
						},
						schema.Keyword{
							Keyword: "storage",
						},
					},
				},
			},
			`attr = unknown`,
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
			"no expr defined",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.OneOf{},
				},
			},
			`attr = keyword1`,
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
