// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func TestCompletionAtPos_BodySchema_Extensions_Count(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		testName           string
		bodySchema         *schema.BodySchema
		referenceTargets   reference.Targets
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"count attribute completion",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name: "type",
							},
							{
								Name: "name",
							},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								Count: true,
							},
						},
					},
				},
			},
			reference.Targets{},
			`resource "aws_instance" "foo" {

}`,
			hcl.Pos{
				Line:   2,
				Column: 1,
				Byte:   32,
			},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label: "count",
					Description: lang.MarkupContent{
						Value: "Total number of instances of this block.\n\n**Note**: A given block cannot use both `count` and `for_each`.",
						Kind:  lang.MarkdownKind,
					},
					Detail: "optional, number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   2,
								Column: 1,
								Byte:   32,
							},
							End: hcl.Pos{
								Line:   2,
								Column: 1,
								Byte:   32,
							},
						},
						NewText: "count",
						Snippet: "count = ",
					},
					TriggerSuggest: true,
					Kind:           lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"count attribute completion with DependentBody",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name:     "type",
								IsDepKey: true,
							},
							{
								Name: "name",
							},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								Count: true,
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "aws_instance"},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"instance_size": {
										IsOptional: true,
										Constraint: schema.LiteralType{Type: cty.String},
									},
								},
							},
						},
					},
				},
			},
			reference.Targets{},
			`resource "aws_instance" "foo" {

}
`,
			hcl.Pos{
				Line:   2,
				Column: 1,
				Byte:   32,
			},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label: "count",
					Description: lang.MarkupContent{
						Value: "Total number of instances of this block.\n\n**Note**: A given block cannot use both `count` and `for_each`.",
						Kind:  lang.MarkdownKind,
					},
					Detail: "optional, number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   2,
								Column: 1,
								Byte:   32,
							},
							End: hcl.Pos{
								Line:   2,
								Column: 1,
								Byte:   32,
							},
						},
						NewText: "count",
						Snippet: "count = ",
					},
					TriggerSuggest: true,
					Kind:           lang.AttributeCandidateKind,
				},
				{
					Label:  "instance_size",
					Detail: "optional, string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   2,
								Column: 1,
								Byte:   32,
							},
							End: hcl.Pos{
								Line:   2,
								Column: 1,
								Byte:   32,
							},
						},
						NewText: "instance_size",
						Snippet: `instance_size = "${1:value}"`,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"count attribute completion does not complete count when extensions not enabled",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name: "type",
							},
							{
								Name: "name",
							},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								Count: false,
							},
						},
					},
				},
			},
			reference.Targets{},
			`resource "aws_instance" "foo" {

}`,
			hcl.Pos{
				Line:   2,
				Column: 1,
				Byte:   32,
			},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"count.index does not complete when not needed",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name: "type",
							},
							{
								Name: "name",
							},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"cpu_count": {
									IsOptional: true,
									Constraint: schema.Reference{
										OfType: cty.Number,
									},
								},
							},
							Extensions: &schema.BodyExtensions{
								Count: true,
							},
						},
					},
				},
			},
			reference.Targets{},
			`resource "aws_instance" "foo" {
	cpu_count =
}`,
			hcl.Pos{
				Line:   2,
				Column: 8,
				Byte:   44,
			},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"count.index value completion",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name: "type",
							},
							{
								Name: "name",
							},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"cpu_count": {
									IsOptional: true,
									Constraint: schema.Reference{
										OfType: cty.Number,
									},
								},
							},
							Extensions: &schema.BodyExtensions{
								Count: true,
							},
						},
					},
				},
			},
			reference.Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "count"},
						lang.AttrStep{Name: "index"},
					},
					Type:        cty.Number,
					Description: lang.PlainText("The distinct index number (starting with 0) corresponding to the instance"),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   34,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 12,
							Byte:   43,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   34,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 8,
							Byte:   39,
						},
					},
				},
			},
			`resource "aws_instance" "foo" {
  count = 4
  cpu_count = 
}`,
			hcl.Pos{Line: 3, Column: 15, Byte: 57},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label: "count.index",
					Description: lang.MarkupContent{
						Value: "The distinct index number (starting with 0) corresponding to the instance",
						Kind:  lang.PlainTextKind,
					},
					Detail: "number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 15, Byte: 57},
							End:      hcl.Pos{Line: 3, Column: 15, Byte: 57},
						},
						NewText: "count.index",
						Snippet: "count.index",
					},
					Kind: lang.ReferenceCandidateKind,
				},
			}),
		},
		{
			"count does not complete more than once",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name: "type",
							},
							{
								Name: "name",
							},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								Count: true,
							},
						},
					},
				},
			},
			reference.Targets{},
			`resource "aws_instance" "foo" {
	count = 4

}`,
			hcl.Pos{
				Line:   3,
				Column: 1,
				Byte:   43,
			},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"count.index does not complete when extension not enabled",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name: "type",
							},
							{
								Name: "name",
							},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"cpu_count": {
									IsOptional: true,
									Constraint: schema.Reference{
										OfType: cty.Number,
									},
								},
							},
							Extensions: &schema.BodyExtensions{
								Count: false,
							},
						},
					},
				},
			},
			reference.Targets{},
			`resource "aws_instance" "foo" {
	cpu_count =
}`,
			hcl.Pos{
				Line:   2,
				Column: 8,
				Byte:   44,
			},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"count.index completes when inside nested blocks",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name: "type",
							},
							{
								Name: "name",
							},
						},
						Body: &schema.BodySchema{
							Blocks: map[string]*schema.BlockSchema{
								"foo": {
									Body: &schema.BodySchema{
										Attributes: map[string]*schema.AttributeSchema{
											"cpu_count": {
												IsOptional: true,
												Constraint: schema.Reference{
													OfType: cty.Number,
												},
											},
										},
									},
								},
							},
							Extensions: &schema.BodyExtensions{
								Count: true,
							},
						},
					},
				},
			},
			reference.Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "count"},
						lang.AttrStep{Name: "index"},
					},
					Type:        cty.Number,
					Description: lang.PlainText("The distinct index number (starting with 0) corresponding to the instance"),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   34,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 12,
							Byte:   43,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   34,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 8,
							Byte:   39,
						},
					},
				},
			},
			`resource "aws_instance" "foo" {
  count = 4
  foo {
    cpu_count =
  }
}`,
			hcl.Pos{Line: 4, Column: 17, Byte: 67},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label: "count.index",
					Description: lang.MarkupContent{
						Value: "The distinct index number (starting with 0) corresponding to the instance",
						Kind:  lang.PlainTextKind,
					},
					Detail: "number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 4, Column: 17, Byte: 67},
							End:      hcl.Pos{Line: 4, Column: 17, Byte: 67},
						},
						NewText: "count.index",
						Snippet: "count.index",
					},
					Kind: lang.ReferenceCandidateKind,
				},
			}),
		},
		{
			"count attribute expression completion",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name: "type",
							},
							{
								Name: "name",
							},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								Count: true,
							},
						},
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					ScopeId: lang.ScopeId("variable"),
					Type:    cty.Number,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   4,
							Column: 1,
							Byte:   45,
						},
						End: hcl.Pos{
							Line:   6,
							Column: 2,
							Byte:   79,
						},
					},
				},
			},
			`resource "aws_instance" "foo" {
  count = 
}
variable "test" {
	type = number
}
`,
			hcl.Pos{
				Line:   2,
				Column: 11,
				Byte:   42,
			},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "var.test",
					Detail: "number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start: hcl.Pos{
								Line:   2,
								Column: 11,
								Byte:   42,
							},
							End: hcl.Pos{
								Line:   2,
								Column: 11,
								Byte:   42,
							},
						},
						NewText: "var.test",
						Snippet: "var.test",
					},
					Kind: lang.ReferenceCandidateKind,
				},
			}),
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)

			d := testPathDecoder(t, &PathContext{
				Schema: tc.bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
				ReferenceTargets: tc.referenceTargets,
			})

			candidates, err := d.CompletionAtPos(ctx, "test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedCandidates, candidates); diff != "" {
				t.Fatalf("unexpected candidates: %s", diff)
			}
		})
	}
}

