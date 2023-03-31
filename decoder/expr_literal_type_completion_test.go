// Copyright (c) HashiCorp, Inc.
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
	"github.com/zclconf/go-cty/cty"
)

func TestCompletionAtPos_exprLiteralType(t *testing.T) {
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
					Constraint: schema.LiteralType{
						Type: cty.Bool,
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
					Constraint: schema.LiteralType{
						Type: cty.Bool,
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
					Constraint: schema.LiteralType{
						Type: cty.String,
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
					Constraint: schema.LiteralType{
						Type: cty.List(cty.String),
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
					Constraint: schema.LiteralType{
						Type: cty.List(cty.Bool),
					},
				},
			},
			`attr = [  ]
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.CompleteCandidates(boolLiteralCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
				End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
			})),
		},
		{
			"inside list of bool multiline",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.Bool),
					},
				},
			},
			`attr = [
  
]
`,
			hcl.Pos{Line: 2, Column: 3, Byte: 11},
			lang.CompleteCandidates(boolLiteralCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
				End:      hcl.Pos{Line: 2, Column: 3, Byte: 11},
			})),
		},
		{
			"inside list next element after space",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.Bool),
					},
				},
			},
			`attr = [ false,  ]
`,
			hcl.Pos{Line: 1, Column: 17, Byte: 16},
			lang.CompleteCandidates(boolLiteralCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 17, Byte: 16},
				End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
			})),
		},
		{
			"inside list next element after newline",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.Bool),
					},
				},
			},
			`attr = [
  false,
  
]
`,
			hcl.Pos{Line: 3, Column: 3, Byte: 20},
			lang.CompleteCandidates(boolLiteralCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 3, Column: 3, Byte: 20},
				End:      hcl.Pos{Line: 3, Column: 3, Byte: 20},
			})),
		},
		{
			"inside list next element after comma",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.Bool),
					},
				},
			},
			`attr = [ false, ]
`,
			hcl.Pos{Line: 1, Column: 16, Byte: 15},
			lang.CompleteCandidates(boolLiteralCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 16, Byte: 15},
				End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
			})),
		},
		{
			"inside list next element near closing bracket",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.Bool),
					},
				},
			},
			`attr = [ false, ]
`,
			hcl.Pos{Line: 1, Column: 17, Byte: 16},
			lang.CompleteCandidates(boolLiteralCandidates("", hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 17, Byte: 16},
				End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
			})),
		},
		{
			"completion inside list with prefix",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.List(cty.Bool),
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
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{cty.Bool}),
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
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{cty.Bool}),
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
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{cty.String, cty.Bool}),
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
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{cty.String, cty.Bool}),
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
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{cty.String, cty.String, cty.Bool}),
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
					Constraint: schema.LiteralType{
						Type: cty.Tuple([]cty.Type{cty.String}),
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
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.Bool),
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
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.Bool),
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
						NewText: "\"key\" = false",
						Snippet: "\"${1:key}\" = ${2:false}",
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
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.Bool),
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
						NewText: "\"key\" = false",
						Snippet: "\"${1:key}\" = ${2:false}",
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
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.Bool),
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
						NewText: "\"key\" = false",
						Snippet: "\"${1:key}\" = ${2:false}",
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
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.Bool),
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
						NewText: `"key" = false`,
						Snippet: `"${1:key}" = ${2:false}`,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"inside map value empty",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.Bool),
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
					Constraint: schema.LiteralType{
						Type: cty.Map(cty.Bool),
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
					Constraint: schema.LiteralType{
						Type: cty.Object(map[string]cty.Type{
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
					Constraint: schema.LiteralType{
						Type: cty.Object(map[string]cty.Type{
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
					Constraint: schema.LiteralType{
						Type: cty.Object(map[string]cty.Type{
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
					Constraint: schema.LiteralType{
						Type: cty.Object(map[string]cty.Type{
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
					Constraint: schema.LiteralType{
						Type: cty.Object(map[string]cty.Type{
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
					Constraint: schema.LiteralType{
						Type: cty.Object(map[string]cty.Type{
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
					Constraint: schema.LiteralType{
						Type: cty.Object(map[string]cty.Type{
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
					Constraint: schema.LiteralType{
						Type: cty.Object(map[string]cty.Type{
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
					Constraint: schema.LiteralType{
						Type: cty.Object(map[string]cty.Type{
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
					Constraint: schema.LiteralType{
						Type: cty.Object(map[string]cty.Type{
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
					Constraint: schema.LiteralType{
						Type: cty.Object(map[string]cty.Type{
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
					Constraint: schema.LiteralType{
						Type: cty.Object(map[string]cty.Type{
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
					Constraint: schema.LiteralType{
						Type: cty.Object(map[string]cty.Type{
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
