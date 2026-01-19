// Copyright IBM Corp. 2020, 2025
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

func TestHoverAtPos_exprObject(t *testing.T) {
	testCases := []struct {
		testName          string
		attrSchema        map[string]*schema.AttributeSchema
		cfg               string
		pos               hcl.Pos
		expectedHoverData *lang.HoverData
	}{
		{
			"empty single-line object without attributes",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{},
				},
			},
			`attr = {}`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_object_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"empty multi-line object without attributes",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{},
				},
			},
			`attr = {
  
}`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("_object_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 3, Column: 2, Byte: 13},
				},
			},
		},
		{
			"empty single-line object with attributes",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								IsOptional: true,
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
							},
						},
					},
				},
			},
			`attr = {}`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("```\n{\n  foo = string # optional\n}\n```\n_object_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"empty single-line object with attributes and overrides",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Name:        "custom",
						Description: lang.Markdown("custom description"),
						Attributes: schema.ObjectAttributes{
							"foo": {
								IsOptional: true,
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
							},
						},
					},
				},
			},
			`attr = {}`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("```\n{\n  foo = string # optional\n}\n```\n_custom_\n\ncustom description"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"empty multi-line object with attributes",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								IsOptional: true,
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
							},
						},
					},
				},
			},
			`attr = {
  
}`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("```\n{\n  foo = string # optional\n}\n```\n_object_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 3, Column: 2, Byte: 13},
				},
			},
		},
		{
			"single item object on valid attribute name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								IsOptional:  true,
								Description: lang.Markdown("kw description"),
								Constraint: schema.Keyword{
									Keyword: "keyword",
								},
							},
						},
					},
				},
			},
			`attr = {
  foo = keyword
}`,
			hcl.Pos{Line: 2, Column: 5, Byte: 13},
			&lang.HoverData{
				Content: lang.Markdown("**foo** _optional, keyword_\n\nkw description"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
					End:      hcl.Pos{Line: 2, Column: 16, Byte: 24},
				},
			},
		},
		{
			"single item object on valid attribute name with custom data",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								IsOptional:   true,
								IsSensitive:  true,
								IsDeprecated: true,
								Description:  lang.Markdown("custom"),
								Constraint: schema.Keyword{
									Keyword:     "keyword",
									Description: lang.Markdown("kw description"),
								},
							},
						},
					},
				},
			},
			`attr = {
  foo = keyword
}`,
			hcl.Pos{Line: 2, Column: 5, Byte: 13},
			&lang.HoverData{
				Content: lang.Markdown("**foo** _optional, sensitive, keyword_\n\ncustom"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
					End:      hcl.Pos{Line: 2, Column: 16, Byte: 24},
				},
			},
		},
		{
			"single item object on invalid attribute name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								IsOptional: true,
								Constraint: schema.Keyword{
									Keyword: "keyword",
								},
							},
						},
					},
				},
			},
			`attr = {
  bar = keyword
}`,
			hcl.Pos{Line: 2, Column: 5, Byte: 13},
			&lang.HoverData{
				Content: lang.Markdown(`_object_`),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 3, Column: 2, Byte: 26},
				},
			},
		},
		{
			"multi item object on valid attribute name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								IsOptional: true,
								Constraint: schema.Keyword{
									Keyword: "keywordfoo",
								},
							},
							"bar": {
								IsRequired: true,
								Constraint: schema.Keyword{
									Keyword: "keywordbar",
								},
							},
							"baz": {
								IsOptional: true,
								Constraint: schema.Keyword{
									Keyword: "keywordbaz",
								},
							},
						},
					},
				},
			},
			`attr = {
  foo = keywordfoo
  bar = keywordbar
  baz = keywordbaz
}`,
			hcl.Pos{Line: 3, Column: 5, Byte: 32},
			&lang.HoverData{
				Content: lang.Markdown("**bar** _required, keyword_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 3, Column: 3, Byte: 30},
					End:      hcl.Pos{Line: 3, Column: 19, Byte: 46},
				},
			},
		},
		{
			"multi item object on matching value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								IsOptional: true,
								Constraint: schema.Keyword{
									Keyword: "keywordfoo",
								},
							},
							"bar": {
								IsOptional: true,
								Constraint: schema.Keyword{
									Keyword: "keywordbar",
								},
							},
							"baz": {
								IsOptional: true,
								Constraint: schema.Keyword{
									Keyword: "keywordbaz",
								},
							},
						},
					},
				},
			},
			`attr = {
  foo = invalid
  bar = keywordbar
  baz = keywordbaz
}`,
			hcl.Pos{Line: 3, Column: 16, Byte: 40},
			&lang.HoverData{
				Content: lang.Markdown("`keywordbar` _keyword_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 3, Column: 9, Byte: 33},
					End:      hcl.Pos{Line: 3, Column: 19, Byte: 43},
				},
			},
		},
		{
			"multi item object on mismatching value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								IsOptional: true,
								Constraint: schema.Keyword{
									Keyword: "keywordfoo",
								},
							},
							"bar": {
								IsOptional: true,
								Constraint: schema.Keyword{
									Keyword: "keywordbar",
								},
							},
							"baz": {
								IsOptional: true,
								Constraint: schema.Keyword{
									Keyword: "keywordbaz",
								},
							},
						},
					},
				},
			},
			`attr = {
  foo = invalid
  bar = keywordbar
  baz = keywordbaz
}`,
			hcl.Pos{Line: 2, Column: 13, Byte: 21},
			nil,
		},
		{
			"multi item object in empty space",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								IsOptional: true,
								Constraint: schema.LiteralType{
									Type: cty.Number,
								},
							},
							"bar": {
								IsOptional: true,
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
							},
							"baz": {
								IsOptional: true,
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
							},
						},
					},
				},
			},
			`attr = {
  bar = "bar"
  baz = "baz"
}`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("```" + `
{
  bar = string # optional
  baz = string # optional
  foo = number # optional
}
` + "```\n_object_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 4, Column: 2, Byte: 38},
				},
			},
		},
		{
			"multi item nested object",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Object{
						Attributes: schema.ObjectAttributes{
							"foo": {
								IsOptional: true,
								Constraint: schema.LiteralType{
									Type: cty.Number,
								},
							},
							"bar": {
								IsOptional: true,
								Constraint: schema.Object{
									Attributes: schema.ObjectAttributes{
										"noot": {
											IsRequired: true,
											Constraint: schema.LiteralType{Type: cty.Bool},
										},
										"animal": {
											IsOptional: true,
											Constraint: schema.LiteralType{Type: cty.String},
										},
									},
								},
							},
							"baz": {
								IsOptional: true,
								Constraint: schema.LiteralType{
									Type: cty.String,
								},
							},
						},
					},
				},
			},
			`attr = {
  bar = {}
  baz = "baz"
}`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("```" + `
{
  bar = {
    animal = string # optional
    noot = bool
  } # optional
  baz = string # optional
  foo = number # optional
}
` + "```\n_object_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 4, Column: 2, Byte: 35},
				},
			},
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
			hoverData, err := d.HoverAtPos(ctx, "test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedHoverData, hoverData); diff != "" {
				t.Fatalf("unexpected hover data: %s", diff)
			}
		})
	}
}
