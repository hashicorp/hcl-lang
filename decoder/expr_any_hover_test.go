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

func TestHoverAtPos_exprAny_functions(t *testing.T) {
	testCases := []struct {
		testName     string
		attrSchema   map[string]*schema.AttributeSchema
		cfg          string
		pos          hcl.Pos
		expectedData *lang.HoverData
	}{
		{
			"over unknown function",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = unknown()
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			nil,
		},
		{
			"over name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = lower("FOO")
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			&lang.HoverData{
				Content: lang.MarkupContent{
					Value: "```terraform\nlower(str string) string\n```\n\n`lower` converts all cased letters in the given string to lowercase.",
					Kind:  lang.MarkdownKind,
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 20, Byte: 19},
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
				Functions: testFunctionSignatures(),
			})

			ctx := context.Background()
			data, err := d.HoverAtPos(ctx, "test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedData, data); diff != "" {
				t.Fatalf("unexpected data: %s", diff)
			}
		})
	}

}
