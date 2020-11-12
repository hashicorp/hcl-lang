package decoder

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestDecoder_CandidatesAtPos_noSchema(t *testing.T) {
	d := NewDecoder()
	f, pDiags := hclsyntax.ParseConfig(testConfig, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.CandidatesAtPos("test.tf", hcl.InitialPos)
	noSchemaErr := &NoSchemaError{}
	if !errors.As(err, &noSchemaErr) {
		t.Fatal("expected NoSchemaError for no schema")
	}
}

func TestDecoder_CandidateAtPos_emptyBody(t *testing.T) {
	d := NewDecoder()
	f := &hcl.File{
		Body: hcl.EmptyBody(),
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.CandidatesAtPos("test.tf", hcl.InitialPos)
	unknownFormatErr := &UnknownFileFormatError{}
	if !errors.As(err, &unknownFormatErr) {
		t.Fatal("expected UnknownFileFormatError for empty body")
	}
}

func TestDecoder_CandidatesAtPos_unknownBlock(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "type"},
		{Name: "name"},
	}
	resourceSchema := &schema.BlockSchema{
		Labels: resourceLabelSchema,
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"count": {ValueType: cty.Number},
			},
		},
	}
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"resource": resourceSchema,
		},
	}

	d := NewDecoder()
	d.SetSchema(bodySchema)
	f, pDiags := hclsyntax.ParseConfig([]byte(`customblock "label1" {

}
`), "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.CandidatesAtPos("test.tf", hcl.Pos{
		Line:   2,
		Column: 1,
		Byte:   23,
	})
	if err == nil {
		t.Fatal("expected error for unknown block")
	}
	if !strings.Contains(err.Error(), "unknown block type") {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
}

func TestDecoder_CandidatesAtPos_prefixNearEOF(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "type"},
		{Name: "name"},
	}
	resourceSchema := &schema.BlockSchema{
		Labels: resourceLabelSchema,
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"count": {ValueType: cty.Number},
			},
		},
	}
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"resource": resourceSchema,
		},
	}

	d := NewDecoder()
	d.SetSchema(bodySchema)
	f, _ := hclsyntax.ParseConfig([]byte(`res`), "test.tf", hcl.InitialPos)

	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	candidates, err := d.CandidatesAtPos("test.tf", hcl.Pos{
		Line:   1,
		Column: 4,
		Byte:   3,
	})
	if err != nil {
		t.Fatal(err)
	}
	expectedCandidates := lang.CompleteCandidates([]lang.Candidate{
		{
			Label:  "resource",
			Detail: "Block",
			TextEdit: lang.TextEdit{
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 1,
						Byte:   0,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 4,
						Byte:   3,
					},
				},
				NewText: "resource",
				Snippet: "resource \"${1:type}\" \"${2:name}\" {\n  ${3}\n}",
			},
			Kind: lang.BlockCandidateKind,
		},
	})
	if diff := cmp.Diff(expectedCandidates, candidates, ctydebug.CmpOptions); diff != "" {
		t.Fatalf("candidates mismatch: %s", diff)
	}
}

func TestDecoder_CandidatesAtPos_invalidBlockPositions(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "type"},
	}
	blockSchema := &schema.BlockSchema{
		Labels: resourceLabelSchema,
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"num_attr": {ValueType: cty.Number},
			},
		},
	}
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"myblock": blockSchema,
		},
	}
	testConfig := []byte(`myblock "foo" {

}
`)
	testCases := []struct {
		name string
		pos  hcl.Pos
	}{
		{
			"whitespace in header",
			hcl.Pos{
				Line:   1,
				Column: 14,
				Byte:   13,
			},
		},
		{
			"opening brace",
			hcl.Pos{
				Line:   1,
				Column: 15,
				Byte:   14,
			},
		},
		{
			"closing brace",
			hcl.Pos{
				Line:   3,
				Column: 1,
				Byte:   17,
			},
		},
	}

	d := NewDecoder()
	d.SetSchema(bodySchema)
	f, pDiags := hclsyntax.ParseConfig(testConfig, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			_, err = d.CandidatesAtPos("test.tf", tc.pos)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), `position outside of "myblock" body`) {
				t.Fatalf("unexpected error message: %q", err.Error())
			}
		})
	}
}

