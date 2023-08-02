// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
		{
			"unknown attribute",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"test": {
						Constraint: schema.LiteralType{Type: cty.Number},
						IsRequired: true,
					},
				},
			},
			`foo = 1`,
			hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unexpected attribute",
					Detail:   "An attribute named \"foo\" is not expected here",
					Subject: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
					},
				},
			},
		},
		{
			"unknown block attribute",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"test": {
									Constraint: schema.LiteralType{Type: cty.Number},
									IsRequired: true,
								},
							},
						},
					},
				},
			},
			`foo {
	foo = 1
}`,
			hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unexpected attribute",
					Detail:   "An attribute named \"foo\" is not expected here",
					Subject: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 2, Byte: 7},
						End:      hcl.Pos{Line: 2, Column: 9, Byte: 14},
					},
				},
			},
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
