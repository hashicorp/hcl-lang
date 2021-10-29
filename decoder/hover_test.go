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
	"github.com/hashicorp/hcl/v2/json"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestDecoder_HoverAtPos_noSchema(t *testing.T) {
	f, pDiags := hclsyntax.ParseConfig(testConfig, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	d := testPathDecoder(t, &PathContext{
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})

	_, err := d.HoverAtPos("test.tf", hcl.InitialPos)
	noSchemaErr := &NoSchemaError{}
	if !errors.As(err, &noSchemaErr) {
		t.Fatal("expected NoSchemaError for no schema")
	}
}

func TestDecoder_HoverAtPos_emptyBody(t *testing.T) {
	f := &hcl.File{
		Body: hcl.EmptyBody(),
	}

	d := testPathDecoder(t, &PathContext{
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})

	_, err := d.HoverAtPos("test.tf", hcl.InitialPos)
	unknownFormatErr := &UnknownFileFormatError{}
	if !errors.As(err, &unknownFormatErr) {
		t.Fatal("expected UnknownFileFormatError for empty body")
	}
}

func TestDecoder_HoverAtPos_json(t *testing.T) {
	f, pDiags := json.Parse([]byte(`{
	"customblock": {
		"label1": {}
	}
}`), "test.tf.json")
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	d := testPathDecoder(t, &PathContext{
		Files: map[string]*hcl.File{
			"test.tf.json": f,
		},
	})

	_, err := d.HoverAtPos("test.tf.json", hcl.InitialPos)
	unknownFormatErr := &UnknownFileFormatError{}
	if !errors.As(err, &unknownFormatErr) {
		t.Fatal("expected UnknownFileFormatError for JSON body")
	}
}

func TestDecoder_HoverAtPos_nilBodySchema(t *testing.T) {
	testCases := []struct {
		name         string
		rootSchema   *schema.BodySchema
		config       string
		pos          hcl.Pos
		expectedData *lang.HoverData
	}{
		{
			"nil static body on type",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type"},
							{Name: "name"},
						},
						Body: nil,
					},
				},
			},
			`resource "label1" {
  count = 1

}
`,
			hcl.Pos{
				Line:   1,
				Column: 2,
				Byte:   1,
			},
			&lang.HoverData{
				Content: lang.Markdown("**resource** _Block_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 1,
						Byte:   0,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 9,
						Byte:   8,
					},
				},
			},
		},
		{
			"nil static body with dependent body on label",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type", IsDepKey: true},
							{Name: "name"},
						},
						Body: nil,
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "label1"},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"one":   {Expr: schema.LiteralTypeOnly(cty.String)},
									"two":   {Expr: schema.LiteralTypeOnly(cty.Number)},
									"three": {Expr: schema.LiteralTypeOnly(cty.Bool)},
								},
							},
						},
					},
				},
			},
			`resource "label1" {
  count = 1
  one = "test"
}
`,
			hcl.Pos{
				Line:   1,
				Column: 13,
				Byte:   12,
			},
			&lang.HoverData{
				Content: lang.Markdown("`label1` type"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 10,
						Byte:   9,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 18,
						Byte:   17,
					},
				},
			},
		},
		{
			"nil static body with dependent body on attribute",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type", IsDepKey: true},
							{Name: "name"},
						},
						Body: nil,
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "label1"},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"one":   {Expr: schema.LiteralTypeOnly(cty.String)},
									"two":   {Expr: schema.LiteralTypeOnly(cty.Number)},
									"three": {Expr: schema.LiteralTypeOnly(cty.Bool)},
								},
							},
						},
					},
				},
			},
			`resource "label1" {
  count = 1
  one = "test"
}
`,
			hcl.Pos{
				Line:   3,
				Column: 4,
				Byte:   35,
			},
			&lang.HoverData{
				Content: lang.Markdown("**one** _string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   3,
						Column: 3,
						Byte:   34,
					},
					End: hcl.Pos{
						Line:   3,
						Column: 15,
						Byte:   46,
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			f, pDiags := hclsyntax.ParseConfig([]byte(tc.config), "test.tf", hcl.InitialPos)
			if len(pDiags) > 0 {
				t.Fatal(pDiags)
			}

			d := testPathDecoder(t, &PathContext{
				Schema: tc.rootSchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			})

			data, err := d.HoverAtPos("test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedData, data); diff != "" {
				t.Fatalf("unexpected data: %s", diff)
			}
		})
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

	f, pDiags := hclsyntax.ParseConfig([]byte(`resource "label1" "test" {
  blablah = 42
}
`), "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	d := testPathDecoder(t, &PathContext{
		Schema: bodySchema,
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})

	_, err := d.HoverAtPos("test.tf", hcl.Pos{
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

	f, pDiags := hclsyntax.ParseConfig([]byte(`customblock "label1" {

}
`), "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	d := testPathDecoder(t, &PathContext{
		Schema: bodySchema,
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})

	_, err := d.HoverAtPos("test.tf", hcl.Pos{
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

	f, pDiags := hclsyntax.ParseConfig(testConfig, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	d := testPathDecoder(t, &PathContext{
		Schema: bodySchema,
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			_, err := d.HoverAtPos("test.tf", tc.pos)
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

	f, _ := hclsyntax.ParseConfig(testConfig, "test.tf", hcl.InitialPos)

	d := testPathDecoder(t, &PathContext{
		Schema: bodySchema,
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})

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
				"object_attr": {
					Expr: schema.LiteralTypeOnly(cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"opt_attr": cty.Bool,
						"req_attr": cty.Number,
					}, []string{"opt_attr"})),
					Description: lang.PlainText("Order attribute"),
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
  object_attr = {
	  opt_attr = false
	  req_attr = 5
  }
}
`)

	f, _ := hclsyntax.ParseConfig(testConfig, "test.tf", hcl.InitialPos)

	d := testPathDecoder(t, &PathContext{
		Schema: bodySchema,
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})

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
		{
			"optional in object",
			hcl.Pos{Line: 6, Column: 10, Byte: 105},
			&lang.HoverData{
				Content: lang.Markdown("```\n{\n  opt_attr = optional, bool\n  req_attr = number\n}\n```\n_object_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 5, Column: 17, Byte: 97},
					End:      hcl.Pos{Line: 8, Column: 4, Byte: 138},
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

func TestDecoder_HoverAtPos_URL(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "type", IsDepKey: true},
		{Name: "name"},
	}
	blockSchema := &schema.BlockSchema{
		Labels:      resourceLabelSchema,
		Description: lang.Markdown("My food block"),
		Body: &schema.BodySchema{
			HoverURL: "https://en.wikipedia.org/wiki/Food",
			Attributes: map[string]*schema.AttributeSchema{
				"any_attr": {Expr: schema.LiteralTypeOnly(cty.Number)},
			},
		},
		DependentBody: map[schema.SchemaKey]*schema.BodySchema{
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "sushi"},
				},
			}): {
				Detail:      "rice, fish etc.",
				HoverURL:    "https://en.wikipedia.org/wiki/Sushi",
				Description: lang.Markdown("Sushi, the Rolls-Rice of Japanese cuisine"),
			},
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "ramen"},
				},
			}): {
				Detail: "noodles, broth etc.",
				DocsLink: &schema.DocsLink{
					URL:     "https://en.wikipedia.org/wiki/Ramen",
					Tooltip: "Ramen docs",
				},
				Description: lang.Markdown("Ramen, a Japanese noodle soup"),
			},
		},
	}
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"myblock": blockSchema,
		},
	}

	testCases := []struct {
		name         string
		cfg          string
		pos          hcl.Pos
		expectedData *lang.HoverData
	}{
		{
			"",
			`myblock "sushi" "salmon" {
  any_attr = 42
}
`,
			hcl.Pos{
				Line:   1,
				Column: 12,
				Byte:   11,
			},
			&lang.HoverData{
				Content: lang.MarkupContent{
					Value: "`sushi`" + ` rice, fish etc.

Sushi, the Rolls-Rice of Japanese cuisine

[` + "`sushi`" + ` on en.wikipedia.org](https://en.wikipedia.org/wiki/Sushi)`,
					Kind: lang.MarkdownKind,
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 9,
						Byte:   8,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 16,
						Byte:   15,
					},
				},
			},
		},
		{
			"",
			`myblock "ramen" "tonkotsu" {
  any_attr = 42
}
`,
			hcl.Pos{
				Line:   1,
				Column: 12,
				Byte:   13,
			},
			&lang.HoverData{
				Content: lang.MarkupContent{
					Value: "`ramen` noodles, broth etc.\n\nRamen, a Japanese noodle soup",
					Kind:  lang.MarkdownKind,
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 9,
						Byte:   8,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 16,
						Byte:   15,
					},
				},
			},
		},
		{
			"",
			`myblock "ramen" "tonkotsu" {
  any_attr = 42
}
`,
			hcl.Pos{
				Line:   1,
				Column: 2,
				Byte:   1,
			},
			&lang.HoverData{
				Content: lang.MarkupContent{
					Value: `**myblock** _Block_

My food block

[` + "`myblock`" + ` on en.wikipedia.org](https://en.wikipedia.org/wiki/Food)`,
					Kind: lang.MarkdownKind,
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 1,
						Byte:   0,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 8,
						Byte:   7,
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)

			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			})

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

