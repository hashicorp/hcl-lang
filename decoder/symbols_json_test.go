package decoder

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/zclconf/go-cty/cty"
)

func TestDecoder_Symbols_json(t *testing.T) {
	f, diags := json.Parse([]byte(``), "test.tf.json")
	if len(diags) == 0 {
		t.Fatal("expected empty JSON file to fail parsing")
	}

	dirPath := t.TempDir()
	d := NewDecoder(&testPathReader{
		paths: map[string]*PathContext{
			dirPath: {
				Files: map[string]*hcl.File{
					"test.tf.json": f,
				},
			},
		},
	})

	symbols, err := d.Symbols(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(symbols) != 0 {
		t.Fatal("expected zero symbols for empty file")
	}
}

func TestDecoder_Symbols_json_emptyFile(t *testing.T) {
	f, diags := json.Parse([]byte(``), "test.tf.json")
	if len(diags) == 0 {
		t.Fatal("expected empty JSON file to fail parsing")
	}

	dirPath := t.TempDir()
	d := NewDecoder(&testPathReader{
		paths: map[string]*PathContext{
			dirPath: {
				Files: map[string]*hcl.File{
					"test.tf.json": f,
				},
			},
		},
	})

	symbols, err := d.Symbols(context.Background(), "test.tf.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(symbols) != 0 {
		t.Fatal("expected zero symbols for empty file")
	}
}

func TestDecoder_Symbols_json_emptyBody(t *testing.T) {
	f, diags := json.Parse([]byte(`{}`), "test.tf.json")
	if len(diags) > 0 {
		t.Fatal(diags)
	}

	dirPath := t.TempDir()
	d := NewDecoder(&testPathReader{
		paths: map[string]*PathContext{
			dirPath: {
				Files: map[string]*hcl.File{
					"test.tf.json": f,
				},
			},
		},
	})

	symbols, err := d.Symbols(context.Background(), "test.tf.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(symbols) != 0 {
		t.Fatal("expected zero symbols for empty body")
	}
}

func TestDecoder_Symbols_json_basic(t *testing.T) {
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"resource": {
				Labels: []*schema.LabelSchema{
					{Name: "type"},
					{Name: "name"},
				},
				Body: &schema.BodySchema{
					Attributes: map[string]*schema.AttributeSchema{
						"cidr_block": {Expr: schema.LiteralTypeOnly(cty.String)},
					},
				},
			},
			"provider": {
				Labels: []*schema.LabelSchema{
					{Name: "name"},
				},
				Body: &schema.BodySchema{
					Attributes: map[string]*schema.AttributeSchema{
						"project": {Expr: schema.LiteralTypeOnly(cty.String)},
						"region":  {Expr: schema.LiteralTypeOnly(cty.String), IsRequired: true},
					},
				},
			},
		},
	}

	testCfg1 := []byte(`{
  "resource": {
    "aws_vpc": {
      "main": {
        "cidr_block": "10.0.0.0/16"
      }
    }
  }
}`)

	f1, pDiags := json.Parse(testCfg1, "first.tf.json")
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	testCfg2 := []byte(`{
  "provider": {
    "google": {
      "project": "my-project-id",
      "region": "us-central1"
    }
  }
}`)
	f2, pDiags := json.Parse(testCfg2, "second.tf.json")
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	dirPath := t.TempDir()
	d := NewDecoder(&testPathReader{
		paths: map[string]*PathContext{
			dirPath: {
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"first.tf.json":  f1,
					"second.tf.json": f2,
				},
			},
		},
	})

	symbols, err := d.Symbols(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}

	expectedSymbols := []Symbol{
		&BlockSymbol{
			Type: "resource",
			Labels: []string{
				"aws_vpc",
				"main",
			},
			path: lang.Path{Path: dirPath},
			rng: hcl.Range{
				Filename: "first.tf.json",
				Start:    hcl.Pos{Line: 4, Column: 15, Byte: 49},
				End:      hcl.Pos{Line: 6, Column: 8, Byte: 94},
			},
			nestedSymbols: []Symbol{
				&AttributeSymbol{
					AttrName: "cidr_block",
					path:     lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "first.tf.json",
						Start:    hcl.Pos{Line: 5, Column: 9, Byte: 59},
						End:      hcl.Pos{Line: 5, Column: 36, Byte: 86},
					},
					nestedSymbols: []Symbol{},
				},
			},
		},
		&BlockSymbol{
			Type: "provider",
			Labels: []string{
				"google",
			},
			path: lang.Path{Path: dirPath},
			rng: hcl.Range{
				Filename: "second.tf.json",
				Start:    hcl.Pos{Line: 3, Column: 15, Byte: 32},
				End:      hcl.Pos{Line: 6, Column: 6, Byte: 103},
			},
			nestedSymbols: []Symbol{
				&AttributeSymbol{
					AttrName: "project",
					path:     lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "second.tf.json",
						Start:    hcl.Pos{Line: 4, Column: 7, Byte: 40},
						End:      hcl.Pos{Line: 4, Column: 33, Byte: 66},
					},
					nestedSymbols: []Symbol{},
				},
				&AttributeSymbol{
					AttrName: "region",
					path:     lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "second.tf.json",
						Start:    hcl.Pos{Line: 5, Column: 7, Byte: 74},
						End:      hcl.Pos{Line: 5, Column: 30, Byte: 97},
					},
					nestedSymbols: []Symbol{},
				},
			},
		},
	}

	diff := cmp.Diff(expectedSymbols, symbols)
	if diff != "" {
		t.Fatalf("unexpected symbols: %s", diff)
	}
}

