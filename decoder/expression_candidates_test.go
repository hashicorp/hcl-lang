package decoder

import (
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
								"first": &schema.AttributeSchema{
									Expr: schema.LiteralTypeOnly(cty.String),
								},
								"second": &schema.AttributeSchema{
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
								"first": &schema.AttributeSchema{
									Expr: schema.LiteralTypeOnly(cty.String),
								},
								"second": &schema.AttributeSchema{
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
								"first": &schema.AttributeSchema{
									Expr: schema.LiteralTypeOnly(cty.String),
								},
								"second": &schema.AttributeSchema{
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
								"first": &schema.AttributeSchema{
									Expr: schema.LiteralTypeOnly(cty.String),
								},
								"second": &schema.AttributeSchema{
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
			"deprecated string value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.LiteralValue{Val: cty.StringVal("first")},
						schema.LiteralValue{Val: cty.StringVal("second"), IsDeprecated: true},
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
					Label:        "second",
					Detail:       "string",
					IsDeprecated: true,
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
			"attribute as list expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.ListExpr{
							Elem: schema.LiteralTypeOnly(cty.String),
						},
					},
				},
			},
			`
`,
			hcl.Pos{Line: 1, Column: 1, Byte: 0},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "attr",
					Detail: "list of string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:      hcl.Pos{Line: 1, Column: 1, Byte: 0},
						},
						NewText: "attr",
						Snippet: "attr = [\n  ${0}\n]",
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"list expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.ListExpr{
							Elem: schema.LiteralTypeOnly(cty.String),
						},
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
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						NewText: "[ ]",
						Snippet: "[ ${0} ]",
					},
					Kind: lang.ListCandidateKind,
				},
			}),
		},
		{
			"list expression inside",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.ListExpr{
							Elem: schema.LiteralTypeOnly(cty.Bool),
						},
					},
				},
			},
			`attr = [  ]
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "true",
					Detail: "bool",
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
								Column: 10,
								Byte:   9,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 10,
								Byte:   9,
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
			"attribute as set expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.SetExpr{
							Elem: schema.LiteralTypeOnly(cty.String),
						},
					},
				},
			},
			`
