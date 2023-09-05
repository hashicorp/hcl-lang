// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
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

	ctx := context.Background()
	_, err := d.HoverAtPos(ctx, "test.tf", hcl.InitialPos)
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

	ctx := context.Background()
	_, err := d.HoverAtPos(ctx, "test.tf", hcl.InitialPos)
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

	ctx := context.Background()
	_, err := d.HoverAtPos(ctx, "test.tf.json", hcl.InitialPos)
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
									"one":   {Constraint: schema.LiteralType{Type: cty.String}},
									"two":   {Constraint: schema.LiteralType{Type: cty.Number}},
									"three": {Constraint: schema.LiteralType{Type: cty.Bool}},
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
									"one":   {Constraint: schema.LiteralType{Type: cty.String}},
									"two":   {Constraint: schema.LiteralType{Type: cty.Number}},
									"three": {Constraint: schema.LiteralType{Type: cty.Bool}},
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
			ctx := context.Background()

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
			data, err := d.HoverAtPos(ctx, "test.tf", tc.pos)
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
				"count": {Constraint: schema.LiteralType{Type: cty.Number}},
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

	ctx := context.Background()
	_, err := d.HoverAtPos(ctx, "test.tf", hcl.Pos{
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
				"count": {Constraint: schema.LiteralType{Type: cty.Number}},
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

	ctx := context.Background()
	_, err := d.HoverAtPos(ctx, "test.tf", hcl.Pos{
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
				"num_attr": {Constraint: schema.LiteralType{Type: cty.Number}},
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
			ctx := context.Background()
			_, err := d.HoverAtPos(ctx, "test.tf", tc.pos)
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
				"num_attr": {Constraint: schema.LiteralType{Type: cty.Number}},
				"str_attr": {Constraint: schema.LiteralType{Type: cty.String}, Description: lang.PlainText("Special attribute")},
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

	ctx := context.Background()
	data, err := d.HoverAtPos(ctx, "test.tf", hcl.Pos{
		Line:   2,
		Column: 17,
		Byte:   32,
	})
	if err != nil {
		t.Fatal(err)
	}
	expectedData := &lang.HoverData{
		Content: lang.Markdown("_string_"),
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
				"num_attr": {Constraint: schema.LiteralType{Type: cty.Number}},
				"str_attr": {
					Constraint:  schema.LiteralType{Type: cty.String},
					IsOptional:  true,
					Description: lang.PlainText("Special attribute"),
				},
				"bool_attr": {
					Constraint:  schema.LiteralType{Type: cty.Bool},
					IsSensitive: true,
					Description: lang.PlainText("Flag attribute"),
				},
				"object_attr": {
					Constraint: schema.LiteralType{Type: cty.ObjectWithOptionalAttrs(map[string]cty.Type{
						"opt_attr": cty.Bool,
						"req_attr": cty.Number,
					}, []string{"opt_attr"})},
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
				Content: lang.Markdown("**opt_attr** _optional, bool_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 6, Column: 4, Byte: 102},
					End:      hcl.Pos{Line: 6, Column: 20, Byte: 118},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			ctx := context.Background()
			data, err := d.HoverAtPos(ctx, "test.tf", tc.pos)
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
				"any_attr": {Constraint: schema.LiteralType{Type: cty.Number}},
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

			ctx := context.Background()
			data, err := d.HoverAtPos(ctx, "test.tf", tc.pos)
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
					Constraint:  schema.TypeDeclaration{},
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
				Content: lang.Markdown("_string_"),
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
				Content: lang.Markdown("_list of string_"),
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
				Content: lang.Markdown("```\n{\n  vegan = bool\n}\n```\n_object_"),
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

			ctx := context.Background()
			pos := hcl.Pos{Line: 2, Column: 11, Byte: 28}
			data, err := d.HoverAtPos(ctx, "test.tf", pos)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.expectedData, data, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("hover data mismatch: %s", diff)
			}
		})
	}
}

func TestDecoder_HoverAtPos_extensions_count(t *testing.T) {
	testCases := []struct {
		name         string
		bodySchema   *schema.BodySchema
		RefTargets   reference.Targets
		origins      reference.Origins
		config       string
		pos          hcl.Pos
		expectedData *lang.HoverData
	}{
		{
			"count attribute name",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"myblock": {
						Labels: []*schema.LabelSchema{
							{Name: "type", IsDepKey: true},
							{Name: "name"},
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
			reference.Origins{},
			`myblock "foo" "bar" {
  count = 1
}
`,
			hcl.Pos{Line: 2, Column: 5, Byte: 24},
			&lang.HoverData{
				Content: lang.Markdown("**count** _optional, number_\n\nTotal number of instances of this block.\n\n**Note**: A given block cannot use both `count` and `for_each`."),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 3, Byte: 24},
					End:      hcl.Pos{Line: 2, Column: 12, Byte: 33},
				},
			},
		},
		{
			"count.index reference",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"myblock": {
						Labels: []*schema.LabelSchema{
							{Name: "type", IsDepKey: true},
							{Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								Count: true,
							},
							Attributes: map[string]*schema.AttributeSchema{
								"foo": {
									IsOptional: true,
									Constraint: schema.OneOf{
										schema.Reference{OfType: cty.Number},
										schema.LiteralType{Type: cty.Number},
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
						lang.RootStep{Name: "count"},
						lang.AttrStep{Name: "index"},
					},
					Type:        cty.Number,
					Description: lang.PlainText("The distinct index number (starting with 0) corresponding to the instance"),
				},
			},
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "count"},
						lang.AttrStep{Name: "index"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 11, Byte: 44},
						End:      hcl.Pos{Line: 3, Column: 22, Byte: 55},
					},
				},
			},
			`myblock "foo" "bar" {
  count = 1
  foo   = count.index
}
`,
			hcl.Pos{Line: 3, Column: 15, Byte: 48},
			&lang.HoverData{
				Content: lang.Markdown("`count.index`\n_number_\n\nThe distinct index number (starting with 0) corresponding to the instance"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 3, Column: 11, Byte: 44},
					End:      hcl.Pos{Line: 3, Column: 22, Byte: 55},
				},
			},
		},
		{
			"count attribute value",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"myblock": {
						Labels: []*schema.LabelSchema{
							{Name: "type", IsDepKey: true},
							{Name: "name"},
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
			reference.Origins{},
			`myblock "foo" "bar" {
  count = 3
}
`,
			hcl.Pos{Line: 2, Column: 11, Byte: 32},
			&lang.HoverData{
				Content: lang.Markdown("_number_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 11, Byte: 32},
					End:      hcl.Pos{Line: 2, Column: 12, Byte: 33},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			ctx := context.Background()

			f, diags := hclsyntax.ParseConfig([]byte(tc.config), "test.tf", hcl.InitialPos)
			if diags != nil {
				t.Fatal(diags)
			}

			d := testPathDecoder(t, &PathContext{
				Schema: tc.bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
				ReferenceTargets: tc.RefTargets,
				ReferenceOrigins: tc.origins,
			})

			data, err := d.HoverAtPos(ctx, "test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.expectedData, data, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("hover data mismatch: %s", diff)
			}
		})
	}
}

func TestDecoder_HoverAtPos_extension_for_each(t *testing.T) {
	testCases := []struct {
		name         string
		bodySchema   *schema.BodySchema
		refTargets   reference.Targets
		origins      reference.Origins
		config       string
		pos          hcl.Pos
		expectedData *lang.HoverData
	}{
		{
			"for_each attribute name",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"myblock": {
						Labels: []*schema.LabelSchema{
							{Name: "type", IsDepKey: true},
							{Name: "name"},
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
			reference.Origins{},
			`myblock "foo" "bar" {
  for_each = {
		
	}
}
`,
			hcl.Pos{Line: 2, Column: 5, Byte: 26},
			&lang.HoverData{
				Content: lang.MarkupContent{
					Value: "**for_each** _optional, map of any single type or set of string_\n\n" +
						"A meta-argument that accepts a map or a set of strings, and creates an instance for each item in that map or set.\n\n" +
						"**Note**: A given block cannot use both `count` and `for_each`.",
					Kind: lang.MarkdownKind,
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 3, Byte: 24},
					End:      hcl.Pos{Line: 4, Column: 3, Byte: 42},
				},
			},
		},
		{
			"each.key attribute key",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"myblock": {
						Labels: []*schema.LabelSchema{
							{Name: "type", IsDepKey: true},
							{Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								ForEach: true,
							},
							Attributes: map[string]*schema.AttributeSchema{
								"foo": {
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
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 21,
							Byte:   20,
						},
						End: hcl.Pos{
							Line:   7,
							Column: 2,
							Byte:   83,
						},
					},
					Type:        cty.String,
					Description: lang.Markdown("The map key (or set member) corresponding to this instance"),
					// TODO: RangePtr & DefRangePtr
				},
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "each"},
						lang.AttrStep{Name: "value"},
					},
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 21,
							Byte:   20,
						},
						End: hcl.Pos{
							Line:   7,
							Column: 2,
							Byte:   83,
						},
					},
					Type:        cty.DynamicPseudoType,
					Description: lang.Markdown("The map value corresponding to this instance. (If a set was provided, this is the same as `each.key`.)"),
					// TODO: RangePtr & DefRangePtr
				},
			},
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "each"},
						lang.AttrStep{Name: "key"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 8, Byte: 29},
						End:      hcl.Pos{Line: 2, Column: 16, Byte: 37},
					},
				},
			},
			`myblock "foo" "bar" {
	foo = each.key
	for_each = {
		thing = "bar"
		woot = 3
	}
}
`,
			hcl.Pos{Line: 2, Column: 8, Byte: 30},
			&lang.HoverData{
				Content: lang.Markdown("`each.key`\n_string_\n\nThe map key (or set member) corresponding to this instance"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 8, Byte: 29},
					End:      hcl.Pos{Line: 2, Column: 16, Byte: 37},
				},
			},
		},
		{
			"each.value attribute value",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"myblock": {
						Labels: []*schema.LabelSchema{
							{Name: "type", IsDepKey: true},
							{Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								ForEach: true,
							},
							Attributes: map[string]*schema.AttributeSchema{
								"foo": {
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
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 21,
							Byte:   20,
						},
						End: hcl.Pos{
							Line:   7,
							Column: 2,
							Byte:   83,
						},
					},
					Type:        cty.String,
					Description: lang.Markdown("The map key (or set member) corresponding to this instance"),
					// TODO: RangePtr & DefRangePtr
				},
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "each"},
						lang.AttrStep{Name: "value"},
					},
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 21,
							Byte:   20,
						},
						End: hcl.Pos{
							Line:   7,
							Column: 2,
							Byte:   83,
						},
					},
					Type:        cty.DynamicPseudoType,
					Description: lang.Markdown("The map value corresponding to this instance. (If a set was provided, this is the same as `each.key`.)"),
					// TODO: RangePtr & DefRangePtr
				},
			},
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "each"},
						lang.AttrStep{Name: "value"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 8, Byte: 29},
						End:      hcl.Pos{Line: 2, Column: 18, Byte: 39},
					},
				},
			},
			`myblock "foo" "bar" {
	foo = each.value
	for_each = {
		thing = "bar"
		woot = 3
	}
}
`,
			hcl.Pos{Line: 2, Column: 8, Byte: 30},
			&lang.HoverData{
				Content: lang.Markdown("`each.value`\n_dynamic_\n\nThe map value corresponding to this instance. (If a set was provided, this is the same as `each.key`.)"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 8, Byte: 29},
					End:      hcl.Pos{Line: 2, Column: 18, Byte: 39},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			ctx := context.Background()

			f, diags := hclsyntax.ParseConfig([]byte(tc.config), "test.tf", hcl.InitialPos)
			if diags != nil {
				t.Fatal(diags)
			}

			d := testPathDecoder(t, &PathContext{
				Schema: tc.bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
				ReferenceTargets: tc.refTargets,
				ReferenceOrigins: tc.origins,
			})

			data, err := d.HoverAtPos(ctx, "test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.expectedData, data, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("hover data mismatch: %s", diff)
			}
		})
	}
}

