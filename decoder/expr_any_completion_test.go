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
	"github.com/zclconf/go-cty/cty/function"
)

func TestCompletionAtPos_exprAny_functions(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		refTargets         reference.Targets
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"list of string functions",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Targets{},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:       "element",
					Detail:      "element(list dynamic, index number) dynamic",
					Description: lang.Markdown("`element` retrieves a single element from a list."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "element()",
						Snippet: "element(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
				{
					Label:       "join",
					Detail:      "join(separator string, …lists list of string) string",
					Description: lang.Markdown("`join` produces a string by concatenating together all elements of a given list of strings with the given delimiter."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "join()",
						Snippet: "join(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
				{
					Label:       "keys",
					Detail:      "keys(inputMap dynamic) dynamic",
					Description: lang.Markdown("`keys` takes a map and returns a list containing the keys from that map."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "keys()",
						Snippet: "keys(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
				{
					Label:       "log",
					Detail:      "log(num number, base number) number",
					Description: lang.Markdown("`log` returns the logarithm of a given number in a given base."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "log()",
						Snippet: "log(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
				{
					Label:       "lower",
					Detail:      "lower(str string) string",
					Description: lang.Markdown("`lower` converts all cased letters in the given string to lowercase."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "lower()",
						Snippet: "lower(${0})",
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
			"list of any functions",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.DynamicPseudoType,
					},
				},
			},
			reference.Targets{},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:       "element",
					Detail:      "element(list dynamic, index number) dynamic",
					Description: lang.Markdown("`element` retrieves a single element from a list."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "element()",
						Snippet: "element(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
				{
					Label:       "join",
					Detail:      "join(separator string, …lists list of string) string",
					Description: lang.Markdown("`join` produces a string by concatenating together all elements of a given list of strings with the given delimiter."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "join()",
						Snippet: "join(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
				{
					Label:       "keys",
					Detail:      "keys(inputMap dynamic) dynamic",
					Description: lang.Markdown("`keys` takes a map and returns a list containing the keys from that map."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "keys()",
						Snippet: "keys(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
				{
					Label:       "log",
					Detail:      "log(num number, base number) number",
					Description: lang.Markdown("`log` returns the logarithm of a given number in a given base."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "log()",
						Snippet: "log(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
				{
					Label:       "lower",
					Detail:      "lower(str string) string",
					Description: lang.Markdown("`lower` converts all cased letters in the given string to lowercase."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "lower()",
						Snippet: "lower(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
				{
					Label:       "split",
					Detail:      "split(separator string, str string) list of string",
					Description: lang.Markdown("`split` produces a list by dividing a given string at all occurrences of a given separator."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "split()",
						Snippet: "split(${0})",
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
			"function by prefix",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Targets{},
			`attr = j
`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:       "join",
					Detail:      "join(separator string, …lists list of string) string",
					Description: lang.Markdown("`join` produces a string by concatenating together all elements of a given list of strings with the given delimiter."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "join()",
						Snippet: "join(${0})",
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
			"first argument of a function",
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
						lang.RootStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.String,
				},
			},
			`attr = element()
`,
			hcl.Pos{Line: 1, Column: 14, Byte: 15},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "foo.bar",
					Detail: "string",
					Kind:   lang.TraversalCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "foo.bar",
						Snippet: "foo.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 14, Byte: 15},
							End:      hcl.Pos{Line: 1, Column: 14, Byte: 15},
						},
					},
				},
				{
					Label:       "element",
					Detail:      "element(list dynamic, index number) dynamic",
					Description: lang.Markdown("`element` retrieves a single element from a list."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "element()",
						Snippet: "element(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 14, Byte: 15},
							End:      hcl.Pos{Line: 1, Column: 14, Byte: 15},
						},
					},
				},
				{
					Label:       "join",
					Detail:      "join(separator string, …lists list of string) string",
					Description: lang.Markdown("`join` produces a string by concatenating together all elements of a given list of strings with the given delimiter."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "join()",
						Snippet: "join(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 14, Byte: 15},
							End:      hcl.Pos{Line: 1, Column: 14, Byte: 15},
						},
					},
				},
				{
					Label:       "keys",
					Detail:      "keys(inputMap dynamic) dynamic",
					Description: lang.Markdown("`keys` takes a map and returns a list containing the keys from that map."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "keys()",
						Snippet: "keys(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 14, Byte: 15},
							End:      hcl.Pos{Line: 1, Column: 14, Byte: 15},
						},
					},
				},
				{
					Label:       "log",
					Detail:      "log(num number, base number) number",
					Description: lang.Markdown("`log` returns the logarithm of a given number in a given base."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "log()",
						Snippet: "log(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 14, Byte: 15},
							End:      hcl.Pos{Line: 1, Column: 14, Byte: 15},
						},
					},
				},
				{
					Label:       "lower",
					Detail:      "lower(str string) string",
					Description: lang.Markdown("`lower` converts all cased letters in the given string to lowercase."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "lower()",
						Snippet: "lower(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 14, Byte: 15},
							End:      hcl.Pos{Line: 1, Column: 14, Byte: 15},
						},
					},
				},
				{
					Label:       "split",
					Detail:      "split(separator string, str string) list of string",
					Description: lang.Markdown("`split` produces a list by dividing a given string at all occurrences of a given separator."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "split()",
						Snippet: "split(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 14, Byte: 15},
							End:      hcl.Pos{Line: 1, Column: 14, Byte: 15},
						},
					},
				},
			}),
		},
		{
			"reference as argument partial",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "lst"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 5, Byte: 27},
						End:      hcl.Pos{Line: 2, Column: 15, Byte: 37},
					},
					Type: cty.List(cty.String),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "lst"},
								lang.IndexStep{Key: cty.NumberIntVal(0)},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 2, Column: 8, Byte: 30},
								End:      hcl.Pos{Line: 2, Column: 13, Byte: 35},
							},
							Type: cty.String,
						},
					},
				},
			},
			`attr = split(va)
`,
			hcl.Pos{Line: 1, Column: 16, Byte: 15},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.lst",
					Detail: "list of string",
					Kind:   lang.TraversalCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 14, Byte: 13},
							End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
						},
						NewText: "var.lst",
						Snippet: "var.lst",
					},
				},
			}),
		},
		{
			"reference as argument with trailing dot",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "obj"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 5, Byte: 27},
						End:      hcl.Pos{Line: 2, Column: 15, Byte: 37},
					},
					Type: cty.List(cty.String),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "obj"},
								lang.AttrStep{Name: "foo"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 2, Column: 8, Byte: 30},
								End:      hcl.Pos{Line: 2, Column: 13, Byte: 35},
							},
							Type: cty.String,
						},
					},
				},
			},
			`attr = split(var.)
`,
			hcl.Pos{Line: 1, Column: 18, Byte: 17},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.obj",
					Detail: "list of string",
					Kind:   lang.TraversalCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 14, Byte: 13},
							End:      hcl.Pos{Line: 1, Column: 18, Byte: 17},
						},
						NewText: "var.obj",
						Snippet: "var.obj",
					},
				},
			}),
		},
		{
			"reference as argument within brackets",
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
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "map"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 5, Byte: 27},
						End:      hcl.Pos{Line: 2, Column: 15, Byte: 37},
					},
					Type: cty.List(cty.String),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "map"},
								lang.IndexStep{Key: cty.StringVal("foo")},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 2, Column: 8, Byte: 30},
								End:      hcl.Pos{Line: 2, Column: 13, Byte: 35},
							},
							Type: cty.String,
						},
					},
				},
			},
			`attr = split(var.map[])
`,
			hcl.Pos{Line: 1, Column: 22, Byte: 21},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `var.map["foo"]`,
					Detail: "string",
					Kind:   lang.TraversalCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 14, Byte: 13},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
						},
						NewText: `var.map["foo"]`,
						Snippet: `var.map["foo"]`,
					},
				},
			}),
		},
		{
			"reference as argument with trailing bracket",
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
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "obj"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 5, Byte: 27},
						End:      hcl.Pos{Line: 2, Column: 15, Byte: 37},
					},
					Type: cty.List(cty.String),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "obj"},
								lang.IndexStep{Key: cty.StringVal("foo")},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 2, Column: 8, Byte: 30},
								End:      hcl.Pos{Line: 2, Column: 13, Byte: 35},
							},
							Type: cty.String,
						},
					},
				},
			},
			`attr = split(var.map[)
`,
			hcl.Pos{Line: 1, Column: 22, Byte: 21},
			lang.CompleteCandidates([]lang.Candidate{
				// TODO: See https://github.com/hashicorp/hcl/issues/604
			}),
		},
		{
			"second number argument of a function",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Targets{},
			`attr = element(["e1", "e2"], )
`,
			hcl.Pos{Line: 1, Column: 28, Byte: 29},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:       "element",
					Detail:      "element(list dynamic, index number) dynamic",
					Description: lang.Markdown("`element` retrieves a single element from a list."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "element()",
						Snippet: "element(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 28, Byte: 29},
							End:      hcl.Pos{Line: 1, Column: 28, Byte: 29},
						},
					},
				},
				{
					Label:       "join",
					Detail:      "join(separator string, …lists list of string) string",
					Description: lang.Markdown("`join` produces a string by concatenating together all elements of a given list of strings with the given delimiter."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "join()",
						Snippet: "join(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 28, Byte: 29},
							End:      hcl.Pos{Line: 1, Column: 28, Byte: 29},
						},
					},
				},
				{
					Label:       "keys",
					Detail:      "keys(inputMap dynamic) dynamic",
					Description: lang.Markdown("`keys` takes a map and returns a list containing the keys from that map."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "keys()",
						Snippet: "keys(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 28, Byte: 29},
							End:      hcl.Pos{Line: 1, Column: 28, Byte: 29},
						},
					},
				},
				{
					Label:       "log",
					Detail:      "log(num number, base number) number",
					Description: lang.Markdown("`log` returns the logarithm of a given number in a given base."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "log()",
						Snippet: "log(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 28, Byte: 29},
							End:      hcl.Pos{Line: 1, Column: 28, Byte: 29},
						},
					},
				},
				{
					Label:       "lower",
					Detail:      "lower(str string) string",
					Description: lang.Markdown("`lower` converts all cased letters in the given string to lowercase."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "lower()",
						Snippet: "lower(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 28, Byte: 29},
							End:      hcl.Pos{Line: 1, Column: 28, Byte: 29},
						},
					},
				},
			}),
		},
		{
			"nested functions with string constraint",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Targets{},
			`attr = join("-", split())
`,
			hcl.Pos{Line: 1, Column: 22, Byte: 23},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:       "element",
					Detail:      "element(list dynamic, index number) dynamic",
					Description: lang.Markdown("`element` retrieves a single element from a list."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "element()",
						Snippet: "element(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 22, Byte: 23},
							End:      hcl.Pos{Line: 1, Column: 22, Byte: 23},
						},
					},
				},
				{
					Label:       "join",
					Detail:      "join(separator string, …lists list of string) string",
					Description: lang.Markdown("`join` produces a string by concatenating together all elements of a given list of strings with the given delimiter."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "join()",
						Snippet: "join(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 22, Byte: 23},
							End:      hcl.Pos{Line: 1, Column: 22, Byte: 23},
						},
					},
				},
				{
					Label:       "keys",
					Detail:      "keys(inputMap dynamic) dynamic",
					Description: lang.Markdown("`keys` takes a map and returns a list containing the keys from that map."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "keys()",
						Snippet: "keys(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 22, Byte: 23},
							End:      hcl.Pos{Line: 1, Column: 22, Byte: 23},
						},
					},
				},
				{
					Label:       "log",
					Detail:      "log(num number, base number) number",
					Description: lang.Markdown("`log` returns the logarithm of a given number in a given base."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "log()",
						Snippet: "log(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 22, Byte: 23},
							End:      hcl.Pos{Line: 1, Column: 22, Byte: 23},
						},
					},
				},
				{
					Label:       "lower",
					Detail:      "lower(str string) string",
					Description: lang.Markdown("`lower` converts all cased letters in the given string to lowercase."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "lower()",
						Snippet: "lower(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 22, Byte: 23},
							End:      hcl.Pos{Line: 1, Column: 22, Byte: 23},
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

			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
				Functions:        testFunctionSignatures(),
				ReferenceTargets: tc.refTargets,
			})

			ctx := context.Background()
			candidates, err := d.CandidatesAtPos(ctx, "test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedCandidates, candidates); diff != "" {
				t.Fatalf("unexpected candidates: %s", diff)
			}
		})
	}
}

