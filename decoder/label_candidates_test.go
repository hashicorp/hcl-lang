package decoder

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestDecoder_CandidateAtPos_incompleteLabels(t *testing.T) {
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"customblock": {
				Labels: []*schema.LabelSchema{
					{
						Name:        "type",
						IsDepKey:    true,
						Completable: true,
					},
				},
				DependentBody: map[schema.SchemaKey]*schema.BodySchema{
					schema.NewSchemaKey(schema.DependencyKeys{
						Labels: []schema.LabelDependent{
							{
								Index: 0,
								Value: "first",
							},
						},
					}): {
						Attributes: map[string]*schema.AttributeSchema{
							"attr1": {Expr: schema.LiteralTypeOnly(cty.Number)},
						},
					},
					schema.NewSchemaKey(schema.DependencyKeys{
						Labels: []schema.LabelDependent{
							{
								Index: 0,
								Value: "second",
							},
						},
					}): {
						Attributes: map[string]*schema.AttributeSchema{
							"attr2": {Expr: schema.LiteralTypeOnly(cty.Number)},
						},
					},
				},
			},
		},
	}

	d := NewDecoder()
	d.maxCandidates = 1
	d.SetSchema(bodySchema)

	f, _ := hclsyntax.ParseConfig([]byte(`customblock "" {
}
`), "test.tf", hcl.InitialPos)
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	candidates, err := d.CandidatesAtPos("test.tf", hcl.Pos{
		Line:   1,
		Column: 14,
		Byte:   13,
	})
	if err != nil {
		t.Fatal(err)
	}
	expectedCandidates := lang.Candidates{
		List: []lang.Candidate{
			{
				Label: "first",
				TextEdit: lang.TextEdit{
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 14,
							Byte:   13,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 14,
							Byte:   13,
						},
					},
					NewText: "first",
					Snippet: "first",
				},
				Kind: lang.LabelCandidateKind,
			},
		},
		IsComplete: false,
	}
	if diff := cmp.Diff(expectedCandidates, candidates); diff != "" {
		t.Fatalf("unexpected candidates: %s", diff)
	}
}