func TestDecoder_CandidatesAtPos_rightHandSide(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "type"},
	}
	blockSchema := &schema.BlockSchema{
		Labels: resourceLabelSchema,
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"num_attr": {ValueType: cty.Number},
				"str_attr": {ValueType: cty.String},
			},
		},
	}
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"myblock": blockSchema,
		},
	}
	testConfig := []byte(`myblock "foo" {
  num_attr = 
}
`)

	d := NewDecoder()
	d.SetSchema(bodySchema)
	f, _ := hclsyntax.ParseConfig(testConfig, "test.tf", hcl.InitialPos)
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.CandidatesAtPos("test.tf", hcl.Pos{
		Line:   2,
		Column: 13,
		Byte:   28,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no candidates for attribute value") {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
}

func TestDecoder_CandidatesAtPos_rightHandSideInString(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "type"},
	}
	blockSchema := &schema.BlockSchema{
		Labels: resourceLabelSchema,
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"num_attr": {ValueType: cty.Number},
				"str_attr": {ValueType: cty.String},
			},
		},
	}
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"myblock": blockSchema,
		},
	}
	testConfig := []byte(`myblock "foo" {
  str_attr = ""
}
`)

	d := NewDecoder()
	d.SetSchema(bodySchema)
	f, _ := hclsyntax.ParseConfig(testConfig, "test.tf", hcl.InitialPos)
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.CandidatesAtPos("test.tf", hcl.Pos{
		Line:   2,
		Column: 15,
		Byte:   30,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no candidates for attribute value") {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
}

func TestDecoder_CandidatesAtPos_endOfLabel(t *testing.T) {
	blockSchema := &schema.BlockSchema{
		Labels: []*schema.LabelSchema{
			{Name: "type"},
		},
		DependentBody: map[schema.SchemaKey]*schema.BodySchema{
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "myfirst"},
				},
			}): {
				Attributes: map[string]*schema.AttributeSchema{
					"one":   {ValueType: cty.String},
					"two":   {ValueType: cty.Number},
					"three": {ValueType: cty.Bool},
				},
			},
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "mysecond"},
				},
			}): {
				Attributes: map[string]*schema.AttributeSchema{
					"four": {ValueType: cty.Number},
					"five": {ValueType: cty.DynamicPseudoType},
				},
			},
		},
	}
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"myblock": blockSchema,
		},
	}
	testConfig := []byte(`myblock "my" {
}
`)

	d := NewDecoder()
	d.SetSchema(bodySchema)
	f, _ := hclsyntax.ParseConfig(testConfig, "test.tf", hcl.InitialPos)
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	candidates, err := d.CandidatesAtPos("test.tf", hcl.Pos{
		Line:   1,
		Column: 12,
		Byte:   11,
	})
	if err != nil {
		t.Fatal(err)
	}
	expectedCandidates := lang.CompleteCandidates([]lang.Candidate{
		{
			Label: "myfirst",
			TextEdit: lang.TextEdit{
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 10,
						Byte:   9,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 12,
						Byte:   11,
					},
				},
				NewText: "myfirst",
				Snippet: "myfirst",
			},
			Kind: lang.LabelCandidateKind,
		},
		{
			Label: "mysecond",
			TextEdit: lang.TextEdit{
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 10,
						Byte:   9,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 12,
						Byte:   11,
					},
				},
				NewText: "mysecond",
				Snippet: "mysecond",
			},
			Kind: lang.LabelCandidateKind,
		},
	})
	if diff := cmp.Diff(expectedCandidates, candidates); diff != "" {
		t.Fatalf("unexpected candidates: %s", diff)
	}
}

func TestDecoder_CandidatesAtPos_zeroByteContent(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "type", IsDepKey: true},
		{Name: "name"},
	}
	resourceSchema := &schema.BlockSchema{
		Labels: resourceLabelSchema,
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"count": {ValueType: cty.Number},
			},
		},
	}
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"resource": resourceSchema,
		},
	}

	d := NewDecoder()
	d.SetSchema(bodySchema)
	f, pDiags := hclsyntax.ParseConfig([]byte{}, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	candidates, err := d.CandidatesAtPos("test.tf", hcl.InitialPos)
	if err != nil {
		t.Fatal(err)
	}
	expectedCandidates := lang.CompleteCandidates([]lang.Candidate{
		{
			Label:  "resource",
			Detail: "Block",
			TextEdit: lang.TextEdit{
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.InitialPos,
					End:      hcl.InitialPos,
				},
				NewText: "resource",
				Snippet: "resource \"${1}\" \"${2:name}\" {\n  ${3}\n}",
			},
			Kind: 2,
		},
	})
	if diff := cmp.Diff(expectedCandidates, candidates); diff != "" {
		t.Fatalf("unexpected candidates: %s", diff)
	}
}

