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
