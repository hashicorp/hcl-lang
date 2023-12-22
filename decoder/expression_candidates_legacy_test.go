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
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestLegacyDecoder_CandidateAtPos_expressions(t *testing.T) {
	ctx := context.Background()
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
					Constraint: schema.LiteralType{Type: cty.String},
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
					Constraint: schema.LiteralValue{
						Value: cty.ObjectVal(map[string]cty.Value{
							"first":  cty.StringVal("blah"),
							"second": cty.NumberIntVal(2345),
						}),
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
  first = "blah"
  second = 2345
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
					Constraint: schema.LiteralType{
						Type: cty.Object(map[string]cty.Type{
							"first":  cty.String,
							"second": cty.Number,
						}),
					},
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
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"first": &schema.AttributeSchema{
								Constraint: schema.LiteralType{Type: cty.String},
							},
							"second": &schema.AttributeSchema{
								Constraint: schema.LiteralType{Type: cty.Number},
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
					Label:  "{…}",
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
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"first": &schema.AttributeSchema{
								Constraint: schema.LiteralType{Type: cty.String},
							},
							"second": &schema.AttributeSchema{
								Constraint: schema.LiteralType{Type: cty.Number},
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
						NewText: `first`,
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
						NewText: "second",
						Snippet: "second = ${1:0}",
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"object as expression - attributes partially declared",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"first": &schema.AttributeSchema{
								Constraint: schema.LiteralType{Type: cty.String},
							},
							"second": &schema.AttributeSchema{
								Constraint: schema.LiteralType{Type: cty.Number},
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
						NewText: "second",
						Snippet: "second = ${1:0}",
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"object as expression - attribute key unknown",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"first": &schema.AttributeSchema{
								Constraint: schema.LiteralType{Type: cty.String},
							},
							"second": &schema.AttributeSchema{
								Constraint: schema.LiteralType{Type: cty.Number},
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
						NewText: "second",
						Snippet: "second = ${1:0}",
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"list as value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ListVal([]cty.Value{
							cty.StringVal("foo"),
							cty.StringVal("bar"),
						}),
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `[ "foo", "bar" ]`,
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
						NewText: `["foo", "bar"]`,
						Snippet: `["foo", "bar"]`,
					},
					Kind: lang.ListCandidateKind,
				},
			}),
		},
		{
			"map as type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{Type: cty.Map(cty.String)},
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
					Constraint: schema.LiteralValue{
						Value: cty.MapVal(map[string]cty.Value{
							"foo": cty.StringVal("moo"),
							"bar": cty.StringVal("boo"),
						}),
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `{ "bar" = "boo", … }`,
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
  "bar" = "boo"
  "foo" = "moo"
}`,
						Snippet: `{
  "bar" = "boo"
  "foo" = "moo"
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
					Constraint: schema.LiteralType{Type: cty.Bool},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
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
						Snippet: "false",
					},
					Kind: lang.BoolCandidateKind,
				},
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
						Snippet: "true",
					},
					Kind: lang.BoolCandidateKind,
				},
			}),
		},
		{
			"string values",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.OneOf{
						schema.LiteralValue{Value: cty.StringVal("first")},
						schema.LiteralValue{Value: cty.StringVal("second")},
						schema.LiteralValue{Value: cty.StringVal("third")},
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
						Snippet: `"first"`},
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
						Snippet: `"second"`},
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
						Snippet: `"third"`},
					Kind: lang.StringCandidateKind,
				},
			}),
		},
		{
			"deprecated string value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.OneOf{
						schema.LiteralValue{Value: cty.StringVal("first")},
						schema.LiteralValue{Value: cty.StringVal("second"), IsDeprecated: true},
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
						Snippet: `"first"`},
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
						Snippet: `"second"`},
					Kind: lang.StringCandidateKind,
				},
			}),
		},
		{
			"tuple as list type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{Type: cty.List(cty.String)},
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
					Constraint: schema.LiteralType{Type: cty.List(cty.String)},
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
					Constraint: schema.List{
						Elem: schema.LiteralType{Type: cty.String},
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
						Snippet: `attr = [ "${1:value}" ]`,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"list expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.List{
						Elem: schema.LiteralType{Type: cty.String},
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
						NewText: `[ "value" ]`,
						Snippet: `[ "${1:value}" ]`,
					},
					Kind: lang.ListCandidateKind,
				},
			}),
		},
		{
			"list expression inside",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.List{
						Elem: schema.LiteralType{Type: cty.Bool},
					},
				},
			},
			`attr = [ ]
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.CompleteCandidates([]lang.Candidate{
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
						Snippet: "false",
					},
					Kind: lang.BoolCandidateKind,
				},
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
						Snippet: "true",
					},
					Kind: lang.BoolCandidateKind,
				},
			}),
		},
		{
			"attribute as set expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.LiteralType{Type: cty.String},
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
						Snippet: `attr = [ "${1:value}" ]`,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"set expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.LiteralType{Type: cty.String},
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
						NewText: `[ "value" ]`,
						Snippet: `[ "${1:value}" ]`,
					},
					Kind: lang.SetCandidateKind,
				},
			}),
		},
		{
			"set expression inside",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.LiteralType{Type: cty.Bool},
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
						Snippet: "false",
					},
					Kind: lang.BoolCandidateKind,
				},
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
						Snippet: "true",
					},
					Kind: lang.BoolCandidateKind,
				},
			}),
		},
		{
			"attribute as tuple expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.LiteralType{Type: cty.String},
							schema.LiteralType{Type: cty.Number},
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
						Snippet: `attr = [ "${1:value}", ${2:0} ]`,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"tuple expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.LiteralType{Type: cty.String},
							schema.LiteralType{Type: cty.Number},
						},
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "[ string, number ]",
					Detail: "tuple",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						NewText: `[ "value", 0 ]`,
						Snippet: `[ "${1:value}", ${2:0} ]`,
					},
					Kind: lang.TupleCandidateKind,
				},
			}),
		},
		{
			"tuple expression inside",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.LiteralType{Type: cty.Bool},
							schema.LiteralType{Type: cty.Number},
						},
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
						Snippet: "false",
					},
					Kind: lang.BoolCandidateKind,
				},
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
						Snippet: "true",
					},
					Kind: lang.BoolCandidateKind,
				},
			}),
		},
		{
			"keyword",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Keyword{
						Keyword: "foobar",
						Name:    "special kw",
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
					Constraint: schema.Map{
						Elem: schema.LiteralType{Type: cty.String},
						Name: "map of something",
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `{ "key" = string }`,
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
						NewText: "{\n  \"name\" = \"value\"\n}",
						Snippet: "{\n  \"${1:name}\" = \"${2:value}\"\n}",
					},
					Kind: lang.MapCandidateKind,
				},
			}),
		},
		{
			"literal of dynamic pseudo type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{Type: cty.DynamicPseudoType},
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
					Constraint: schema.TypeDeclaration{},
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
						},
						NewText: "bool",
						Snippet: "bool",
					},
					Kind: lang.BoolCandidateKind,
				},
				{
					Label:  "number",
					Detail: "number",
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
						NewText: "number",
						Snippet: "number",
					},
					Kind: lang.NumberCandidateKind,
				},
				{
					Label:  "string",
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
						NewText: "string",
						Snippet: "string",
					},
					Kind: lang.StringCandidateKind,
				},
				{
					Label:  "list(…)",
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
						NewText: "list()",
						Snippet: "list(${0})",
					},
					TriggerSuggest: true,
					Kind:           lang.ListCandidateKind,
				},
				{
					Label:  "set(…)",
					Detail: "set",
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
						NewText: "set()",
						Snippet: "set(${0})",
					},
					TriggerSuggest: true,
					Kind:           lang.SetCandidateKind,
				},
				{
					Label:  "tuple([…])",
					Detail: "tuple",
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
						NewText: "tuple([])",
						Snippet: "tuple([ ${0} ])",
					},
					TriggerSuggest: true,
					Kind:           lang.TupleCandidateKind,
				},
				{
					Label:  "map(…)",
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
						NewText: "map()",
						Snippet: "map(${0})",
					},
					TriggerSuggest: true,
					Kind:           lang.MapCandidateKind,
				},
				{
					Label:  "object({…})",
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
						NewText: "object({\n\n})",
						Snippet: "object({\n  ${1:name} = ${2}\n})",
					},
					Kind: lang.ObjectCandidateKind,
				},
			}),
		},
		{
			"empty list",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.List{},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "[ ]",
					Detail: "list",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						NewText: "[ ]",
						Snippet: "[ ${1} ]",
					},
					Kind: lang.ListCandidateKind,
				},
			}),
		},
		{
			"multiple traversal constraints in set",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.OneOf{
							schema.Reference{OfScopeId: lang.ScopeId("one")},
							schema.Reference{OfScopeId: lang.ScopeId("two")},
						},
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "[ reference ]",
					Detail: "set of reference",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						NewText: "[ ]",
						Snippet: "[ ${1} ]",
					},
					Kind:           lang.SetCandidateKind,
					TriggerSuggest: true,
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

func TestLegacyDecoder_CandidateAtPos_traversalExpressions(t *testing.T) {
	ctx := context.Background()
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
						Constraint: schema.Reference{OfType: cty.String},
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
						Constraint: schema.Reference{OfType: cty.List(cty.String)},
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
						Constraint: schema.Reference{OfType: cty.String},
					},
				},
			},
			reference.Targets{
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Type: cty.List(cty.Number),
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
						Constraint: schema.Reference{OfType: cty.DynamicPseudoType},
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
						Constraint: schema.Reference{OfType: cty.String},
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
						Constraint: schema.Reference{OfScopeId: lang.ScopeId("test")},
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
						Constraint: schema.Reference{
							Address: &schema.ReferenceAddrSchema{
								ScopeId: lang.ScopeId("blah"),
							},
							Name: "test",
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
						Constraint: schema.Set{
							Elem: schema.Reference{
								Address: &schema.ReferenceAddrSchema{
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
									Constraint: schema.LiteralType{Type: cty.String},
									IsOptional: true,
								},
								"blah": {
									Constraint: schema.LiteralType{Type: cty.Bool},
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
									Constraint: schema.Reference{OfType: cty.String},
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
						Constraint: schema.Reference{OfType: cty.String},
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
						Constraint: schema.Reference{OfType: cty.String},
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
						Constraint: schema.Reference{OfType: cty.String},
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
						Constraint: schema.Reference{OfType: cty.String},
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
						Constraint: schema.Reference{OfType: cty.String},
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
						Constraint: schema.Reference{OfType: cty.String},
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

func TestLegacyDecoder_CandidateAtPos_expressions_crossFileTraversal(t *testing.T) {
	ctx := context.Background()
	f1, _ := hclsyntax.ParseConfig([]byte(`variable "aaa" {}