func TestDecoder_Symbols_json_dependentBody(t *testing.T) {
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"resource": {
				Labels: []*schema.LabelSchema{
					{Name: "type", IsDepKey: true},
					{Name: "name"},
				},
				Body: &schema.BodySchema{},
				DependentBody: map[schema.SchemaKey]*schema.BodySchema{
					schema.NewSchemaKey(schema.DependencyKeys{
						Labels: []schema.LabelDependent{
							{
								Index: 0,
								Value: "aws_instance",
							},
						},
					}): {
						Attributes: map[string]*schema.AttributeSchema{
							"subnet_ids": {
								IsOptional: true,
								Expr: schema.ExprConstraints{
									schema.ListExpr{Elem: schema.LiteralTypeOnly(cty.String)},
								},
							},
							"random_kw": {
								IsOptional: true,
								Expr: schema.ExprConstraints{
									schema.KeywordExpr{
										Keyword: "foo",
									},
								},
							},
						},
						Blocks: map[string]*schema.BlockSchema{
							"configuration": {
								Type: schema.BlockTypeObject,
								Body: &schema.BodySchema{
									Attributes: map[string]*schema.AttributeSchema{
										"name":     {IsOptional: true, Expr: schema.LiteralTypeOnly(cty.String)},
										"num":      {IsOptional: true, Expr: schema.LiteralTypeOnly(cty.Number)},
										"boolattr": {IsOptional: true, Expr: schema.LiteralTypeOnly(cty.Bool)},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	testCfg := []byte(`{
  "resource": {
    "aws_instance": {
      "test": {
        "subnet_ids": [ "one-1", "two-2" ],
        "configuration": {
          "name": "blah",
          "num": 42,
          "boolattr": true
        },
        "random_kw": "${foo}"
      }
    }
  }
}`)
	f, pDiags := json.Parse(testCfg, "test.tf.json")
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	dirPath := t.TempDir()
	d := NewDecoder(&testPathReader{
		paths: map[string]*PathContext{
			dirPath: {
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.tf.json": f,
				},
			},
		},
	})

	symbols, err := d.Symbols(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}

	expectedSymbols := []Symbol{
		&BlockSymbol{
			Type:   "resource",
			Labels: []string{"aws_instance", "test"},
			path:   lang.Path{Path: dirPath},
			rng: hcl.Range{
				Filename: "test.tf.json",
				Start:    hcl.Pos{Line: 4, Column: 15, Byte: 54},
				End:      hcl.Pos{Line: 12, Column: 8, Byte: 249},
			},
			nestedSymbols: []Symbol{
				&AttributeSymbol{
					AttrName: "subnet_ids",
					path:     lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 5, Column: 9, Byte: 64},
						End:      hcl.Pos{Line: 5, Column: 43, Byte: 98},
					},
					nestedSymbols: []Symbol{},
				},
				&BlockSymbol{
					Type:   "configuration",
					Labels: []string{},
					path:   lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 6, Column: 26, Byte: 125},
						End:      hcl.Pos{Line: 10, Column: 10, Byte: 210},
					},
					nestedSymbols: []Symbol{
						&AttributeSymbol{
							AttrName: "name",
							path:     lang.Path{Path: dirPath},
							rng: hcl.Range{
								Filename: "test.tf.json",
								Start:    hcl.Pos{Line: 7, Column: 11, Byte: 137},
								End:      hcl.Pos{Line: 7, Column: 25, Byte: 151},
							},
							nestedSymbols: []Symbol{},
						},
						&AttributeSymbol{
							AttrName: "num",
							path:     lang.Path{Path: dirPath},
							rng: hcl.Range{
								Filename: "test.tf.json",
								Start:    hcl.Pos{Line: 8, Column: 11, Byte: 163},
								End:      hcl.Pos{Line: 8, Column: 20, Byte: 172},
							},
							nestedSymbols: []Symbol{},
						},
						&AttributeSymbol{
							AttrName: "boolattr",
							path:     lang.Path{Path: dirPath},
							rng: hcl.Range{
								Filename: "test.tf.json",
								Start:    hcl.Pos{Line: 9, Column: 11, Byte: 184},
								End:      hcl.Pos{Line: 9, Column: 27, Byte: 200},
							},
							nestedSymbols: []Symbol{},
						},
					},
				},
				&AttributeSymbol{
					AttrName: "random_kw",
					path:     lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 11, Column: 9, Byte: 220},
						End:      hcl.Pos{Line: 11, Column: 30, Byte: 241},
					},
					nestedSymbols: []Symbol{},
				},
			},
		},
	}

	diff := cmp.Diff(expectedSymbols, symbols)
	if diff != "" {
		t.Fatalf("unexpected symbols:\n%s", diff)
	}
}

