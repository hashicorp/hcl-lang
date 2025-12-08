// Copyright IBM Corp. 2020, 2025
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
				{
					Label:       "namespaced::function",
					Description: lang.Markdown("Example for hcl valid namespaced function"),
					Detail:      "namespaced::function() bool",
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "namespaced::function()",
						Snippet: "namespaced::function(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
				{
					Label:       "provider::framework::example",
					Detail:      "provider::framework::example(input string) string",
					Description: lang.Markdown("Echoes given argument as result"),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "provider::framework::example()",
						Snippet: "provider::framework::example(${0})",
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
					Label:       "namespaced::function",
					Description: lang.Markdown("Example for hcl valid namespaced function"),
					Detail:      "namespaced::function() bool",
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "namespaced::function()",
						Snippet: "namespaced::function(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
				{
					Label:       "provider::framework::example",
					Detail:      "provider::framework::example(input string) string",
					Description: lang.Markdown("Echoes given argument as result"),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "provider::framework::example()",
						Snippet: "provider::framework::example(${0})",
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
					Kind:   lang.ReferenceCandidateKind,
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
					Label:       "namespaced::function",
					Description: lang.Markdown("Example for hcl valid namespaced function"),
					Detail:      "namespaced::function() bool",
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "namespaced::function()",
						Snippet: "namespaced::function(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 14, Byte: 15},
							End:      hcl.Pos{Line: 1, Column: 14, Byte: 15},
						},
					},
				},
				{
					Label:       "provider::framework::example",
					Detail:      "provider::framework::example(input string) string",
					Description: lang.Markdown("Echoes given argument as result"),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "provider::framework::example()",
						Snippet: "provider::framework::example(${0})",
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
					Kind:   lang.ReferenceCandidateKind,
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
					Kind:   lang.ReferenceCandidateKind,
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
					Type: cty.Map(cty.String),
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
					Kind:   lang.ReferenceCandidateKind,
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
				{
					Label:  `var.map`,
					Detail: "map of string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 22, Byte: 21},
							End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
						},
						NewText: `var.map`,
						Snippet: `var.map`,
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
							Start:    hcl.Pos{Line: 1, Column: 22, Byte: 21},
							End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
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
							Start:    hcl.Pos{Line: 1, Column: 22, Byte: 21},
							End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
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
							Start:    hcl.Pos{Line: 1, Column: 22, Byte: 21},
							End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
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
							Start:    hcl.Pos{Line: 1, Column: 22, Byte: 21},
							End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
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
							Start:    hcl.Pos{Line: 1, Column: 22, Byte: 21},
							End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
						},
					},
				},
				{
					Label:       "namespaced::function",
					Description: lang.Markdown("Example for hcl valid namespaced function"),
					Detail:      "namespaced::function() bool",
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "namespaced::function()",
						Snippet: "namespaced::function(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 22, Byte: 21},
							End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
						},
					},
				},
				{
					Label:       "provider::framework::example",
					Detail:      "provider::framework::example(input string) string",
					Description: lang.Markdown("Echoes given argument as result"),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "provider::framework::example()",
						Snippet: "provider::framework::example(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 22, Byte: 21},
							End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
						},
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
				{
					Label:       "provider::framework::example",
					Detail:      "provider::framework::example(input string) string",
					Description: lang.Markdown("Echoes given argument as result"),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "provider::framework::example()",
						Snippet: "provider::framework::example(${0})",
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
				{
					Label:       "namespaced::function",
					Description: lang.Markdown("Example for hcl valid namespaced function"),
					Detail:      "namespaced::function() bool",
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "namespaced::function()",
						Snippet: "namespaced::function(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 22, Byte: 23},
							End:      hcl.Pos{Line: 1, Column: 22, Byte: 23},
						},
					},
				},
				{
					Label:       "provider::framework::example",
					Detail:      "provider::framework::example(input string) string",
					Description: lang.Markdown("Echoes given argument as result"),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "provider::framework::example()",
						Snippet: "provider::framework::example(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 22, Byte: 23},
							End:      hcl.Pos{Line: 1, Column: 22, Byte: 23},
						},
					},
				},
			}),
		},
		{
			"in front of prefix (with space)",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Targets{},
			`attr = t 
`,
			hcl.Pos{Line: 1, Column: 7, Byte: 6},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"in front of non-empty [ ]",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Targets{},
			`attr = t [x]
`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"in incomplete namespaced function call",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Targets{},
			`attr = provider::framewo`,
			hcl.Pos{Line: 1, Column: 18, Byte: 17}, // cursor is after second colon
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:       "provider::framework::example",
					Detail:      "provider::framework::example(input string) string",
					Description: lang.Markdown("Echoes given argument as result"),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "provider::framework::example()",
						Snippet: "provider::framework::example(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 18, Byte: 17},
						},
					},
				},
			}),
		},
		{
			"in namespaced function call with only one namespace",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Targets{},
			`attr = namespaced::function()`,
			hcl.Pos{Line: 1, Column: 23, Byte: 22}, // within "function"
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:       "namespaced::function",
					Description: lang.Markdown("Example for hcl valid namespaced function"),
					Detail:      "namespaced::function() bool",
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "namespaced::function()",
						Snippet: "namespaced::function(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 30, Byte: 29},
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
				Functions:        testFunctionSignatures(),
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

func TestCompletionAtPos_exprAny_combinedExpressions(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		refTargets         reference.Targets
		funcSignatures     map[string]schema.FunctionSignature
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"any matching expression empty",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.Bool,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.Number,
				},
			},
			testFunctionSignatures(),
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "local.foo",
					Detail: "bool",
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
					Label:       "namespaced::function",
					Description: lang.Markdown("Example for hcl valid namespaced function"),
					Detail:      "namespaced::function() bool",
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "namespaced::function()",
						Snippet: "namespaced::function(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
				{
					Label:       "provider::framework::example",
					Detail:      "provider::framework::example(input string) string",
					Description: lang.Markdown("Echoes given argument as result"),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "provider::framework::example()",
						Snippet: "provider::framework::example(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
				{
					Label:  "false",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "false",
						Snippet: "false",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
				{
					Label:  "true",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "true",
						Snippet: "true",
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
			"any matching expression by prefix",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "toot"},
						lang.AttrStep{Name: "noot"},
					},
					Type: cty.Bool,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.Bool,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "too"},
						lang.AttrStep{Name: "not"},
					},
					Type: cty.Number,
				},
			},
			map[string]schema.FunctionSignature{
				"tobool": {
					Params: []function.Parameter{
						{
							Name: "v",
							Type: cty.DynamicPseudoType,
						},
					},
					ReturnType:  cty.Bool,
					Description: "`tobool` converts its argument to a boolean value.",
				},
				"substr": {
					Params: []function.Parameter{
						{
							Name: "str",
							Type: cty.String,
						},
						{
							Name: "offset",
							Type: cty.Number,
						},
						{
							Name: "length",
							Type: cty.Number,
						},
					},
					ReturnType:  cty.String,
					Description: "`substr` extracts a substring from a given string by offset and (maximum) length.",
				},
			},
			`attr = t
`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "toot.noot",
					Detail: "bool",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "toot.noot",
						Snippet: "toot.noot",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
				{
					Label:       "tobool",
					Detail:      "tobool(v dynamic) bool",
					Description: lang.Markdown("`tobool` converts its argument to a boolean value."),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "tobool()",
						Snippet: "tobool(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
				{
					Label:  "true",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "true",
						Snippet: "true",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
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
				Functions:        tc.funcSignatures,
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
		"provider::framework::example": {
			Params: []function.Parameter{
				{
					Name:        "input",
					Type:        cty.String,
					Description: "String to echo",
				},
			},
			ReturnType:  cty.String,
			Description: "Echoes given argument as result",
			Detail:      "bflad/framework 0.2.0",
		},
		"namespaced::function": {
			Params:      []function.Parameter{},
			ReturnType:  cty.Bool,
			Description: "Example for hcl valid namespaced function",
		},
	}
}

