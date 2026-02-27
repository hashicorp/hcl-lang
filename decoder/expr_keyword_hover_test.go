// Copyright IBM Corp. 2020, 2026
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

func TestHoverAtPos_exprKeyword(t *testing.T) {
	testCases := []struct {
		testName          string
		attrSchema        map[string]*schema.AttributeSchema
		cfg               string
		pos               hcl.Pos
		expectedHoverData *lang.HoverData
	}{
		{
			"mismatching expression type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Keyword{
						Keyword: "foobar",
					},
				},
			},
			`attr = "foobar"`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			nil,
		},
		{
			"mismatching keyword",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Keyword{
						Keyword: "foobar",
					},
				},
			},
			`attr = barfoo`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			nil,
		},
		{
			"matching keyword",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Keyword{
						Keyword: "foobar",
					},
				},
			},
			`attr = foobar`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			&lang.HoverData{
				Content: lang.Markdown("`foobar` _keyword_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
				},
			},
		},
		{
			"matching keyword with all metadata",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Keyword{
						Keyword:     "foobar",
						Name:        "custom name",
						Description: lang.Markdown("custom _description_"),
					},
				},
			},
			`attr = foobar`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			&lang.HoverData{
				Content: lang.Markdown("`foobar` _custom name_\n\ncustom _description_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
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
