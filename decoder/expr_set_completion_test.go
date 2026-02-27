// Copyright IBM Corp. 2020, 2026
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

func TestCompletionAtPos_exprSet(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"attribute as set expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.LiteralType{
							Type: cty.String,
						},
					},
				},
			},
			`
`,
			hcl.Pos{Line: 1, Column: 1, Byte: 0},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `attr`,
					Detail: "set of string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:      hcl.Pos{Line: 1, Column: 1, Byte: 0},
						},
						NewText: "attr",
						Snippet: "attr = [ \"${1:value}\" ]",
					},
					Kind:           lang.AttributeCandidateKind,
					TriggerSuggest: false,
				},
			}),
		},
		{
			"empty expression no element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `[ ]`,
					Detail: "set",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						NewText: "[ ]",
						Snippet: "[ ${1} ]",
					},
					Kind: lang.SetCandidateKind,
				},
			}),
		},
		{
			"empty expression with element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.LiteralType{
							Type: cty.String,
						},
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `[ string ]`,
					Detail: "set of string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						NewText: "[ \"value\" ]",
						Snippet: "[ \"${1:value}\" ]",
					},
					Kind:           lang.SetCandidateKind,
					TriggerSuggest: false,
				},
			}),
		},

		// single line tests
		{
			"inside brackets single-line",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = [  ]
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
							End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"inside single-line partial element near end",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = [ key ]
`,
			hcl.Pos{Line: 1, Column: 13, Byte: 12},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
							End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			// TODO: revisit if we allow only unique elements in sets
			"inside single-line complete element in the middle",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = [ keyword, ]
`,
			hcl.Pos{Line: 1, Column: 13, Byte: 12},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
							End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			// TODO: revisit if we allow only unique elements in sets
			"inside single-line after previous element with comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = [ keyword, ]
`,
			hcl.Pos{Line: 1, Column: 19, Byte: 18},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 19, Byte: 18},
							End:      hcl.Pos{Line: 1, Column: 19, Byte: 18},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			// TODO: revisit if we allow only unique elements in sets
			"inside single-line after previous element without comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = [ keyword  ]
`,
			hcl.Pos{Line: 1, Column: 19, Byte: 18},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 19, Byte: 18},
							End:      hcl.Pos{Line: 1, Column: 19, Byte: 18},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			// TODO: revisit if we allow only unique elements in sets
			"inside single-line between elements with commas",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = [ keyword,  , keyword ]
`,
			hcl.Pos{Line: 1, Column: 19, Byte: 18},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 19, Byte: 18},
							End:      hcl.Pos{Line: 1, Column: 19, Byte: 18},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},

		// multi line tests
		{
			"inside brackets multi-line",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = [
  
]
`,
			hcl.Pos{Line: 2, Column: 3, Byte: 11},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
							End:      hcl.Pos{Line: 2, Column: 3, Byte: 11},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"inside multi-line partial element near end",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = [
  key
]
`,
			hcl.Pos{Line: 2, Column: 6, Byte: 14},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
							End:      hcl.Pos{Line: 2, Column: 6, Byte: 14},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			// TODO: revisit if we allow only unique elements in sets
			"inside multi-line complete element in the middle",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = [
  keyword,
]
`,
			hcl.Pos{Line: 2, Column: 6, Byte: 14},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
							End:      hcl.Pos{Line: 2, Column: 10, Byte: 18},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			// TODO: revisit if we allow only unique elements in sets
			"inside multi-line new line before existing element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = [
  
  keyword,
]
`,
			hcl.Pos{Line: 2, Column: 3, Byte: 11},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
							End:      hcl.Pos{Line: 2, Column: 3, Byte: 11},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			// TODO: revisit if we allow only unique elements in sets
			"inside multi-line partial element before existing element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = [
  key
  keyword,
]
`,
			hcl.Pos{Line: 2, Column: 6, Byte: 14},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
							End:      hcl.Pos{Line: 2, Column: 6, Byte: 14},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			// TODO: revisit if we allow only unique elements in sets
			"inside multi-line after previous element with comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = [
  keyword,
  
]
`,
			hcl.Pos{Line: 3, Column: 3, Byte: 22},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 22},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 22},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			// TODO: revisit if we allow only unique elements in sets
			"inside multi-line after previous element without comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = [
  keyword
  
]
`,
			hcl.Pos{Line: 3, Column: 3, Byte: 21},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 21},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 21},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			// TODO: revisit if we allow only unique elements in sets
			"inside multi-line between elements with commas",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = [
  keyword,
  
  keyword,
]
`,
			hcl.Pos{Line: 3, Column: 3, Byte: 22},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 22},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 22},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			// TODO: revisit if we allow only unique elements in sets
			"inside multi-line after previous element with comma same line",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = [
  keyword, 
]
`,
			hcl.Pos{Line: 2, Column: 12, Byte: 20},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 12, Byte: 20},
							End:      hcl.Pos{Line: 2, Column: 12, Byte: 20},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
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
			})

			ctx := context.Background()
			candidates, err := d.CompletionAtPos(ctx, "test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedCandidates, candidates); diff != "" {
				t.Logf("position: %#v in config: %s", tc.pos, tc.cfg)
				t.Fatalf("unexpected candidates: %s", diff)
			}
		})
	}
}

func TestCompletionAtPos_exprSet_references(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		refTargets         reference.Targets
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"single-line with trailing dot",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.Reference{OfScopeId: lang.ScopeId("variable")},
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "bar"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					ScopeId: lang.ScopeId("variable"),
				},
			},
			`attr = [ var. ]
`,
			hcl.Pos{Line: 1, Column: 14, Byte: 13},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.bar",
					Detail: "reference",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.bar",
						Snippet: "var.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
							End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
						},
					},
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
			})

			ctx := context.Background()
			candidates, err := d.CompletionAtPos(ctx, "test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedCandidates, candidates); diff != "" {
				t.Logf("position: %#v in config: %s", tc.pos, tc.cfg)
				t.Fatalf("unexpected candidates: %s", diff)
			}
		})
	}
}