func TestCompletionAtPos_exprAny_literalTypes(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"bool",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "false",
					Detail: cty.Bool.FriendlyNameForConstraint(),
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "false",
						Snippet: "false",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
				{
					Label:  "true",
					Detail: cty.Bool.FriendlyNameForConstraint(),
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "true",
						Snippet: "true",
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
			"bool by prefix",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
					},
				},
			},
			`attr = f
`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "false",
					Detail: cty.Bool.FriendlyNameForConstraint(),
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "false",
						Snippet: "false",
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
			"string",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"list of strings",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "[ string ]",
					Detail: "list of string",
					Kind:   lang.ListCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: `[ "" ]`,
						Snippet: `[ "${1:value}" ]`,
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
			"inside list of bool",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.Bool),
					},
				},
			},
			`attr = [  ]
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.CompleteCandidates(boolLiteralTypeCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
				End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
			})),
		},
		{
			"inside list of bool multiline",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.Bool),
					},
				},
			},
			`attr = [
  
]
`,
			hcl.Pos{Line: 2, Column: 3, Byte: 11},
			lang.CompleteCandidates(boolLiteralTypeCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
				End:      hcl.Pos{Line: 2, Column: 3, Byte: 11},
			})),
		},
		{
			"inside list next element after space",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.Bool),
					},
				},
			},
			`attr = [ false,  ]
`,
			hcl.Pos{Line: 1, Column: 17, Byte: 16},
			lang.CompleteCandidates(boolLiteralTypeCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 17, Byte: 16},
				End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
			})),
		},
		{
			"inside list next element after newline",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.Bool),
					},
				},
			},
			`attr = [
  false,
  
]
`,
			hcl.Pos{Line: 3, Column: 3, Byte: 20},
			lang.CompleteCandidates(boolLiteralTypeCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 3, Column: 3, Byte: 20},
				End:      hcl.Pos{Line: 3, Column: 3, Byte: 20},
			})),
		},
		{
			"inside list next element after comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.Bool),
					},
				},
			},
			`attr = [ false, ]
`,
			hcl.Pos{Line: 1, Column: 16, Byte: 15},
			lang.CompleteCandidates(boolLiteralTypeCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 16, Byte: 15},
				End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
			})),
		},
		{
			"inside list next element near closing bracket",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.Bool),
					},
				},
			},
			`attr = [ false, ]
`,
			hcl.Pos{Line: 1, Column: 17, Byte: 16},
			lang.CompleteCandidates(boolLiteralTypeCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 17, Byte: 16},
				End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
			})),
		},
		{
			"completion inside list with prefix",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.Bool),
					},
				},
			},
			`attr = [ f ]
