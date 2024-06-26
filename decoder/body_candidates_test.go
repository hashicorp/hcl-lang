// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestDecoder_CandidateAtPos_incompleteAttributes(t *testing.T) {
	ctx := context.Background()
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"customblock": {
				Labels: []*schema.LabelSchema{
					{Name: "type"},
				},
				Body: &schema.BodySchema{
					Attributes: map[string]*schema.AttributeSchema{
						"attr1":           {Constraint: schema.LiteralType{Type: cty.Number}},
						"attr2":           {Constraint: schema.LiteralType{Type: cty.Number}},
						"some_other_attr": {Constraint: schema.LiteralType{Type: cty.Number}},
						"another_attr":    {Constraint: schema.LiteralType{Type: cty.Number}},
					},
				},
			},
		},
	}

	f, _ := hclsyntax.ParseConfig([]byte(`customblock "label1" {
  attr
}
`), "test.tf", hcl.InitialPos)

	d := testPathDecoder(t, &PathContext{
		Schema: bodySchema,
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})
	d.maxCandidates = 1

	candidates, err := d.CompletionAtPos(ctx, "test.tf", hcl.Pos{
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
				Detail: "number",
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
					Snippet: "attr1 = ${1:0}",
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
	ctx := context.Background()
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"customblock": {
				Labels: []*schema.LabelSchema{
					{Name: "type"},
				},
				Body: &schema.BodySchema{
					Attributes: map[string]*schema.AttributeSchema{
						"attr1":           {Constraint: schema.LiteralType{Type: cty.Number}, IsComputed: true},
						"attr2":           {Constraint: schema.LiteralType{Type: cty.Number}, IsComputed: true, IsOptional: true},
						"some_other_attr": {Constraint: schema.LiteralType{Type: cty.Number}},
						"another_attr":    {Constraint: schema.LiteralType{Type: cty.Number}},
					},
				},
			},
		},
	}

	f, _ := hclsyntax.ParseConfig([]byte(`customblock "label1" {
  attr
}
`), "test.tf", hcl.InitialPos)
	d := testPathDecoder(t, &PathContext{
		Schema: bodySchema,
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})

	candidates, err := d.CompletionAtPos(ctx, "test.tf", hcl.Pos{
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
				Detail: "optional, number",
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
					Snippet: "attr2 = ${1:0}",
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
	ctx := context.Background()
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

	f, _ := hclsyntax.ParseConfig([]byte(`customblock "label1" {
  block1 {}
  block
}
`), "test.tf", hcl.InitialPos)
	d := testPathDecoder(t, &PathContext{
		Schema: bodySchema,
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})
	d.maxCandidates = 1

	candidates, err := d.CompletionAtPos(ctx, "test.tf", hcl.Pos{
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

func TestDecoder_CandidateAtPos_duplicateNames(t *testing.T) {
	ctx := context.Background()
	bodySchema := &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"ingress": {
				IsOptional: true,
				Constraint: schema.LiteralType{
					Type: cty.Object(map[string]cty.Type{
						"attr1": cty.String,
						"attr2": cty.Number,
					}),
				},
			},
		},
		Blocks: map[string]*schema.BlockSchema{
			"ingress": {
				Body: &schema.BodySchema{
					Attributes: map[string]*schema.AttributeSchema{
						"attr1": {Constraint: schema.LiteralType{Type: cty.String}, IsRequired: true},
						"attr2": {Constraint: schema.LiteralType{Type: cty.Number}, IsRequired: true},
					},
				},
			},
		},
	}

	f, _ := hclsyntax.ParseConfig([]byte("\n"), "test.tf", hcl.InitialPos)

	d := testPathDecoder(t, &PathContext{
		Schema: bodySchema,
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})
	d.PrefillRequiredFields = true

	candidates, err := d.CompletionAtPos(ctx, "test.tf", hcl.InitialPos)
	if err != nil {
		t.Fatal(err)
	}
	expectedCandidates := lang.Candidates{
		List: []lang.Candidate{
			{
				Label:  "ingress",
				Detail: "optional, object",
				TextEdit: lang.TextEdit{
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End:      hcl.InitialPos,
					},
					NewText: "ingress",
					Snippet: `ingress = {
  attr1 = "${1:value}"
  attr2 = ${2:0}
}`,
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
