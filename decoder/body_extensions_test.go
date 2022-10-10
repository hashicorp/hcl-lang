package decoder

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func TestCompletionAtPos_BodySchema_Extensions(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		testName           string
		bodySchema         *schema.BodySchema
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
						Kind:  lang.PlainTextKind,
					},
					Detail: "optional, number",
					TextEdit: lang.TextEdit{Range: hcl.Range{
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
					}, NewText: "count", Snippet: "count = ${1:1}"},
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
