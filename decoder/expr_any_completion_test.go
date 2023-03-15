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
	"github.com/zclconf/go-cty/cty/function"
)

func TestCompletionAtPos_exprAny_functions(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"list of functions",
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
			"function by prefix",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
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
			"first argument of a function ",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = element()
`,
			hcl.Pos{Line: 1, Column: 14, Byte: 15},
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
			}),
		},
		{
			"second argument of a function ",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
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
			}),
		},
		{
			"nested functions",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
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
				Functions: testFunctionSignatures(),
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