func TestDecoder_CandidatesAtPos_endOfFilePos(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "type", IsDepKey: true},
		{Name: "name"},
	}
	resourceSchema := &schema.BlockSchema{
		Labels: resourceLabelSchema,
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"count": {ValueType: cty.Number},
			},
		},
	}
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"resource": resourceSchema,
		},
	}

	cfg := []byte(`resource "azurerm_subnet" "example" {
  count = 3
}
`)

	d := NewDecoder()
	d.SetSchema(bodySchema)
	f, pDiags := hclsyntax.ParseConfig([]byte(cfg), "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	candidates, err := d.CandidatesAtPos("test.tf", hcl.Pos{Line: 4, Column: 1, Byte: 52})
	if err != nil {
		t.Fatal(err)
	}
	expectedCandidates := lang.Candidates{
		List: []lang.Candidate{
			{
				Label:  "resource",
				Detail: "Block",
				TextEdit: lang.TextEdit{
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 4, Column: 1, Byte: 52},
						End:      hcl.Pos{Line: 4, Column: 1, Byte: 52},
					},
					NewText: "resource",
					Snippet: "resource \"${1}\" \"${2:name}\" {\n  ${3}\n}",
				},
				Kind: lang.BlockCandidateKind,
			},
		},
		IsComplete: true,
	}
	if diff := cmp.Diff(expectedCandidates, candidates); diff != "" {
		t.Fatalf("unexpected candidates: %s", diff)
	}
}

func TestDecoder_CandidatesAtPos_emptyLabel(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "type", IsDepKey: true},
		{Name: "name"},
	}
	resourceSchema := &schema.BlockSchema{
		Labels: resourceLabelSchema,
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"count": {ValueType: cty.Number},
			},
		},
		DependentBody: map[schema.SchemaKey]*schema.BodySchema{
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "azurerm_subnet"},
				},
			}): {
				Attributes: map[string]*schema.AttributeSchema{
					"one":   {ValueType: cty.String, IsRequired: true},
					"two":   {ValueType: cty.Number},
					"three": {ValueType: cty.Bool},
				},
			},
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "random_resource"},
				},
			}): {
				Attributes: map[string]*schema.AttributeSchema{
					"four": {ValueType: cty.Number},
					"five": {ValueType: cty.DynamicPseudoType},
				},
			},
		},
	}
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"resource": resourceSchema,
		},
	}

	cfg := []byte(`resource "" "" {
}
`)

	d := NewDecoder()
	d.SetSchema(bodySchema)
	f, pDiags := hclsyntax.ParseConfig([]byte(cfg), "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	candidates, err := d.CandidatesAtPos("test.tf", hcl.Pos{Line: 1, Column: 11, Byte: 10})
	if err != nil {
		t.Fatal(err)
	}
	expectedCandidates := lang.CompleteCandidates([]lang.Candidate{
		{
			Label: "azurerm_subnet",
			TextEdit: lang.TextEdit{
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
					End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
				},
				NewText: "azurerm_subnet",
				Snippet: "azurerm_subnet",
			},
			Kind: lang.LabelCandidateKind,
		},
		{
			Label: "random_resource",
			TextEdit: lang.TextEdit{
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
					End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
				},
				NewText: "random_resource",
				Snippet: "random_resource",
			},
			Kind: lang.LabelCandidateKind,
		},
	})
	if diff := cmp.Diff(expectedCandidates, candidates); diff != "" {
		t.Fatalf("unexpected candidates: %s", diff)
	}
}