func TestCompletionAtPos_BodySchema_Extensions_ForEach(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		testName           string
		bodySchema         *schema.BodySchema
		referenceTargets   reference.Targets
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"foreach attribute completion",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type"}, {Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								ForEach: true,
							},
						},
					},
				},
			},
			reference.Targets{},
			`resource "aws_instance" "foo" {

}`,
			hcl.Pos{Line: 2, Column: 1, Byte: 32},
			lang.CompleteCandidates([]lang.Candidate{{
				Label: "for_each",
				Description: lang.MarkupContent{
					Value: "A meta-argument that accepts a map or a set of strings, and creates an instance for each item in that map or set.\n\n**Note**: A given block cannot use both `count` and `for_each`.",
					Kind:  lang.MarkdownKind,
				},
				Detail: "optional, map of any single type or set of string or object",
				TextEdit: lang.TextEdit{
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 32},
						End:      hcl.Pos{Line: 2, Column: 1, Byte: 32},
					},
					NewText: "for_each",
					Snippet: "for_each = ",
				},
				Kind:           lang.AttributeCandidateKind,
				TriggerSuggest: true,
			},
			}),
		},
		{
			"foreach attribute completion does not complete foreach when extensions not enabled",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type"}, {Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								ForEach: false,
							},
						},
					},
				},
			},
			reference.Targets{},
			`resource "aws_instance" "foo" {

}`,
			hcl.Pos{Line: 2, Column: 1, Byte: 32},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"each.* value completion",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type"}, {Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								ForEach: true,
							},
							Attributes: map[string]*schema.AttributeSchema{
								"thing": {
									IsOptional: true,
									Constraint: schema.Reference{
										OfType: cty.String,
									},
								},
							},
						},
					},
				},
			},
			reference.Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "each"},
						lang.AttrStep{Name: "key"},
					},
					Type:        cty.String,
					Description: lang.Markdown("The map key (or set member) corresponding to this instance"),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 32},
						End:      hcl.Pos{Line: 5, Column: 2, Byte: 93},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 32},
						End:      hcl.Pos{Line: 2, Column: 9, Byte: 40},
					},
				},
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "each"},
						lang.AttrStep{Name: "value"},
					},
					Type:        cty.DynamicPseudoType,
					Description: lang.Markdown("The map value corresponding to this instance. (If a set was provided, this is the same as `each.key`.)"),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 32},
						End:      hcl.Pos{Line: 5, Column: 2, Byte: 93},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 32},
						End:      hcl.Pos{Line: 2, Column: 9, Byte: 40},
					},
				},
			},
			`resource "aws_instance" "foo" {
for_each = {
	a_group = "eastus"
	another_group = "westus2"
}
thing = 
}`,
			hcl.Pos{Line: 6, Column: 9, Byte: 102},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "each.key",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					Description: lang.MarkupContent{
						Value: "The map key (or set member) corresponding to this instance",
						Kind:  lang.MarkdownKind,
					},
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 6, Column: 9, Byte: 102},
							End:      hcl.Pos{Line: 6, Column: 9, Byte: 102},
						},
						NewText: "each.key",
						Snippet: "each.key",
					},
				},
				{
					Label:  "each.value",
					Detail: "dynamic",
					Kind:   lang.ReferenceCandidateKind,
					Description: lang.MarkupContent{
						Value: "The map value corresponding to this instance. (If a set was provided, this is the same as `each.key`.)",
						Kind:  lang.MarkdownKind,
					},
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 6, Column: 9, Byte: 102},
							End:      hcl.Pos{Line: 6, Column: 9, Byte: 102},
						},
						NewText: "each.value",
						Snippet: "each.value",
					},
				},
			}),
		},
		{
			"each.* does not complete when not needed",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type"}, {Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								ForEach: true,
							},
							Attributes: map[string]*schema.AttributeSchema{
								"thing": {
									IsOptional: true,
									Constraint: schema.Reference{
										OfType: cty.String,
									},
								},
							},
						},
					},
				},
			},
			reference.Targets{},
			`resource "aws_instance" "foo" {
thing = 
}`,
			hcl.Pos{Line: 2, Column: 9, Byte: 40},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"each.* does not complete when extension not enabled",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type"}, {Name: "name"},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"thing": {
									IsOptional: true,
									Constraint: schema.Reference{
										OfType: cty.Number,
									},
								},
							},
							Extensions: &schema.BodyExtensions{
								ForEach: false,
							},
						},
					},
				},
			},
			reference.Targets{},
			`resource "aws_instance" "foo" {
	thing = 
}`,
			hcl.Pos{Line: 2, Column: 8, Byte: 41},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"for_each does not complete more than once",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type"}, {Name: "name"},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"thing": {
									IsOptional: true,
									Constraint: schema.Reference{
										OfType: cty.Number,
									},
								},
							},
							Extensions: &schema.BodyExtensions{
								ForEach: true,
							},
						},
					},
				},
			},
			reference.Targets{},
			`resource "aws_instance" "foo" {
for_each = {
	a_group = "eastus"
	another_group = "westus2"
}

}`,
			hcl.Pos{Line: 6, Column: 1, Byte: 94},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "thing",
					Detail: "optional, number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 6, Column: 1, Byte: 94},
							End:      hcl.Pos{Line: 6, Column: 1, Byte: 94},
						},
						NewText: "thing",
						Snippet: "thing = ",
					},
					TriggerSuggest: true,
					Kind:           lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"each.* completes when inside nested blocks",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type"}, {Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								ForEach: true,
							},
							Blocks: map[string]*schema.BlockSchema{
								"foo": {
									Body: &schema.BodySchema{
										Attributes: map[string]*schema.AttributeSchema{
											"thing": {
												IsOptional: true,
												Constraint: schema.Reference{
													OfType: cty.DynamicPseudoType,
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
			reference.Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "each"},
						lang.AttrStep{Name: "key"},
					},
					Type:        cty.String,
					Description: lang.Markdown("The map key (or set member) corresponding to this instance"),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 32},
						End:      hcl.Pos{Line: 5, Column: 2, Byte: 93},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 32},
						End:      hcl.Pos{Line: 2, Column: 9, Byte: 40},
					},
				},
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "each"},
						lang.AttrStep{Name: "value"},
					},
					Type:        cty.DynamicPseudoType,
					Description: lang.Markdown("The map value corresponding to this instance. (If a set was provided, this is the same as `each.key`.)"),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 32},
						End:      hcl.Pos{Line: 5, Column: 2, Byte: 93},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 32},
						End:      hcl.Pos{Line: 2, Column: 9, Byte: 40},
					},
				},
			},
			`resource "aws_instance" "foo" {
for_each = {
	a_group = "eastus"
	another_group = "westus2"
}
foo {
	thing = 
}
}`,
			hcl.Pos{Line: 7, Column: 13, Byte: 109},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "each.key",
					Detail: "string",
					Kind:   lang.ReferenceCandidateKind,
					Description: lang.MarkupContent{
						Value: "The map key (or set member) corresponding to this instance",
						Kind:  lang.MarkdownKind,
					},
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 7, Column: 13, Byte: 109},
							End:      hcl.Pos{Line: 7, Column: 13, Byte: 109},
						},
						NewText: "each.key",
						Snippet: "each.key",
					},
				},
				{
					Label:  "each.value",
					Detail: "dynamic",
					Kind:   lang.ReferenceCandidateKind,
					Description: lang.MarkupContent{
						Value: "The map value corresponding to this instance. (If a set was provided, this is the same as `each.key`.)",
						Kind:  lang.MarkdownKind,
					},
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 7, Column: 13, Byte: 109},
							End:      hcl.Pos{Line: 7, Column: 13, Byte: 109},
						},
						NewText: "each.value",
						Snippet: "each.value",
					},
				},
			}),
		},
		{
			"each.* does not complete for for_each",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type"}, {Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								ForEach: true,
							},
							Attributes: map[string]*schema.AttributeSchema{
								"thing": {
									IsOptional: true,
									Constraint: schema.Reference{
										OfType: cty.String,
									},
								},
							},
						},
					},
				},
			},
			reference.Targets{},
			`resource "aws_instance" "foo" {
for_each = 
}`,
			hcl.Pos{Line: 2, Column: 12, Byte: 43},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  `{ "key" = any type }`,
					Detail: "map of dynamic",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 12, Byte: 43},
							End:      hcl.Pos{Line: 2, Column: 12, Byte: 43},
						},
						NewText: "{\n  \"key\" = \n}",
						Snippet: "{\n  \"${1:key}\" = ${2}\n}",
					},
					Kind: lang.MapCandidateKind,
				},
				{
					Label:  "[ string ]",
					Detail: "set of string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 12, Byte: 43},
							End:      hcl.Pos{Line: 2, Column: 12, Byte: 43},
						},
						NewText: `[ "" ]`,
						Snippet: `[ "${1:value}" ]`,
					},
					Kind: lang.SetCandidateKind,
				},
				{
					Label:  "{  }",
					Detail: "object",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 12, Byte: 43},
							End:      hcl.Pos{Line: 2, Column: 12, Byte: 43},
						},
						NewText: "{\n}",
						Snippet: "{\n}",
					},
					Kind: lang.ObjectCandidateKind,
				},
			}),
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)

			d := testPathDecoder(t, &PathContext{
				Schema: tc.bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
				ReferenceTargets: tc.referenceTargets,
			})

			candidates, err := d.CompletionAtPos(ctx, "test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedCandidates, candidates); diff != "" {
				t.Fatalf("unexpected candidates: %s", diff)
			}
		})
	}
}