func TestDecoder_HoverAtPos_typeDeclaration(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "name", IsDepKey: true},
	}
	blockSchema := &schema.BlockSchema{
		Labels:      resourceLabelSchema,
		Description: lang.Markdown("My special block"),
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"type": {
					Expr:        schema.ExprConstraints{schema.TypeDeclarationExpr{}},
					IsOptional:  true,
					Description: lang.PlainText("Special attribute"),
				},
			},
		},
	}
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"myblock": blockSchema,
		},
	}

	testCases := []struct {
		name         string
		cfg          string
		expectedData *lang.HoverData
	}{
		{
			"primitive type",
			`myblock "sushi" {
  type = string
}
`,
			&lang.HoverData{
				Content: lang.Markdown("Type declaration"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 10, Byte: 27},
					End:      hcl.Pos{Line: 2, Column: 16, Byte: 33},
				},
			},
		},
		{
			"capsule type",
			`myblock "sushi" {
  type = list(string)
}
`,
			&lang.HoverData{
				Content: lang.Markdown("Type declaration"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 10, Byte: 27},
					End:      hcl.Pos{Line: 2, Column: 22, Byte: 39},
				},
			},
		},
		{
			"object type",
			`myblock "sushi" {
  type = object({
	  vegan = bool
  })
}
`,
			&lang.HoverData{
				Content: lang.Markdown("Type declaration"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 10, Byte: 27},
					End:      hcl.Pos{Line: 4, Column: 5, Byte: 56},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			testConfig := []byte(tc.cfg)
			f, _ := hclsyntax.ParseConfig(testConfig, "test.tf", hcl.InitialPos)

			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			})

			pos := hcl.Pos{Line: 2, Column: 6, Byte: 32}
			data, err := d.HoverAtPos("test.tf", pos)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.expectedData, data, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("hover data mismatch: %s", diff)
			}
		})
	}
}