`,
			hcl.Pos{Line: 1, Column: 1, Byte: 0},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "attr",
					Detail: "set of string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:      hcl.Pos{Line: 1, Column: 1, Byte: 0},
						},
						NewText: "attr",
						Snippet: "attr = [\n  ${0}\n]",
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"set expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.SetExpr{
							Elem: schema.LiteralTypeOnly(cty.String),
						},
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "[ string ]",
					Detail: "set of string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						NewText: "[ ]",
						Snippet: "[ ${0} ]",
					},
					Kind: lang.SetCandidateKind,
				},
			}),
		},
		{
			"set expression inside",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.SetExpr{
							Elem: schema.LiteralTypeOnly(cty.Bool),
						},
					},
				},
			},
			`attr = [  ]
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "true",
					Detail: "bool",
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
								Column: 10,
								Byte:   9,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 10,
								Byte:   9,
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
			"attribute as tuple expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.TupleExpr{
							Elems: []schema.ExprConstraints{
								schema.LiteralTypeOnly(cty.String),
								schema.LiteralTypeOnly(cty.Number),
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
					Label:  "attr",
					Detail: "tuple",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:      hcl.Pos{Line: 1, Column: 1, Byte: 0},
						},
						NewText: "attr",
						Snippet: "attr = [\n  ${0}\n]",
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"tuple expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.TupleExpr{
							Elems: []schema.ExprConstraints{
								schema.LiteralTypeOnly(cty.String),
								schema.LiteralTypeOnly(cty.Number),
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
					Label:  "[ string ]",
					Detail: "tuple",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						NewText: "[ ]",
						Snippet: "[ ${0} ]",
					},
					Kind: lang.TupleCandidateKind,
				},
			}),
		},
		{
			"tuple expression inside",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.TupleExpr{
							Elems: []schema.ExprConstraints{
								schema.LiteralTypeOnly(cty.Bool),
								schema.LiteralTypeOnly(cty.Number),
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
					Label:  "true",
					Detail: "bool",
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
								Column: 10,
								Byte:   9,
							},
							End: hcl.Pos{
								Line:   1,
								Column: 10,
								Byte:   9,
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
					Kind: lang.MapCandidateKind,
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
					Kind: lang.MapCandidateKind,
				},
			}),
		},
		{
			"literal of dynamic pseudo type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.LiteralTypeExpr{Type: cty.DynamicPseudoType},
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.ZeroCandidates(),
		},
		{
			"type declaration",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.TypeDeclarationExpr{},
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "bool",
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
						}, NewText: "bool", Snippet: "bool"},
					Kind: lang.AttributeCandidateKind,
				},
				{
					Label:  "number",
					Detail: "number",
					TextEdit: lang.TextEdit{Range: hcl.Range{
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
					}, NewText: "number", Snippet: "number"},
					Kind: lang.AttributeCandidateKind,
				},
				{
					Label:  "string",
					Detail: "string",
					TextEdit: lang.TextEdit{Range: hcl.Range{
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
					}, NewText: "string", Snippet: "string"},
					Kind: lang.AttributeCandidateKind,
				},
				{
					Label:  "list()",
					Detail: "list()",
					TextEdit: lang.TextEdit{Range: hcl.Range{
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
					}, NewText: "list()", Snippet: "list(${0})"},
					Kind: lang.AttributeCandidateKind,
				},
				{
					Label:  "set()",
					Detail: "set()",
					TextEdit: lang.TextEdit{Range: hcl.Range{
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
					}, NewText: "set()", Snippet: "set(${0})"},
					Kind: lang.AttributeCandidateKind,
				},
				{
					Label:  "tuple()",
					Detail: "tuple()",
					TextEdit: lang.TextEdit{Range: hcl.Range{
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
					}, NewText: "tuple()", Snippet: "tuple(${0})"},
					Kind: lang.AttributeCandidateKind,
				},
				{
					Label:  "map()",
					Detail: "map()",
					TextEdit: lang.TextEdit{Range: hcl.Range{
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
					}, NewText: "map()", Snippet: "map(${0})"},
					Kind: lang.AttributeCandidateKind,
				},
				{
					Label:  "object({})",
					Detail: "object({})",
					TextEdit: lang.TextEdit{Range: hcl.Range{
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
					}, NewText: "object({})", Snippet: "object({\n ${1:name} = ${2}\n})"},
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

func TestDecoder_CandidateAtPos_traversalExpressions(t *testing.T) {
	testCases := []struct {
		testName           string
		bodySchema         *schema.BodySchema
		builtinRefs        reference.Targets
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"no references",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{OfType: cty.String},
						},
					},
				},
			},
			reference.Targets{},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.ZeroCandidates(),
		},

		{
			"no matching references",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{OfType: cty.String},
						},
					},
				},
			},
			reference.Targets{
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.RootStep{Name: "first"},
					},
					Type: cty.Bool,
				},
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.RootStep{Name: "second"},
					},
					Type: cty.Number,
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.ZeroCandidates(),
		},
		{
			"type matching references",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{OfType: cty.String},
						},
					},
				},
			},
			reference.Targets{
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Type: cty.Bool,
				},
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
					Type: cty.String,
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.second",
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
						NewText: "var.second",
						Snippet: "var.second",
					},
					Kind: lang.TraversalCandidateKind,
				},
			}),
		},
		{
			"reference of any type",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{OfType: cty.DynamicPseudoType},
						},
					},
				},
			},
			reference.Targets{
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Type: cty.Bool,
				},
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
					Type: cty.String,
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.first",
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
						NewText: "var.first",
						Snippet: "var.first",
					},
					Kind: lang.TraversalCandidateKind,
				},
				{
					Label:  "var.second",
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
						NewText: "var.second",
						Snippet: "var.second",
					},
					Kind: lang.TraversalCandidateKind,
				},
			}),
		},
		{
			"reference targets of any type",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{OfType: cty.String},
						},
					},
				},
			},
			reference.Targets{
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Type: cty.DynamicPseudoType,
				},
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
					Type: cty.DynamicPseudoType,
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.first",
					Detail: "dynamic",
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
						NewText: "var.first",
						Snippet: "var.first",
					},
					Kind: lang.TraversalCandidateKind,
				},
				{
					Label:  "var.second",
					Detail: "dynamic",
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
						NewText: "var.second",
						Snippet: "var.second",
					},
					Kind: lang.TraversalCandidateKind,
				},
			}),
		},
		{
			"scope matching references",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{OfScopeId: lang.ScopeId("test")},
						},
					},
				},
			},
			reference.Targets{
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					ScopeId: lang.ScopeId("test"),
				},
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
					ScopeId: lang.ScopeId("second"),
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.first",
					Detail: "reference",
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
						NewText: "var.first",
						Snippet: "var.first",
					},
					Kind: lang.TraversalCandidateKind,
				},
			}),
		},
		{
			"no candidates for addressable traversal",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{
								Address: &schema.TraversalAddrSchema{
									ScopeId: lang.ScopeId("blah"),
								},
								Name: "test",
							},
						},
					},
				},
			},
			reference.Targets{
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "zero"},
					},
					Type: cty.Number,
				},
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					ScopeId: lang.ScopeId("blah"),
				},
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
					ScopeId: lang.ScopeId("another"),
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.ZeroCandidates(),
		},
		{
			"no candidates for addressable traversal in set",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Expr: schema.ExprConstraints{
							schema.SetExpr{
								Elem: schema.ExprConstraints{
									schema.TraversalExpr{
										Address: &schema.TraversalAddrSchema{
											ScopeId: lang.ScopeId("blah"),
										},
										Name: "test",
									},
								},
							},
						},
					},
				},
			},
			reference.Targets{
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "zero"},
					},
					Type: cty.Number,
				},
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					ScopeId: lang.ScopeId("blah"),
				},
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
					ScopeId: lang.ScopeId("another"),
				},
			},
			`attr = [  ]
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.ZeroCandidates(),
		},
		{
			"range filtered references",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"custom": {
						Labels: []*schema.LabelSchema{
							{Name: "name"},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"greeting": {
									Expr:       schema.LiteralTypeOnly(cty.String),
									IsOptional: true,
								},
								"blah": {
									Expr:       schema.LiteralTypeOnly(cty.Bool),
									IsComputed: true,
								},
							},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "custom"},
								schema.LabelStep{Index: 0},
							},
							ScopeId:    lang.ScopeId("custom"),
							BodyAsData: true,
							InferBody:  true,
						},
					},
					"another_block": {
						Labels: []*schema.LabelSchema{
							{Name: "name"},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"one": {
									Expr: schema.ExprConstraints{
										schema.TraversalExpr{OfType: cty.String},
									},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			reference.Targets{},
			`custom "test" {
  greeting = "hello"
}

