// Copyright IBM Corp. 2026
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

func TestCompletionAtPos_exprAny_index(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		refTargets         reference.Targets
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"empty complex index",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_instance"},
						lang.AttrStep{Name: "name"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"tags": cty.List(cty.String),
					}, []string{"tags"}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_instance"},
								lang.AttrStep{Name: "name"},
								lang.AttrStep{Name: "tags"},
							},
							RangePtr: &hcl.Range{
								Filename: "variables.tf",
								Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
								End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
							},
							Type: cty.List(cty.String),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws_instance"},
										lang.AttrStep{Name: "name"},
										lang.AttrStep{Name: "tags"},
										lang.IndexStep{Key: cty.StringVal("name")},
									},
									RangePtr: &hcl.Range{
										Filename: "variables.tf",
										Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
										End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
									},
									Type: cty.String,
								},
							},
						},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "name"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.String,
				},
			},
			`attr = aws_instance.name.tags[]
`,
			hcl.Pos{Line: 1, Column: 31, Byte: 30},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `aws_instance.name.tags["name"]`,
					Detail: "string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 32, Byte: 31},
						},
						NewText: `aws_instance.name.tags["name"]`,
						Snippet: `aws_instance.name.tags["name"]`,
					},
					Kind: lang.ReferenceCandidateKind,
				},
				{
					Label:  `aws_instance.name`,
					Detail: "object",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 31, Byte: 30},
							End:      hcl.Pos{Line: 1, Column: 31, Byte: 30},
						},
						NewText: `aws_instance.name`,
						Snippet: `aws_instance.name`,
					},
					Kind: lang.ReferenceCandidateKind,
				},
				{
					Label:  `local.name`,
					Detail: "string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 31, Byte: 30},
							End:      hcl.Pos{Line: 1, Column: 31, Byte: 30},
						},
						NewText: `local.name`,
						Snippet: `local.name`,
					},
					Kind: lang.ReferenceCandidateKind,
				},
			}),
		},
		{
			"prefix complex index",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_instance"},
						lang.AttrStep{Name: "name"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"tags": cty.List(cty.String),
					}, []string{"tags"}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_instance"},
								lang.AttrStep{Name: "name"},
								lang.AttrStep{Name: "tags"},
							},
							RangePtr: &hcl.Range{
								Filename: "variables.tf",
								Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
								End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
							},
							Type: cty.List(cty.String),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws_instance"},
										lang.AttrStep{Name: "name"},
										lang.AttrStep{Name: "tags"},
										lang.IndexStep{Key: cty.StringVal("name")},
									},
									RangePtr: &hcl.Range{
										Filename: "variables.tf",
										Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
										End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
									},
									Type: cty.String,
								},
							},
						},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "name"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.String,
				},
			},
			`attr = aws_instance.name.tags[l]
`,
			hcl.Pos{Line: 1, Column: 32, Byte: 31},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `local.name`,
					Detail: "string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 31, Byte: 30},
							End:      hcl.Pos{Line: 1, Column: 32, Byte: 31},
						},
						NewText: `local.name`,
						Snippet: `local.name`,
					},
					Kind: lang.ReferenceCandidateKind,
				},
			}),
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
				ReferenceTargets: tc.refTargets,
				// Functions:        testFunctionSignatures(),
			})

			ctx := context.Background()
			candidates, err := d.CompletionAtPos(ctx, "test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedCandidates, candidates); diff != "" {
				t.Fatalf("unexpected candidates: %s", diff)
			}
		})
	}
}

func TestHoverAtPos_exprAny_index(t *testing.T) {
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
			"empty index",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{},
			reference.Targets{},
			`attr = aws_instance.name.tags[]
`,
			hcl.Pos{Line: 1, Column: 31, Byte: 30},
			nil,
		},
		{
			"local reference in index",
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
						lang.AttrStep{Name: "name"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 31, Byte: 30},
						End:      hcl.Pos{Line: 1, Column: 41, Byte: 40},
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
						lang.AttrStep{Name: "name"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.String,
				},
			},
			`attr = aws_instance.name.tags[local.name]
`,
			hcl.Pos{Line: 1, Column: 35, Byte: 34},
			&lang.HoverData{
				Content: lang.Markdown("`local.name`\n_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 31, Byte: 30},
					End:      hcl.Pos{Line: 1, Column: 41, Byte: 40},
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

func TestSemanticTokens_exprAny_index(t *testing.T) {
	testCases := []struct {
		testName               string
		attrSchema             map[string]*schema.AttributeSchema
		refOrigins             reference.Origins
		refTargets             reference.Targets
		cfg                    string
		expectedSemanticTokens []lang.SemanticToken
	}{
		{
			"simple conditional",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{},
			reference.Targets{},
			`attr = true ? "t" : 422
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
					Type:      lang.TokenBool,
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
						Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
						End:      hcl.Pos{Line: 1, Column: 18, Byte: 17},
					},
				},
				{
					Type:      lang.TokenNumber,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 21, Byte: 20},
						End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
					},
				},
			},
		},
		{
			"empty index",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{},
			reference.Targets{},
			`attr = aws_instance.name.tags[local.name]
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
			"local reference in index",
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
						lang.AttrStep{Name: "name"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 31, Byte: 30},
						End:      hcl.Pos{Line: 1, Column: 41, Byte: 40},
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
						lang.AttrStep{Name: "name"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.String,
				},
			},
			`attr = aws_instance.name.tags[local.name]
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
					Type:      lang.TokenReferenceStep,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 31, Byte: 30},
						End:      hcl.Pos{Line: 1, Column: 36, Byte: 35},
					},
				},
				{
					Type:      lang.TokenReferenceStep,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 37, Byte: 36},
						End:      hcl.Pos{Line: 1, Column: 41, Byte: 40},
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
