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
	"github.com/zclconf/go-cty/cty"
)

func TestLegacyDecoder_CandidateAtPos_expressions_Hooks(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		pos                hcl.Pos
		completionHooks    CompletionFuncMap
		expectedCandidates lang.Candidates
	}{
		{
			"simple hook",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{Type: cty.String},
					CompletionHooks: lang.CompletionHooks{
						{
							Name: "TestCompletionHook",
						},
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			CompletionFuncMap{
				"TestCompletionHook": func(ctx context.Context, value cty.Value) ([]Candidate, error) {
					candidates := []Candidate{
						{
							Label:         "\"label\"",
							Detail:        "detail",
							Kind:          lang.StringCandidateKind,
							Description:   lang.PlainText("description"),
							RawInsertText: "\"insertText\"",
						},
					}
					return candidates, nil
				},
			},
			lang.IncompleteCandidates([]lang.Candidate{
				{
					Label:       "\"label\"",
					Detail:      "detail",
					Kind:        lang.StringCandidateKind,
					Description: lang.PlainText("description"),
					TextEdit: lang.TextEdit{
						NewText: "\"insertText\"",
						Snippet: "\"insertText\"",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
			}),
		},
		{
			"hook with prefix",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{Type: cty.String},
					CompletionHooks: lang.CompletionHooks{
						{
							Name: "TestCompletionHook",
						},
					},
				},
			},
			`attr = "pa"
`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			CompletionFuncMap{
				"TestCompletionHook": func(ctx context.Context, value cty.Value) ([]Candidate, error) {
					candidates := []Candidate{
						{
							Label:         value.AsString(),
							Kind:          lang.StringCandidateKind,
							RawInsertText: value.AsString(),
						},
					}
					return candidates, nil
				},
			},
			lang.IncompleteCandidates([]lang.Candidate{
				{
					Label: "pa",
					Kind:  lang.StringCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "pa",
						Snippet: "pa",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
						},
					},
				},
			}),
		},
		{
			"incomplete attr value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{Type: cty.String},
					CompletionHooks: lang.CompletionHooks{
						{
							Name: "TestCompletionHook",
						},
					},
				},
			},
			`attr = "pa

`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			CompletionFuncMap{
				"TestCompletionHook": func(ctx context.Context, value cty.Value) ([]Candidate, error) {
					candidates := []Candidate{
						{
							Label:         value.AsString(),
							Kind:          lang.StringCandidateKind,
							RawInsertText: value.AsString(),
						},
					}
					return candidates, nil
				},
			},
			lang.IncompleteCandidates([]lang.Candidate{
				{
					Label: "pa",
					Kind:  lang.StringCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "pa",
						Snippet: "pa",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
						},
					},
				},
			}),
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			bodySchema := &schema.BodySchema{
				Attributes: tc.attrSchema,
			}

			// We're ignoring diagnostics here, since some test cases may contain invalid HCL
			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			})
			for n, h := range tc.completionHooks {
				d.decoderCtx.CompletionHooks[n] = h
			}

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

func TestLegacyDecoder_CandidateAtPos_maxCandidates(t *testing.T) {
	ctx := context.Background()
	bodySchema := &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"attr": {
				Constraint: schema.LiteralType{Type: cty.String},
				CompletionHooks: lang.CompletionHooks{
					{
						Name: "TestCompletionHook50",
					},
					{
						Name: "TestCompletionHook70",
					},
				},
			},
		},
	}

	// We're ignoring diagnostics here, since our config contains invalid HCL
	f, _ := hclsyntax.ParseConfig([]byte(`attr = `), "test.tf", hcl.InitialPos)
	d := testPathDecoder(t, &PathContext{
		Schema: bodySchema,
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})
	d.decoderCtx.CompletionHooks["TestCompletionHook50"] = func(ctx context.Context, value cty.Value) ([]Candidate, error) {
		candidates := make([]Candidate, 0)
		for i := 0; i < 50; i++ {
			candidates = append(candidates, Candidate{
				Label: fmt.Sprintf("\"Label %d\"", i),
				Kind:  lang.StringCandidateKind,
			})
		}
		return candidates, nil
	}
	d.decoderCtx.CompletionHooks["TestCompletionHook70"] = func(ctx context.Context, value cty.Value) ([]Candidate, error) {
		candidates := make([]Candidate, 0)
		for i := 0; i < 70; i++ {
			candidates = append(candidates, Candidate{
				Label: fmt.Sprintf("\"Label %d\"", i),
				Kind:  lang.StringCandidateKind,
			})
		}
		return candidates, nil
	}

	candidates, err := d.CompletionAtPos(ctx, "test.tf", hcl.Pos{Line: 1, Column: 8, Byte: 7})
	if err != nil {
		t.Fatal(err)
	}
	count := len(candidates.List)

	if uint(count) != d.maxCandidates {
		t.Fatalf("unexpected candidates count: %d", count)
	}
}