another_block "meh" {
  one = 
}
`,
			hcl.Pos{Line: 6, Column: 9, Byte: 70},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "custom.test",
					Detail: "object",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   6,
								Column: 9,
								Byte:   70,
							},
							End: hcl.Pos{
								Line:   6,
								Column: 9,
								Byte:   70,
							},
						},
						NewText: `custom.test`,
						Snippet: `custom.test`,
					},
					Kind: lang.TraversalCandidateKind,
				},
			}),
		},
		{
			"partially matching references before dot",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{OfType: cty.String},
						},
					},
				},
			},
			reference.Targets{
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Type: cty.String,
				},
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
					Type: cty.String,
				},
			},
			`attr = var
`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.first",
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
								Column: 11,
								Byte:   10,
							},
						},
						NewText: "var.first",
						Snippet: "var.first",
					},
					Kind: lang.TraversalCandidateKind,
				},
				{
					Label:  "var.second",
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
								Column: 11,
								Byte:   10,
							},
						},
						NewText: "var.second",
						Snippet: "var.second",
					},
					Kind: lang.TraversalCandidateKind,
				},
			}),
		},
		{
			"partially matching references after dot",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{OfType: cty.String},
						},
					},
				},
			},
			reference.Targets{
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Type: cty.String,
				},
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
					Type: cty.String,
				},
			},
			`attr = var.
