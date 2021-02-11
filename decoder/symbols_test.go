package decoder

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestDecoder_SymbolsInFile_emptyBody(t *testing.T) {
	d := NewDecoder()
	f := &hcl.File{
		Body: hcl.EmptyBody(),
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.SymbolsInFile("test.tf")
	unknownFormatErr := &UnknownFileFormatError{}
	if !errors.As(err, &unknownFormatErr) {
		t.Fatal("expected UnknownFileFormatError for empty body")
	}
}

func TestDecoder_SymbolsInFile_zeroByteContent(t *testing.T) {
	d := NewDecoder()
	f, pDiags := hclsyntax.ParseConfig([]byte{}, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	symbols, err := d.SymbolsInFile("test.tf")
	if err != nil {
		t.Fatal(err)
	}
	expectedSymbols := []Symbol{}
	if diff := cmp.Diff(expectedSymbols, symbols); diff != "" {
		t.Fatalf("unexpected symbols: %s", diff)
	}
}

func TestDecoder_SymbolsInFile_fileNotFound(t *testing.T) {
	d := NewDecoder()
	f, pDiags := hclsyntax.ParseConfig([]byte{}, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.SymbolsInFile("foobar.tf")
	notFoundErr := &FileNotFoundError{}
	if !errors.As(err, &notFoundErr) {
		t.Fatal("expected FileNotFoundError for non-existent file")
	}
}

func TestDecoder_SymbolsInFile_basic(t *testing.T) {
	d := NewDecoder()
	f, pDiags := hclsyntax.ParseConfig(testConfig, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}
	err := d.LoadFile("test.tf", f)
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
				"azurerm_subnet",
				"example",
			},
			rng: hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Column: 1, Line: 1, Byte: 0},
				End:      hcl.Pos{Column: 2, Line: 3, Byte: 51},
			},
			nestedSymbols: []Symbol{
				&AttributeSymbol{
					AttrName: "count",
					ExprKind: lang.LiteralTypeKind{Type: cty.Number},
					rng: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Column: 3, Line: 2, Byte: 40},
						End:      hcl.Pos{Column: 12, Line: 2, Byte: 49},
					},
					nestedSymbols: []Symbol{},
				},
			},
		},
		&BlockSymbol{
			Type: "resource",
			Labels: []string{
				"random_resource",
				"test",
			},
			rng: hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Column: 1, Line: 5, Byte: 53},
				End:      hcl.Pos{Column: 2, Line: 7, Byte: 101},
			},
			nestedSymbols: []Symbol{
				&AttributeSymbol{
					AttrName: "arg",
					ExprKind: lang.LiteralTypeKind{Type: cty.String},
					rng: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Column: 3, Line: 6, Byte: 91},
						End:      hcl.Pos{Column: 11, Line: 6, Byte: 99},
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

func TestDecoder_Symbols_basic(t *testing.T) {
	d := NewDecoder()

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

	err := d.LoadFile("first.tf", f1)
	if err != nil {
		t.Fatal(err)
	}
	err = d.LoadFile("second.tf", f2)
	if err != nil {
		t.Fatal(err)
	}

	symbols, err := d.Symbols("")
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
			rng: hcl.Range{
				Filename: "first.tf",
				Start:    hcl.Pos{Line: 2, Column: 1, Byte: 1},
				End:      hcl.Pos{Line: 4, Column: 2, Byte: 59},
			},
			nestedSymbols: []Symbol{
				&AttributeSymbol{
					AttrName: "cidr_block",
					ExprKind: lang.LiteralTypeKind{Type: cty.String},
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
			rng: hcl.Range{
				Filename: "second.tf",
				Start:    hcl.Pos{Line: 2, Column: 1, Byte: 1},
				End:      hcl.Pos{Line: 5, Column: 2, Byte: 84},
			},
			nestedSymbols: []Symbol{
				&AttributeSymbol{
					AttrName: "project",
					ExprKind: lang.LiteralTypeKind{Type: cty.String},
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

func TestDecoder_Symbols_expressions(t *testing.T) {
	d := NewDecoder()

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

	err := d.LoadFile("test.tf", f)
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
			rng: hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 2, Column: 1, Byte: 1},
				End:      hcl.Pos{Line: 11, Column: 2, Byte: 180},
			},
			nestedSymbols: []Symbol{
				&AttributeSymbol{
					AttrName: "subnet_ids",
					ExprKind: lang.TupleConsExprKind{},
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

func TestDecoder_Symbols_query(t *testing.T) {
	d := NewDecoder()

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

	err := d.LoadFile("first.tf", f1)
	if err != nil {
		t.Fatal(err)
	}
	err = d.LoadFile("second.tf", f2)
	if err != nil {
		t.Fatal(err)
	}

	symbols, err := d.Symbols("google")
	if err != nil {
		t.Fatal(err)
	}

	expectedSymbols := []Symbol{
		&BlockSymbol{
			Type: "provider",
			Labels: []string{
				"google",
			},
			rng: hcl.Range{
				Filename: "second.tf",
				Start:    hcl.Pos{Line: 2, Column: 1, Byte: 1},
				End:      hcl.Pos{Line: 5, Column: 2, Byte: 84},
			},
			nestedSymbols: []Symbol{
				&AttributeSymbol{
					AttrName: "project",
					ExprKind: lang.LiteralTypeKind{Type: cty.String},
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