func testFunctionSignatures() map[string]schema.FunctionSignature {
	return map[string]schema.FunctionSignature{
		"element": {
			Params: []function.Parameter{
				{
					Name: "list",
					Type: cty.DynamicPseudoType,
				},
				{
					Name: "index",
					Type: cty.Number,
				},
			},
			ReturnType:  cty.DynamicPseudoType,
			Description: "`element` retrieves a single element from a list.",
		},
		"join": {
			Params: []function.Parameter{
				{
					Name:        "separator",
					Description: "Delimiter to insert between the given strings.",
					Type:        cty.String,
				},
			},
			VarParam: &function.Parameter{
				Name:        "lists",
				Description: "One or more lists of strings to join.",
				Type:        cty.List(cty.String),
			},
			ReturnType:  cty.String,
			Description: "`join` produces a string by concatenating together all elements of a given list of strings with the given delimiter.",
		},
		"keys": {
			Params: []function.Parameter{
				{
					Name:        "inputMap",
					Description: "The map to extract keys from. May instead be an object-typed value, in which case the result is a tuple of the object attributes.",
					Type:        cty.DynamicPseudoType,
				},
			},
			ReturnType:  cty.DynamicPseudoType,
			Description: "`keys` takes a map and returns a list containing the keys from that map.",
		},
		"log": {
			Params: []function.Parameter{
				{
					Name: "num",
					Type: cty.Number,
				},
				{
					Name: "base",
					Type: cty.Number,
				},
			},
			ReturnType:  cty.Number,
			Description: "`log` returns the logarithm of a given number in a given base.",
		},
		"lower": {
			Params: []function.Parameter{
				{
					Name: "str",
					Type: cty.String,
				},
			},
			ReturnType:  cty.String,
			Description: "`lower` converts all cased letters in the given string to lowercase.",
		},
		"split": {
			Params: []function.Parameter{
				{
					Name: "separator",
					Type: cty.String,
				},
				{
					Name: "str",
					Type: cty.String,
				},
			},
			ReturnType:  cty.List(cty.String),
			Description: "`split` produces a list by dividing a given string at all occurrences of a given separator.",
		},
	}
}