`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.first",
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
								Column: 12,
								Byte:   11,
							},
						},
						NewText: "var.first",
						Snippet: "var.first",
					},
					Kind: lang.TraversalCandidateKind,
				},
				{
					Label:  "var.second",
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
								Column: 12,
								Byte:   11,
							},
						},
						NewText: "var.second",
						Snippet: "var.second",
					},
					Kind: lang.TraversalCandidateKind,
				},
			}),
		},
		{
			"step-based completion - top level",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{OfType: cty.String},
						},
					},
				},
			},
			reference.Targets{
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Type: cty.Object(map[string]cty.Type{
						"nested": cty.String,
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "first"},
								lang.AttrStep{Name: "nested"},
							},
							Type: cty.String,
						},
					},
				},
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
					Type: cty.String,
				},
			},
			`attr = var.
`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.first",
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
								Column: 12,
								Byte:   11,
							},
						},
						NewText: "var.first",
						Snippet: "var.first",
					},
					Kind: lang.TraversalCandidateKind,
				},
				{
					Label:  "var.second",
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
								Column: 12,
								Byte:   11,
							},
						},
						NewText: "var.second",
						Snippet: "var.second",
					},
					Kind: lang.TraversalCandidateKind,
				},
			}),
		},
		{
			"step-based completion - inside object",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{OfType: cty.String},
						},
					},
				},
			},
			reference.Targets{
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Type: cty.Object(map[string]cty.Type{
						"nested": cty.String,
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "first"},
								lang.AttrStep{Name: "nested"},
							},
							Type: cty.String,
						},
					},
				},
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
					Type: cty.String,
				},
			},
			`attr = var.first.
`,
			hcl.Pos{Line: 1, Column: 18, Byte: 17},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.first.nested",
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
								Column: 18,
								Byte:   17,
							},
						},
						NewText: "var.first.nested",
						Snippet: "var.first.nested",
					},
					Kind: lang.TraversalCandidateKind,
				},
			}),
		},
		{
			"step-based completion - inside list",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{OfType: cty.String},
						},
					},
				},
			},
			reference.Targets{
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Type: cty.List(cty.Object(map[string]cty.Type{
						"nested": cty.String,
					})),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "first"},
								lang.IndexStep{Key: cty.NumberIntVal(0)},
							},
							Type: cty.Object(map[string]cty.Type{
								"nested": cty.String,
							}),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "var"},
										lang.AttrStep{Name: "first"},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
										lang.AttrStep{Name: "nested"},
									},
									Type: cty.String,
								},
							},
						},
					},
				},
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
					Type: cty.String,
				},
			},
			`attr = var.
`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.first",
					Detail: "list of object",
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
								Column: 12,
								Byte:   11,
							},
						},
						NewText: "var.first",
						Snippet: "var.first",
					},
					Kind: lang.TraversalCandidateKind,
				},
				{
					Label:  "var.second",
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
								Column: 12,
								Byte:   11,
							},
						},
						NewText: "var.second",
						Snippet: "var.second",
					},
					Kind: lang.TraversalCandidateKind,
				},
			}),
		},
		{
			"step-based completion - inside map",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{OfType: cty.String},
						},
					},
				},
			},
			reference.Targets{
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Type: cty.Map(cty.String),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "first"},
								lang.IndexStep{Key: cty.StringVal("foo")},
							},
							Type: cty.String,
						},
					},
				},
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
					Type: cty.String,
				},
			},
			`attr = var.
`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.first",
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
								Column: 12,
								Byte:   11,
							},
						},
						NewText: "var.first",
						Snippet: `var.first`,
					},
					Kind: lang.TraversalCandidateKind,
				},
				{
					Label:  "var.second",
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
								Column: 12,
								Byte:   11,
							},
						},
						NewText: "var.second",
						Snippet: "var.second",
					},
					Kind: lang.TraversalCandidateKind,
				},
			}),
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)

			testDir := t.TempDir()
			dirReader := &testPathReader{
				paths: map[string]*PathContext{
					testDir: {
						Schema: tc.bodySchema,
						Files: map[string]*hcl.File{
							"test.tf": f,
						},
						ReferenceTargets: tc.builtinRefs,
					},
				},
			}
			decoder := NewDecoder(dirReader)
			d, err := decoder.Path(lang.Path{Path: testDir})
			if err != nil {
				t.Fatal(err)
			}
			refTargets, err := d.CollectReferenceTargets()
			if err != nil {
				t.Fatal(err)
			}

			dirReader.paths[testDir].ReferenceTargets = append(dirReader.paths[testDir].ReferenceTargets, refTargets...)

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