`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "false",
					Detail: cty.Bool.FriendlyNameForConstraint(),
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "false",
						Snippet: "false",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
							End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
						},
					},
				},
			}),
		},
		{
			"tuple",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{cty.Bool}),
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "[ bool ]",
					Detail: "tuple",
					Kind:   lang.TupleCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "[ false ]",
						Snippet: "[ ${1:false} ]",
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
			"inside tuple",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{cty.Bool}),
					},
				},
			},
			`attr = [  ]
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "false",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "false",
						Snippet: "false",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
							End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
						},
					},
				},
				{
					Label:  "true",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "true",
						Snippet: "true",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
							End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
						},
					},
				},
			}),
		},
		{
			"inside tuple next element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{cty.String, cty.Bool}),
					},
				},
			},
			`attr = [ "",  ]
`,
			hcl.Pos{Line: 1, Column: 14, Byte: 13},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "false",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "false",
						Snippet: "false",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 14, Byte: 13},
							End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
						},
					},
				},
				{
					Label:  "true",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "true",
						Snippet: "true",
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
			"inside tuple next element without comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{cty.String, cty.Bool}),
					},
				},
			},
			`attr = [ ""  ]
`,
			hcl.Pos{Line: 1, Column: 13, Byte: 12},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"inside tuple in space between elements",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{cty.String, cty.String, cty.Bool}),
					},
				},
			},
			`attr = [ "", ""  ]
`,
			hcl.Pos{Line: 1, Column: 13, Byte: 12},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"inside tuple next element which does not exist",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Tuple([]cty.Type{cty.String}),
					},
				},
			},
			`attr = [ "",  ]
`,
			hcl.Pos{Line: 1, Column: 14, Byte: 13},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"map",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.Bool),
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `{ "key" = bool }`,
					Detail: "map of bool",
					Kind:   lang.MapCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "{\n  \"key\" = false\n}",
						Snippet: "{\n  \"${1:key}\" = ${2:false}\n}",
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
			"inside empty map",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.Bool),
					},
				},
			},
			`attr = {
  
}
`,
			hcl.Pos{Line: 2, Column: 3, Byte: 11},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = bool`,
					Detail: "bool",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "\"key\" = ",
						Snippet: "\"${1:key}\" = ",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
							End:      hcl.Pos{Line: 2, Column: 3, Byte: 11},
						},
					},
				},
			}),
		},
		{
			"inside map after first item",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.Bool),
					},
				},
			},
			`attr = {
  "key" = true
  
}
`,
			hcl.Pos{Line: 3, Column: 3, Byte: 26},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = bool`,
					Detail: "bool",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "\"key\" = ",
						Snippet: "\"${1:key}\" = ",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 26},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 26},
						},
					},
				},
			}),
		},
		{
			"inside map between items",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.Bool),
					},
				},
			},
			`attr = {
  "key" = true
  
  "another" = false
}
`,
			hcl.Pos{Line: 3, Column: 3, Byte: 26},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = bool`,
					Detail: "bool",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "\"key\" = ",
						Snippet: "\"${1:key}\" = ",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 26},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 26},
						},
					},
				},
			}),
		},
		{
			"inside map before item",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.Bool),
					},
				},
			},
			`attr = {
  "key" = true
}
`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = bool`,
					Detail: "bool",
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
			"inside map value empty",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.Bool),
					},
				},
			},
			`attr = {
  "key" = 
}
`,
			hcl.Pos{Line: 2, Column: 11, Byte: 19},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "false",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "false",
						Snippet: "false",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 11, Byte: 19},
							End:      hcl.Pos{Line: 2, Column: 11, Byte: 19},
						},
					},
				},
				{
					Label:  "true",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "true",
						Snippet: "true",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 11, Byte: 19},
							End:      hcl.Pos{Line: 2, Column: 11, Byte: 19},
						},
					},
				},
			}),
		},
		{
			"inside map value with prefix",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.Bool),
					},
				},
			},
			`attr = {
  "key" = f
}
`,
			hcl.Pos{Line: 2, Column: 12, Byte: 20},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "false",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "false",
						Snippet: "false",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 11, Byte: 19},
							End:      hcl.Pos{Line: 2, Column: 12, Byte: 20},
						},
					},
				},
			}),
		},
		{
			"object",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Bool,
							"baz": cty.Number,
						}),
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `{ bar = bool, … }`,
					Detail: "object",
					Kind:   lang.ObjectCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "{\n  bar = false\n  baz = 1\n  foo = \"\"\n}",
						Snippet: "{\n  bar = ${1:false}\n  baz = ${2:1}\n  foo = \"${3:value}\"\n}",
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
			"inside empty object",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Bool,
							"baz": cty.Number,
						}),
					},
				},
			},
			`attr = {

}
`,
			hcl.Pos{Line: 2, Column: 1, Byte: 9},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `bar`,
					Detail: "required, bool",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "bar",
						Snippet: "bar = ${1:false}",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 1, Byte: 9},
							End:      hcl.Pos{Line: 2, Column: 1, Byte: 9},
						},
					},
				},
				{
					Label:  `baz`,
					Detail: "required, number",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "baz",
						Snippet: "baz = ${1:0}",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 1, Byte: 9},
							End:      hcl.Pos{Line: 2, Column: 1, Byte: 9},
						},
					},
				},
				{
					Label:  `foo`,
					Detail: "required, string",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "foo",
						Snippet: "foo = \"${1:value}\"",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 1, Byte: 9},
							End:      hcl.Pos{Line: 2, Column: 1, Byte: 9},
						},
					},
				},
			}),
		},
		{
			"inside object after first item",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Bool,
						}),
					},
				},
			},
			`attr = {
  foo = "baz"
  
}
`,
			hcl.Pos{Line: 3, Column: 3, Byte: 25},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `bar`,
					Detail: "required, bool",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "bar",
						Snippet: "bar = ${1:false}",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 25},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 25},
						},
					},
				},
			}),
		},
		{
			"inside object between items",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Bool,
							"baz": cty.Number,
						}),
					},
				},
			},
			`attr = {
  foo = "baz"
  
  baz = 42
}
`,
			hcl.Pos{Line: 3, Column: 3, Byte: 25},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `bar`,
					Detail: "required, bool",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "bar",
						Snippet: "bar = ${1:false}",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 25},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 25},
						},
					},
				},
			}),
		},
		{
			"inside object before item",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Bool,
						}),
					},
				},
			},
			`attr = {
  foo = "baz"
}
`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"inside object key",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Bool,
							"baz": cty.Number,
						}),
					},
				},
			},
			`attr = {
  bar = true
}
`,
			hcl.Pos{Line: 2, Column: 5, Byte: 13},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `bar`,
					Detail: "required, bool",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "bar",
						Snippet: "bar = ${1:false}",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
							End:      hcl.Pos{Line: 2, Column: 13, Byte: 21},
						},
					},
				},
				{
					Label:  `baz`,
					Detail: "required, number",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "baz",
						Snippet: "baz = ${1:0}",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
							End:      hcl.Pos{Line: 2, Column: 13, Byte: 21},
						},
					},
				},
			}),
		},
		{
			"inside object value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Bool,
							"baz": cty.Number,
						}),
					},
				},
			},
			`attr = {
  bar = false
}
`,
			hcl.Pos{Line: 2, Column: 10, Byte: 18},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `false`,
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "false",
						Snippet: "false",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 9, Byte: 17},
							End:      hcl.Pos{Line: 2, Column: 14, Byte: 22},
						},
					},
				},
			}),
		},
		{
			"inside object with incomplete key",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Bool,
							"baz": cty.Number,
						}),
					},
				},
			},
			`attr = {
  ba
}
`,
			hcl.Pos{Line: 2, Column: 5, Byte: 13},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `bar`,
					Detail: "required, bool",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "bar",
						Snippet: "bar = ${1:false}",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
							End:      hcl.Pos{Line: 2, Column: 5, Byte: 13},
						},
					},
				},
				{
					Label:  `baz`,
					Detail: "required, number",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "baz",
						Snippet: "baz = ${1:0}",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
							End:      hcl.Pos{Line: 2, Column: 5, Byte: 13},
						},
					},
				},
			}),
		},
		{
			"inside object with no value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Bool,
							"baz": cty.Number,
						}),
					},
				},
			},
			`attr = {
  bar = 
}
`,
			hcl.Pos{Line: 2, Column: 9, Byte: 17},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `false`,
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "false",
						Snippet: "false",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 9, Byte: 17},
							End:      hcl.Pos{Line: 2, Column: 9, Byte: 17},
						},
					},
				},
				{
					Label:  `true`,
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "true",
						Snippet: "true",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 9, Byte: 17},
							End:      hcl.Pos{Line: 2, Column: 9, Byte: 17},
						},
					},
				},
			}),
		},
		{
			"inside object with incomplete value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"foo": cty.String,
							"bar": cty.Bool,
							"baz": cty.Number,
						}),
					},
				},
			},
			`attr = {
  bar = f
}
`,
			hcl.Pos{Line: 2, Column: 10, Byte: 18},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `false`,
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "false",
						Snippet: "false",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 9, Byte: 17},
							End:      hcl.Pos{Line: 2, Column: 10, Byte: 18},
						},
					},
				},
			}),
		},

		{
			"map expr inside object",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"mymap": cty.Map(cty.String),
						}),
					},
				},
			},
			`attr = {

}
`,
			hcl.Pos{Line: 2, Column: 1, Byte: 9},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "mymap",
					Detail: "required, map of string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 1, Byte: 9},
							End:      hcl.Pos{Line: 2, Column: 1, Byte: 9},
						},
						NewText: "mymap",
						Snippet: "mymap = {\n  \"${1:name}\" = \"${2:value}\"\n}",
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"new map entry inside object",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"mymap": cty.Map(cty.String),
						}),
					},
				},
			},
			`attr = {
  mymap = 
}
`,
			hcl.Pos{Line: 2, Column: 11, Byte: 19},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `{ "key" = string }`,
					Detail: "map of string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 11, Byte: 19},
							End:      hcl.Pos{Line: 2, Column: 11, Byte: 19},
						},
						NewText: "{\n  \"key\" = \"\"\n}",
						Snippet: "{\n  \"${1:key}\" = \"${2:value}\"\n}",
					},
					Kind: lang.MapCandidateKind,
				},
			}),
		},
		{
			"inside map expr inside object",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"mymap": cty.Map(cty.String),
						}),
					},
				},
			},
			`attr = {
  mymap = {
    
  }
}
`,
			hcl.Pos{Line: 3, Column: 5, Byte: 25},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "\"key\" = string",
					Detail: "string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 5, Byte: 25},
							End:      hcl.Pos{Line: 3, Column: 5, Byte: 25},
						},
						NewText: "\"key\" = \"value\"",
						Snippet: "\"${1:key}\" = \"${2:value}\"",
					},
					Kind: lang.AttributeCandidateKind,
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
				t.Fatalf("unexpected candidates: %s", diff)
			}
		})
	}
}

