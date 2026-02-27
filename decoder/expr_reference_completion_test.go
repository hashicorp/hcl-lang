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

func TestCompletionAtPos_exprReference(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		refTargets         reference.Targets
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"no expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
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
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.List(cty.Number),
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "baz"},
					},
					Type: cty.Number,
				},
			},
			`attr = `,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "local.foo",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "local.foo",
						Snippet: "local.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
				{
					Label:  "local.baz",
					Detail: "number",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "local.baz",
						Snippet: "local.baz",
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
			"matching prefix",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.Number,
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.List(cty.String),
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.Number,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "data"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.Number,
				},
			},
			`attr = local`,
			hcl.Pos{Line: 1, Column: 13, Byte: 12},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "local.bar",
					Detail: "number",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "local.bar",
						Snippet: "local.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
						},
					},
				},
			}),
		},
		{
			"matching prefix in the middle",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
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
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.List(cty.Number),
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "data"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.Number,
				},
			},
			`attr = local`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "local.foo",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "local.foo",
						Snippet: "local.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
						},
					},
				},
			}),
		},
		{
			"matching prefix after trailing dot",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
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
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.List(cty.Number),
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "data"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.Number,
				},
			},
			`attr = local.`,
			hcl.Pos{Line: 1, Column: 14, Byte: 13},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "local.foo",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "local.foo",
						Snippet: "local.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
						},
					},
				},
			}),
		},
		{
			"mismatching prefix",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.Number,
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
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.Number,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "data"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.Number,
				},
			},
			`attr = x`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			lang.CompleteCandidates([]lang.Candidate{}),
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
				ReferenceTargets: tc.refTargets,
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
