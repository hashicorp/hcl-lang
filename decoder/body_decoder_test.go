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

func TestDecoder_CandidateAtPos_incompleteAttributes(t *testing.T) {
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"customblock": {
				Labels: []*schema.LabelSchema{
					{Name: "type"},
				},
				Body: &schema.BodySchema{
					Attributes: map[string]*schema.AttributeSchema{
						"attr1":           {Expr: schema.LiteralTypeOnly(cty.Number)},
						"attr2":           {Expr: schema.LiteralTypeOnly(cty.Number)},
						"some_other_attr": {Expr: schema.LiteralTypeOnly(cty.Number)},
						"another_attr":    {Expr: schema.LiteralTypeOnly(cty.Number)},
					},
				},
			},
		},
	}

	d := NewDecoder()
	d.maxCandidates = 1

	d.SetSchema(bodySchema)

	f, _ := hclsyntax.ParseConfig([]byte(`customblock "label1" {
  attr
}
`), "test.tf", hcl.InitialPos)
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	candidates, err := d.CandidatesAtPos("test.tf", hcl.Pos{
		Line:   2,
		Column: 7,
		Byte:   29,
	})
	if err != nil {
		t.Fatal(err)
	}
	expectedCandidates := lang.Candidates{
		List: []lang.Candidate{
			{
				Label:  "attr1",
				Detail: "Optional, number",
				TextEdit: lang.TextEdit{
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   25,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 7,
							Byte:   29,
						},
					},
					NewText: "attr1",
					Snippet: "attr1 = ${1:1}",
				},
				Kind: lang.AttributeCandidateKind,
			},
		},
		IsComplete: false,
	}
	if diff := cmp.Diff(expectedCandidates, candidates); diff != "" {
		t.Fatalf("unexpected candidates: %s", diff)
	}
}

func TestDecoder_CandidateAtPos_computedAttributes(t *testing.T) {
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"customblock": {
				Labels: []*schema.LabelSchema{
					{Name: "type"},
				},
				Body: &schema.BodySchema{
					Attributes: map[string]*schema.AttributeSchema{
						"attr1":           {Expr: schema.LiteralTypeOnly(cty.Number), IsComputed: true},
						"attr2":           {Expr: schema.LiteralTypeOnly(cty.Number), IsComputed: true, IsOptional: true},
						"some_other_attr": {Expr: schema.LiteralTypeOnly(cty.Number)},
						"another_attr":    {Expr: schema.LiteralTypeOnly(cty.Number)},
					},
				},
			},
		},
	}

	d := NewDecoder()
	d.SetSchema(bodySchema)

	f, _ := hclsyntax.ParseConfig([]byte(`customblock "label1" {
  attr
}
`), "test.tf", hcl.InitialPos)
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	candidates, err := d.CandidatesAtPos("test.tf", hcl.Pos{
		Line:   2,
		Column: 7,
		Byte:   29,
	})
	if err != nil {
		t.Fatal(err)
	}
	expectedCandidates := lang.Candidates{
		List: []lang.Candidate{
			{
				Label:  "attr2",
				Detail: "Optional, number",
				TextEdit: lang.TextEdit{
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   25,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 7,
							Byte:   29,
						},
					},
					NewText: "attr2",
					Snippet: "attr2 = ${1:1}",
				},
				Kind: lang.AttributeCandidateKind,
			},
		},
		IsComplete: true,
	}
	if diff := cmp.Diff(expectedCandidates, candidates); diff != "" {
		t.Fatalf("unexpected candidates: %s", diff)
	}
}

func TestDecoder_CandidateAtPos_incompleteBlocks(t *testing.T) {
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"customblock": {
				Labels: []*schema.LabelSchema{
					{Name: "type"},
				},
				Body: &schema.BodySchema{
					Blocks: map[string]*schema.BlockSchema{
						"block1":           {MaxItems: 1},
						"block2":           {},
						"block3":           {},
						"some_other_block": {},
						"another_block":    {},
					},
				},
			},
		},
	}

	d := NewDecoder()
	d.maxCandidates = 1

	d.SetSchema(bodySchema)

	f, _ := hclsyntax.ParseConfig([]byte(`customblock "label1" {
  block1 {}
  block
}
`), "test.tf", hcl.InitialPos)
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	candidates, err := d.CandidatesAtPos("test.tf", hcl.Pos{
		Line:   3,
		Column: 8,
		Byte:   42,
	})
	if err != nil {
		t.Fatal(err)
	}
	expectedCandidates := lang.Candidates{
		List: []lang.Candidate{
			{
				Label:  "block2",
				Detail: "Block",
				TextEdit: lang.TextEdit{
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   3,
							Column: 3,
							Byte:   37,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 8,
							Byte:   42,
						},
					},
					NewText: "block2",
					Snippet: "block2 {\n  ${1}\n}",
				},
				Kind: lang.BlockCandidateKind,
			},
		},
		IsComplete: false,
	}
	if diff := cmp.Diff(expectedCandidates, candidates); diff != "" {
		t.Fatalf("unexpected candidates: %s", diff)
	}
}
