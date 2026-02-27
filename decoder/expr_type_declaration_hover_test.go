// Copyright IBM Corp. 2026
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

func TestHoverAtPos_exprTypeDeclaration(t *testing.T) {
	testCases := []struct {
		testName          string
		attrSchema        map[string]*schema.AttributeSchema
		cfg               string
		pos               hcl.Pos
		expectedHoverData *lang.HoverData
	}{
		{
			"primitive type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = string`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown(`_string_`),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
				},
			},
		},
		{
			"list type on list",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = list(string)`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			&lang.HoverData{
				Content: lang.Markdown(`_list of string_`),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 20, Byte: 19},
				},
			},
		},
		{
			"list type on element type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = list(string)`,
			hcl.Pos{Line: 1, Column: 16, Byte: 15},
			&lang.HoverData{
				Content: lang.Markdown(`_string_`),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 13, Byte: 12},
					End:      hcl.Pos{Line: 1, Column: 19, Byte: 18},
				},
			},
		},
		{
			"tuple type on tuple",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = tuple([string])`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			&lang.HoverData{
				Content: lang.Markdown(`_tuple_`),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
				},
			},
		},
		{
			"object type on object",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  foo = string
})
`,
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("```\n{\n  foo = string\n}\n```\n_object_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 3, Column: 3, Byte: 33},
				},
			},
		},
		{
			"object type on attribute name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  foo = string
})
`,
			hcl.Pos{Line: 2, Column: 5, Byte: 20},
			&lang.HoverData{
				Content: lang.Markdown("`foo` = _string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 3, Byte: 18},
					End:      hcl.Pos{Line: 2, Column: 15, Byte: 30},
				},
			},
		},
		{
			"object type on attribute value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.TypeDeclaration{},
				},
			},
			`attr = object({
  foo = string
})
`,
			hcl.Pos{Line: 2, Column: 11, Byte: 26},
			&lang.HoverData{
				Content: lang.Markdown("_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 9, Byte: 24},
					End:      hcl.Pos{Line: 2, Column: 15, Byte: 30},
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
