package decoder

import (
	"errors"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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