func TestDecoder_HoverAtPos_extensions_dynamic(t *testing.T) {
	testCases := []struct {
		name         string
		bodySchema   *schema.BodySchema
		config       string
		pos          hcl.Pos
		expectedData *lang.HoverData
	}{
		{
			"dynamic block",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"myblock": {
						Labels: []*schema.LabelSchema{
							{Name: "type", IsDepKey: true}, {Name: "name"},
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
									{Index: 0, Value: "foo"},
								},
							}): {
								Blocks: map[string]*schema.BlockSchema{
									"thing": {
										Body: schema.NewBodySchema(),
									},
								},
							},
						},
					},
				},
			},
			`myblock "foo" "bar" {
  dynamic "thing" {	}
}
`,
			hcl.Pos{Line: 2, Column: 5, Byte: 26},
			&lang.HoverData{
				Content: lang.MarkupContent{
					Value: "**dynamic** _Block_\n\n" +
						"A dynamic block to produce blocks dynamically by iterating over a given complex value",
					Kind: lang.MarkdownKind,
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 3, Byte: 24},
					End:      hcl.Pos{Line: 2, Column: 10, Byte: 31},
				},
			},
		},
		{
			"hover on nested dynamic blocks",
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
							},
						},
					},
				},
			},
			`resource "aws_instance" "example" {
  dynamic "foo" {
    content {
      dynamic "bar" {
        content {
        }
      }
    }
  }
}`,
			hcl.Pos{Line: 5, Column: 13, Byte: 102},
			&lang.HoverData{
				Content: lang.MarkupContent{
					Value: "**content** _Block, min: 1, max: 1_\n\n" +
						"The body of each generated block",
					Kind: lang.MarkdownKind,
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 5, Column: 9, Byte: 98},
					End:      hcl.Pos{Line: 5, Column: 16, Byte: 105},
				},
			},
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
      dynamic "baz" {

      }
    }
  }
}`,
			hcl.Pos{Line: 4, Column: 10, Byte: 63},
			&lang.HoverData{
				Content: lang.MarkupContent{
					Value: "**dynamic** _Block_\n\n" +
						"A dynamic block to produce blocks dynamically by iterating over a given complex value",
					Kind: lang.MarkdownKind,
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 4, Column: 7, Byte: 60},
					End:      hcl.Pos{Line: 4, Column: 14, Byte: 67},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			ctx := context.Background()

			f, diags := hclsyntax.ParseConfig([]byte(tc.config), "test.tf", hcl.InitialPos)
			if diags != nil {
				t.Fatal(diags)
			}

			d := testPathDecoder(t, &PathContext{
				Schema: tc.bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			})

			data, err := d.HoverAtPos(ctx, "test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.expectedData, data, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("hover data mismatch: %s", diff)
			}
		})
	}
}

func TestDecoder_HoverAtPos_extensions_references(t *testing.T) {
	testCases := []struct {
		name         string
		bodySchema   *schema.BodySchema
		config       string
		pos          hcl.Pos
		expectedData *lang.HoverData
	}{
		{
			"for_each var reference",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"myblock": {
						Labels: []*schema.LabelSchema{
							{Name: "type", IsDepKey: true},
							{Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								ForEach: true,
							},
							Attributes: map[string]*schema.AttributeSchema{
								"foo": {
									IsOptional: true,
									Constraint: schema.Reference{
										OfType: cty.String,
									},
								},
							},
						},
					},
					"variable": {
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "var"},
								schema.LabelStep{Index: 0},
							},
							ScopeId:     lang.ScopeId("variable"),
							AsReference: true,
							AsTypeOf: &schema.BlockAsTypeOf{
								AttributeExpr:  "type",
								AttributeValue: "default",
							},
						},
						Labels: []*schema.LabelSchema{
							{Name: "name"},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"type": {
									Constraint: schema.TypeDeclaration{},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`myblock "foo" "bar" {
  foo = each.value
  for_each = var.name
}
variable "name" {
  value = { key = "value" }
}
`,
			hcl.Pos{Line: 3, Column: 19, Byte: 59},
			&lang.HoverData{
				Content: lang.Markdown("`var.name`\n_dynamic_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 3, Column: 14, Byte: 54},
					End:      hcl.Pos{Line: 3, Column: 22, Byte: 62},
				},
			},
		},
		{
			"count var reference",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"myblock": {
						Labels: []*schema.LabelSchema{
							{Name: "type", IsDepKey: true},
							{Name: "name"},
						},
						Body: &schema.BodySchema{
							Extensions: &schema.BodyExtensions{
								Count: true,
							},
							Attributes: map[string]*schema.AttributeSchema{
								"foo": {
									IsOptional: true,
									Constraint: schema.Reference{
										OfType: cty.Number,
									},
								},
							},
						},
					},
					"variable": {
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "var"},
								schema.LabelStep{Index: 0},
							},
							ScopeId:     lang.ScopeId("variable"),
							AsReference: true,
							AsTypeOf: &schema.BlockAsTypeOf{
								AttributeExpr:  "type",
								AttributeValue: "default",
							},
						},
						Labels: []*schema.LabelSchema{
							{Name: "name"},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"type": {
									Constraint: schema.TypeDeclaration{},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`myblock "foo" "bar" {
  foo = count.index
  count = var.name
}
variable "name" {
  value = 4
}
`,
			hcl.Pos{Line: 3, Column: 16, Byte: 57},
			&lang.HoverData{
				Content: lang.Markdown("`var.name`\n_dynamic_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 3, Column: 11, Byte: 52},
					End:      hcl.Pos{Line: 3, Column: 19, Byte: 60},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			ctx := context.Background()

			f, diags := hclsyntax.ParseConfig([]byte(tc.config), "test.tf", hcl.InitialPos)
			if diags != nil {
				t.Fatal(diags)
			}

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
			origins, err := d.CollectReferenceOrigins()
			if err != nil {
				t.Fatal(err)
			}
			d = testPathDecoder(t, &PathContext{
				Schema: tc.bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
				ReferenceTargets: targets,
				ReferenceOrigins: origins,
			})

			data, err := d.HoverAtPos(ctx, "test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.expectedData, data, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("hover data mismatch: %s", diff)
			}
		})
	}
}