func TestCompletionAtPos_BodySchema_Extensions_SelfRef(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		testName           string
		bodySchema         *schema.BodySchema
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		// nested block
		{
			"target self addr enabled but no extension enabled",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name:        "type",
								IsDepKey:    true,
								Completable: true,
							},
							{Name: "name"},
						},
						Body: schema.NewBodySchema(),
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{
										Index: 0,
										Value: "aws_instance",
									},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"cpu_count": {
										IsOptional: true,
										Constraint: schema.OneOf{
											schema.Reference{
												OfType: cty.Number,
											},
											schema.LiteralType{
												Type: cty.Number,
											},
										},
									},
									"fox": {
										IsOptional: true,
										Constraint: schema.OneOf{
											schema.Reference{
												OfType: cty.Number,
											},
											schema.LiteralType{
												Type: cty.Number,
											},
										},
									},
								},
							},
						},
						Address: &schema.BlockAddrSchema{
							DependentBodyAsData:  true,
							InferDependentBody:   true,
							DependentBodySelfRef: true,
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
								schema.LabelStep{Index: 1},
							},
						},
					},
				},
			},
			`resource "aws_instance" "foo" {
  cpu_count = 4
  fox =
}`,
			hcl.Pos{Line: 3, Column: 8, Byte: 55},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"target self addr enabled and extension enabled",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name:        "type",
								IsDepKey:    true,
								Completable: true,
							},
							{Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								SelfRefs: true,
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{
										Index: 0,
										Value: "aws_instance",
									},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"cpu_count": {
										IsOptional: true,
										Constraint: schema.OneOf{
											schema.Reference{
												OfType: cty.Number,
											},
											schema.LiteralType{
												Type: cty.Number,
											},
										},
									},
									"fox": {
										IsOptional: true,
										Constraint: schema.OneOf{
											schema.Reference{
												OfType: cty.Number,
											},
											schema.LiteralType{
												Type: cty.Number,
											},
										},
									},
								},
							},
						},
						Address: &schema.BlockAddrSchema{
							DependentBodyAsData:  true,
							InferDependentBody:   true,
							DependentBodySelfRef: true,
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
								schema.LabelStep{Index: 1},
							},
						},
					},
				},
			},
			`resource "aws_instance" "foo" {
  cpu_count = 4
  fox =
}`,
			hcl.Pos{Line: 3, Column: 8, Byte: 55},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "self",
					Detail: "object",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 8, Byte: 55},
							End:      hcl.Pos{Line: 3, Column: 8, Byte: 55},
						},

						NewText: "self",
						Snippet: "self",
					},
					Kind: lang.ReferenceCandidateKind,
				},
			}),
		},
		{
			"target self addr disabled and extension enabled",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name:        "type",
								IsDepKey:    true,
								Completable: true,
							},
							{Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								SelfRefs: true,
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{
										Index: 0,
										Value: "aws_instance",
									},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"cpu_count": {
										IsOptional: true,
										Constraint: schema.OneOf{
											schema.Reference{
												OfType: cty.Number,
											},
											schema.LiteralType{
												Type: cty.Number,
											},
										},
									},
									"fox": {
										IsOptional: true,
										Constraint: schema.OneOf{
											schema.Reference{
												OfType: cty.Number,
											},
											schema.LiteralType{
												Type: cty.Number,
											},
										},
									},
								},
							},
						},
						Address: &schema.BlockAddrSchema{
							DependentBodyAsData: true,
							InferDependentBody:  true,
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
								schema.LabelStep{Index: 1},
							},
						},
					},
				},
			},
			`resource "aws_instance" "foo" {
  cpu_count = 4
  fox =
}`,
			hcl.Pos{Line: 3, Column: 8, Byte: 55},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"no cyclical completion (attr = self.attr)",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name:        "type",
								IsDepKey:    true,
								Completable: true,
							},
							{Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								SelfRefs: true,
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{
										Index: 0,
										Value: "aws_instance",
									},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"cpu_count": {
										IsOptional: true,
										Constraint: schema.OneOf{
											schema.Reference{
												OfType: cty.Number,
											},
											schema.LiteralType{
												Type: cty.Number,
											},
										},
									},
								},
							},
						},
						Address: &schema.BlockAddrSchema{
							DependentBodyAsData:  true,
							InferDependentBody:   true,
							DependentBodySelfRef: true,
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
								schema.LabelStep{Index: 1},
							},
						},
					},
				},
			},
			`resource "aws_instance" "foo" {
  cpu_count = 
}`,
			hcl.Pos{Line: 2, Column: 15, Byte: 46},
			lang.CompleteCandidates([]lang.Candidate{}),
		},
		{
			"completion with prefix",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name:        "type",
								IsDepKey:    true,
								Completable: true,
							},
							{Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								SelfRefs: true,
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{
										Index: 0,
										Value: "aws_instance",
									},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"cpu_count": {
										IsOptional: true,
										Constraint: schema.OneOf{
											schema.Reference{
												OfType: cty.Number,
											},
											schema.LiteralType{
												Type: cty.Number,
											},
										},
									},
									"fox": {
										IsOptional: true,
										Constraint: schema.OneOf{
											schema.Reference{
												OfType: cty.Number,
											},
											schema.LiteralType{
												Type: cty.Number,
											},
										},
									},
								},
							},
						},
						Address: &schema.BlockAddrSchema{
							DependentBodyAsData:  true,
							InferDependentBody:   true,
							DependentBodySelfRef: true,
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
								schema.LabelStep{Index: 1},
							},
						},
					},
				},
			},
			`resource "aws_instance" "foo" {
  cpu_count = 4
  fox = self.
}`,
			hcl.Pos{Line: 3, Column: 14, Byte: 61},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "self.cpu_count",
					Detail: "number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 9, Byte: 56},
							End:      hcl.Pos{Line: 3, Column: 14, Byte: 61},
						},
						NewText: "self.cpu_count",
						Snippet: "self.cpu_count",
					},
					Kind: lang.ReferenceCandidateKind,
				},
			}),
		},
		{
			"target self addr enabled and extension enabled within a block",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name:        "type",
								IsDepKey:    true,
								Completable: true,
							},
							{Name: "name"},
						},
						Body: &schema.BodySchema{
							Blocks: map[string]*schema.BlockSchema{
								"animal": {
									Body: &schema.BodySchema{
										Extensions: &schema.BodyExtensions{
											SelfRefs: true,
										},
										Attributes: map[string]*schema.AttributeSchema{
											"fox": {
												IsOptional: true,
												Constraint: schema.OneOf{
													schema.Reference{
														OfType: cty.Number,
													},
													schema.LiteralType{
														Type: cty.Number,
													},
												},
											},
										},
									},
								},
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{
										Index: 0,
										Value: "aws_instance",
									},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"cpu_count": {
										IsOptional: true,
										Constraint: schema.OneOf{
											schema.Reference{
												OfType: cty.Number,
											},
											schema.LiteralType{
												Type: cty.Number,
											},
										},
									},
								},
							},
						},
						Address: &schema.BlockAddrSchema{
							DependentBodyAsData:  true,
							InferDependentBody:   true,
							DependentBodySelfRef: true,
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
								schema.LabelStep{Index: 1},
							},
						},
					},
				},
			},
			`resource "aws_instance" "foo" {
  cpu_count = 4
  animal {
    fox =
  }
}`,
			hcl.Pos{Line: 4, Column: 10, Byte: 68},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "self",
					Detail: "object",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 4, Column: 10, Byte: 68},
							End:      hcl.Pos{Line: 4, Column: 10, Byte: 68},
						},

						NewText: "self",
						Snippet: "self",
					},
					Kind: lang.ReferenceCandidateKind,
				},
			}),
		},
		{
			"regular block with selfref enabled",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						Labels: []*schema.LabelSchema{
							{Name: "name"},
						},
						Address: &schema.BlockAddrSchema{
							BodyAsData:  true,
							InferBody:   true,
							BodySelfRef: true,
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "foo"},
								schema.LabelStep{Index: 0},
							},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"static": {
									IsOptional: true,
									Constraint: schema.OneOf{
										schema.Reference{
											OfType: cty.Number,
										},
										schema.LiteralType{
											Type: cty.Number,
										},
									},
								},
								"fox": {
									IsOptional: true,
									Constraint: schema.OneOf{
										schema.Reference{
											OfType: cty.Number,
										},
										schema.LiteralType{
											Type: cty.Number,
										},
									},
								},
							},
							Extensions: &schema.BodyExtensions{
								SelfRefs: true,
							},
						},
					},
				},
			},
			`foo "bar" {
  static = 4
  fox = 
}`,
			hcl.Pos{Line: 3, Column: 14, Byte: 33},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "self.static",
					Detail: "number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 14, Byte: 33},
							End:      hcl.Pos{Line: 3, Column: 14, Byte: 33},
						},
						NewText: "self.static",
						Snippet: "self.static",
					},
					Kind: lang.ReferenceCandidateKind,
				},
			}),
		},
		{
			"regular block with selfref enabled - prefix",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"foo": {
						Labels: []*schema.LabelSchema{
							{Name: "name"},
						},
						Address: &schema.BlockAddrSchema{
							BodyAsData:  true,
							InferBody:   true,
							BodySelfRef: true,
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "foo"},
								schema.LabelStep{Index: 0},
							},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"static": {
									IsOptional: true,
									Constraint: schema.OneOf{
										schema.Reference{
											OfType: cty.Number,
										},
										schema.LiteralType{
											Type: cty.Number,
										},
									},
								},
								"fox": {
									IsOptional: true,
									Constraint: schema.OneOf{
										schema.Reference{
											OfType: cty.Number,
										},
										schema.LiteralType{
											Type: cty.Number,
										},
									},
								},
							},
							Extensions: &schema.BodyExtensions{
								SelfRefs: true,
							},
						},
					},
				},
			},
			`foo "bar" {
  static = 4
  fox = self.
}`,
			hcl.Pos{Line: 3, Column: 14, Byte: 38},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "self.static",
					Detail: "number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 9, Byte: 33},
							End:      hcl.Pos{Line: 3, Column: 14, Byte: 38},
						},
						NewText: "self.static",
						Snippet: "self.static",
					},
					Kind: lang.ReferenceCandidateKind,
				},
			}),
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)

			d := testPathDecoder(t, &PathContext{
				Schema: tc.bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			})
			targets, err := d.CollectReferenceTargets()
			if err != nil {
				t.Fatal(err)
			}
			d = testPathDecoder(t, &PathContext{
				Schema: tc.bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
				ReferenceTargets: targets,
			})

			candidates, err := d.CompletionAtPos(ctx, "test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedCandidates, candidates); diff != "" {
				t.Fatalf("unexpected candidates: %s", diff)
			}
		})
	}
}

