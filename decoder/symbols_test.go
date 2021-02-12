package decoder

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
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
					Type:     cty.Number,
					rng: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Column: 3, Line: 2, Byte: 40},
						End:      hcl.Pos{Column: 12, Line: 2, Byte: 49},
					},
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
					Type:     cty.String,
					rng: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Column: 3, Line: 6, Byte: 91},
						End:      hcl.Pos{Column: 11, Line: 6, Byte: 99},
					},
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
					Type:     cty.String,
					rng: hcl.Range{
						Filename: "first.tf",
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 31},
						End:      hcl.Pos{Line: 3, Column: 29, Byte: 57},
					},
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
					Type:     cty.String,
					rng: hcl.Range{
						Filename: "second.tf",
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 23},
						End:      hcl.Pos{Line: 3, Column: 32, Byte: 52},
					},
				},
				&AttributeSymbol{
					AttrName: "region",
					Type:     cty.String,
					rng: hcl.Range{
						Filename: "second.tf",
						Start:    hcl.Pos{Line: 4, Column: 3, Byte: 55},
						End:      hcl.Pos{Line: 4, Column: 30, Byte: 82},
					},
				},
			},
		},
	}

	diff := cmp.Diff(expectedSymbols, symbols)
	if diff != "" {
		t.Fatalf("unexpected symbols: %s", diff)
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
					Type:     cty.String,
					rng: hcl.Range{
						Filename: "second.tf",
						Start:    hcl.Pos{Line: 3, Column: 3, Byte: 23},
						End:      hcl.Pos{Line: 3, Column: 32, Byte: 52},
					},
				},
				&AttributeSymbol{
					AttrName: "region",
					Type:     cty.String,
					rng: hcl.Range{
						Filename: "second.tf",
						Start:    hcl.Pos{Line: 4, Column: 3, Byte: 55},
						End:      hcl.Pos{Line: 4, Column: 30, Byte: 82},
					},
				},
			},
		},
	}

	diff := cmp.Diff(expectedSymbols, symbols)
	if diff != "" {
		t.Fatalf("unexpected symbols: %s", diff)
	}
}
