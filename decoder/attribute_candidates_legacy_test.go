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

func TestLegacyDecoder_CompletionAtPos_EmptyCompletionData(t *testing.T) {
	testCases := []struct {
		testName           string
		attrName           string
		attrSchema         *schema.AttributeSchema
		expectedCandidates lang.Candidates
	}{
		{
			"primitive type",
			"primitive",
			&schema.AttributeSchema{
				Constraint: schema.LiteralType{Type: cty.String},
			},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "primitive",
					Detail: "string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.InitialPos,
							End:      hcl.InitialPos,
						},
						NewText: "primitive",
						Snippet: `primitive = "${1:value}"`,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"map of strings",
			"mymap",
			&schema.AttributeSchema{
				Constraint: schema.LiteralType{Type: cty.Map(cty.String)},
			},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "mymap",
					Detail: "map of string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.InitialPos,
							End:      hcl.InitialPos,
						},
						NewText: "mymap",
						Snippet: `mymap = {
  "${1:name}" = "${2:value}"
}`,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"map of numbers",
			"mymap",
			&schema.AttributeSchema{
				Constraint: schema.LiteralType{Type: cty.Map(cty.Number)},
			},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "mymap",
					Detail: "map of number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.InitialPos,
							End:      hcl.InitialPos,
						},
						NewText: "mymap",
						Snippet: `mymap = {
  "${1:name}" = ${2:0}
}`,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"list of numbers",
			"mylist",
			&schema.AttributeSchema{
				Constraint: schema.LiteralType{Type: cty.List(cty.Number)},
			},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "mylist",
					Detail: "list of number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.InitialPos,
							End:      hcl.InitialPos,
						},
						NewText: "mylist",
						Snippet: `mylist = [ ${1:0} ]`,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"list of objects",
			"mylistobj",
			&schema.AttributeSchema{
				Constraint: schema.LiteralType{
					Type: cty.List(cty.Object(map[string]cty.Type{
						"first":  cty.String,
						"second": cty.Number,
					})),
				},
			},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "mylistobj",
					Detail: "list of object",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.InitialPos,
							End:      hcl.InitialPos,
						},
						NewText: "mylistobj",
						Snippet: `mylistobj = [ {
  first = "${1:value}"
  second = ${2:0}
} ]`,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"set of numbers",
			"myset",
			&schema.AttributeSchema{
				Constraint: schema.LiteralType{Type: cty.Set(cty.Number)},
			},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "myset",
					Detail: "set of number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.InitialPos,
							End:      hcl.InitialPos,
						},
						NewText: "myset",
						Snippet: `myset = [ ${1:0} ]`,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"object",
			"myobj",
			&schema.AttributeSchema{
				Constraint: schema.LiteralType{
					Type: cty.Object(map[string]cty.Type{
						"keystr":  cty.String,
						"keynum":  cty.Number,
						"keybool": cty.Bool,
					}),
				},
			},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "myobj",
					Detail: "object",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.InitialPos,
							End:      hcl.InitialPos,
						},
						NewText: "myobj",
						Snippet: `myobj = {
  keybool = ${1:false}
  keynum = ${2:0}
  keystr = "${3:value}"
}`,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"unknown type",
			"mynil",
			&schema.AttributeSchema{
				Constraint: schema.LiteralType{Type: cty.DynamicPseudoType},
			},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "mynil",
					Detail: "any type",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.InitialPos,
							End:      hcl.InitialPos,
						},
						NewText: "mynil",
						Snippet: `mynil = `,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"nested object",
			"myobj",
			&schema.AttributeSchema{
				Constraint: schema.LiteralType{
					Type: cty.Object(map[string]cty.Type{
						"keystr": cty.String,
						"another": cty.Object(map[string]cty.Type{
							"nestedstr":     cty.String,
							"nested_number": cty.Number,
						}),
					}),
				},
			},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "myobj",
					Detail: "object",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.InitialPos,
							End:      hcl.InitialPos,
						},
						NewText: "myobj",
						Snippet: `myobj = {
  another = {
    nested_number = ${1:0}
    nestedstr = "${2:value}"
  }
  keystr = "${3:value}"
}`,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			f, _ := hclsyntax.ParseConfig([]byte("\n"), "test.tf", hcl.InitialPos)
			d := testPathDecoder(t, &PathContext{
				Schema: &schema.BodySchema{
					Attributes: map[string]*schema.AttributeSchema{
						tc.attrName: tc.attrSchema,
					},
				},
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			})
			d.PrefillRequiredFields = true

			ctx := context.Background()
			candidates, err := d.CompletionAtPos(ctx, "test.tf", hcl.InitialPos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedCandidates, candidates); diff != "" {
				t.Fatalf("unexpected candidates: %s", diff)
			}
		})
	}
}