func TestCompletionAtPos_exprAny_references(t *testing.T) {
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
					Constraint: schema.AnyExpression{
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
					Constraint: schema.AnyExpression{
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
					Constraint: schema.AnyExpression{
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
					Constraint: schema.AnyExpression{
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
					Constraint: schema.AnyExpression{
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

func TestCompletionAtPos_exprAny_skipComplex(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		funcSignatures     map[string]schema.FunctionSignature
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"complex map expression",
			map[string]*schema.AttributeSchema{
				"tags": {
					Constraint: schema.OneOf{
						schema.AnyExpression{
							OfType:                  cty.Map(cty.String),
							SkipLiteralComplexTypes: true,
						},
						schema.Map{
							Elem: schema.AnyExpression{OfType: cty.String},
						},
					},
				},
			},
			map[string]schema.FunctionSignature{},
			`tags = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `{ "key" = string }`,
					Detail: "map of string",
					Kind:   lang.MapCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "{\n  \n}",
						Snippet: "{\n  ${1}\n}",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
					TriggerSuggest: true,
				},
			}),
		},
		{
			"complex map expression inside brackets",
			map[string]*schema.AttributeSchema{
				"tags": {
					Constraint: schema.OneOf{
						schema.AnyExpression{
							OfType:                  cty.Map(cty.String),
							SkipLiteralComplexTypes: true,
						},
						schema.Map{
							Elem: schema.AnyExpression{OfType: cty.String},
						},
					},
				},
			},
			testFunctionSignatures(),
			`tags = {
  
}
`,
			hcl.Pos{Line: 2, Column: 1, Byte: 11},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `"key" = string`,
					Detail: "string",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: `"key" = `,
						Snippet: `"${1:key}" = `,
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 1, Byte: 11},
							End:      hcl.Pos{Line: 2, Column: 1, Byte: 11},
						},
					},
				},
			}),
		},
		{
			"complex map expression with prefix",
			map[string]*schema.AttributeSchema{
				"tags": {
					Constraint: schema.OneOf{
						schema.AnyExpression{
							OfType:                  cty.Map(cty.String),
							SkipLiteralComplexTypes: true,
						},
						schema.Map{
							Elem: schema.AnyExpression{OfType: cty.String},
						},
					},
				},
			},
			testFunctionSignatures(),
			`tags = {
  "attr" = j
}
`,
			hcl.Pos{Line: 2, Column: 13, Byte: 21},
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
							Start:    hcl.Pos{Line: 2, Column: 12, Byte: 20},
							End:      hcl.Pos{Line: 2, Column: 13, Byte: 21},
						},
					},
				},
			}),
		},
		// TODO test for object
		// TODO test for list
		// TODO test for set
		// TODO test for tuple
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
				Functions: tc.funcSignatures,
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

func TestCompletionAtPos_exprAny_operators(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		refTargets         reference.Targets
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"logical OR",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
					},
				},
			},
			reference.Targets{},
			`attr = true || f
`,
			hcl.Pos{Line: 1, Column: 17, Byte: 16},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "false",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "false",
						Snippet: "false",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 16, Byte: 15},
							End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
						},
					},
				},
			}),
		},
		{
			"logical OR mismatching constraint",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Number,
					},
				},
			},
			reference.Targets{},
			`attr = true || f
`,
			hcl.Pos{Line: 1, Column: 17, Byte: 16},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"adding",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Number,
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
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.String,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.Number,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "list"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.List(cty.String),
				},
			},
			`attr = 42 + v
`,
			hcl.Pos{Line: 1, Column: 14, Byte: 13},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.bar",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.bar",
						Snippet: "var.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 13, Byte: 12},
							End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
						},
					},
				},
				{
					Label:  "var.foo",
					Detail: "number",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.foo",
						Snippet: "var.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 13, Byte: 12},
							End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
						},
					},
				},
			}),
		},
		{
			"negation",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
					},
				},
			},
			reference.Targets{},
			`attr = !f
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "false",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "false",
						Snippet: "false",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
							End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
						},
					},
				},
			}),
		},
		{
			"unterminated expression with unary op",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
					},
				},
			},
			reference.Targets{},
			`attr = !
`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
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
							Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
							End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
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
							Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
							End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
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
							Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
							End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
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
							Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
							End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
				{
					Label:       "namespaced::function",
					Description: lang.Markdown("Example for hcl valid namespaced function"),
					Detail:      "namespaced::function() bool",
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "namespaced::function()",
						Snippet: "namespaced::function(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
							End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
				{
					Label:       "provider::framework::example",
					Detail:      "provider::framework::example(input string) string",
					Description: lang.Markdown("Echoes given argument as result"),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "provider::framework::example()",
						Snippet: "provider::framework::example(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
							End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
				{
					Label:  "false",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "false",
						Snippet: "false",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
							End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
				{
					Label:  "true",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "true",
						Snippet: "true",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
							End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
			}),
		},
		{
			"unterminated expression with unary op and trailing dot",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
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
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.Bool,
				},
			},
			`attr = !var.
`,
			hcl.Pos{Line: 1, Column: 13, Byte: 12},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.bar",
					Detail: "bool",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.bar",
						Snippet: "var.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
							End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
						},
					},
				},
			}),
		},
		{
			"unterminated expression with binary op",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Number,
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
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.String,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.Number,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "list"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.List(cty.String),
				},
			},
			`attr = 42 +
`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			lang.CompleteCandidates([]lang.Candidate{
				// TODO: See https://github.com/hashicorp/hcl-lang/issues/321
			}),
		},
		{
			"unterminated expression with binary op and trailing dot",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Number,
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
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.String,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.Number,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "list"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.List(cty.String),
				},
			},
			`attr = 42 + var.
`,
			hcl.Pos{Line: 1, Column: 17, Byte: 16},
			lang.CompleteCandidates([]lang.Candidate{
				// TODO: See https://github.com/hashicorp/hcl-lang/issues/321
			}),
		},
		{
			"unary expression with function",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.String,
				},
			},
			`attr = !lower()
`,
			hcl.Pos{Line: 1, Column: 15, Byte: 14},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.foo",
						Snippet: "var.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
							End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
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
							Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
							End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
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
							Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
							End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
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
							Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
							End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
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
							Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
							End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
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
							Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
							End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
						},
					},
				},
				{
					Label:       "namespaced::function",
					Description: lang.Markdown("Example for hcl valid namespaced function"),
					Detail:      "namespaced::function() bool",
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "namespaced::function()",
						Snippet: "namespaced::function(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
							End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
						},
					},
				},
				{
					Label:       "provider::framework::example",
					Detail:      "provider::framework::example(input string) string",
					Description: lang.Markdown("Echoes given argument as result"),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "provider::framework::example()",
						Snippet: "provider::framework::example(${0})",
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
			"unary expression with unterminated function",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.String,
				},
			},
			`attr = !lower(
`,
			hcl.Pos{Line: 1, Column: 15, Byte: 14},
			lang.CompleteCandidates([]lang.Candidate{
				// TODO: See https://github.com/hashicorp/hcl-lang/issues/325
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
				Functions:        testFunctionSignatures(),
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

func TestCompletionAtPos_exprAny_parentheses(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		refTargets         reference.Targets
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"empty bool parentheses",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
					},
				},
			},
			reference.Targets{},
			`attr = ()
`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "false",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "false",
						Snippet: "false",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
							End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
				{
					Label:  "true",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "true",
						Snippet: "true",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
							End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
			}),
		},
		{
			"empty number parentheses",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Number,
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.Number,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.Bool,
				},
			},
			`attr = ()
`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "number",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.foo",
						Snippet: "var.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
							End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
			}),
		},
		{
			"unterminated empty parentheses",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
					},
				},
			},
			reference.Targets{},
			`attr = (
`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			lang.CompleteCandidates([]lang.Candidate{
				// TODO: See https://github.com/hashicorp/hcl/issues/649
			}),
		},
		{
			"parentheses with prefix",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
					},
				},
			},
			reference.Targets{},
			`attr = (t)
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "true",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "true",
						Snippet: "true",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
							End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
						},
					},
				},
			}),
		},
		{
			"parentheses with prefix with trailing dot",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.Bool,
				},
			},
			`attr = (var.)
