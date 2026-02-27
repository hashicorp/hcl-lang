// Copyright IBM Corp. 2026
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

func TestCompletionAtPos_exprMap(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"empty expression no element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `{ "key" = value }`,
					Detail: "map",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						NewText: "{\n  \n}",
						Snippet: "{\n  ${1}\n}",
					},
					Kind: lang.MapCandidateKind,
				},
			}),
		},
		{
			"empty expression with element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "kw",
						},
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `{ "key" = keyword }`,
					Detail: "map of keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						NewText: "{\n  \n}",
						Snippet: "{\n  ${1}\n}",
					},
					Kind:           lang.MapCandidateKind,
					TriggerSuggest: true,
				},
			}),
		},

		// single line tests
		{
			"inside braces single-line",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "kw",
						},
					},
				},
			},
			`attr = {  }
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
							End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
						},
						NewText: `"key" = `,
						Snippet: `"${1:key}" = `,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"single-line new element inside quoted key name with no equal sign",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "kw",
						},
					},
				},
			},
			`attr = { "foo" }
`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"single-line new element inside key name with no equal sign",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "kw",
						},
					},
				},
			},
			`attr = { foo }
`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"single-line new element value after equal sign",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "kw",
						},
					},
				},
			},
			`attr = { foo =  }
`,
			hcl.Pos{Line: 1, Column: 16, Byte: 15},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `kw`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 16, Byte: 15},
							End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
						},
						NewText: `kw`,
						Snippet: `kw`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"single-line new quoted element value after equal sign",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "kw",
						},
					},
				},
			},
			`attr = { "foo" =  }
`,
			hcl.Pos{Line: 1, Column: 18, Byte: 17},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `kw`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 18, Byte: 17},
							End:      hcl.Pos{Line: 1, Column: 18, Byte: 17},
						},
						NewText: `kw`,
						Snippet: `kw`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"single-line new element partial value near end",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "kw",
						},
					},
				},
			},
			`attr = { foo = k }
`,
			hcl.Pos{Line: 1, Column: 17, Byte: 16},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `kw`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 16, Byte: 15},
							End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
						},
						NewText: `kw`,
						Snippet: `kw`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"single-line new element partial value in the middle of value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = { foo = keyword }
`,
			hcl.Pos{Line: 1, Column: 18, Byte: 17},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 16, Byte: 15},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"single-line quoted element partial value in the middle of value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = { "foo" = keyword }