variable "bbb" {}
variable "ccc" {}
`), "test1.tf", hcl.InitialPos)
	f2, _ := hclsyntax.ParseConfig([]byte(`value = 
`), "test2.tf", hcl.InitialPos)

	bodySchema := &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"value": {
				IsOptional: true,
				Constraint: schema.Reference{
					OfScopeId: lang.ScopeId("variable"),
					OfType:    cty.DynamicPseudoType,
				},
			},
		},
		Blocks: map[string]*schema.BlockSchema{
			"variable": {
				Labels: []*schema.LabelSchema{
					{Name: "name"},
				},
				Address: &schema.BlockAddrSchema{
					Steps: []schema.AddrStep{
						schema.StaticStep{Name: "var"},
						schema.LabelStep{Index: 0},
					},
					FriendlyName: "variable",
					ScopeId:      lang.ScopeId("variable"),
					AsTypeOf: &schema.BlockAsTypeOf{
						AttributeExpr: "type",
					},
				},
			},
		},
	}

	testDir := t.TempDir()
	dirReader := &testPathReader{
		paths: map[string]*PathContext{
			testDir: {
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test1.tf": f1,
					"test2.tf": f2,
				},
				ReferenceTargets: reference.Targets{},
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

	expectedTargets := reference.Targets{
		{
			Addr:    lang.Address{lang.RootStep{Name: "var"}, lang.AttrStep{Name: "aaa"}},
			ScopeId: lang.ScopeId("variable"),
			RangePtr: &hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:      hcl.Pos{Line: 1, Column: 18, Byte: 17},
			},
			DefRangePtr: &hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
			},
			Type: cty.DynamicPseudoType,
		},
		{
			Addr:    lang.Address{lang.RootStep{Name: "var"}, lang.AttrStep{Name: "bbb"}},
			ScopeId: lang.ScopeId("variable"),
			RangePtr: &hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 2, Column: 1, Byte: 18},
				End:      hcl.Pos{Line: 2, Column: 18, Byte: 35},
			},
			DefRangePtr: &hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 2, Column: 1, Byte: 18},
				End:      hcl.Pos{Line: 2, Column: 15, Byte: 32},
			},
			Type: cty.DynamicPseudoType,
		},
		{
			Addr:    lang.Address{lang.RootStep{Name: "var"}, lang.AttrStep{Name: "ccc"}},
			ScopeId: lang.ScopeId("variable"),
			RangePtr: &hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 3, Column: 1, Byte: 36},
				End:      hcl.Pos{Line: 3, Column: 18, Byte: 53},
			},
			DefRangePtr: &hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 3, Column: 1, Byte: 36},
				End:      hcl.Pos{Line: 3, Column: 15, Byte: 50},
			},
			Type: cty.DynamicPseudoType,
		},
	}
	if diff := cmp.Diff(expectedTargets, refTargets, ctydebug.CmpOptions); diff != "" {
		t.Fatalf("unexpected targets: %s", diff)
	}

	dirReader.paths[testDir].ReferenceTargets = refTargets

	candidates, err := d.CompletionAtPos(ctx, "test2.tf", hcl.Pos{
		Line:   1,
		Column: 9,
		Byte:   8,
	})
	if err != nil {
		t.Fatal(err)
	}

	expectedCandidates := lang.Candidates{
		List: []lang.Candidate{
			{
				Label:  "var.aaa",
				Detail: "dynamic",
				TextEdit: lang.TextEdit{
					Range: hcl.Range{
						Filename: "test2.tf",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
					},
					NewText: "var.aaa",
					Snippet: "var.aaa",
				},
				Kind: lang.TraversalCandidateKind,
			},
			{
				Label:  "var.bbb",
				Detail: "dynamic",
				TextEdit: lang.TextEdit{
					Range: hcl.Range{
						Filename: "test2.tf",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
					},
					NewText: "var.bbb",
					Snippet: "var.bbb",
				},
				Kind: lang.TraversalCandidateKind,
			},
			{
				Label:  "var.ccc",
				Detail: "dynamic",
				TextEdit: lang.TextEdit{
					Range: hcl.Range{
						Filename: "test2.tf",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
					},
					NewText: "var.ccc",
					Snippet: "var.ccc",
				},
				Kind: lang.TraversalCandidateKind,
			},
		},
		IsComplete: true,
	}
	if diff := cmp.Diff(expectedCandidates, candidates); diff != "" {
		t.Fatalf("unexpected candidates: %s", diff)
	}
}

func TestLegacyDecoder_CandidateAtPos_expressions_Hooks(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		pos                hcl.Pos
		completionHooks    CompletionFuncMap
		expectedCandidates lang.Candidates
	}{
		{
			"simple hook",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{Type: cty.String},
					CompletionHooks: lang.CompletionHooks{
						{
							Name: "TestCompletionHook",
						},
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			CompletionFuncMap{
				"TestCompletionHook": func(ctx context.Context, value cty.Value) ([]Candidate, error) {
					candidates := []Candidate{
						{
							Label:         "\"label\"",
							Detail:        "detail",
							Kind:          lang.StringCandidateKind,
							Description:   lang.PlainText("description"),
							RawInsertText: "\"insertText\"",
						},
					}
					return candidates, nil
				},
			},
			lang.IncompleteCandidates([]lang.Candidate{
				{
					Label:       "\"label\"",
					Detail:      "detail",
					Kind:        lang.StringCandidateKind,
					Description: lang.PlainText("description"),
					TextEdit: lang.TextEdit{
						NewText: "\"insertText\"",
						Snippet: "\"insertText\"",
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
			"hook with prefix",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{Type: cty.String},
					CompletionHooks: lang.CompletionHooks{
						{
							Name: "TestCompletionHook",
						},
					},
				},
			},
			`attr = "pa"
