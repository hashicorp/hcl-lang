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
