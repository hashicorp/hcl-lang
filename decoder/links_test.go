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

func TestLinksInFile(t *testing.T) {
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
				"str_attr": {Expr: schema.LiteralTypeOnly(cty.String), Description: lang.PlainText("Special attribute")},
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

	d := NewDecoder()
	f, pDiags := hclsyntax.ParseConfig(testConfig, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	d.SetSchema(bodySchema)

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
