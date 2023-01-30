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

func TestSemanticTokens_exprMap(t *testing.T) {
	testCases := []struct {
		testName               string
		attrSchema             map[string]*schema.AttributeSchema
		cfg                    string
		expectedSemanticTokens []lang.SemanticToken
	}{
		{
			"undefined element constraint",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{},
				},
			},
			`attr = {}`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
			},
		},
		{
			"single-line with mismatching expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "foobar",
						},
					},
				},
			},
			`attr = [ foobar ]`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
			},
		},
		{
			"single-line with mismatching key type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "foobar",
						},
					},
				},
			},
			`attr = { 422 = foobar }`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
			},
		},
		{
			"single-line with valid item",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "foobar",
						},
					},
				},
			},
			`attr = { foo = foobar }`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
						End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
					},
				},
				{
					Type:      lang.TokenKeyword,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 16, Byte: 15},
						End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
					},
				},
			},
		},
		{
			"single-line with valid multiple items",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "foobar",
						},
					},
				},
			},
			`attr = { foo = foobar, bar = foobar }`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
						End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
					},
				},
				{
					Type:      lang.TokenKeyword,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 16, Byte: 15},
						End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 24, Byte: 23},
						End:      hcl.Pos{Line: 1, Column: 27, Byte: 26},
					},
				},
				{
					Type:      lang.TokenKeyword,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 30, Byte: 29},
						End:      hcl.Pos{Line: 1, Column: 36, Byte: 35},
					},
				},
			},
		},
		{
			"single-line with valid multiple items with quoted keys",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "foobar",
						},
					},
				},
			},
			`attr = { "foo" = foobar, "bar" = foobar }`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
						End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
					},
				},
				{
					Type:      lang.TokenKeyword,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 18, Byte: 17},
						End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 26, Byte: 25},
						End:      hcl.Pos{Line: 1, Column: 31, Byte: 30},
					},
				},
				{
					Type:      lang.TokenKeyword,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 34, Byte: 33},
						End:      hcl.Pos{Line: 1, Column: 40, Byte: 39},
					},
				},
			},
		},
		{
			"single-line with multiple items and one mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "foobar",
						},
					},
				},
			},
			`attr = { foo = bar, bar = foobar }`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
						End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 21, Byte: 20},
						End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
					},
				},
				{
					Type:      lang.TokenKeyword,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 27, Byte: 26},
						End:      hcl.Pos{Line: 1, Column: 33, Byte: 32},
					},
				},
			},
		},
		{
			"multi-line with valid item",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "foobar",
						},
					},
				},
			},
			`attr = {
  foo = foobar
}`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 6, Byte: 14},
					},
				},
				{
					Type:      lang.TokenKeyword,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 9, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 15, Byte: 23},
					},
				},
			},
		},
		{
			"multi-line with valid multiple items",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "foobar",
						},
					},
				},
			},
			`attr = {
  foo = foobar
  bar = foobar
}`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 6, Byte: 14},
					},
				},
				{
					Type:      lang.TokenKeyword,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 9, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 15, Byte: 23},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 26},
						End:      hcl.Pos{Line: 3, Column: 6, Byte: 29},
					},
				},
				{
					Type:      lang.TokenKeyword,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 9, Byte: 32},
						End:      hcl.Pos{Line: 3, Column: 15, Byte: 38},
					},
				},
			},
		},
		{
			"multi-line with multiple items and one mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Keyword{
							Keyword: "foobar",
						},
					},
				},
			},
			`attr = {
  foo = bar
  bar = foobar
}`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 5, Byte: 4},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 11},
						End:      hcl.Pos{Line: 2, Column: 6, Byte: 14},
					},
				},
				{
					Type:      lang.TokenMapKey,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 23},
						End:      hcl.Pos{Line: 3, Column: 6, Byte: 26},
					},
				},
				{
					Type:      lang.TokenKeyword,
					Modifiers: lang.SemanticTokenModifiers{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 9, Byte: 29},
						End:      hcl.Pos{Line: 3, Column: 15, Byte: 35},
					},
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
			tokens, err := d.SemanticTokensInFile(ctx, "test.tf")
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedSemanticTokens, tokens); diff != "" {
				t.Fatalf("unexpected tokens: %s", diff)
			}
		})
	}
}