func TestDecoder_CandidatesAtPos_basic(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "type", IsDepKey: true},
		{Name: "name"},
	}

	resourceSchema := &schema.BlockSchema{
		Labels: resourceLabelSchema,
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"count": {ValueType: cty.Number},
			},
		},
		DependentBody: map[schema.SchemaKey]*schema.BodySchema{
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "azurerm_subnet"},
				},
			}): {
				Attributes: map[string]*schema.AttributeSchema{
					"one":   {ValueType: cty.String, IsRequired: true},
					"two":   {ValueType: cty.Number},
					"three": {ValueType: cty.Bool},
				},
			},
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "random_resource"},
				},
			}): {
				Attributes: map[string]*schema.AttributeSchema{
					"four": {ValueType: cty.Number},
					"five": {ValueType: cty.DynamicPseudoType},
				},
			},
		},
	}

	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"resource": resourceSchema,
		},
	}

	cfg := []byte(`resource "azurerm_subnet" "example" {
  count = 3
}

resource "random_resource" "test" {
  arg = ""
}
`)

	d := NewDecoder()
	d.SetSchema(bodySchema)

	f, pDiags := hclsyntax.ParseConfig(cfg, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name               string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"in-between blocks",
			hcl.Pos{Column: 1, Line: 4, Byte: 52},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "resource",
					Detail: "Block",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 4, Column: 1, Byte: 52},
							End:      hcl.Pos{Line: 4, Column: 1, Byte: 52},
						},
						NewText: "resource",
						Snippet: "resource \"${1}\" \"${2:name}\" {\n  ${3}\n}",
					},
					Kind: lang.BlockCandidateKind,
				},
			}),
		},
		{
			"block type",
			hcl.Pos{Line: 1, Column: 2, Byte: 1},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "resource",
					Detail: "Block",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:      hcl.Pos{Line: 3, Column: 2, Byte: 51},
						},
						NewText: "resource",
						Snippet: "resource \"${1}\" \"${2:name}\" {\n  ${3}\n}",
					},
					Kind: lang.BlockCandidateKind,
				},
			}),
		},
		{
			"first label",
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label: "azurerm_subnet",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
							End:      hcl.Pos{Line: 1, Column: 25, Byte: 24},
						},
						NewText: "azurerm_subnet",
						Snippet: "azurerm_subnet",
					},
					Kind: lang.LabelCandidateKind,
				},
				{
					Label: "random_resource",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
							End:      hcl.Pos{Line: 1, Column: 25, Byte: 24},
						},
						NewText: "random_resource",
						Snippet: "random_resource",
					},
					Kind: lang.LabelCandidateKind,
				},
			}),
		},
		{
			"first block body",
			hcl.Pos{Line: 2, Column: 1, Byte: 38},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "one",
					Detail: "Required, string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 1, Byte: 38},
							End:      hcl.Pos{Line: 2, Column: 1, Byte: 38},
						},
						NewText: "one",
						Snippet: `one = "${1:value}"`,
					},
					Kind: lang.AttributeCandidateKind,
				},
				{
					Label:  "three",
					Detail: "Optional, bool",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 1, Byte: 38},
							End:      hcl.Pos{Line: 2, Column: 1, Byte: 38},
						},
						NewText: "three",
						Snippet: "three = ${1:false}",
					},
					Kind: lang.AttributeCandidateKind,
				},
				{
					Label:  "two",
					Detail: "Optional, number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 1, Byte: 38},
							End:      hcl.Pos{Line: 2, Column: 1, Byte: 38},
						},
						NewText: "two",
						Snippet: "two = ${1:1}",
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"first block attribute",
			hcl.Pos{Line: 2, Column: 3, Byte: 40},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "one",
					Detail: "Required, string",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 40},
							End:      hcl.Pos{Line: 2, Column: 12, Byte: 49},
						},
						NewText: "one",
						Snippet: `one = "${1:value}"`,
					},
					Kind: lang.AttributeCandidateKind,
				},
				{
					Label:  "three",
					Detail: "Optional, bool",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 40},
							End:      hcl.Pos{Line: 2, Column: 12, Byte: 49},
						},
						NewText: "three",
						Snippet: "three = ${1:false}",
					},
					Kind: lang.AttributeCandidateKind,
				},
				{
					Label:  "two",
					Detail: "Optional, number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 3, Byte: 40},
							End:      hcl.Pos{Line: 2, Column: 12, Byte: 49},
						},
						NewText: "two",
						Snippet: "two = ${1:1}",
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			candidates, err := d.CandidatesAtPos("test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}
			expectedCandidates := tc.expectedCandidates

			diff := cmp.Diff(expectedCandidates, candidates, ctydebug.CmpOptions)
			if diff != "" {
				t.Fatalf("unexpected schema for %s: %s", stringPos(tc.pos), diff)
			}
		})
	}
}

