package decoder

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestValidate_schema(t *testing.T) {
	testCases := []struct {
		testName            string
		bodySchema          *schema.BodySchema
		cfg                 string
		expectedDiagnostics hcl.Diagnostics
	}{
		{
			"empty schema",
			schema.NewBodySchema(),
			``,
			hcl.Diagnostics{},
		},
		{
			"valid schema",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"test": {
						Constraint: schema.LiteralType{Type: cty.Number},
						IsRequired: true,
					},
				},
			},
			`test = 1`,
			hcl.Diagnostics{},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%2d-%s", i, tc.testName), func(t *testing.T) {
			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			d := testPathDecoder(t, &PathContext{
				Schema: tc.bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			})

			ctx := context.Background()
			diags, err := d.Validate(ctx)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedDiagnostics, diags); diff != "" {
				t.Fatalf("unexpected diagnostics: %s", diff)
			}
		})
	}
}
