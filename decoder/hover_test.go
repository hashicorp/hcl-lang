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

func TestDecoder_HoverAtPos_noSchema(t *testing.T) {
	d := NewDecoder()
	f, pDiags := hclsyntax.ParseConfig(testConfig, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.HoverAtPos("test.tf", hcl.InitialPos)
	noSchemaErr := &NoSchemaError{}
	if !errors.As(err, &noSchemaErr) {
		t.Fatal("expected NoSchemaError for no schema")
	}
}

func TestDecoder_HoverAtPos_emptyBody(t *testing.T) {
	d := NewDecoder()
	f := &hcl.File{
		Body: hcl.EmptyBody(),
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.HoverAtPos("test.tf", hcl.InitialPos)
	unknownFormatErr := &UnknownFileFormatError{}
	if !errors.As(err, &unknownFormatErr) {
		t.Fatal("expected UnknownFileFormatError for empty body")
	}
}

func TestDecoder_HoverAtPos_unknownAttribute(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "type"},
		{Name: "name"},
	}
	resourceSchema := &schema.BlockSchema{
		Labels: resourceLabelSchema,
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"count": {Expr: schema.LiteralTypeOnly(cty.Number)},
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
	f, pDiags := hclsyntax.ParseConfig([]byte(`resource "label1" "test" {
  blablah = 42
}
`), "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.HoverAtPos("test.tf", hcl.Pos{
		Line:   2,
		Column: 6,
		Byte:   32,
	})
	if err == nil {
		t.Fatal("expected error for unknown attribute")
	}
	if !strings.Contains(err.Error(), "unknown attribute") {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
}

func TestDecoder_HoverAtPos_unknownBlock(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "type"},
		{Name: "name"},
	}
	resourceSchema := &schema.BlockSchema{
		Labels: resourceLabelSchema,
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"count": {Expr: schema.LiteralTypeOnly(cty.Number)},
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

	_, err = d.HoverAtPos("test.tf", hcl.Pos{
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

func TestDecoder_HoverAtPos_invalidBlockPositions(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "type"},
	}
	blockSchema := &schema.BlockSchema{
		Labels: resourceLabelSchema,
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"num_attr": {Expr: schema.LiteralTypeOnly(cty.Number)},
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
			_, err = d.HoverAtPos("test.tf", tc.pos)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), `position outside of "myblock" body`) {
				t.Fatalf("unexpected error message: %q", err.Error())
			}
		})
	}
}

func TestDecoder_HoverAtPos_rightHandSide(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "type"},
	}
	blockSchema := &schema.BlockSchema{
		Labels: resourceLabelSchema,
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"num_attr": {Expr: schema.LiteralTypeOnly(cty.Number)},
				"str_attr": {Expr: schema.LiteralTypeOnly(cty.String), Description: lang.PlainText("Special attribute")},
			},
		},
	}
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"myblock": blockSchema,
		},
	}
	testConfig := []byte(`myblock "foo" {
  str_attr = "test"
}
`)

	d := NewDecoder()
	d.SetSchema(bodySchema)
	f, _ := hclsyntax.ParseConfig(testConfig, "test.tf", hcl.InitialPos)
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	data, err := d.HoverAtPos("test.tf", hcl.Pos{
		Line:   2,
		Column: 17,
		Byte:   32,
	})
	if err != nil {
		t.Fatal(err)
	}
	expectedData := &lang.HoverData{
		Content: lang.Markdown("`\"test\"` _string_"),
		Range: hcl.Range{
			Filename: "test.tf",
			Start:    hcl.Pos{Line: 2, Column: 14, Byte: 29},
			End:      hcl.Pos{Line: 2, Column: 20, Byte: 35},
		},
	}
	if diff := cmp.Diff(expectedData, data, ctydebug.CmpOptions); diff != "" {
		t.Fatalf("hover data mismatch: %s", diff)
	}
}

func TestDecoder_HoverAtPos_basic(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "type", IsDepKey: true},
		{Name: "name"},
	}
	blockSchema := &schema.BlockSchema{
		Labels:      resourceLabelSchema,
		Description: lang.Markdown("My special block"),
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"num_attr": {Expr: schema.LiteralTypeOnly(cty.Number)},
				"str_attr": {
					Expr:        schema.LiteralTypeOnly(cty.String),
					IsOptional:  true,
					Description: lang.PlainText("Special attribute"),
				},
				"bool_attr": {
					Expr:        schema.LiteralTypeOnly(cty.Bool),
					IsSensitive: true,
					Description: lang.PlainText("Flag attribute"),
				},
			},
		},
		DependentBody: map[schema.SchemaKey]*schema.BodySchema{
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "sushi"},
				},
			}): {
				Detail:      "rice, fish etc.",
				Description: lang.Markdown("Sushi, the Rolls-Rice of Japanese cuisine"),
			},
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "ramen"},
				},
			}): {
				Detail:      "noodles, broth etc.",
				Description: lang.Markdown("Ramen, a Japanese noodle soup"),
			},
		},
	}
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"myblock": blockSchema,
		},
	}
	testConfig := []byte(`myblock "sushi" "salmon" {
  str_attr = "test"
  num_attr = 4
  bool_attr = true
}
`)
	d := NewDecoder()
	d.SetSchema(bodySchema)

	f, _ := hclsyntax.ParseConfig(testConfig, "test.tf", hcl.InitialPos)
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name         string
		pos          hcl.Pos
		expectedData *lang.HoverData
	}{
		{
			"optional attribute name",
			hcl.Pos{Line: 2, Column: 6, Byte: 32},
			&lang.HoverData{
				Content: lang.Markdown("**str_attr** _optional, string_\n\nSpecial attribute"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 3, Byte: 29},
					End:      hcl.Pos{Line: 2, Column: 20, Byte: 46},
				},
			},
		},
		{
			"sensitive attribute name",
			hcl.Pos{Line: 4, Column: 6, Byte: 68},
			&lang.HoverData{
				Content: lang.Markdown("**bool_attr** _sensitive, bool_\n\nFlag attribute"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 4, Column: 3, Byte: 64},
					End:      hcl.Pos{Line: 4, Column: 19, Byte: 80},
				},
			},
		},
		{
			"block type",
			hcl.Pos{Line: 1, Column: 3, Byte: 2},
			&lang.HoverData{
				Content: lang.Markdown("**myblock** _Block_\n\nMy special block"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
				},
			},
		},
		{
			"dependent label",
			hcl.Pos{Line: 1, Column: 11, Byte: 10},
			&lang.HoverData{
				Content: lang.Markdown("`sushi` rice, fish etc.\n\nSushi, the Rolls-Rice of Japanese cuisine"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
					End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
				},
			},
		},
		{
			"independent label",
			hcl.Pos{Line: 1, Column: 21, Byte: 20},
			&lang.HoverData{
				Content: lang.Markdown(`"salmon" (name)`),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 17, Byte: 16},
					End:      hcl.Pos{Line: 1, Column: 25, Byte: 24},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			data, err := d.HoverAtPos("test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.expectedData, data, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("hover data mismatch: %s", diff)
			}
		})
	}
}
