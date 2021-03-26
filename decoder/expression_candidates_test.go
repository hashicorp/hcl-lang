package decoder

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestDecoder_CandidateAtPos_expressions(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"string type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.LiteralTypeOnly(cty.String),
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.ZeroCandidates(),
		},
		{
			"object as value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.LiteralValue{
							Val: cty.ObjectVal(map[string]cty.Value{
								"first":  cty.StringVal("blah"),
								"second": cty.NumberIntVal(2345),
							}),
						},
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "{ first = \"blah\", … }",
					Detail: "object",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
						},
						NewText: `{
  first = "blah"
  second = 2345
}`,
						Snippet: `{
  first = "${1:blah}"
  second = ${2:2345}
}`,
					},
					Kind: lang.ObjectCandidateKind,
				},
			}),
		},
		{
			"object as type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{
						"first":  cty.String,
						"second": cty.Number,
					})),
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "{ first = string, … }",
					Detail: "object",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
						},
						NewText: `{
  first = ""
  second = 1
}`,
						Snippet: `{
  first = "${1:value}"
  second = ${2:1}
}`,
					},
					Kind: lang.ObjectCandidateKind,
				},
			}),
		},
		{
			"object as expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.ObjectExpr{
							Attributes: schema.ObjectExprAttributes{
								"first": schema.ObjectAttribute{
									Expr: schema.LiteralTypeOnly(cty.String),
								},
								"second": schema.ObjectAttribute{
									Expr: schema.LiteralTypeOnly(cty.Number),
								},
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
					Label:  "{ }",
					Detail: "object",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
						},
						NewText: "{\n  \n}",
						Snippet: "{\n  ${1}\n}",
					},
					Kind:           lang.ObjectCandidateKind,
					TriggerSuggest: true,
				},
			}),
		},
		{
			"object as expression - attribute",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.ObjectExpr{
							Attributes: schema.ObjectExprAttributes{
								"first": schema.ObjectAttribute{
									Expr: schema.LiteralTypeOnly(cty.String),
								},
								"second": schema.ObjectAttribute{
									Expr: schema.LiteralTypeOnly(cty.Number),
								},
							},
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
					Label:  "first",
					Detail: "string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   2,
								Column: 3,
								Byte:   11,
							},
							End: hcl.Pos{
								Line:   2,
								Column: 3,
								Byte:   11,
							},
						},
						NewText: `first = ""`,
						Snippet: `first = "${1:value}"`,
					},
					Kind: lang.AttributeCandidateKind,
				},
				{
					Label:  "second",
					Detail: "number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   2,
								Column: 3,
								Byte:   11,
							},
							End: hcl.Pos{
								Line:   2,
								Column: 3,
								Byte:   11,
							},
						},
						NewText: "second = 1",
						Snippet: "second = ${1:1}",
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"object as expression - attributes partially declared",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.ObjectExpr{
							Attributes: schema.ObjectExprAttributes{
								"first": schema.ObjectAttribute{
									Expr: schema.LiteralTypeOnly(cty.String),
								},
								"second": schema.ObjectAttribute{
									Expr: schema.LiteralTypeOnly(cty.Number),
								},
							},
						},
					},
				},
			},
			`attr = {
  first = "blah"
  
}
`,
			hcl.Pos{Line: 3, Column: 3, Byte: 28},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "second",
					Detail: "number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   3,
								Column: 3,
								Byte:   28,
							},
							End: hcl.Pos{
								Line:   3,
								Column: 3,
								Byte:   28,
							},
						},
						NewText: "second = 1",
						Snippet: "second = ${1:1}",
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"object as expression - attribute key unknown",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.ObjectExpr{
							Attributes: schema.ObjectExprAttributes{
								"first": schema.ObjectAttribute{
									Expr: schema.LiteralTypeOnly(cty.String),
								},
								"second": schema.ObjectAttribute{
									Expr: schema.LiteralTypeOnly(cty.Number),
								},
							},
						},
					},
				},
			},
			`attr = {
  first = "blah"
  var.test = "foo"
  "${var.env}.${another}" = "prod"
  
}
`,
			hcl.Pos{Line: 5, Column: 3, Byte: 82},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "second",
					Detail: "number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   5,
								Column: 3,
								Byte:   82,
							},
							End: hcl.Pos{
								Line:   5,
								Column: 3,
								Byte:   82,
							},
						},
						NewText: "second = 1",
						Snippet: "second = ${1:1}",
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"list as value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.LiteralValue{
							Val: cty.ListVal([]cty.Value{
								cty.StringVal("foo"),
								cty.StringVal("bar"),
							}),
						},
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `[ "foo", "bar" ]`,
					Detail: "list",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
						},
						NewText: "[\n  \"foo\",\n  \"bar\",\n]",
						Snippet: "[\n  \"${1:foo}\",\n  \"${2:bar}\",\n]",
					},
					Kind: lang.ListCandidateKind,
				},
			}),
		},
		{
			"map as type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.LiteralTypeOnly(cty.Map(cty.String)),
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `{ "key" = string }`,
					Detail: "map of string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
						},
						NewText: `{
  "key" = ""
}`,
						Snippet: `{
  "${1:key}" = "${2:value}"
}`,
					},
					Kind: lang.MapCandidateKind,
				},
			}),
		},
		{
			"map as value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.LiteralValue{
							Val: cty.MapVal(map[string]cty.Value{
								"foo": cty.StringVal("moo"),
								"bar": cty.StringVal("boo"),
							}),
						},
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `{ "bar" = "boo", … }`,
					Detail: "map",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
						},
						NewText: `{
  "bar" = "boo"
  "foo" = "moo"
}`,
						Snippet: `{
  "${1:bar}" = "${2:boo}"
  "${3:foo}" = "${4:moo}"
}`,
					},
					Kind: lang.MapCandidateKind,
				},
			}),
		},
		{
			"bool type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.LiteralTypeOnly(cty.Bool),
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "true",
					Detail: "bool",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
						},
						NewText: "true",
						Snippet: "${1:true}",
					},
					Kind: lang.BoolCandidateKind,
				},
				{
					Label:  "false",
					Detail: "bool",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
						},
						NewText: "false",
						Snippet: "${1:false}",
					},
					Kind: lang.BoolCandidateKind,
				},
			}),
		},
		{
			"string values",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.LiteralValue{Val: cty.StringVal("first")},
						schema.LiteralValue{Val: cty.StringVal("second")},
						schema.LiteralValue{Val: cty.StringVal("third")},
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "first",
					Detail: "string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
						},
						NewText: `"first"`,
						Snippet: `"${1:first}"`},
					Kind: lang.StringCandidateKind,
				},
				{
					Label:  "second",
					Detail: "string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
						},
						NewText: `"second"`,
						Snippet: `"${1:second}"`},
					Kind: lang.StringCandidateKind,
				},
				{
					Label:  "third",
					Detail: "string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
						},
						NewText: `"third"`,
						Snippet: `"${1:third}"`},
					Kind: lang.StringCandidateKind,
				},
			}),
		},
		{
			"tuple constant expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.TupleConsExpr{
							AnyElem: schema.ExprConstraints{
								schema.LiteralValue{Val: cty.StringVal("one")},
								schema.LiteralValue{Val: cty.StringVal("two")},
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
					Label: "[  ]",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
						},
						NewText: "[ ]",
						Snippet: "[ ${0} ]",
					},
					Kind:           lang.TupleCandidateKind,
					TriggerSuggest: true,
				},
			}),
		},
		{
			"tuple constant expression inside",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.TupleConsExpr{
							AnyElem: schema.ExprConstraints{
								schema.LiteralValue{Val: cty.StringVal("one")},
								schema.LiteralValue{Val: cty.StringVal("two")},
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
					Label:  "one",
					Detail: "string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   1,
								Column: 10,
								Byte:   9,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 10,
								Byte:   9,
							},
						},
						NewText: `"one"`,
						Snippet: `"${1:one}"`,
					},
					Kind: lang.StringCandidateKind,
				},
				{
					Label:  "two",
					Detail: "string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   1,
								Column: 10,
								Byte:   9,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 10,
								Byte:   9,
							},
						},
						NewText: `"two"`,
						Snippet: `"${1:two}"`,
					},
					Kind: lang.StringCandidateKind,
				},
			}),
		},
		{
			"tuple as list type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.LiteralTypeOnly(cty.List(cty.String)),
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "[ string ]",
					Detail: "list of string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
						},
						NewText: `[ "" ]`,
						Snippet: `[ "${1:value}" ]`,
					},
					Kind: lang.ListCandidateKind,
				},
			}),
		},
		{
			"tuple as list type inside",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.LiteralTypeOnly(cty.List(cty.String)),
				},
			},
			`attr = [  ]
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.ZeroCandidates(),
		},
		{
			"keyword",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.KeywordExpr{
							Keyword: "foobar",
							Name:    "special kw",
						},
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "foobar",
					Detail: "special kw",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
						},
						NewText: "foobar",
						Snippet: "foobar",
					},
					Kind: lang.KeywordCandidateKind,
				},
			}),
		},
		{
			"map expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.MapExpr{
							Elem: schema.LiteralTypeOnly(cty.String),
							Name: "map of something",
						},
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "{ key = string }",
					Detail: "map of something",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
						},
						NewText: "{\n  name = \"\"\n}",
						Snippet: "{\n  ${1:name} = \"${1:value}\"\n}",
					},
					Kind:           lang.MapCandidateKind,
					TriggerSuggest: true,
				},
			}),
		},
		{
			"map expression of tuple expr",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.MapExpr{
							Elem: schema.ExprConstraints{
								schema.TupleConsExpr{
									Name:    "special tuple",
									AnyElem: schema.LiteralTypeOnly(cty.Number),
								},
							},
							Name: "special map",
						},
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "{ key = [ number ] }",
					Detail: "special map",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 8,
								Byte:   7,
							},
						},
						NewText: `{
  name = [  ]
}`,
						Snippet: `{
  ${1:name} = [ ${2} ]
}`,
					},
					Kind:           lang.MapCandidateKind,
					TriggerSuggest: true,
				},
			}),
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			d := NewDecoder()
			d.SetSchema(&schema.BodySchema{
				Attributes: tc.attrSchema,
			})

			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			err := d.LoadFile("test.tf", f)
			if err != nil {
				t.Fatal(err)
			}

			candidates, err := d.CandidatesAtPos("test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedCandidates, candidates); diff != "" {
				t.Fatalf("unexpected candidates: %s", diff)
			}
		})
	}
}
