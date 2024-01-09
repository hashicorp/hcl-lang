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

func TestCompletionAtPos_exprTuple(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"attribute as tuple expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.LiteralType{
								Type: cty.String,
							},
							schema.LiteralType{
								Type: cty.Bool,
							},
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
					Detail: "tuple",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:      hcl.Pos{Line: 1, Column: 1, Byte: 0},
						},
						NewText: "attr",
						Snippet: "attr = [ \"${1:value}\", ${2:false} ]",
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
					Constraint: schema.Tuple{},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `[ ]`,
					Detail: "tuple",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						NewText: "[ ]",
						Snippet: "[ ${1} ]",
					},
					Kind: lang.TupleCandidateKind,
				},
			}),
		},
		{
			"empty expression with elements",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.LiteralType{
								Type: cty.String,
							},
							schema.LiteralType{
								Type: cty.Bool,
							},
						},
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `[ string, bool ]`,
					Detail: "tuple",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						NewText: `[ "value", false ]`,
						Snippet: `[ "${1:value}", ${2:false} ]`,
					},
					Kind:           lang.TupleCandidateKind,
					TriggerSuggest: false,
				},
			}),
		},
		{
			"empty expression with many elements",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.LiteralType{
								Type: cty.String,
							},
							schema.LiteralType{
								Type: cty.Bool,
							},
							schema.LiteralType{
								Type: cty.Number,
							},
							schema.LiteralType{
								Type: cty.String,
							},
						},
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `[ string, bool, numbe… ]`,
					Detail: "tuple",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						NewText: `[ "value", false, 0, "value" ]`,
						Snippet: `[ "${1:value}", ${2:false}, ${3:0}, "${4:value}" ]`,
					},
					Kind:           lang.TupleCandidateKind,
					TriggerSuggest: false,
				},
			}),
		},

		// single line tests
		{
			"inside brackets single-line",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
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
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
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
			"inside single-line complete element in the middle",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
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
			"inside single-line before existing element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
						},
					},
				},
			},
			`attr = [  keyword, ]
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"inside single-line after previous element with comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
							schema.Keyword{
								Keyword: "drowyek",
							},
						},
					},
				},
			},
			`attr = [ keyword, ]
`,
			hcl.Pos{Line: 1, Column: 19, Byte: 18},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `drowyek`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 19, Byte: 18},
							End:      hcl.Pos{Line: 1, Column: 19, Byte: 18},
						},
						NewText: `drowyek`,
						Snippet: `drowyek`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"inside single-line after previous element without comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
							schema.Keyword{
								Keyword: "drowyek",
							},
						},
					},
				},
			},
			`attr = [ keyword  ]
`,
			hcl.Pos{Line: 1, Column: 19, Byte: 18},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"inside single-line between elements with commas",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
							schema.Keyword{
								Keyword: "drowyek",
							},
							schema.Keyword{
								Keyword: "valid",
							},
						},
					},
				},
			},
			`attr = [ keyword,  , valid ]
`,
			hcl.Pos{Line: 1, Column: 19, Byte: 18},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `drowyek`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 19, Byte: 18},
							End:      hcl.Pos{Line: 1, Column: 19, Byte: 18},
						},
						NewText: `drowyek`,
						Snippet: `drowyek`,
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
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
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
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
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
			"inside multi-line complete element in the middle",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
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
			"inside multi-line same line before existing element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
						},
					},
				},
			},
			`attr = [
  keyword,
]
`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"inside multi-line new line before existing element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "drowyek",
							},
							schema.Keyword{
								Keyword: "keyword",
							},
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
					Label:  `drowyek`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
							End:      hcl.Pos{Line: 2, Column: 3, Byte: 11},
						},
						NewText: `drowyek`,
						Snippet: `drowyek`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"inside multi-line partial element before existing element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "drowyek",
							},
							schema.Keyword{
								Keyword: "keyword",
							},
						},
					},
				},
			},
			`attr = [
  dro
  keyword,
]
`,
			hcl.Pos{Line: 2, Column: 6, Byte: 14},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `drowyek`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
							End:      hcl.Pos{Line: 2, Column: 6, Byte: 14},
						},
						NewText: `drowyek`,
						Snippet: `drowyek`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"inside multi-line after previous element with comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
							schema.Keyword{
								Keyword: "drowyek",
							},
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
					Label:  `drowyek`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 22},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 22},
						},
						NewText: `drowyek`,
						Snippet: `drowyek`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"inside multi-line after previous element without comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
							schema.Keyword{
								Keyword: "drowyek",
							},
						},
					},
				},
			},
			`attr = [
  keyword
  
]
`,
			hcl.Pos{Line: 3, Column: 3, Byte: 21},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"inside multi-line between elements with commas",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
							schema.Keyword{
								Keyword: "drowyek",
							},
							schema.Keyword{
								Keyword: "valid",
							},
						},
					},
				},
			},
			`attr = [
  keyword,
  
  valid,
]
`,
			hcl.Pos{Line: 3, Column: 3, Byte: 22},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `drowyek`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 22},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 22},
						},
						NewText: `drowyek`,
						Snippet: `drowyek`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"inside multi-line after previous element with comma same line",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
							schema.Keyword{
								Keyword: "drowyek",
							},
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
					Label:  `drowyek`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 12, Byte: 20},
							End:      hcl.Pos{Line: 2, Column: 12, Byte: 20},
						},
						NewText: `drowyek`,
						Snippet: `drowyek`,
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

func TestCompletionAtPos_exprTuple_references(t *testing.T) {
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
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Reference{OfScopeId: lang.ScopeId("variable")},
						},
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
