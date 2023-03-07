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

func TestHoverAtPos_exprTuple(t *testing.T) {
	testCases := []struct {
		testName          string
		attrSchema        map[string]*schema.AttributeSchema
		cfg               string
		pos               hcl.Pos
		expectedHoverData *lang.HoverData
	}{
		{
			"empty single-line tuple without element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{},
				},
			},
			`attr = []`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_tuple_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"empty multi-line tuple without element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{},
				},
			},
			`attr = [
  
]`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("_tuple_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 3, Column: 2, Byte: 13},
				},
			},
		},
		{
			"empty single-line tuple with one element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
							schema.Keyword{
								Keyword: "drowyek",
							},
						},
					},
				},
			},
			`attr = []`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_tuple_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"empty single-line tuple with elements and description",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
							schema.Keyword{
								Keyword: "drowyek",
							},
						},
						Description: lang.Markdown("description"),
					},
				},
			},
			`attr = []`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_tuple_\n\ndescription"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"empty multi-line tuple with one element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
						},
					},
				},
			},
			`attr = [
  
]`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("_tuple_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 3, Column: 2, Byte: 13},
				},
			},
		},
		{
			"single element single-line tuple on element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword: "keyword",
							},
							schema.Keyword{
								Keyword: "drowyek",
							},
						},
						Description: lang.Markdown("description"),
					},
				},
			},
			`attr = [keyword]`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			&lang.HoverData{
				Content: lang.Markdown("`keyword` _keyword_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
					End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
				},
			},
		},
		{
			"single element single-line tuple on element with custom data",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword:     "keyword",
								Description: lang.Markdown("key description"),
							},
							schema.Keyword{
								Keyword:     "drowyek",
								Description: lang.Markdown("dro description"),
							},
						},
						Description: lang.Markdown("description"),
					},
				},
			},
			`attr = [keyword]`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			&lang.HoverData{
				Content: lang.Markdown("`keyword` _keyword_\n\nkey description"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
					End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
				},
			},
		},
		{
			"multi-element single-line tuple on tuple",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.LiteralType{
								Type: cty.String,
							},
							schema.LiteralType{
								Type: cty.String,
							},
						},
						Description: lang.Markdown("description"),
					},
				},
			},
			`attr = [ "one", "two" ]`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			&lang.HoverData{
				Content: lang.Markdown("_tuple_\n\ndescription"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 8,
						Byte:   7,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 24,
						Byte:   23,
					},
				},
			},
		},
		{
			"single element multi-line tuple on element with custom data",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword:     "keyword",
								Description: lang.Markdown("key description"),
							},
							schema.Keyword{
								Keyword:     "drowyek",
								Description: lang.Markdown("dro description"),
							},
						},
						Description: lang.Markdown("description"),
					},
				},
			},
			`attr = [
  keyword,
]`,
			hcl.Pos{Line: 2, Column: 6, Byte: 14},
			&lang.HoverData{
				Content: lang.Markdown("`keyword` _keyword_\n\nkey description"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
					End:      hcl.Pos{Line: 2, Column: 10, Byte: 18},
				},
			},
		},
		{
			"multi-element multi-line tuple on invalid element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword:     "keyword",
								Description: lang.Markdown("key description"),
							},
							schema.Keyword{
								Keyword:     "drowyek",
								Description: lang.Markdown("dro description"),
							},
						},
						Description: lang.Markdown("description"),
					},
				},
			},
			`attr = [
  "foo",
  keyword,
]`,
			hcl.Pos{Line: 2, Column: 6, Byte: 14},
			nil,
		},
		{
			"multi-element multi-line tuple on second element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Keyword{
								Keyword:     "keyword",
								Description: lang.Markdown("key description"),
							},
							schema.Keyword{
								Keyword:     "drowyek",
								Description: lang.Markdown("dro description"),
							},
						},
						Description: lang.Markdown("description"),
					},
				},
			},
			`attr = [
  keyword,
  drowyek,
]`,
			hcl.Pos{Line: 3, Column: 6, Byte: 25},
			&lang.HoverData{
				Content: lang.Markdown("`drowyek` _keyword_\n\ndro description"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 3, Column: 3, Byte: 22},
					End:      hcl.Pos{Line: 3, Column: 10, Byte: 29},
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