func TestDecoder_CandidatesAtPos_AnyAttribute(t *testing.T) {
	providersSchema := &schema.BlockSchema{
		Body: &schema.BodySchema{
			AnyAttribute: &schema.AttributeSchema{
				ValueType: cty.Object(map[string]cty.Type{
					"source":  cty.String,
					"version": cty.String,
				}),
			},
		},
	}

	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"required_providers": providersSchema,
		},
	}

	cfg := []byte(`required_providers {

}
`)

	d := NewDecoder()
	d.SetSchema(bodySchema)

	f, pDiags := hclsyntax.ParseConfig(cfg, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	pos := hcl.Pos{Line: 2, Column: 1, Byte: 21}
	candidates, err := d.CandidatesAtPos("test.tf", pos)
	if err != nil {
		t.Fatal(err)
	}
	expectedCandidates := lang.CompleteCandidates([]lang.Candidate{
		{
			Label:  "name",
			Detail: "Optional, object",
			TextEdit: lang.TextEdit{
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 1, Byte: 21},
					End:      hcl.Pos{Line: 2, Column: 1, Byte: 21},
				},
				NewText: "name",
				Snippet: "name = {\n  source = \"${1:value}\"\n  version = \"${2:value}\"\n}",
			},
			Kind: lang.AttributeCandidateKind,
		},
	})

	diff := cmp.Diff(expectedCandidates, candidates, ctydebug.CmpOptions)
	if diff != "" {
		t.Fatalf("unexpected schema for %s: %s", stringPos(pos), diff)
	}
}

func TestDecoder_CandidatesAtPos_multipleTypes(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "type", IsDepKey: true},
		{Name: "name"},
	}

	resourceSchema := &schema.BlockSchema{
		Labels: resourceLabelSchema,
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"for_each": {
					ValueTypes: schema.ValueTypes{
						cty.Set(cty.DynamicPseudoType),
						cty.Map(cty.DynamicPseudoType),
					},
				},
			},
		},
	}

	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"resource": resourceSchema,
		},
	}

	cfg := []byte(`resource "azurerm_subnet" "example" {

}
`)

	d := NewDecoder()
	d.SetSchema(bodySchema)

	f, pDiags := hclsyntax.ParseConfig(cfg, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	pos := hcl.Pos{Line: 2, Column: 1, Byte: 38}
	candidates, err := d.CandidatesAtPos("test.tf", pos)
	if err != nil {
		t.Fatal(err)
	}
	expectedCandidates := lang.CompleteCandidates([]lang.Candidate{
		{
			Label:  "for_each",
			Detail: "Optional, set of dynamic or map of dynamic",
			TextEdit: lang.TextEdit{
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 1, Byte: 38},
					End:      hcl.Pos{Line: 2, Column: 1, Byte: 38},
				},
				NewText: "for_each",
				Snippet: "for_each = [ ${1} ]",
			},
			Kind: lang.AttributeCandidateKind,
		},
	})

	diff := cmp.Diff(expectedCandidates, candidates, ctydebug.CmpOptions)
	if diff != "" {
		t.Fatalf("unexpected schema for %s: %s", stringPos(pos), diff)
	}
}