`,
			hcl.Pos{Line: 1, Column: 13, Byte: 12},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "bool",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.foo",
						Snippet: "var.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
							End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
						},
					},
				},
			}),
		},
		{
			"empty parentheses as map key",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
				},
			},
			`attr = {
  () = "foo"
}
`,
			hcl.Pos{Line: 2, Column: 4, Byte: 12},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.foo",
						Snippet: "var.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 4, Byte: 12},
							End:      hcl.Pos{Line: 2, Column: 4, Byte: 12},
						},
					},
				},
			}),
		},
		{
			"parentheses with prefix as map key",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
				},
			},
			`attr = {
  (var) = "foo"
}
`,
			hcl.Pos{Line: 2, Column: 7, Byte: 15},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.foo",
						Snippet: "var.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 4, Byte: 12},
							End:      hcl.Pos{Line: 2, Column: 7, Byte: 15},
						},
					},
				},
			}),
		},
		{
			"empty parentheses as map key in static map",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.LiteralType{Type: cty.String},
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
				},
			},
			`attr = {
  () = "foo"
}
`,
			hcl.Pos{Line: 2, Column: 4, Byte: 12},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"parentheses with prefix as map key in static map",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.LiteralType{Type: cty.String},
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
				},
			},
			`attr = {
  (var) = "foo"
}
`,
			hcl.Pos{Line: 2, Column: 7, Byte: 15},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"empty parentheses as object attribute name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"bar": cty.String,
						}),
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
				},
			},
			`attr = {
  () = "foo"
}
`,
			hcl.Pos{Line: 2, Column: 4, Byte: 12},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.foo",
						Snippet: "var.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 4, Byte: 12},
							End:      hcl.Pos{Line: 2, Column: 4, Byte: 12},
						},
					},
				},
			}),
		},
		{
			"parentheses with prefix as object attribute name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"bar": cty.String,
						}),
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
				},
			},
			`attr = {
  (var) = "foo"
}
`,
			hcl.Pos{Line: 2, Column: 7, Byte: 15},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.foo",
						Snippet: "var.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 4, Byte: 12},
							End:      hcl.Pos{Line: 2, Column: 7, Byte: 15},
						},
					},
				},
			}),
		},
		{
			"empty parentheses as object attribute name in static object",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": &schema.AttributeSchema{
								Constraint: schema.LiteralType{Type: cty.String},
							},
						},
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
				},
			},
			`attr = {
  () = "foo"
}
`,
			hcl.Pos{Line: 2, Column: 4, Byte: 12},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"parentheses with prefix as map key in static map",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": &schema.AttributeSchema{
								Constraint: schema.LiteralType{Type: cty.String},
							},
						},
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
				},
			},
			`attr = {
  (var) = "foo"
}
`,
			hcl.Pos{Line: 2, Column: 7, Byte: 15},
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

func TestCompletionAtPos_exprAny_forExpr(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		refTargets         reference.Targets
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		// list
		{
			"list on collection",
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
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.List(cty.String),
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "mystring"},
					},
					Type: cty.String,
				},
			},
			`attr = [for i, v in var: v]
