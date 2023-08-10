// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"
	"sort"
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
		expectedDiagnostics map[string]hcl.Diagnostics
	}{
		{
			"empty schema",
			schema.NewBodySchema(),
			``,
			map[string]hcl.Diagnostics{},
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
			map[string]hcl.Diagnostics{},
		},
		// attributes
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
			`test = 1
	foo = 1`,
			map[string]hcl.Diagnostics{
				"test.tf": {
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unexpected attribute",
						Detail:   "An attribute named \"foo\" is not expected here",
						Subject: &hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 2, Byte: 10},
							End:      hcl.Pos{Line: 2, Column: 9, Byte: 17},
						},
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
	test = 1
	foo = 1
}`,
			map[string]hcl.Diagnostics{
				"test.tf": {
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unexpected attribute",
						Detail:   "An attribute named \"foo\" is not expected here",
						Subject: &hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 2, Byte: 17},
							End:      hcl.Pos{Line: 3, Column: 9, Byte: 24},
						},
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
			map[string]hcl.Diagnostics{
				"test.tf": {
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
		},
		// blocks
		{
			"missing required attribute",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"wakka": {
						IsRequired: true,
						Constraint: schema.LiteralType{Type: cty.String},
					},
					"bar": {
						Constraint: schema.LiteralType{Type: cty.String},
					},
				},
			},
			`bar = "baz"`,
			map[string]hcl.Diagnostics{
				"test.tf": {
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Required attribute \"wakka\" not specified",
						Detail:   "An attribute named \"wakka\" is required here",
						Subject: &hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
						},
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
			map[string]hcl.Diagnostics{
				"test.tf": {
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
			`foo {
	test =1
}`,
			map[string]hcl.Diagnostics{
				"test.tf": {
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
			`foo "expected" "notExpected" {
	test = 1
}`,
			map[string]hcl.Diagnostics{
				"test.tf": {
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
			`foo "expected" {
	test = 1
}`,
			map[string]hcl.Diagnostics{
				"test.tf": {
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
		},
		{
			"too many blocks",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						Body: &schema.BodySchema{
							Blocks: map[string]*schema.BlockSchema{
								"bar": {
									MaxItems: 1,
								},
								"two": {},
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
				bar {}
				bar {}
				two {}
				test = 1
			}`,
			map[string]hcl.Diagnostics{
				"test.tf": {
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Too many blocks specified for \"bar\"",
						Detail:   "Only 1 block(s) are expected for \"bar\"",
						Subject: &hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 5, Byte: 4},
							End:      hcl.Pos{Line: 6, Column: 5, Byte: 56},
						},
					},
				},
			},
		},
		// either min or max is in schema but no blocks specified
		{
			"too few blocks",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						Body: &schema.BodySchema{
							Blocks: map[string]*schema.BlockSchema{
								"one": {
									MinItems: 2,
								},
								"two": {},
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
				test = 1
			}`,
			map[string]hcl.Diagnostics{
				"test.tf": {
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Too few blocks specified for \"one\"",
						Detail:   "At least 2 block(s) are expected for \"one\"",
						Subject: &hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 5, Byte: 4},
							End:      hcl.Pos{Line: 5, Column: 5, Byte: 45},
						},
					},
				},
			},
		},
		{
			"minitems with no blocks",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						Body: &schema.BodySchema{
							Blocks: map[string]*schema.BlockSchema{
								"one": {
									MinItems: 2,
								},
								"two": {},
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
				test = 1
			}`,
			map[string]hcl.Diagnostics{
				"test.tf": {
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Too few blocks specified for \"one\"",
						Detail:   "At least 2 block(s) are expected for \"one\"",
						Subject: &hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 5, Byte: 4},
							End:      hcl.Pos{Line: 3, Column: 5, Byte: 23},
						},
					},
				},
			},
		},
		{
			"min and max items with enough blocks for minitems",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						Body: &schema.BodySchema{
							Blocks: map[string]*schema.BlockSchema{
								"one": {
									MinItems: 2,
									MaxItems: 4,
								},
								"two": {},
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
				one {}
				one {}
				test = 1
			}`,
			map[string]hcl.Diagnostics{},
		},
		{
			"min and max set on two different blocks with correct number",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						Body: &schema.BodySchema{
							Blocks: map[string]*schema.BlockSchema{
								"one": {
									MinItems: 2,
								},
								"two": {
									MaxItems: 1,
								},
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
				one {}
				two {}
				test = 1
			}`,
			map[string]hcl.Diagnostics{},
		},
		{
			"min and max set on two different blocks with incorrect number",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						Body: &schema.BodySchema{
							Blocks: map[string]*schema.BlockSchema{
								"one": {
									MinItems: 2,
								},
								"two": {
									MaxItems: 1,
								},
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
				two {}
				test = 1
			}`,
			map[string]hcl.Diagnostics{
				"test.tf": {
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Too few blocks specified for \"one\"",
						Detail:   "At least 2 block(s) are expected for \"one\"",
						Subject: &hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 5, Byte: 4},
							End:      hcl.Pos{Line: 6, Column: 5, Byte: 56},
						},
					},
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Too many blocks specified for \"two\"",
						Detail:   "Only 1 block(s) are expected for \"two\"",
						Subject: &hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 5, Byte: 4},
							End:      hcl.Pos{Line: 6, Column: 5, Byte: 56},
						},
					},
				},
			},
		},
		{
			"max is in schema but no blocks specified",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						Body: &schema.BodySchema{
							Blocks: map[string]*schema.BlockSchema{
								"one": {
									MaxItems: 4,
								},
								"two": {},
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
				test = 1
			}`,
			map[string]hcl.Diagnostics{},
		},
		{
			"any attribute",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						Body: &schema.BodySchema{
							AnyAttribute: &schema.AttributeSchema{
								Constraint: schema.LiteralType{Type: cty.Number},
								IsOptional: true,
							},
						},
					},
				},
			},
			`foo {
				test = 1
			}`,
			map[string]hcl.Diagnostics{},
		},
		{
			"deprecated any attribute",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						Body: &schema.BodySchema{
							AnyAttribute: &schema.AttributeSchema{
								Constraint:   schema.LiteralType{Type: cty.Number},
								IsOptional:   true,
								IsDeprecated: true,
							},
						},
					},
				},
			},
			`foo {
				test = 1
			}`,
			map[string]hcl.Diagnostics{
				"test.tf": {
					{
						Severity: hcl.DiagWarning,
						Summary:  `"test" is deprecated`,
						Detail:   `Reason: ""`,
						Subject: &hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 5, Byte: 10},
							End:      hcl.Pos{Line: 2, Column: 13, Byte: 18},
						},
					},
				},
			},
		},
		{
			"count attribute present no diags",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								Count: true,
							},
						},
					},
				},
			},
			`foo {
				count = 1
			}`,
			map[string]hcl.Diagnostics{},
		},
		{
			"for_each attribute present no diags",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								ForEach: true,
							},
						},
					},
				},
			},
			`foo {
				for_each = []
			}`,
			map[string]hcl.Diagnostics{},
		},
		{
			"dynamic block disabled and present",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								DynamicBlocks: false,
							},
						},
						Labels: []*schema.LabelSchema{
							{
								Name:     "key",
								IsDepKey: true,
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "toot"},
								},
							}): {
								Blocks: map[string]*schema.BlockSchema{
									"noot": {
										Body: &schema.BodySchema{
											Attributes: map[string]*schema.AttributeSchema{
												"test": {
													IsOptional: true,
													Constraint: schema.LiteralType{Type: cty.String},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			`foo "toot" {
				dynamic "test" {
					for_each = {}
				}
			}`,
			map[string]hcl.Diagnostics{
				"test.tf": {
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unexpected block",
						Detail:   `Blocks of type "dynamic" are not expected here`,
						Subject: &hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 5, Byte: 17},
							End:      hcl.Pos{Line: 2, Column: 12, Byte: 24},
						},
					},
				},
			},
		},
		{
			"dynamic block enabled and present",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								DynamicBlocks: true,
							},
						},
						Labels: []*schema.LabelSchema{
							{
								Name:     "key",
								IsDepKey: true,
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "toot"},
								},
							}): {
								Blocks: map[string]*schema.BlockSchema{
									"noot": {
										Body: &schema.BodySchema{
											Attributes: map[string]*schema.AttributeSchema{
												"test": {
													IsOptional: true,
													Constraint: schema.LiteralType{Type: cty.String},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			`foo "toot" {
				dynamic "test" {
					for_each = {}
				}
			}`,
			map[string]hcl.Diagnostics{},
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

			sortDiagnostics(diags["test.tf"])

			if diff := cmp.Diff(tc.expectedDiagnostics["test.tf"], diags["test.tf"]); diff != "" {
				t.Fatalf("unexpected diagnostics: %s", diff)
			}
		})
	}
}

func sortDiagnostics(diags hcl.Diagnostics) {
	sort.Slice(diags, func(i, j int) bool {
		return diags[i].Subject.Start.Byte < diags[j].Subject.Start.Byte ||
			diags[i].Summary < diags[j].Summary
	})
}
