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
)

func TestHoverAtPos_exprOneOf(t *testing.T) {
	testCases := []struct {
		testName          string
		attrSchema        map[string]*schema.AttributeSchema
		cfg               string
		pos               hcl.Pos
		expectedHoverData *lang.HoverData
	}{
		{
			"matching first expr",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.OneOf{
						schema.Keyword{
							Keyword: "keyword1",
						},
						schema.Keyword{
							Keyword: "keyword2",
						},
						schema.Keyword{
							Keyword: "keyword3",
						},
					},
				},
			},
			`attr = keyword1`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("`keyword1` _keyword_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
				},
			},
		},
		{
			"matching second expr",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.OneOf{
						schema.Keyword{
							Keyword: "keyword1",
						},
						schema.Keyword{
							Keyword: "keyword2",
						},
					},
				},
			},
			`attr = keyword2`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("`keyword2` _keyword_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
				},
			},
		},
		{
			"no matching expr",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.OneOf{
						schema.Keyword{
							Keyword: "keyword1",
						},
						schema.Keyword{
							Keyword: "keyword2",
						},
					},
				},
			},
			`attr = keyword3`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			nil,
		},
		{
			"no expr defined",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.OneOf{},
				},
			},
			`attr = keyword1`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
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