`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			CompletionFuncMap{
				"TestCompletionHook": func(ctx context.Context, value cty.Value) ([]Candidate, error) {
					candidates := []Candidate{
						{
							Label:         value.AsString(),
							Kind:          lang.StringCandidateKind,
							RawInsertText: value.AsString(),
						},
					}
					return candidates, nil
				},
			},
			lang.IncompleteCandidates([]lang.Candidate{
				{
					Label: "pa",
					Kind:  lang.StringCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "pa",
						Snippet: "pa",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
						},
					},
				},
			}),
		},
		{
			"incomplete attr value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{Type: cty.String},
					CompletionHooks: lang.CompletionHooks{
						{
							Name: "TestCompletionHook",
						},
					},
				},
			},
			`attr = "pa

`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			CompletionFuncMap{
				"TestCompletionHook": func(ctx context.Context, value cty.Value) ([]Candidate, error) {
					candidates := []Candidate{
						{
							Label:         value.AsString(),
							Kind:          lang.StringCandidateKind,
							RawInsertText: value.AsString(),
						},
					}
					return candidates, nil
				},
			},
			lang.IncompleteCandidates([]lang.Candidate{
				{
					Label: "pa",
					Kind:  lang.StringCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "pa",
						Snippet: "pa",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
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

			// We're ignoring diagnostics here, since some test cases may contain invalid HCL
			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			})
			for n, h := range tc.completionHooks {
				d.decoderCtx.CompletionHooks[n] = h
			}

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

func TestLegacyDecoder_CandidateAtPos_maxCandidates(t *testing.T) {
	ctx := context.Background()
	bodySchema := &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"attr": {
				Constraint: schema.LiteralType{Type: cty.String},
				CompletionHooks: lang.CompletionHooks{
					{
						Name: "TestCompletionHook50",
					},
					{
						Name: "TestCompletionHook70",
					},
				},
			},
		},
	}

	// We're ignoring diagnostics here, since our config contains invalid HCL
	f, _ := hclsyntax.ParseConfig([]byte(`attr = `), "test.tf", hcl.InitialPos)
	d := testPathDecoder(t, &PathContext{
		Schema: bodySchema,
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})
	d.decoderCtx.CompletionHooks["TestCompletionHook50"] = func(ctx context.Context, value cty.Value) ([]Candidate, error) {
		candidates := make([]Candidate, 0)
		for i := 0; i < 50; i++ {
			candidates = append(candidates, Candidate{
				Label: fmt.Sprintf("\"Label %d\"", i),
				Kind:  lang.StringCandidateKind,
			})
		}
		return candidates, nil
	}
	d.decoderCtx.CompletionHooks["TestCompletionHook70"] = func(ctx context.Context, value cty.Value) ([]Candidate, error) {
		candidates := make([]Candidate, 0)
		for i := 0; i < 70; i++ {
			candidates = append(candidates, Candidate{
				Label: fmt.Sprintf("\"Label %d\"", i),
				Kind:  lang.StringCandidateKind,
			})
		}
		return candidates, nil
	}

	candidates, err := d.CompletionAtPos(ctx, "test.tf", hcl.Pos{Line: 1, Column: 8, Byte: 7})
	if err != nil {
		t.Fatal(err)
	}
	count := len(candidates.List)

	if uint(count) != d.maxCandidates {
		t.Fatalf("unexpected candidates count: %d", count)
	}
}