func TestCompletionAtPos_BodySchema_Extensions_DynamicBlock(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		testName           string
		bodySchema         *schema.BodySchema
		cfg                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
		expectedErr        string
	}{
		{
			"dynamic block does not complete if not enabled",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name: "type",
							},
							{
								Name: "name",
							},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								DynamicBlocks: false,
							},
						},
					},
				},
			},
			`
resource "aws_elastic_beanstalk_environment" "example" {
	name = "example"
	
}`,
			hcl.Pos{
				Line:   4,
				Column: 3,
				Byte:   77,
			},
			lang.CompleteCandidates([]lang.Candidate{}),
			"",
		},
		{
			"dynamic block does not complete without blocks",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name:     "type",
								IsDepKey: true,
							}, {Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								DynamicBlocks: true,
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "aws_instance"},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"instance_size": {
										IsOptional: true,
										Constraint: schema.LiteralType{Type: cty.String},
									},
								},
							},
						},
					},
				},
			},
			`resource "aws_instance" "example" {
	name = "example"
	
}`,
			hcl.Pos{Line: 3, Column: 3, Byte: 55},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "instance_size",
					Detail: "optional, string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 55},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 55},
						},
						NewText: "instance_size",
						Snippet: `instance_size = "${1:value}"`,
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
			"",
		},
		{
			"dynamic block completion",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name:     "type",
								IsDepKey: true,
							}, {Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								DynamicBlocks: true,
							},
							Blocks:     make(map[string]*schema.BlockSchema, 0),
							Attributes: make(map[string]*schema.AttributeSchema, 0),
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "aws_instance"},
								},
							}): {
								Blocks: map[string]*schema.BlockSchema{
									"foo": {
										Body: schema.NewBodySchema(),
									},
								},
								Attributes: map[string]*schema.AttributeSchema{
									"instance_size": {
										IsOptional: true,
										Constraint: schema.LiteralType{Type: cty.String},
									},
								},
							},
						},
					},
				},
			},
			`resource "aws_instance" "example" {
	name = "example"
	
}`,
			hcl.Pos{Line: 3, Column: 3, Byte: 55},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label: "dynamic",
					Description: lang.MarkupContent{
						Value: "A dynamic block to produce blocks dynamically by iterating over a given complex value",
						Kind:  lang.MarkdownKind,
					},
					Detail:         "Block",
					Kind:           lang.BlockCandidateKind,
					TriggerSuggest: true,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 55},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 55},
						},
						NewText: "dynamic",
						Snippet: "dynamic \"${1}\" {\n  ${2}\n}",
					},
				},
				{
					Label:  "foo",
					Detail: "Block",
					Kind:   lang.BlockCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 55},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 55},
						},
						NewText: "foo",
						Snippet: "foo {\n  ${1}\n}",
					},
				},
				{
					Label:  "instance_size",
					Detail: "optional, string",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 55},
							End:      hcl.Pos{Line: 3, Column: 3, Byte: 55},
						},
						NewText: "instance_size",
						Snippet: `instance_size = "${1:value}"`,
					},
				},
			}),
			"",
		},
		{
			"dynamic block inner completion",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name:     "type",
								IsDepKey: true,
							}, {Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								DynamicBlocks: true,
							},
							Blocks:     make(map[string]*schema.BlockSchema, 0),
							Attributes: make(map[string]*schema.AttributeSchema, 0),
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "aws_instance"},
								},
							}): {
								Blocks: map[string]*schema.BlockSchema{
									"foo": {
										Body: schema.NewBodySchema(),
									},
								},
								Attributes: map[string]*schema.AttributeSchema{
									"instance_size": {
										IsOptional: true,
										Constraint: schema.LiteralType{Type: cty.String},
									},
								},
							},
						},
					},
				},
			},
			`resource "aws_instance" "example" {
	name = "example"
	dynamic "foo" {
		
	}
}`,
			hcl.Pos{Line: 4, Column: 5, Byte: 73},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label: "content",
					Description: lang.MarkupContent{
						Value: "The body of each generated block",
						Kind:  lang.PlainTextKind,
					},
					Detail: "Block, min: 1, max: 1",
					Kind:   lang.BlockCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 4, Column: 5, Byte: 73},
							End:      hcl.Pos{Line: 4, Column: 5, Byte: 73},
						},
						NewText: "content",
						Snippet: "content {\n  ${1}\n}",
					},
				},
				{
					Label: "for_each",
					Description: lang.MarkupContent{
						Value: "A meta-argument that accepts a list, map or a set of strings, and creates an instance for each item in that list, map or set.",
						Kind:  lang.MarkdownKind,
					},
					Detail:         "required, map of any single type or list of any single type or set of string",
					Kind:           lang.AttributeCandidateKind,
					TriggerSuggest: true,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 4, Column: 5, Byte: 73},
							End:      hcl.Pos{Line: 4, Column: 5, Byte: 73},
						},
						NewText: "for_each",
						Snippet: "for_each = ",
					},
				},
				{
					Label: "iterator",
					Description: lang.MarkupContent{
						Value: "The name of a temporary variable that represents the current element of the complex value. Defaults to the label of the dynamic block.",
						Kind:  lang.MarkdownKind,
					},
					Detail: "optional, string",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 4, Column: 5, Byte: 73},
							End:      hcl.Pos{Line: 4, Column: 5, Byte: 73},
						},
						NewText: "iterator",
						Snippet: `iterator = "${1:value}"`,
					},
				},
				{
					Label: "labels",
					Description: lang.MarkupContent{
						Value: "A list of strings that specifies the block labels, in order, to use for each generated block.",
						Kind:  lang.MarkdownKind,
					},
					Detail: "optional, list of string",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 4, Column: 5, Byte: 73},
							End:      hcl.Pos{Line: 4, Column: 5, Byte: 73},
						},
						NewText: "labels",
						Snippet: "labels = ",
					},
					TriggerSuggest: true,
				},
			}),
			"",
		},
		{
			"dynamic block content attribute completion",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name:     "type",
								IsDepKey: true,
							}, {Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								DynamicBlocks: true,
							},
							Blocks:     make(map[string]*schema.BlockSchema, 0),
							Attributes: make(map[string]*schema.AttributeSchema, 0),
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "aws_instance"},
								},
							}): {
								Blocks: map[string]*schema.BlockSchema{
									"foo": {
										Body: &schema.BodySchema{
											Attributes: map[string]*schema.AttributeSchema{
												"thing": {
													IsOptional: true,
													Constraint: schema.LiteralType{Type: cty.String},
												},
											},
										},
									},
								},
								Attributes: map[string]*schema.AttributeSchema{
									"instance_size": {
										IsOptional: true,
										Constraint: schema.LiteralType{Type: cty.String},
									},
								},
							},
						},
					},
				},
			},
			`resource "aws_instance" "example" {
	name = "example"
	dynamic "foo" {
		content {
			
		}
	}
}`,
			hcl.Pos{Line: 5, Column: 7, Byte: 86},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "thing",
					Detail: "optional, string",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 5, Column: 7, Byte: 86},
							End:      hcl.Pos{Line: 5, Column: 7, Byte: 86},
						},
						NewText: "thing",
						Snippet: `thing = "${1:value}"`,
					},
				},
			}),
			"",
		},
		{
			"dynamic block label only completes dependent blocks",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name:        "type",
								IsDepKey:    true,
								Completable: true,
							},
							{Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								DynamicBlocks: true,
							},
							Blocks: map[string]*schema.BlockSchema{
								"lifecycle": {
									Body: schema.NewBodySchema(),
								},
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "aws_instance"},
								},
							}): {
								Blocks: map[string]*schema.BlockSchema{
									"foo": {
										Body: schema.NewBodySchema(),
									},
								},
							},
						},
					},
				},
			},
			`resource "aws_instance" "example" {
  name = "example"
  dynamic "" {
    
  }
}`,
			hcl.Pos{Line: 3, Column: 12, Byte: 66},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label: "foo",
					Kind:  lang.LabelCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 12, Byte: 66},
							End:      hcl.Pos{Line: 3, Column: 12, Byte: 66},
						},
						NewText: "foo",
						Snippet: "foo",
					},
				},
			}),
			"",
		},
		{
			"dynamic block label never completes static blocks",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name:        "type",
								IsDepKey:    true,
								Completable: true,
							},
							{Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								DynamicBlocks: true,
							},
							Blocks: map[string]*schema.BlockSchema{
								"lifecycle": {
									Body: schema.NewBodySchema(),
								},
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "aws_instance"},
								},
							}): {
								Blocks: make(map[string]*schema.BlockSchema, 0),
							},
						},
					},
				},
			},
			`resource "aws_instance" "example" {
  name = "example"
  dynamic "" {
    
  }
}`,
			hcl.Pos{Line: 3, Column: 12, Byte: 66},
			lang.CompleteCandidates([]lang.Candidate{}),
			`test.tf (3,12): unknown block type "dynamic"`,
		},
		// completion nesting should work
		{
			"dynamic block completion nesting should work",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type", IsDepKey: true}, {Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								DynamicBlocks: true,
							},
							Blocks:     make(map[string]*schema.BlockSchema, 0),
							Attributes: make(map[string]*schema.AttributeSchema, 0),
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "aws_instance"},
								},
							}): {
								Blocks: map[string]*schema.BlockSchema{
									"foo": {
										Body: &schema.BodySchema{
											Blocks: map[string]*schema.BlockSchema{
												"bar": {
													Body: schema.NewBodySchema(),
												},
											},
											Attributes: map[string]*schema.AttributeSchema{
												"thing": {
													IsOptional: true,
													Constraint: schema.LiteralType{Type: cty.String},
												},
											},
										},
									},
								},
								Attributes: map[string]*schema.AttributeSchema{
									"instance_size": {
										IsOptional: true,
										Constraint: schema.LiteralType{Type: cty.String},
									},
								},
							},
						},
					},
				},
			},
			`resource "aws_instance" "example" {
	name = "example"
	dynamic "foo" {
		content {
			
		}
	}
}`,
			hcl.Pos{Line: 5, Column: 7, Byte: 86},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "bar",
					Detail: "Block",
					Kind:   lang.BlockCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 5, Column: 7, Byte: 86},
							End:      hcl.Pos{Line: 5, Column: 7, Byte: 86},
						},
						NewText: "bar",
						Snippet: "bar {\n  ${1}\n}",
					},
				},
				{
					Label: "dynamic",
					Description: lang.MarkupContent{
						Value: "A dynamic block to produce blocks dynamically by iterating over a given complex value",
						Kind:  lang.MarkdownKind,
					},
					Detail:         "Block",
					Kind:           lang.BlockCandidateKind,
					TriggerSuggest: true,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 5, Column: 7, Byte: 86},
							End:      hcl.Pos{Line: 5, Column: 7, Byte: 86},
						},
						NewText: "dynamic",
						Snippet: "dynamic \"${1}\" {\n  ${2}\n}",
					},
				},
				{
					Label:  "thing",
					Detail: "optional, string",
					Kind:   lang.AttributeCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 5, Column: 7, Byte: 86},
							End:      hcl.Pos{Line: 5, Column: 7, Byte: 86},
						},
						NewText: "thing",
						Snippet: `thing = "${1:value}"`,
					},
				},
			}),
			"",
		},
		{
			"dynamic block completion after attribute equal sign",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name:     "type",
								IsDepKey: true,
							}, {Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								DynamicBlocks: true,
							},
							Blocks:     make(map[string]*schema.BlockSchema, 0),
							Attributes: make(map[string]*schema.AttributeSchema, 0),
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "aws_instance"},
								},
							}): {
								Blocks: map[string]*schema.BlockSchema{
									"foo": {
										Body: &schema.BodySchema{
											Attributes: map[string]*schema.AttributeSchema{
												"thing": {
													IsOptional: true,
													Constraint: schema.LiteralType{Type: cty.Bool},
												},
											},
										},
									},
								},
								Attributes: map[string]*schema.AttributeSchema{
									"instance_size": {
										IsOptional: true,
										Constraint: schema.LiteralType{Type: cty.String},
									},
								},
							},
						},
					},
				},
			},
			`resource "aws_instance" "example" {
	name = "example"
	dynamic "foo" {
		content {
			thing = 
		}
	}
}`,
			hcl.Pos{Line: 5, Column: 15, Byte: 94},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "false",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 5, Column: 15, Byte: 94},
							End:      hcl.Pos{Line: 5, Column: 15, Byte: 94},
						},
						NewText: "false",
						Snippet: `false`,
					},
				},
				{
					Label:  "true",
					Detail: "bool",
					Kind:   lang.BoolCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 5, Column: 15, Byte: 94},
							End:      hcl.Pos{Line: 5, Column: 15, Byte: 94},
						},
						NewText: "true",
						Snippet: `true`,
					},
				},
			}),
			"",
		},
		// check allows more than one dynamic
		{
			"allows more than one dynamic",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type", IsDepKey: true}, {Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								DynamicBlocks: true,
							},
							Blocks:     make(map[string]*schema.BlockSchema, 0),
							Attributes: make(map[string]*schema.AttributeSchema, 0),
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "aws_instance"},
								},
							}): {
								Blocks: map[string]*schema.BlockSchema{
									"foo": {
										Body: schema.NewBodySchema(),
									},
								},
							},
						},
					},
				},
			},
			`resource "aws_instance" "example" {
	name = "example"
	dynamic "foo" {
		
	}
	
}`,
			hcl.Pos{Line: 6, Column: 3, Byte: 78},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label: "dynamic",
					Description: lang.MarkupContent{
						Value: "A dynamic block to produce blocks dynamically by iterating over a given complex value",
						Kind:  lang.MarkdownKind,
					},
					Detail:         "Block",
					Kind:           lang.BlockCandidateKind,
					TriggerSuggest: true,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 6, Column: 3, Byte: 78},
							End:      hcl.Pos{Line: 6, Column: 3, Byte: 78},
						},
						NewText: "dynamic",
						Snippet: "dynamic \"${1}\" {\n  ${2}\n}",
					},
				},
				{
					Label:  "foo",
					Detail: "Block",
					Kind:   lang.BlockCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 6, Column: 3, Byte: 78},
							End:      hcl.Pos{Line: 6, Column: 3, Byte: 78},
						},
						NewText: "foo",
						Snippet: "foo {\n  ${1}\n}",
					},
				},
			}),
			"",
		},
		// allows dynamic blocks in blocks
		{
			"allows dynamic blocks in blocks",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type", IsDepKey: true}, {Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								DynamicBlocks: true,
							},
							Blocks:     make(map[string]*schema.BlockSchema, 0),
							Attributes: make(map[string]*schema.AttributeSchema, 0),
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "aws_instance"},
								},
							}): {
								Blocks: map[string]*schema.BlockSchema{
									"foo": {
										Body: &schema.BodySchema{
											Blocks: map[string]*schema.BlockSchema{
												"bar": {
													Body: schema.NewBodySchema(),
												},
											},
										},
									},
								},
								Attributes: map[string]*schema.AttributeSchema{
									"instance_size": {
										IsOptional: true,
										Constraint: schema.LiteralType{Type: cty.String},
									},
								},
							},
						},
					},
				},
			},
			`resource "aws_instance" "example" {
	name = "example"
	foo {
		
	}	
}`,
			hcl.Pos{Line: 4, Column: 5, Byte: 63},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "bar",
					Detail: "Block",
					Kind:   lang.BlockCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 4, Column: 5, Byte: 63},
							End:      hcl.Pos{Line: 4, Column: 5, Byte: 63},
						},
						NewText: "bar",
						Snippet: "bar {\n  ${1}\n}",
					},
				},
				{
					Label: "dynamic",
					Description: lang.MarkupContent{
						Value: "A dynamic block to produce blocks dynamically by iterating over a given complex value",
						Kind:  lang.MarkdownKind,
					},
					Detail:         "Block",
					Kind:           lang.BlockCandidateKind,
					TriggerSuggest: true,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 4, Column: 5, Byte: 63},
							End:      hcl.Pos{Line: 4, Column: 5, Byte: 63},
						},
						NewText: "dynamic",
						Snippet: "dynamic \"${1}\" {\n  ${2}\n}",
					},
				},
			}),
			"",
		},
		{
			"never complete dynamic as a dynamic label",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name:        "type",
								IsDepKey:    true,
								Completable: true,
							},
							{Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								DynamicBlocks: true,
							},
							Blocks: make(map[string]*schema.BlockSchema, 0),
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "aws_instance"},
								},
							}): {
								Blocks: map[string]*schema.BlockSchema{
									"foo": {
										Body: &schema.BodySchema{
											Blocks: map[string]*schema.BlockSchema{
												"bar": {
													Body: schema.NewBodySchema(),
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
			`resource "aws_instance" "example" {
  foo {
    dynamic "" {
      
    }
  }
}`,
			hcl.Pos{Line: 3, Column: 14, Byte: 57},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label: "bar",
					Kind:  lang.LabelCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 14, Byte: 57},
							End:      hcl.Pos{Line: 3, Column: 14, Byte: 57},
						},
						NewText: "bar",
						Snippet: "bar",
					},
				},
			}),
			"",
		},
		{
			"deeper nesting support",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{
								Name:        "type",
								IsDepKey:    true,
								Completable: true,
							},
							{Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								DynamicBlocks: true,
							},
							Blocks: make(map[string]*schema.BlockSchema, 0),
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "aws_instance"},
								},
							}): {
								Blocks: map[string]*schema.BlockSchema{
									"foo": {
										Body: &schema.BodySchema{
											Blocks: map[string]*schema.BlockSchema{
												"bar": {
													Body: &schema.BodySchema{
														Blocks: map[string]*schema.BlockSchema{
															"baz": {
																Body: schema.NewBodySchema(),
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
					},
				},
			},
			`resource "aws_instance" "example" {
  foo {
    bar {
      
    }
  }
}`,
			hcl.Pos{Line: 4, Column: 7, Byte: 60},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "baz",
					Detail: "Block",
					Kind:   lang.BlockCandidateKind,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 4, Column: 7, Byte: 60},
							End:      hcl.Pos{Line: 4, Column: 7, Byte: 60},
						},
						NewText: "baz",
						Snippet: "baz {\n  ${1}\n}",
					},
				},
				{
					Label: "dynamic",
					Description: lang.MarkupContent{
						Value: "A dynamic block to produce blocks dynamically by iterating over a given complex value",
						Kind:  lang.MarkdownKind,
					},
					Detail:         "Block",
					Kind:           lang.BlockCandidateKind,
					TriggerSuggest: true,
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 4, Column: 7, Byte: 60},
							End:      hcl.Pos{Line: 4, Column: 7, Byte: 60},
						},
						NewText: "dynamic",
						Snippet: "dynamic \"${1}\" {\n  ${2}\n}",
					},
				},
			}),
			"",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)

			d := testPathDecoder(t, &PathContext{
				Schema: tc.bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			})

			candidates, err := d.CompletionAtPos(ctx, "test.tf", tc.pos)

			if err != nil {
				if tc.expectedErr != "" && err.Error() != tc.expectedErr {
					t.Fatalf("unexpected error: %q\nexpected: %q\n",
						err, tc.expectedErr)
				} else if tc.expectedErr == "" {
					t.Fatal(err)
				}
			} else if tc.expectedErr != "" {
				t.Fatalf("expected error: %q", tc.expectedErr)
			}

			if diff := cmp.Diff(tc.expectedCandidates, candidates); diff != "" {
				t.Fatalf("unexpected candidates: %s", diff)
			}
		})
	}
}