`,
			hcl.Pos{Line: 1, Column: 24, Byte: 23},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "list of string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.foo",
						Snippet: "var.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 21, Byte: 20},
							End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
						},
					},
				},
			}),
		},
		{
			"list on value",
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
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.List(cty.String),
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.String,
				},
			},
			`attr = [for i, v in var: v]
`,
			hcl.Pos{Line: 1, Column: 27, Byte: 26},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.bar",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.bar",
						Snippet: "var.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 26, Byte: 25},
							End:      hcl.Pos{Line: 1, Column: 27, Byte: 26},
						},
					},
				},
			}),
		},
		{
			"list on condition",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
					},
				},
			},
			reference.Targets{},
			`attr = [for i, v in var: v if t]
`,
			hcl.Pos{Line: 1, Column: 32, Byte: 31},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "true",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "true",
						Snippet: "true",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 31, Byte: 30},
							End:      hcl.Pos{Line: 1, Column: 32, Byte: 31},
						},
					},
				},
			}),
		},

		// set
		{
			"set on collection",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Set(cty.String),
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.Set(cty.String),
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "mystring"},
					},
					Type: cty.String,
				},
			},
			`attr = [for i, v in var: v]
`,
			hcl.Pos{Line: 1, Column: 24, Byte: 23},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "set of string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.foo",
						Snippet: "var.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 21, Byte: 20},
							End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
						},
					},
				},
			}),
		},
		{
			"set on value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Set(cty.String),
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.Set(cty.String),
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.String,
				},
			},
			`attr = [for i, v in var: v]
`,
			hcl.Pos{Line: 1, Column: 27, Byte: 26},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.bar",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.bar",
						Snippet: "var.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 26, Byte: 25},
							End:      hcl.Pos{Line: 1, Column: 27, Byte: 26},
						},
					},
				},
			}),
		},
		{
			"set on condition",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Set(cty.String),
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.Set(cty.String),
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.String,
				},
			},
			`attr = [for i, v in var: v if t]
`,
			hcl.Pos{Line: 1, Column: 32, Byte: 31},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "true",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "true",
						Snippet: "true",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 31, Byte: 30},
							End:      hcl.Pos{Line: 1, Column: 32, Byte: 31},
						},
					},
				},
			}),
		},

		// tuple
		{
			"tuple on collection",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.EmptyTuple,
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.EmptyTuple,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "mystring"},
					},
					Type: cty.String,
				},
			},
			`attr = [for k, v in var: v]
`,
			hcl.Pos{Line: 1, Column: 24, Byte: 23},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "tuple",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.foo",
						Snippet: "var.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 21, Byte: 20},
							End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
						},
					},
				},
			}),
		},
		{
			"tuple on value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.EmptyTuple,
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.EmptyTuple,
				},
			},
			`attr = [for k, v in var: v]
`,
			hcl.Pos{Line: 1, Column: 27, Byte: 26},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.foo",
						Snippet: "var.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 26, Byte: 25},
							End:      hcl.Pos{Line: 1, Column: 27, Byte: 26},
						},
					},
				},
				{
					Label:  "var.bar",
					Detail: "tuple",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.bar",
						Snippet: "var.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 26, Byte: 25},
							End:      hcl.Pos{Line: 1, Column: 27, Byte: 26},
						},
					},
				},
			}),
		},
		{
			"tuple on condition",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.EmptyTuple,
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.EmptyTuple,
				},
			},
			`attr = [for k, v in var: v if f]
`,
			hcl.Pos{Line: 1, Column: 32, Byte: 31},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "false",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "false",
						Snippet: "false",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 31, Byte: 30},
							End:      hcl.Pos{Line: 1, Column: 32, Byte: 31},
						},
					},
				},
			}),
		},

		// map
		{
			"map on collection",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.Map(cty.String),
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "mystring"},
					},
					Type: cty.String,
				},
			},
			`attr = {for k, v in var: k => v}
`,
			hcl.Pos{Line: 1, Column: 24, Byte: 23},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "map of string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.foo",
						Snippet: "var.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 21, Byte: 20},
							End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
						},
					},
				},
			}),
		},
		{
			"map on key",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.Number),
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.List(cty.String),
				},
			},
			`attr = {for k, v in var: v => val}
`,
			hcl.Pos{Line: 1, Column: 27, Byte: 26},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.foo",
						Snippet: "var.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 26, Byte: 25},
							End:      hcl.Pos{Line: 1, Column: 27, Byte: 26},
						},
					},
				},
			}),
		},
		{
			"map on value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.List(cty.String),
				},
			},
			`attr = {for k, v in var: k => v}
`,
			hcl.Pos{Line: 1, Column: 32, Byte: 31},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.foo",
						Snippet: "var.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 31, Byte: 30},
							End:      hcl.Pos{Line: 1, Column: 32, Byte: 31},
						},
					},
				},
			}),
		},
		{
			"map on condition",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
					},
				},
			},
			reference.Targets{},
			`attr = {for k, v in var: v if f}
`,
			hcl.Pos{Line: 1, Column: 32, Byte: 31},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "false",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "false",
						Snippet: "false",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 31, Byte: 30},
							End:      hcl.Pos{Line: 1, Column: 32, Byte: 31},
						},
					},
				},
			}),
		},

		// object
		{
			"object on collection",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.EmptyObject,
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.EmptyObject,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "mystring"},
					},
					Type: cty.String,
				},
			},
			`attr = {for k, v in var: k => v}
`,
			hcl.Pos{Line: 1, Column: 24, Byte: 23},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "object",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.foo",
						Snippet: "var.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 21, Byte: 20},
							End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
						},
					},
				},
			}),
		},
		{
			"object on key",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.EmptyObject,
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.List(cty.String),
				},
			},
			`attr = {for k, v in var: v => val}
`,
			hcl.Pos{Line: 1, Column: 27, Byte: 26},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.foo",
						Snippet: "var.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 26, Byte: 25},
							End:      hcl.Pos{Line: 1, Column: 27, Byte: 26},
						},
					},
				},
			}),
		},
		{
			"object on value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.EmptyObject,
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.List(cty.String),
				},
			},
			`attr = {for k, v in var: k => v}
`,
			hcl.Pos{Line: 1, Column: 32, Byte: 31},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.foo",
						Snippet: "var.foo",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 31, Byte: 30},
							End:      hcl.Pos{Line: 1, Column: 32, Byte: 31},
						},
					},
				},
				{
					Label:  "var.bar",
					Detail: "list of string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.bar",
						Snippet: "var.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 31, Byte: 30},
							End:      hcl.Pos{Line: 1, Column: 32, Byte: 31},
						},
					},
				},
			}),
		},
		{
			"object on condition",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.EmptyObject,
					},
				},
			},
			reference.Targets{},
			`attr = {for k, v in var: v if f}