func TestDecoder_Symbols_json_unknownExpression(t *testing.T) {
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"resource": {
				Labels: []*schema.LabelSchema{
					{Name: "type", IsDepKey: true},
					{Name: "name"},
				},
				Body: &schema.BodySchema{
					Attributes: map[string]*schema.AttributeSchema{
						"subnet_ids": {
							IsOptional: true,
							Expr: schema.ExprConstraints{
								schema.ListExpr{Elem: schema.LiteralTypeOnly(cty.String)},
							},
						},
						"random_kw": {
							IsOptional: true,
							Expr: schema.ExprConstraints{
								schema.KeywordExpr{
									Keyword: "foo",
								},
							},
						},
					},
					Blocks: map[string]*schema.BlockSchema{
						"configuration": {
							Type: schema.BlockTypeObject,
							Body: &schema.BodySchema{
								Attributes: map[string]*schema.AttributeSchema{
									"name":     {IsOptional: true, Expr: schema.LiteralTypeOnly(cty.String)},
									"num":      {IsOptional: true, Expr: schema.LiteralTypeOnly(cty.Number)},
									"boolattr": {IsOptional: true, Expr: schema.LiteralTypeOnly(cty.Bool)},
								},
							},
						},
					},
				},
			},
		},
	}

	testCfg := []byte(`{
  "resource": {
    "aws_instance": {
      "test": {
        "subnet_ids": [ "${var.test}", "two-2" ],
        "configuration": {
          "${var.key}": "blah",
          "num": "${var.value}",
          "${var.env}.${another}": "prod",
          "${foo(var.arg)}": "bar"
        },
        "random_kw": "${var.value}"
      }
    }
  }
}`)
	f, pDiags := json.Parse(testCfg, "test.tf.json")
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	dirPath := t.TempDir()
	d := NewDecoder(&testPathReader{
		paths: map[string]*PathContext{
			dirPath: {
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"first.tf.json": f,
				},
			},
		},
	})

	symbols, err := d.Symbols(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}

	expectedSymbols := []Symbol{
		&BlockSymbol{
			Type:   "resource",
			Labels: []string{"aws_instance", "test"},
			path:   lang.Path{Path: dirPath},
			rng: hcl.Range{
				Filename: "test.tf.json",
				Start:    hcl.Pos{Line: 4, Column: 15, Byte: 54},
				End:      hcl.Pos{Line: 13, Column: 8, Byte: 330},
			},
			nestedSymbols: []Symbol{
				&AttributeSymbol{
					AttrName: "subnet_ids",
					path:     lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 5, Column: 9, Byte: 64},
						End:      hcl.Pos{Line: 5, Column: 49, Byte: 104},
					},
					nestedSymbols: []Symbol{},
				},
				&BlockSymbol{
					Type:   "configuration",
					Labels: []string{},
					path:   lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 6, Column: 26, Byte: 131},
						End:      hcl.Pos{Line: 11, Column: 10, Byte: 285},
					},
					nestedSymbols: []Symbol{
						&AttributeSymbol{
							AttrName: "num",
							path:     lang.Path{Path: dirPath},
							rng: hcl.Range{
								Filename: "test.tf.json",
								Start:    hcl.Pos{Line: 8, Column: 11, Byte: 175},
								End:      hcl.Pos{Line: 8, Column: 32, Byte: 196},
							},
							nestedSymbols: []Symbol{},
						},
					},
				},
				&AttributeSymbol{
					AttrName: "random_kw",
					path:     lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 12, Column: 9, Byte: 295},
						End:      hcl.Pos{Line: 12, Column: 36, Byte: 322},
					},
					nestedSymbols: []Symbol{},
				},
			},
		},
	}

	diff := cmp.Diff(expectedSymbols, symbols)
	if diff != "" {
		t.Fatalf("unexpected symbols:\n%s", diff)
	}
}
