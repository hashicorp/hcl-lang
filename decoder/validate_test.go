// Copyright (c) HashiCorp, Inc.
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
		{
			"deprecated attribute",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"test": {
						Constraint: schema.LiteralType{Type: cty.Number},
						IsRequired: true,
					},
					"wakka": {
						Constraint:   schema.LiteralType{Type: cty.Number},
						IsRequired:   false,
						IsDeprecated: true,
						Description: lang.MarkupContent{
							Value: "Use `wakka_wakka` instead",
							Kind:  lang.MarkdownKind,
						},
					},
				},
			},
			`test = 1
wakka = 2
`,
			hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagWarning,
					Summary:  "\"wakka\" is deprecated",
					Detail:   "Reason: \"Use `wakka_wakka` instead\"",
					Subject: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 9},
						End:      hcl.Pos{Line: 2, Column: 10, Byte: 18},
					},
				},
			},
		},
		{
			"unknown block",
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
			`bar {}`,
			hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unexpected block",
					Detail:   "Blocks of type \"bar\" are not expected here",
					Subject: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
		},

		{
			"deprecated block",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						IsDeprecated: true,
						Description: lang.MarkupContent{
							Value: "Use `wakka` instead",
							Kind:  lang.MarkdownKind,
						},
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
			`foo {}`,
			hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagWarning,
					Summary:  "\"foo\" is deprecated",
					Detail:   "Reason: \"Use `wakka` instead\"",
					Subject: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
		},

		{
			"extra block labels",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						Labels: []*schema.LabelSchema{
							{
								Name: "expected",
							},
						},
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
			`foo "expected" "notExpected" {}`,
			hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Too many labels specified for \"foo\"",
					Detail:   "Only 1 label(s) are expected for \"foo\" blocks",
					Subject: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 16, Byte: 15},
						End:      hcl.Pos{Line: 1, Column: 29, Byte: 28},
					},
				},
			},
		},

		{
			"too few block labels",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						Labels: []*schema.LabelSchema{
							{
								Name: "expected",
							},
							{
								Name: "expected2",
							},
						},
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
			`foo "expected" {}`,
			hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Not enough labels specified for \"foo\"",
					Detail:   "All \"foo\" blocks must have 2 label(s)",
					Subject: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
		},

		{
			"too many blocks",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						MaxItems: 1,
						Body: &schema.BodySchema{
							Blocks: map[string]*schema.BlockSchema{
								"one": &schema.BlockSchema{},
								"two": &schema.BlockSchema{},
							},
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
				one {}
				two {}
			}`,
			hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Too many blocks specified for \"foo\"",
					Detail:   "Only 1 block(s) are expected for \"foo\"",
					Subject: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
		},

		{
			"too few blocks",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						MinItems: 2,
						Body: &schema.BodySchema{
							Blocks: map[string]*schema.BlockSchema{
								"one": &schema.BlockSchema{},
								"two": &schema.BlockSchema{},
							},
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
				one {}
			}`,
			hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Too few blocks specified for \"foo\"",
					Detail:   "At least 2 block(s) are expected for \"foo\"",
					Subject: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
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
