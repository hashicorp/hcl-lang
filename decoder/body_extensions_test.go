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

func TestCompletionAtPos_BodySchema_Extensions(t *testing.T) {
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
						Value: "The distinct index number (starting with 0) corresponding to the instance",
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
										Expr:       schema.LiteralTypeOnly(cty.String),
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
						Value: "The distinct index number (starting with 0) corresponding to the instance",
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
									Expr: schema.ExprConstraints{
										schema.TraversalExpr{
											OfType: cty.Number,
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
									Expr: schema.ExprConstraints{
										schema.TraversalExpr{
											OfType: cty.Number,
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
			reference.Targets{},
			`resource "aws_instance" "foo" {
	count = 4
	cpu_count =
}`,
			hcl.Pos{
				Line:   3,
				Column: 14,
				Byte:   55,
			},
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
							Start: hcl.Pos{
								Line:   3,
								Column: 13,
								Byte:   55,
							},
							End: hcl.Pos{
								Line:   3,
								Column: 13,
								Byte:   55,
							},
						},
						NewText: "count.index",
						Snippet: "count.index",
					},
					Kind: lang.TraversalCandidateKind,
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
									Expr: schema.ExprConstraints{
										schema.TraversalExpr{
											OfType: cty.Number,
										},
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
												Expr: schema.ExprConstraints{
													schema.TraversalExpr{
														OfType: cty.Number,
													},
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
			reference.Targets{},
			`resource "aws_instance" "foo" {
  count = 4
  foo {
    cpu_count =
  }
}`,
			hcl.Pos{
				Line:   4,
				Column: 17,
				Byte:   67,
			},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label: "count.index",
					Description: lang.MarkupContent{
						Value: "The distinct index number (starting with 0) corresponding to the instance",
						Kind:  lang.PlainTextKind,
					},
					Detail: "number",
					TextEdit: lang.TextEdit{Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   4,
							Column: 16,
							Byte:   67,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 16,
							Byte:   67,
						},
					}, NewText: "count.index", Snippet: "count.index"},
					Kind: lang.TraversalCandidateKind,
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
					Kind: lang.TraversalCandidateKind,
				},
			}),
		},
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
				Detail: "optional, map of any single type or set of string",
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
									Expr: schema.ExprConstraints{
										schema.TraversalExpr{
											OfType: cty.String,
										},
									},
								},
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
thing = 
}`,
			hcl.Pos{Line: 6, Column: 8, Byte: 101},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "each.key",
					Detail: "string",
					Kind:   lang.TraversalCandidateKind,
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
					Detail: "any type",
					Kind:   lang.TraversalCandidateKind,
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
									Expr: schema.ExprConstraints{
										schema.TraversalExpr{
											OfType: cty.String,
										},
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
									Expr: schema.ExprConstraints{
										schema.TraversalExpr{
											OfType: cty.Number,
										},
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
									Expr: schema.ExprConstraints{
										schema.TraversalExpr{
											OfType: cty.Number,
										},
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
												Expr: schema.ExprConstraints{
													schema.TraversalExpr{
														OfType: cty.Number,
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
			reference.Targets{},
			`resource "aws_instance" "foo" {
for_each = {
	a_group = "eastus"
	another_group = "westus2"
}
foo {
	thing = 
}
}`,
			hcl.Pos{Line: 7, Column: 11, Byte: 109},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "each.key",
					Detail: "string",
					Kind:   lang.TraversalCandidateKind,
					Description: lang.MarkupContent{
						Value: "The map key (or set member) corresponding to this instance",
						Kind:  lang.MarkdownKind,
					},
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 7, Column: 10, Byte: 109},
							End:      hcl.Pos{Line: 7, Column: 10, Byte: 109},
						},
						NewText: "each.key",
						Snippet: "each.key",
					},
				},
				{
					Label:  "each.value",
					Detail: "any type",
					Kind:   lang.TraversalCandidateKind,
					Description: lang.MarkupContent{
						Value: "The map value corresponding to this instance. (If a set was provided, this is the same as `each.key`.)",
						Kind:  lang.MarkdownKind,
					},
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 7, Column: 10, Byte: 109},
							End:      hcl.Pos{Line: 7, Column: 10, Byte: 109},
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
									Expr: schema.ExprConstraints{
										schema.TraversalExpr{
											OfType: cty.String,
										},
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
					Detail: "map of any single type",
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

			candidates, err := d.CandidatesAtPos(ctx, "test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedCandidates, candidates); diff != "" {
				t.Fatalf("unexpected candidates: %s", diff)
			}
		})
	}
}
