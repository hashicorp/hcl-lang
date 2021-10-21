package decoder

import (
	"errors"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func TestDecoder_SymbolsInFile_emptyBody(t *testing.T) {
	f := &hcl.File{
		Body: hcl.EmptyBody(),
	}

	d := testPathDecoder(t, &PathContext{
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})

	_, err := d.SymbolsInFile("test.tf")
	unknownFormatErr := &UnknownFileFormatError{}
	if !errors.As(err, &unknownFormatErr) {
		t.Fatal("expected UnknownFileFormatError for empty body")
	}
}

func TestDecoder_SymbolsInFile_fileNotFound(t *testing.T) {
	f, pDiags := hclsyntax.ParseConfig([]byte{}, "test.tf", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	d := testPathDecoder(t, &PathContext{
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})

	_, err := d.SymbolsInFile("foobar.tf")
	notFoundErr := &FileNotFoundError{}
	if !errors.As(err, &notFoundErr) {
		t.Fatal("expected FileNotFoundError for non-existent file")
	}
}