`,
			hcl.Pos{Line: 1, Column: 32, Byte: 31},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "false",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "false",
						Snippet: "false",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 31, Byte: 30},
							End:      hcl.Pos{Line: 1, Column: 32, Byte: 31},
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
				t.Fatalf("unexpected candidates: %s", diff)
			}
		})
	}
}

func TestCompletionAtPos_exprAny_template(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		refTargets         reference.Targets
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"simple empty template (wrapped)",
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
						lang.AttrStep{Name: "bar"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.String,
				},
			},
			`attr = "${}"
`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.bar",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.bar",
						Snippet: "var.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
							End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
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
							Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
							End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
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
							Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
							End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
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
							Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
							End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
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
							Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
							End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
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
							Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
							End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
						},
					},
				},
				{
					Label:       "namespaced::function",
					Description: lang.Markdown("Example for hcl valid namespaced function"),
					Detail:      "namespaced::function() bool",
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "namespaced::function()",
						Snippet: "namespaced::function(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
							End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
						},
					},
				},
				{
					Label:       "provider::framework::example",
					Detail:      "provider::framework::example(input string) string",
					Description: lang.Markdown("Echoes given argument as result"),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "provider::framework::example()",
						Snippet: "provider::framework::example(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
							End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
						},
					},
				},
			}),
		},
		{
			"simple empty template with prefix (wrapped)",
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
						lang.AttrStep{Name: "bar"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.String,
				},
			},
			`attr = "${v}"
`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.bar",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.bar",
						Snippet: "var.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
							End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
						},
					},
				},
			}),
		},
		{
			"simple empty template with prefix trailing dot (wrapped)",
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
						lang.AttrStep{Name: "bar"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.String,
				},
			},
			`attr = "${var.}"
`,
			hcl.Pos{Line: 1, Column: 15, Byte: 14},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.bar",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.bar",
						Snippet: "var.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
							End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
						},
					},
				},
			}),
		},
		{
			"multi line template",
			map[string]*schema.AttributeSchema{
				"content": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
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
					Type: cty.String,
				},
			},
			`content = <<-EOT
  [Service]
  User=${}
EOT
`,
			hcl.Pos{Line: 3, Column: 10, Byte: 38},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.bar",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.bar",
						Snippet: "var.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 10, Byte: 38},
							End:      hcl.Pos{Line: 3, Column: 10, Byte: 38},
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
							Start:    hcl.Pos{Line: 3, Column: 10, Byte: 38},
							End:      hcl.Pos{Line: 3, Column: 10, Byte: 38},
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
							Start:    hcl.Pos{Line: 3, Column: 10, Byte: 38},
							End:      hcl.Pos{Line: 3, Column: 10, Byte: 38},
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
							Start:    hcl.Pos{Line: 3, Column: 10, Byte: 38},
							End:      hcl.Pos{Line: 3, Column: 10, Byte: 38},
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
							Start:    hcl.Pos{Line: 3, Column: 10, Byte: 38},
							End:      hcl.Pos{Line: 3, Column: 10, Byte: 38},
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
							Start:    hcl.Pos{Line: 3, Column: 10, Byte: 38},
							End:      hcl.Pos{Line: 3, Column: 10, Byte: 38},
						},
					},
				},
				{
					Label:       "namespaced::function",
					Description: lang.Markdown("Example for hcl valid namespaced function"),
					Detail:      "namespaced::function() bool",
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "namespaced::function()",
						Snippet: "namespaced::function(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 10, Byte: 38},
							End:      hcl.Pos{Line: 3, Column: 10, Byte: 38},
						},
					},
				},
				{
					Label:       "provider::framework::example",
					Detail:      "provider::framework::example(input string) string",
					Description: lang.Markdown("Echoes given argument as result"),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "provider::framework::example()",
						Snippet: "provider::framework::example(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 10, Byte: 38},
							End:      hcl.Pos{Line: 3, Column: 10, Byte: 38},
						},
					},
				},
			}),
		},
		{
			"multiple expression, partial completion",
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
						lang.AttrStep{Name: "bar"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.String,
				},
			},
			`attr = "foo-${var.bar}-bar-${l}"
`,
			hcl.Pos{Line: 1, Column: 31, Byte: 30},
			lang.CompleteCandidates([]lang.Candidate{
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
							Start:    hcl.Pos{Line: 1, Column: 30, Byte: 29},
							End:      hcl.Pos{Line: 1, Column: 31, Byte: 30},
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
							Start:    hcl.Pos{Line: 1, Column: 30, Byte: 29},
							End:      hcl.Pos{Line: 1, Column: 31, Byte: 30},
						},
					},
				},
			}),
		},
		{
			"no completion between $ and {",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Targets{},
			`attr = "fo${var.bar}bar"
`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"no completion behind {",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			reference.Targets{},
			`attr = "fo${var.bar}"
`,
			hcl.Pos{Line: 1, Column: 21, Byte: 20},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"completion within function within expression",
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
						lang.AttrStep{Name: "bar"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.String,
				},
			},
			`attr = "foo-${element([0], )}-bar"
`,
			hcl.Pos{Line: 1, Column: 28, Byte: 27},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.bar",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.bar",
						Snippet: "var.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
							End:      hcl.Pos{Line: 1, Column: 28, Byte: 27},
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
							Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
							End:      hcl.Pos{Line: 1, Column: 28, Byte: 27},
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
							Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
							End:      hcl.Pos{Line: 1, Column: 28, Byte: 27},
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
							Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
							End:      hcl.Pos{Line: 1, Column: 28, Byte: 27},
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
							Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
							End:      hcl.Pos{Line: 1, Column: 28, Byte: 27},
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
							Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
							End:      hcl.Pos{Line: 1, Column: 28, Byte: 27},
						},
					},
				},
				{
					Label:       "provider::framework::example",
					Detail:      "provider::framework::example(input string) string",
					Description: lang.Markdown("Echoes given argument as result"),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "provider::framework::example()",
						Snippet: "provider::framework::example(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
							End:      hcl.Pos{Line: 1, Column: 28, Byte: 27},
						},
					},
				},
			}),
		},
		{
			"expression within function expression",
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
						lang.AttrStep{Name: "bar"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.String,
				},
			},
			`attr = lower("${var.}-${md5(local.name)}")
`,
			hcl.Pos{Line: 1, Column: 21, Byte: 20},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.bar",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.bar",
						Snippet: "var.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 17, Byte: 16},
							End:      hcl.Pos{Line: 1, Column: 21, Byte: 20},
						},
					},
				},
			}),
		},
		{
			"expression with strip markers",
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
						lang.AttrStep{Name: "bar"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.String,
				},
			},
			`attr = "Hello-${~ v}"
