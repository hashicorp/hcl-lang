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
)

func TestHoverAtPos_exprMap(t *testing.T) {
	testCases := []struct {
		testName          string
		attrSchema        map[string]*schema.AttributeSchema
		cfg               string
		pos               hcl.Pos
		expectedHoverData *lang.HoverData
	}{
		{
			"empty single-line map without element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{},
				},
			},
			`attr = {}`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_map_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"empty multi-line map without element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{},
				},
			},
			`attr = {
  
}`,
			hcl.Pos{Line: 2, Column: 2, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("_map_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 3, Column: 2, Byte: 13},
				},
			},
		},
		{
			"empty single-line map with element",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = {}`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_map of keyword_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"empty single-line map with element and extra data",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Name:        "custom",
						Description: lang.Markdown("custom description"),
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
						MinItems: 1,
						MaxItems: 3,
					},
				},
			},
			`attr = {}`,
			hcl.Pos{Line: 1, Column: 9, Byte: 8},
			&lang.HoverData{
				Content: lang.Markdown("_custom_\n\ncustom description"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
				},
			},
		},
		{
			"single item map on key name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = {
  foo = keyword
}`,
			hcl.Pos{Line: 2, Column: 5, Byte: 13},
			nil,
		},
		{
			"single item map on invalid key type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = {
  422 = keyword
}`,
			hcl.Pos{Line: 2, Column: 5, Byte: 13},
			nil,
		},
		{
			"multi item map on valid key type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keyword",
						},
					},
				},
			},
			`attr = {
  422 = keywordfoo
  bar = keywordbar
  432 = keywordbaz
}`,
			hcl.Pos{Line: 3, Column: 5, Byte: 32},
			nil,
		},
		{
			"multi item map on matching value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keywordbar",
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
			"multi item map on mismatching value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "keywordbar",
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
