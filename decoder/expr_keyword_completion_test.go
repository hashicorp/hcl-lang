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

func TestCompletionAtPos_exprKeyword(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"no expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Keyword{Keyword: "foobar"},
				},
			},
			`attr = `,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "foobar",
					Detail: "keyword",
					Kind:   lang.KeywordCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "foobar",
						Snippet: "foobar",
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
					Constraint: schema.Keyword{Keyword: "foobar"},
				},
			},
			`attr = f`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "foobar",
					Detail: "keyword",
					Kind:   lang.KeywordCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "foobar",
						Snippet: "foobar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
			}),
		},
		{
			"matching prefix with dot",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Keyword{Keyword: "foobar"},
				},
			},
			`attr = f.`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"matching prefix in the middle",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Keyword{Keyword: "foobar"},
				},
			},
			`attr = foo`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "foobar",
					Detail: "keyword",
					Kind:   lang.KeywordCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "foobar",
						Snippet: "foobar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
						},
					},
				},
			}),
		},
		{
			"mismatching prefix",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Keyword{Keyword: "foobar"},
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
