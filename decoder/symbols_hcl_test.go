package decoder

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestDecoder_SymbolsInFile_hcl_zeroByteContent(t *testing.T) {
	f, pDiags := hclsyntax.ParseConfig([]byte{}, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	d := testPathDecoder(t, &PathContext{
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})

	symbols, err := d.SymbolsInFile("test.tf")
	if err != nil {
		t.Fatal(err)
	}
	expectedSymbols := []Symbol{}
	if diff := cmp.Diff(expectedSymbols, symbols); diff != "" {
		t.Fatalf("unexpected symbols: %s", diff)
	}
}

func TestDecoder_Symbols_hcl_basic(t *testing.T) {
	testCfg1 := []byte(`
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}
`)
	f1, pDiags := hclsyntax.ParseConfig(testCfg1, "first.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	testCfg2 := []byte(`
provider "google" {
  project     = "my-project-id"
  region      = "us-central1"
}
`)
	f2, pDiags := hclsyntax.ParseConfig(testCfg2, "second.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	dirPath := t.TempDir()
	d := NewDecoder(&testPathReader{
		paths: map[string]*PathContext{
			dirPath: {
				Files: map[string]*hcl.File{
					"first.tf":  f1,
					"second.tf": f2,
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
				Filename: "first.tf",
				Start:    hcl.Pos{Line: 2, Column: 1, Byte: 1},
				End:      hcl.Pos{Line: 4, Column: 2, Byte: 59},
			},
			nestedSymbols: []Symbol{
				&AttributeSymbol{
					AttrName: "cidr_block",
					ExprKind: lang.LiteralTypeKind{Type: cty.String},
					path:     lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "first.tf",
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 31},
						End:      hcl.Pos{Line: 3, Column: 29, Byte: 57},
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
				Filename: "second.tf",
				Start:    hcl.Pos{Line: 2, Column: 1, Byte: 1},
				End:      hcl.Pos{Line: 5, Column: 2, Byte: 84},
			},
			nestedSymbols: []Symbol{
				&AttributeSymbol{
					AttrName: "project",
					ExprKind: lang.LiteralTypeKind{Type: cty.String},
					path:     lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "second.tf",
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 23},
						End:      hcl.Pos{Line: 3, Column: 32, Byte: 52},
					},
					nestedSymbols: []Symbol{},
				},
				&AttributeSymbol{
					AttrName: "region",
					ExprKind: lang.LiteralTypeKind{Type: cty.String},
					path:     lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "second.tf",
						Start:    hcl.Pos{Line: 4, Column: 3, Byte: 55},
						End:      hcl.Pos{Line: 4, Column: 30, Byte: 82},
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

func TestDecoder_SymbolsInFile_hcl(t *testing.T) {
	testCfg := []byte(`
resource "aws_instance" "test" {
  subnet_ids = [ "one-1", "two-2" ]
  configuration = {
  	name = "blah"
  	num = 42
  	boolattr = true
  	foo(42) = "bar"
  }
  random_kw = foo
}
`)
	f, pDiags := hclsyntax.ParseConfig(testCfg, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	dirPath := t.TempDir()
	d, err := NewDecoder(&testPathReader{
		paths: map[string]*PathContext{
			dirPath: {
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			},
		},
	}).Path(lang.Path{Path: dirPath})
	if err != nil {
		t.Fatal(err)
	}

	symbols, err := d.SymbolsInFile("test.tf")
	if err != nil {
		t.Fatal(err)
	}

	expectedSymbols := []Symbol{
		&BlockSymbol{
			Type: "resource",
			Labels: []string{
				"aws_instance",
				"test",
			},
			path: lang.Path{Path: dirPath},
			rng: hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 2, Column: 1, Byte: 1},
				End:      hcl.Pos{Line: 11, Column: 2, Byte: 180},
			},
			nestedSymbols: []Symbol{
				&AttributeSymbol{
					AttrName: "subnet_ids",
					ExprKind: lang.TupleConsExprKind{},
					path:     lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   3,
							Column: 3,
							Byte:   36,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 36,
							Byte:   69,
						},
					},
					nestedSymbols: []Symbol{
						&ExprSymbol{
							ExprName: "0",
							ExprKind: lang.LiteralTypeKind{
								Type: cty.String,
							},
							path: lang.Path{Path: dirPath},
							rng: hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 18,
									Byte:   51,
								},
								End: hcl.Pos{
									Line:   3,
									Column: 25,
									Byte:   58,
								},
							},
							nestedSymbols: []Symbol{},
						},
						&ExprSymbol{
							ExprName: "1",
							ExprKind: lang.LiteralTypeKind{
								Type: cty.String,
							},
							path: lang.Path{Path: dirPath},
							rng: hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 27,
									Byte:   60,
								},
								End: hcl.Pos{
									Line:   3,
									Column: 34,
									Byte:   67,
								},
							},
							nestedSymbols: []Symbol{},
						},
					},
				},
				&AttributeSymbol{
					AttrName: "configuration",
					ExprKind: lang.ObjectConsExprKind{},
					path:     lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   4,
							Column: 3,
							Byte:   72,
						},
						End: hcl.Pos{
							Line:   9,
							Column: 4,
							Byte:   160,
						},
					},
					nestedSymbols: []Symbol{
						&ExprSymbol{
							ExprName: "name",
							ExprKind: lang.LiteralTypeKind{
								Type: cty.String,
							},
							path: lang.Path{Path: dirPath},
							rng: hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   5,
									Column: 4,
									Byte:   93,
								},
								End: hcl.Pos{
									Line:   5,
									Column: 17,
									Byte:   106,
								},
							},
							nestedSymbols: []Symbol{},
						},
						&ExprSymbol{
							ExprName: "num",
							ExprKind: lang.LiteralTypeKind{
								Type: cty.Number,
							},
							path: lang.Path{Path: dirPath},
							rng: hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   6,
									Column: 4,
									Byte:   110,
								},
								End: hcl.Pos{
									Line:   6,
									Column: 12,
									Byte:   118,
								},
							},
							nestedSymbols: []Symbol{},
						},
						&ExprSymbol{
							ExprName: "boolattr",
							ExprKind: lang.LiteralTypeKind{
								Type: cty.Bool,
							},
							path: lang.Path{Path: dirPath},
							rng: hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   7,
									Column: 4,
									Byte:   122,
								},
								End: hcl.Pos{
									Line:   7,
									Column: 19,
									Byte:   137,
								},
							},
							nestedSymbols: []Symbol{},
						},
					},
				},
				&AttributeSymbol{
					AttrName: "random_kw",
					ExprKind: lang.TraversalExprKind{},
					path:     lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   10,
							Column: 3,
							Byte:   163,
						},
						End: hcl.Pos{
							Line:   10,
							Column: 18,
							Byte:   178,
						},
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

func TestDecoder_SymbolsInFile_hcl_unknownExpression(t *testing.T) {
	testCfg := []byte(`
resource "aws_instance" "test" {
  subnet_ids = [ var.test, "two-2" ]
  configuration = {
  	var.key = "blah"
  	num = var.value
    "${var.env}.${another}" = "prod"
  	foo(var.arg) = "bar"
  }
  random_kw = var.value
}
`)
	f, pDiags := hclsyntax.ParseConfig(testCfg, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	dirPath := t.TempDir()
	d, err := NewDecoder(&testPathReader{
		paths: map[string]*PathContext{
			dirPath: {
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			},
		},
	}).Path(lang.Path{Path: dirPath})
	if err != nil {
		t.Fatal(err)
	}

	symbols, err := d.SymbolsInFile("test.tf")
	if err != nil {
		t.Fatal(err)
	}

	expectedSymbols := []Symbol{
		&BlockSymbol{
			Type: "resource",
			Labels: []string{
				"aws_instance",
				"test",
			},
			path: lang.Path{Path: dirPath},
			rng: hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 2, Column: 1, Byte: 1},
				End:      hcl.Pos{Line: 11, Column: 2, Byte: 220},
			},
			nestedSymbols: []Symbol{
				&AttributeSymbol{
					AttrName: "subnet_ids",
					ExprKind: lang.TupleConsExprKind{},
					path:     lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   3,
							Column: 3,
							Byte:   36,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 37,
							Byte:   70,
						},
					},
					nestedSymbols: []Symbol{
						&ExprSymbol{
							ExprName: "0",
							ExprKind: lang.TraversalExprKind{},
							path:     lang.Path{Path: dirPath},
							rng: hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 18,
									Byte:   51,
								},
								End: hcl.Pos{
									Line:   3,
									Column: 26,
									Byte:   59,
								},
							},
							nestedSymbols: []Symbol{},
						},
						&ExprSymbol{
							ExprName: "1",
							ExprKind: lang.LiteralTypeKind{
								Type: cty.String,
							},
							path: lang.Path{Path: dirPath},
							rng: hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 28,
									Byte:   61,
								},
								End: hcl.Pos{
									Line:   3,
									Column: 35,
									Byte:   68,
								},
							},
							nestedSymbols: []Symbol{},
						},
					},
				},
				&AttributeSymbol{
					AttrName: "configuration",
					ExprKind: lang.ObjectConsExprKind{},
					path:     lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   4,
							Column: 3,
							Byte:   73,
						},
						End: hcl.Pos{
							Line:   9,
							Column: 4,
							Byte:   194,
						},
					},
					nestedSymbols: []Symbol{
						&ExprSymbol{
							ExprName: "num",
							ExprKind: lang.TraversalExprKind{},
							path:     lang.Path{Path: dirPath},
							rng: hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   6,
									Column: 4,
									Byte:   114,
								},
								End: hcl.Pos{
									Line:   6,
									Column: 19,
									Byte:   129,
								},
							},
							nestedSymbols: []Symbol{},
						},
					},
				},
				&AttributeSymbol{
					AttrName: "random_kw",
					ExprKind: lang.TraversalExprKind{},
					path:     lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   10,
							Column: 3,
							Byte:   197,
						},
						End: hcl.Pos{
							Line:   10,
							Column: 24,
							Byte:   218,
						},
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

func TestDecoder_Symbols_hcl_query(t *testing.T) {
	testCfg1 := []byte(`
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}
`)
	f1, pDiags := hclsyntax.ParseConfig(testCfg1, "first.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	testCfg2 := []byte(`
provider "google" {
  project     = "my-project-id"
  region      = "us-central1"
}
`)
	f2, pDiags := hclsyntax.ParseConfig(testCfg2, "second.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	dirPath := t.TempDir()
	d := NewDecoder(&testPathReader{
		paths: map[string]*PathContext{
			dirPath: {
				Files: map[string]*hcl.File{
					"first.tf":  f1,
					"second.tf": f2,
				},
			},
		},
	})

	symbols, err := d.Symbols(context.Background(), "google")
	if err != nil {
		t.Fatal(err)
	}

	expectedSymbols := []Symbol{
		&BlockSymbol{
			Type: "provider",
			Labels: []string{
				"google",
			},
			path: lang.Path{Path: dirPath},
			rng: hcl.Range{
				Filename: "second.tf",
				Start:    hcl.Pos{Line: 2, Column: 1, Byte: 1},
				End:      hcl.Pos{Line: 5, Column: 2, Byte: 84},
			},
			nestedSymbols: []Symbol{
				&AttributeSymbol{
					AttrName: "project",
					ExprKind: lang.LiteralTypeKind{Type: cty.String},
					path:     lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "second.tf",
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 23},
						End:      hcl.Pos{Line: 3, Column: 32, Byte: 52},
					},
					nestedSymbols: []Symbol{},
				},
				&AttributeSymbol{
					AttrName: "region",
					ExprKind: lang.LiteralTypeKind{Type: cty.String},
					path:     lang.Path{Path: dirPath},
					rng: hcl.Range{
						Filename: "second.tf",
						Start:    hcl.Pos{Line: 4, Column: 3, Byte: 55},
						End:      hcl.Pos{Line: 4, Column: 30, Byte: 82},
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
