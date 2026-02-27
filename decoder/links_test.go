// Copyright IBM Corp. 2020, 2026
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/zclconf/go-cty/cty"
)

func TestLinksInFileBlock(t *testing.T) {
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
				"str_attr": {Constraint: schema.LiteralType{Type: cty.String}, Description: lang.PlainText("Special attribute")},
			},
		},
		DependentBody: map[schema.SchemaKey]*schema.BodySchema{
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "sushi"},
				},
			}): {
				DocsLink:    &schema.DocsLink{URL: "https://en.wikipedia.org/wiki/Sushi"},
				Detail:      "rice, fish etc.",
				Description: lang.Markdown("Sushi, the Rolls-Rice of Japanese cuisine"),
			},
			schema.NewSchemaKey(schema.DependencyKeys{
				Labels: []schema.LabelDependent{
					{Index: 0, Value: "ramen"},
				},
			}): {
				DocsLink:    &schema.DocsLink{URL: "https://en.wikipedia.org/wiki/Ramen"},
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
}
`)

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

	links, err := d.LinksInFile("test.tf")
	if err != nil {
		t.Fatal(err)
	}

	expectedLinks := []lang.Link{
		{
			URI: "https://en.wikipedia.org/wiki/Sushi",
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
	}

	diff := cmp.Diff(expectedLinks, links)
	if diff != "" {
		t.Fatalf("unexpected links: %s", diff)
	}
}

func TestLinksInFileAttribute(t *testing.T) {
	resourceLabelSchema := []*schema.LabelSchema{
		{Name: "name"},
	}
	blockSchema := &schema.BlockSchema{
		Labels:      resourceLabelSchema,
		Description: lang.Markdown("My special block"),
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"num_attr": {Constraint: schema.LiteralType{Type: cty.Number}},
				"source": {
					Constraint:  schema.LiteralType{Type: cty.String},
					Description: lang.PlainText("Special attribute"),
					IsDepKey:    true,
				},
			},
		},
		DependentBody: map[schema.SchemaKey]*schema.BodySchema{
			schema.NewSchemaKey(schema.DependencyKeys{
				Attributes: []schema.AttributeDependent{
					{
						Name: "source",
						Expr: schema.ExpressionValue{
							Static: cty.StringVal("example.com/source"),
						},
					},
				},
			}): {
				DocsLink: &schema.DocsLink{URL: "https://example.com/some/source"},
			},
		},
	}
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"myblock": blockSchema,
		},
	}
	testConfig := []byte(`myblock "example" {
  source = "example.com/source"
  num_attr = 4
}
`)

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

	links, err := d.LinksInFile("test.tf")
	if err != nil {
		t.Fatal(err)
	}

	expectedLinks := []lang.Link{
		{
			URI: "https://example.com/some/source",
			Range: hcl.Range{
				Filename: "test.tf",
				Start: hcl.Pos{
					Line:   2,
					Column: 12,
					Byte:   31,
				},
				End: hcl.Pos{
					Line:   2,
					Column: 32,
					Byte:   51,
				},
			},
		},
	}

	diff := cmp.Diff(expectedLinks, links)
	if diff != "" {
		t.Fatalf("unexpected links: %s", diff)
	}
}

func TestLinksInFile_json(t *testing.T) {
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

	// We never want to provide links in JSON configs
	_, err := d.LinksInFile("test.tf.json")
	unknownFormatErr := &UnknownFileFormatError{}
	if !errors.As(err, &unknownFormatErr) {
		t.Fatal("expected UnknownFileFormatError for JSON body")
	}
}
