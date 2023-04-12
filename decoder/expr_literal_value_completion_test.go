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

func TestCompletionAtPos_exprLiteralValue(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		// extra metadata
		{
			"bool",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value:        cty.StringVal("foo"),
						IsDeprecated: true,
						Description:  lang.Markdown("foobar"),
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:        "foo",
					Detail:       "string",
					Kind:         lang.StringCandidateKind,
					IsDeprecated: true,
					Description:  lang.Markdown("foobar"),
					TextEdit: lang.TextEdit{
						NewText: `"foo"`,
						Snippet: `"foo"`,
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
			}),
		},
		// primitive types
		{
			"bool",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.BoolVal(true),
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
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
			"bool partial",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.BoolVal(true),
					},
				},
			},
			`attr = tr
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.CompleteCandidates([]lang.Candidate{
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
							End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
						},
					},
				},
			}),
		},
		{
			"bool partial middle",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.BoolVal(true),
					},
				},
			},
			`attr = true
`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			lang.CompleteCandidates([]lang.Candidate{
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
							End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
						},
					},
				},
			}),
		},
		{
			"string",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.StringVal("foo"),
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "foo",
					Detail: "string",
					Kind:   lang.StringCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: `"foo"`,
						Snippet: `"foo"`,
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
			"string partial before closing quote",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.StringVal("foobar"),
					},
				},
			},
			`attr = "foo"
`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "foobar",
					Detail: "string",
					Kind:   lang.StringCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: `"foobar"`,
						Snippet: `"foobar"`,
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
			"string partial without closing quote",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.StringVal("foobar"),
					},
				},
			},
			`attr = "foo
`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "foobar",
					Detail: "string",
					Kind:   lang.StringCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: `"foobar"`,
						Snippet: `"foobar"`,
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
			"string partial after closing quote",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.StringVal("foobar"),
					},
				},
			},
			`attr = "foo"
`,
			hcl.Pos{Line: 1, Column: 13, Byte: 12},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "foobar",
					Detail: "string",
					Kind:   lang.StringCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: `"foobar"`,
						Snippet: `"foobar"`,
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
			"whole number",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.NumberIntVal(1),
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "1",
					Detail: "number",
					Kind:   lang.NumberCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: `1`,
						Snippet: `1`,
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
			"whole number partial",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.NumberIntVal(1189998819991197253),
					},
				},
			},
			`attr = 118999
`,
			hcl.Pos{Line: 1, Column: 14, Byte: 13},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "1189998819991197253",
					Detail: "number",
					Kind:   lang.NumberCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "1189998819991197253",
						Snippet: "1189998819991197253",
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
			"whole number partial middle",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.NumberIntVal(1189998819991197253),
					},
				},
			},
			`attr = 1189998819991197253
`,
			hcl.Pos{Line: 1, Column: 14, Byte: 13},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "1189998819991197253",
					Detail: "number",
					Kind:   lang.NumberCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: "1189998819991197253",
						Snippet: "1189998819991197253",
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 27, Byte: 26},
						},
					},
				},
			}),
		},
		{
			"fractional number",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.NumberFloatVal(42.223),
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "42.223",
					Detail: "number",
					Kind:   lang.NumberCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: `42.223`,
						Snippet: `42.223`,
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
			"fractional number partial",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.NumberFloatVal(42.223),
					},
				},
			},
			`attr = 42.
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "42.223",
					Detail: "number",
					Kind:   lang.NumberCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: `42.223`,
						Snippet: `42.223`,
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
			"fractional number partial middle",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.NumberFloatVal(42.223),
					},
				},
			},
			`attr = 42.223
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "42.223",
					Detail: "number",
					Kind:   lang.NumberCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: `42.223`,
						Snippet: `42.223`,
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
						},
					},
				},
			}),
		},

		// complex types
		{
			"map",
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
					Kind:   lang.MapCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: `{
  "bar" = "boo"
  "foo" = "moo"
}`,
						Snippet: `{
  "bar" = "boo"
  "foo" = "moo"
}`,
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
			"object",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ObjectVal(map[string]cty.Value{
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
					Label:  `{ bar = "boo", … }`,
					Detail: "object",
					Kind:   lang.ObjectCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: `{
  bar = "boo"
  foo = "moo"
}`,
						Snippet: `{
  bar = "boo"
  foo = "moo"
}`,
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
			"list",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.ListVal([]cty.Value{
							cty.BoolVal(true),
						}),
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `[ true ]`,
					Detail: "list of bool",
					Kind:   lang.ListCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: `[true]`,
						Snippet: `[true]`,
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
			"set",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.SetVal([]cty.Value{
							cty.BoolVal(false),
						}),
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `[ false ]`,
					Detail: "set of bool",
					Kind:   lang.SetCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: `[false]`,
						Snippet: `[false]`,
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
			"tuple",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.LiteralValue{
						Value: cty.TupleVal([]cty.Value{
							cty.BoolVal(true),
						}),
					},
				},
			},
			`attr = 
`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `[ true ]`,
					Detail: "tuple",
					Kind:   lang.TupleCandidateKind,
					TextEdit: lang.TextEdit{
						NewText: `[true]`,
						Snippet: `[true]`,
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
					},
				},
			}),
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%02d-%s", i, tc.testName), func(t *testing.T) {
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