`,
			hcl.Pos{Line: 1, Column: 20, Byte: 19},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.bar",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.bar",
						Snippet: "var.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 19, Byte: 18},
							End:      hcl.Pos{Line: 1, Column: 20, Byte: 19},
						},
					},
				},
			}),
		},
		{
			"expression within list",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.List{
						Elem: schema.AnyExpression{OfType: cty.String},
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
					Type: cty.String,
				},
			},
			`attr = ["foo", "bar-${}"]
`,
			hcl.Pos{Line: 1, Column: 23, Byte: 22},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.bar",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.bar",
						Snippet: "var.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
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
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
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
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
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
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
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
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
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
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
						},
					},
				},
				{
					Label:       "namespaced::function",
					Description: lang.Markdown("Example for hcl valid namespaced function"),
					Detail:      "namespaced::function() bool",
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "namespaced::function()",
						Snippet: "namespaced::function(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
						},
					},
				},
				{
					Label:       "provider::framework::example",
					Detail:      "provider::framework::example(input string) string",
					Description: lang.Markdown("Echoes given argument as result"),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "provider::framework::example()",
						Snippet: "provider::framework::example(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
						},
					},
				},
			}),
		},
		{
			"expression within set",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.AnyExpression{OfType: cty.String},
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
					Type: cty.String,
				},
			},
			`attr = ["foo", "bar-${}"]
`,
			hcl.Pos{Line: 1, Column: 23, Byte: 22},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.bar",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.bar",
						Snippet: "var.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
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
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
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
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
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
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
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
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
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
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
						},
					},
				},
				{
					Label:       "namespaced::function",
					Description: lang.Markdown("Example for hcl valid namespaced function"),
					Detail:      "namespaced::function() bool",
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "namespaced::function()",
						Snippet: "namespaced::function(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
						},
					},
				},
				{
					Label:       "provider::framework::example",
					Detail:      "provider::framework::example(input string) string",
					Description: lang.Markdown("Echoes given argument as result"),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "provider::framework::example()",
						Snippet: "provider::framework::example(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
						},
					},
				},
			}),
		},
		{
			"expression within tuple",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.AnyExpression{OfType: cty.String},
							schema.AnyExpression{OfType: cty.String},
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
					Type: cty.String,
				},
			},
			`attr = ["foo", "bar-${}"]
`,
			hcl.Pos{Line: 1, Column: 23, Byte: 22},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.bar",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "var.bar",
						Snippet: "var.bar",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
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
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
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
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
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
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
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
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
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
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
						},
					},
				},
				{
					Label:       "namespaced::function",
					Description: lang.Markdown("Example for hcl valid namespaced function"),
					Detail:      "namespaced::function() bool",
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "namespaced::function()",
						Snippet: "namespaced::function(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
						},
					},
				},
				{
					Label:       "provider::framework::example",
					Detail:      "provider::framework::example(input string) string",
					Description: lang.Markdown("Echoes given argument as result"),
					Kind:        lang.FunctionCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "provider::framework::example()",
						Snippet: "provider::framework::example(${0})",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
							End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
						},
					},
				},
			}),
		},
		// TODO: test for directive after https://github.com/hashicorp/terraform-ls/issues/527 lands
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
				Functions:        testFunctionSignatures(),
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

func TestCompletionAtPos_exprAny_conditional(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		refTargets         reference.Targets
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"condition part",
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
						lang.AttrStep{Name: "bar"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.Number,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.Bool,
				},
			},
			`attr = v ? bar : baz`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "bool",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
						NewText: "var.foo",
						Snippet: "var.foo",
					},
					Kind: lang.ReferenceCandidateKind,
				},
			}),
		},
		{
			"true part",
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
						lang.AttrStep{Name: "bar"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.Number,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.Bool,
				},
			},
			`attr = cond ? v : baz`,
			hcl.Pos{Line: 1, Column: 16, Byte: 15},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.bar",
					Detail: "number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
							End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
						},
						NewText: "var.bar",
						Snippet: "var.bar",
					},
					Kind: lang.ReferenceCandidateKind,
				},
				{
					Label:  "var.foo",
					Detail: "bool",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 15, Byte: 14},
							End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
						},
						NewText: "var.foo",
						Snippet: "var.foo",
					},
					Kind: lang.ReferenceCandidateKind,
				},
			}),
		},
		{
			"false part",
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
						lang.AttrStep{Name: "bar"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.Number,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.Bool,
				},
			},
			`attr = cond ? bar : v`,
			hcl.Pos{Line: 1, Column: 22, Byte: 21},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.bar",
					Detail: "number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 21, Byte: 20},
							End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
						},
						NewText: "var.bar",
						Snippet: "var.bar",
					},
					Kind: lang.ReferenceCandidateKind,
				},
				{
					Label:  "var.foo",
					Detail: "bool",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 21, Byte: 20},
							End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
						},
						NewText: "var.foo",
						Snippet: "var.foo",
					},
					Kind: lang.ReferenceCandidateKind,
				},
			}),
		},
		{
			"condition in template",
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
						lang.AttrStep{Name: "bar"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.Number,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.Bool,
				},
			},
			`attr = "x-${foo ? bar : v}"`,
			hcl.Pos{Line: 1, Column: 26, Byte: 25},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.bar",
					Detail: "number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 25, Byte: 24},
							End:      hcl.Pos{Line: 1, Column: 26, Byte: 25},
						},
						NewText: "var.bar",
						Snippet: "var.bar",
					},
					Kind: lang.ReferenceCandidateKind,
				},
				{
					Label:  "var.foo",
					Detail: "bool",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 25, Byte: 24},
							End:      hcl.Pos{Line: 1, Column: 26, Byte: 25},
						},
						NewText: "var.foo",
						Snippet: "var.foo",
					},
					Kind: lang.ReferenceCandidateKind,
				},
			}),
		},
		{
			"condition as directive",
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
						lang.AttrStep{Name: "bar"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.Number,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					RangePtr: &hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 3, Byte: 19},
					},
					Type: cty.Bool,
				},
			},
			`attr = "x-%{ if v }bar%{ else }baz%{ endif }"`,
			hcl.Pos{Line: 1, Column: 18, Byte: 17},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.foo",
					Detail: "bool",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 17, Byte: 16},
							End:      hcl.Pos{Line: 1, Column: 18, Byte: 17},
						},
						NewText: "var.foo",
						Snippet: "var.foo",
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
				Functions:        testFunctionSignatures(),
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