func TestDecoder_CandidatesAtPos_incompleteAttrOrBlock(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "type"},
		{Name: "name"},
	}

	resourceSchema := &schema.BlockSchema{
		Labels: resourceLabelSchema,
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"count": {ValueType: cty.Number},
			},
		},
	}

	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"resource": resourceSchema,
		},
	}

	d := NewDecoder()
	d.SetSchema(bodySchema)

	testCases := []struct {
		name               string
		src                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"new root block or attribute",
			`
res
`,
			hcl.Pos{Line: 2, Column: 4, Byte: 4},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "resource",
					Detail: "Block",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 1, Byte: 1},
							End:      hcl.Pos{Line: 2, Column: 4, Byte: 4},
						},
						NewText: "resource",
						Snippet: "resource \"${1:type}\" \"${2:name}\" {\n  ${3}\n}",
					},
					Kind: lang.BlockCandidateKind,
				},
			}),
		},
		{
			"new block or attribute inside a block",
			`
resource "any" "ref" {
  co
}
`,
			hcl.Pos{Line: 3, Column: 5, Byte: 28},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "count",
					Detail: "Optional, number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 26},
							End:      hcl.Pos{Line: 3, Column: 5, Byte: 28},
						},
						NewText: "count",
						Snippet: "count = ${1:1}",
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"new block or attribute inside a block",
			`
resource "any" "ref" {
  co
}
`,
			hcl.Pos{Line: 3, Column: 5, Byte: 28},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "count",
					Detail: "Optional, number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 26},
							End:      hcl.Pos{Line: 3, Column: 5, Byte: 28},
						},
						NewText: "count",
						Snippet: "count = ${1:1}",
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			f, _ := hclsyntax.ParseConfig([]byte(tc.src), "test.tf", hcl.InitialPos)

			err := d.LoadFile("test.tf", f)
			if err != nil {
				t.Fatal(err)
			}

			candidates, err := d.CandidatesAtPos("test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}
			expectedCandidates := tc.expectedCandidates

			diff := cmp.Diff(expectedCandidates, candidates, ctydebug.CmpOptions)
			if diff != "" {
				t.Fatalf("unexpected schema for %s: %s", stringPos(tc.pos), diff)
			}
		})
	}
}

func TestDecoder_CandidatesAtPos_incompleteLabel(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "type"},
		{Name: "name"},
	}

	resourceSchema := &schema.BlockSchema{
		Labels: resourceLabelSchema,
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"count": {ValueType: cty.Number},
			},
		},
		DependentBody: map[schema.SchemaKey]*schema.BodySchema{
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "mycloud_instance"},
				},
			}): {},
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "mycloud_bucket"},
				},
			}): {},
		},
	}

	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"resource": resourceSchema,
		},
	}

	d := NewDecoder()
	d.SetSchema(bodySchema)

	testCases := []struct {
		name               string
		src                string
		pos                hcl.Pos
		expectedCandidates lang.Candidates
	}{
		{
			"",
			`
res
`,
			hcl.Pos{Line: 2, Column: 4, Byte: 4},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "resource",
					Detail: "Block",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 2, Column: 1, Byte: 1},
							End:      hcl.Pos{Line: 2, Column: 4, Byte: 4},
						},
						NewText: "resource",
						Snippet: "resource \"${1:type}\" \"${2:name}\" {\n  ${3}\n}",
					},
					Kind: lang.BlockCandidateKind,
				},
			}),
		},
		{
			"new block or attribute inside a block",
			`
resource "any" "ref" {
  co
}
`,
			hcl.Pos{Line: 3, Column: 5, Byte: 28},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "count",
					Detail: "Optional, number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 26},
							End:      hcl.Pos{Line: 3, Column: 5, Byte: 28},
						},
						NewText: "count",
						Snippet: "count = ${1:1}",
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
		{
			"new block or attribute inside a block",
			`
resource "any" "ref" {
  co
}
`,
			hcl.Pos{Line: 3, Column: 5, Byte: 28},
			lang.CompleteCandidates([]lang.Candidate{
				{
					Label:  "count",
					Detail: "Optional, number",
					TextEdit: lang.TextEdit{
						Range: hcl.Range{
							Filename: "test.tf",
							Start:    hcl.Pos{Line: 3, Column: 3, Byte: 26},
							End:      hcl.Pos{Line: 3, Column: 5, Byte: 28},
						},
						NewText: "count",
						Snippet: "count = ${1:1}",
					},
					Kind: lang.AttributeCandidateKind,
				},
			}),
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			f, _ := hclsyntax.ParseConfig([]byte(tc.src), "test.tf", hcl.InitialPos)

			err := d.LoadFile("test.tf", f)
			if err != nil {
				t.Fatal(err)
			}

			candidates, err := d.CandidatesAtPos("test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}
			expectedCandidates := tc.expectedCandidates

			diff := cmp.Diff(expectedCandidates, candidates, ctydebug.CmpOptions)
			if diff != "" {
				t.Fatalf("unexpected schema for %s: %s", stringPos(tc.pos), diff)
			}
		})
	}
}

var testConfig = []byte(`resource "azurerm_subnet" "example" {
  count = 3
}

resource "random_resource" "test" {
  arg = ""
}
`)