`,
			hcl.Pos{Line: 1, Column: 21, Byte: 20},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 18, Byte: 17},
							End:      hcl.Pos{Line: 1, Column: 25, Byte: 24},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"single-line element partial value in the middle of key name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = { foo = keyword }
`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"single-line quoted element partial value in the middle of key name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = { "foo" = keyword }
`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"single-line before existing item",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "kw",
						},
					},
				},
			},
			`attr = {  foo = kw }
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
							End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
						},
						NewText: `"key" = `,
						Snippet: `"${1:key}" = `,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"single-line after previous item with comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "kw",
						},
					},
				},
			},
			`attr = { foo = kw, }
`,
			hcl.Pos{Line: 1, Column: 19, Byte: 18},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 19, Byte: 18},
							End:      hcl.Pos{Line: 1, Column: 19, Byte: 18},
						},
						NewText: `"key" = `,
						Snippet: `"${1:key}" = `,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"single-line after previous item without comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "kw",
						},
					},
				},
			},
			`attr = { foo = kw  }
`,
			hcl.Pos{Line: 1, Column: 19, Byte: 18},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 19, Byte: 18},
							End:      hcl.Pos{Line: 1, Column: 19, Byte: 18},
						},
						NewText: `"key" = `,
						Snippet: `"${1:key}" = `,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"single-line between items with commas",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "kw",
						},
					},
				},
			},
			`attr = { foo = kw,  , bar = kw }
`,
			hcl.Pos{Line: 1, Column: 20, Byte: 19},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 20, Byte: 19},
							End:      hcl.Pos{Line: 1, Column: 20, Byte: 19},
						},
						NewText: `"key" = `,
						Snippet: `"${1:key}" = `,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"single-line between items without commas",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "kw",
						},
					},
				},
			},
			`attr = { foo = kw,   bar = kw }
`,
			hcl.Pos{Line: 1, Column: 20, Byte: 19},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 20, Byte: 19},
							End:      hcl.Pos{Line: 1, Column: 20, Byte: 19},
						},
						NewText: `"key" = `,
						Snippet: `"${1:key}" = `,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},

		// multi line tests
		{
			"inside braces multi-line",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "kw",
						},
					},
				},
			},
			`attr = {
  
}
`,
			hcl.Pos{Line: 2, Column: 3, Byte: 11},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
							End:      hcl.Pos{Line: 2, Column: 3, Byte: 11},
						},
						NewText: `"key" = `,
						Snippet: `"${1:key}" = `,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"multi-line new element value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "kw",
						},
					},
				},
			},
			`attr = {
  foo = 
}
`,
			hcl.Pos{Line: 2, Column: 9, Byte: 17},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `kw`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 9, Byte: 17},
							End:      hcl.Pos{Line: 2, Column: 9, Byte: 17},
						},
						NewText: `kw`,
						Snippet: `kw`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"inside multi-line map partial key",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "kw",
						},
					},
				},
			},
			`attr = {
  b
}`,
			hcl.Pos{Line: 2, Column: 4, Byte: 12},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"inside multi-line partial new element value near end",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "kw",
						},
					},
				},
			},
			`attr = {
  foo = k
}
`,
			hcl.Pos{Line: 2, Column: 10, Byte: 18},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `kw`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 9, Byte: 17},
							End:      hcl.Pos{Line: 2, Column: 10, Byte: 18},
						},
						NewText: `kw`,
						Snippet: `kw`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"inside multi-line partial new element value in the middle",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = {
  foo = keyword
}
`,
			hcl.Pos{Line: 2, Column: 12, Byte: 20},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 9, Byte: 17},
							End:      hcl.Pos{Line: 2, Column: 16, Byte: 24},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"multi-line value after existing attribute",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = {
  foo = keyword
  bar = 
}
`,
			hcl.Pos{Line: 3, Column: 9, Byte: 33},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 9, Byte: 33},
							End:      hcl.Pos{Line: 3, Column: 9, Byte: 33},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"multi-line value before existing attribute",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = {
  bar = 
  foo = keyword
}
`,
			hcl.Pos{Line: 2, Column: 9, Byte: 17},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 9, Byte: 17},
							End:      hcl.Pos{Line: 2, Column: 9, Byte: 17},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"multi-line value before between attributes",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = {
  bar = keyword
  baz = 
  foo = keyword
}
`,
			hcl.Pos{Line: 3, Column: 9, Byte: 33},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 9, Byte: 33},
							End:      hcl.Pos{Line: 3, Column: 9, Byte: 33},
						},
						NewText: `keyword`,
						Snippet: `keyword`,
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"multi-line key between attributes",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = {
  bar = keyword
  baz
  foo = keyword
}
`,
			hcl.Pos{Line: 3, Column: 5, Byte: 29},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"multi-line before attribute same line",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = {
  foo = keyword
}
`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 2, Byte: 10},
							End:      hcl.Pos{Line: 2, Column: 2, Byte: 10},
						},
						NewText: `"key" = `,
						Snippet: `"${1:key}" = `,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"multi-line before attribute separate line",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = {
  
  foo = keyword
}
`,
			hcl.Pos{Line: 2, Column: 3, Byte: 11},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
							End:      hcl.Pos{Line: 2, Column: 3, Byte: 11},
						},
						NewText: `"key" = `,
						Snippet: `"${1:key}" = `,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"multi-line after attribute",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = {
  foo = keyword
  
}
`,
			hcl.Pos{Line: 3, Column: 3, Byte: 27},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 27},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 27},
						},
						NewText: `"key" = `,
						Snippet: `"${1:key}" = `,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"multi-line after attribute with comma newline",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = {
  foo = keyword,
  
}
`,
			hcl.Pos{Line: 3, Column: 3, Byte: 28},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 28},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 28},
						},
						NewText: `"key" = `,
						Snippet: `"${1:key}" = `,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"multi-line after attribute with comma same line",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = {
  foo = keyword, 
}
`,
			hcl.Pos{Line: 2, Column: 18, Byte: 26},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 18, Byte: 26},
							End:      hcl.Pos{Line: 2, Column: 18, Byte: 26},
						},
						NewText: `"key" = `,
						Snippet: `"${1:key}" = `,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"multi-line after attribute without comma same line",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = {
  foo = keyword 
}
`,
			hcl.Pos{Line: 2, Column: 17, Byte: 25},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 17, Byte: 25},
							End:      hcl.Pos{Line: 2, Column: 17, Byte: 25},
						},
						NewText: `"key" = `,
						Snippet: `"${1:key}" = `,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"multi-line between attributes without commas",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = {
  foo = keyword
  
  bar = keyword
}
`,
			hcl.Pos{Line: 3, Column: 3, Byte: 27},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 27},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 27},
						},
						NewText: `"key" = `,
						Snippet: `"${1:key}" = `,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"multi-line between attributes with comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = {
  foo = keyword, 
  bar = keyword
}
`,
			hcl.Pos{Line: 2, Column: 18, Byte: 26},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = keyword`,
					Detail: "keyword",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 18, Byte: 26},
							End:      hcl.Pos{Line: 2, Column: 18, Byte: 26},
						},
						NewText: `"key" = `,
						Snippet: `"${1:key}" = `,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
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
				t.Logf("pos: %#v, config: %s\n", tc.pos, tc.cfg)
				t.Fatalf("unexpected candidates: %s", diff)
			}
		})
	}
}
