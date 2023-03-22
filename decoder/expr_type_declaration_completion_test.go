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

func TestCompletionAtPos_exprTypeDeclaration(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"all types",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates(allTypeDeclarationsAsCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
				End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
			})),
		},
		{
			"inside list",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = list()
`,
			hcl.Pos{Line: 1, Column: 13, Byte: 12},
			lang.CompleteCandidates(allTypeDeclarationsAsCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 13, Byte: 12},
				End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
			})),
		},
		{
			"inside set name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = set()
`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "string",
					Detail: "string",
					Kind:   lang.StringCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "string",
						Snippet: "string",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
						},
					},
				},
				{
					Label:  "set(…)",
					Detail: "set",
					Kind:   lang.SetCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "set()",
						Snippet: fmt.Sprintf("set(${%d})", 0),
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
						},
					},
					TriggerSuggest: true,
				},
			}),
		},
		{
			"partial string name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = st
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "string",
					Detail: "string",
					Kind:   lang.StringCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "string",
						Snippet: "string",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
						},
					},
				},
			}),
		},
		{
			"partial list name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = li
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "list(…)",
					Detail: "list",
					Kind:   lang.ListCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "list()",
						Snippet: fmt.Sprintf("list(${%d})", 0),
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
						},
					},
					TriggerSuggest: true,
				},
			}),
		},
		{
			"inside tuple",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = tuple()
`,
			hcl.Pos{Line: 1, Column: 14, Byte: 13},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "[…]",
					Detail: "tuple",
					Kind:   lang.TupleCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "[]",
						Snippet: "[ ${0} ]",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 14, Byte: 13},
							End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
						},
					},
				},
			}),
		},
		{
			"tuple inside brackets single-line",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = tuple([  ])
`,
			hcl.Pos{Line: 1, Column: 15, Byte: 14},
			lang.CompleteCandidates(allTypeDeclarationsAsCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
				End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
			})),
		},
		{
			"inside tuple - second type after comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = tuple([string,])
`,
			hcl.Pos{Line: 1, Column: 22, Byte: 21},
			lang.CompleteCandidates(allTypeDeclarationsAsCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 22, Byte: 21},
				End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
			})),
		},
		{
			"inside tuple - missing brackets second type after comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = tuple(string,)
`,
			hcl.Pos{Line: 1, Column: 21, Byte: 20},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"inside tuple - second type after space",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = tuple([string, ])
`,
			hcl.Pos{Line: 1, Column: 23, Byte: 22},
			lang.CompleteCandidates(allTypeDeclarationsAsCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
				End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
			})),
		},
		{
			"inside tuple - missing brackets second type after space",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = tuple(string, )
`,
			hcl.Pos{Line: 1, Column: 22, Byte: 21},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"inside tuple - second partial type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = tuple([string, s])
`,
			hcl.Pos{Line: 1, Column: 24, Byte: 23},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "string",
					Detail: "string",
					Kind:   lang.StringCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "string",
						Snippet: "string",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
						},
					},
				},
				{
					Label:  "set(…)",
					Detail: "set",
					Kind:   lang.SetCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "set()",
						Snippet: "set(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
						},
					},
					TriggerSuggest: true,
				},
			}),
		},
		{
			"inside set - invalid second argument",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = set(string,)
`,
			hcl.Pos{Line: 1, Column: 19, Byte: 18},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		// object tests
		{
			"inside object without braces",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object()
`,
			hcl.Pos{Line: 1, Column: 15, Byte: 14},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "{…}",
					Detail: "object",
					Kind:   lang.ObjectCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "{\n\n}",
						Snippet: fmt.Sprintf("{\n  ${%d:name} = ${%d}\n}", 1, 2),
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
							End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
						},
					},
				},
			}),
		},
		{
			"single-line inside object braces",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({})
`,
			hcl.Pos{Line: 1, Column: 16, Byte: 15},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "name = type",
					Detail: "type",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "name = ",
						Snippet: fmt.Sprintf("${%d:name} = ", 1),
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 16, Byte: 15},
							End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
						},
					},
				},
			}),
		},
		{
			"missing object notation single-line new element inside quoted key name with no equal sign",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = { "foo" }
`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"missing object notation single-line new element inside key name with no equal sign",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = { foo }
`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"missing object notation single-line new element value after equal sign",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = { foo =  }
`,
			hcl.Pos{Line: 1, Column: 16, Byte: 15},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"single-line object value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({ name =  })
`,
			hcl.Pos{Line: 1, Column: 24, Byte: 23},
			lang.CompleteCandidates(allTypeDeclarationsAsCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 24, Byte: 23},
				End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
			})),
		},
		{
			"inside single-line object partial value near end",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({ foo = s })
`,
			hcl.Pos{Line: 1, Column: 24, Byte: 23},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "string",
					Detail: "string",
					Kind:   lang.StringCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "string",
						Snippet: "string",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
						},
					},
				},
				{
					Label:  "set(…)",
					Detail: "set",
					Kind:   lang.SetCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "set()",
						Snippet: "set(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
						},
					},
					TriggerSuggest: true,
				},
			}),
		},
		{
			"inside single-line object partial value in the middle",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({ foo = string })
`,
			hcl.Pos{Line: 1, Column: 22, Byte: 23},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "string",
					Detail: "string",
					Kind:   lang.StringCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "string",
						Snippet: "string",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 29, Byte: 28},
						},
					},
				},
				{
					Label:  "set(…)",
					Detail: "set",
					Kind:   lang.SetCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "set()",
						Snippet: "set(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 29, Byte: 28},
						},
					},
					TriggerSuggest: true,
				},
			}),
		},
		{
			"single-line before attribute",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({  foo = string })
`,
			hcl.Pos{Line: 1, Column: 15, Byte: 16},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"single-line after attribute and comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({ foo = string,  })
`,
			hcl.Pos{Line: 1, Column: 29, Byte: 30},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "name = type",
					Detail: "type",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "name = ",
						Snippet: fmt.Sprintf("${%d:name} = ", 1),
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 29, Byte: 30},
							End:      hcl.Pos{Line: 1, Column: 29, Byte: 30},
						},
					},
				},
			}),
		},
		{
			"single-line after attribute without comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({ foo = string  })
`,
			hcl.Pos{Line: 1, Column: 28, Byte: 29},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"single-line between attributes with commas",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({ foo = string,  , bar = string })
`,
			hcl.Pos{Line: 1, Column: 29, Byte: 30},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "name = type",
					Detail: "type",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "name = ",
						Snippet: fmt.Sprintf("${%d:name} = ", 1),
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 29, Byte: 30},
							End:      hcl.Pos{Line: 1, Column: 29, Byte: 30},
						},
					},
				},
			}),
		},
		{
			"single-line between attributes without commas",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({ foo = string,  bar = string })
`,
			hcl.Pos{Line: 1, Column: 29, Byte: 30},
			lang.CompleteCandidates([]lang.Candidate{}),
		},

		// multi-line object tests
		{
			"multi-line inside object braces",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  
})
`,
			hcl.Pos{Line: 2, Column: 3, Byte: 18},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "name = type",
					Detail: "type",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "name = ",
						Snippet: fmt.Sprintf("${%d:name} = ", 1),
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 18},
							End:      hcl.Pos{Line: 2, Column: 3, Byte: 18},
						},
					},
				},
			}),
		},
		{
			"multi-line object value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  name = 
})
`,
			hcl.Pos{Line: 2, Column: 10, Byte: 25},
			lang.CompleteCandidates(allTypeDeclarationsAsCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 2, Column: 10, Byte: 25},
				End:      hcl.Pos{Line: 2, Column: 10, Byte: 25},
			})),
		},
		{
			"inside multi-line object partial value near end",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  foo = s
})
`,
			hcl.Pos{Line: 2, Column: 10, Byte: 25},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "string",
					Detail: "string",
					Kind:   lang.StringCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "string",
						Snippet: "string",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 9, Byte: 24},
							End:      hcl.Pos{Line: 2, Column: 10, Byte: 25},
						},
					},
				},
				{
					Label:  "set(…)",
					Detail: "set",
					Kind:   lang.SetCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "set()",
						Snippet: "set(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 9, Byte: 24},
							End:      hcl.Pos{Line: 2, Column: 10, Byte: 25},
						},
					},
					TriggerSuggest: true,
				},
			}),
		},
		{
			"inside multi-line object partial value in the middle",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  foo = string
})
`,
			hcl.Pos{Line: 2, Column: 10, Byte: 25},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "string",
					Detail: "string",
					Kind:   lang.StringCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "string",
						Snippet: "string",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 9, Byte: 24},
							End:      hcl.Pos{Line: 2, Column: 15, Byte: 30},
						},
					},
				},
				{
					Label:  "set(…)",
					Detail: "set",
					Kind:   lang.SetCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "set()",
						Snippet: "set(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 9, Byte: 24},
							End:      hcl.Pos{Line: 2, Column: 15, Byte: 30},
						},
					},
					TriggerSuggest: true,
				},
			}),
		},
		{
			"multi-line object value after existing attribute",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  foo = string
  bar = 
})
`,
			hcl.Pos{Line: 3, Column: 9, Byte: 39},
			lang.CompleteCandidates(allTypeDeclarationsAsCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 3, Column: 9, Byte: 39},
				End:      hcl.Pos{Line: 3, Column: 9, Byte: 39},
			})),
		},
		{
			"multi-line object value before existing attribute",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  bar = 
  foo = string
})
`,
			hcl.Pos{Line: 2, Column: 9, Byte: 24},
			lang.CompleteCandidates(allTypeDeclarationsAsCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 2, Column: 9, Byte: 24},
				End:      hcl.Pos{Line: 2, Column: 9, Byte: 24},
			})),
		},
		{
			"multi-line object value between existing attributes",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  bar = number
  baz = 
  foo = string
})
`,
			hcl.Pos{Line: 3, Column: 9, Byte: 39},
			lang.CompleteCandidates(allTypeDeclarationsAsCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 3, Column: 9, Byte: 39},
				End:      hcl.Pos{Line: 3, Column: 9, Byte: 39},
			})),
		},
		{
			"multi-line before attribute same line",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  foo = string
})
`,
			hcl.Pos{Line: 2, Column: 1, Byte: 16},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"multi-line before attribute separate line",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  
  foo = string
})
`,
			hcl.Pos{Line: 2, Column: 3, Byte: 18},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "name = type",
					Detail: "type",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "name = ",
						Snippet: fmt.Sprintf("${%d:name} = ", 1),
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 18},
							End:      hcl.Pos{Line: 2, Column: 3, Byte: 18},
						},
					},
				},
			}),
		},
		{
			"multi-line after attribute",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  foo = string
  
})
`,
			hcl.Pos{Line: 3, Column: 3, Byte: 33},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "name = type",
					Detail: "type",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "name = ",
						Snippet: fmt.Sprintf("${%d:name} = ", 1),
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 33},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 33},
						},
					},
				},
			}),
		},
		{
			"multi-line after attribute with comma newline",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  foo = string,
  
})
`,
			hcl.Pos{Line: 3, Column: 3, Byte: 34},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "name = type",
					Detail: "type",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "name = ",
						Snippet: fmt.Sprintf("${%d:name} = ", 1),
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 34},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 34},
						},
					},
				},
			}),
		},
		{
			"multi-line after attribute with comma same line",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  foo = string, 
})
`,
			hcl.Pos{Line: 2, Column: 17, Byte: 32},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "name = type",
					Detail: "type",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "name = ",
						Snippet: fmt.Sprintf("${%d:name} = ", 1),
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 17, Byte: 32},
							End:      hcl.Pos{Line: 2, Column: 17, Byte: 32},
						},
					},
				},
			}),
		},
		{
			"multi-line after attribute without comma same line",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  foo = string 
})
`,
			hcl.Pos{Line: 2, Column: 16, Byte: 31},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"multi-line between attributes without commas",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  foo = string
  
  bar = string
})
`,
			hcl.Pos{Line: 3, Column: 3, Byte: 33},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "name = type",
					Detail: "type",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "name = ",
						Snippet: fmt.Sprintf("${%d:name} = ", 1),
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 33},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 33},
						},
					},
				},
			}),
		},
		{
			"multi-line between attributes with comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  foo = string, 
  bar = string
})
`,
			hcl.Pos{Line: 2, Column: 17, Byte: 32},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "name = type",
					Detail: "type",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "name = ",
						Snippet: fmt.Sprintf("${%d:name} = ", 1),
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 17, Byte: 32},
							End:      hcl.Pos{Line: 2, Column: 17, Byte: 32},
						},
					},
				},
			}),
		},
		{
			"multi-line inside attribute",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  s = 
})`,
			hcl.Pos{Line: 2, Column: 4, Byte: 19},
			lang.CompleteCandidates([]lang.Candidate{}),
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
			candidates, err := d.CandidatesAtPos(ctx, "test.tf", tc.pos)
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
