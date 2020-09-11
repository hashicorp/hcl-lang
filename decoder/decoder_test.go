package decoder

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
)

func TestDecoder_LoadFile_nilFile(t *testing.T) {
	d := NewDecoder()
	err := d.LoadFile("test.tf", nil)
	if err == nil {
		t.Fatal("expected error for nil file")
	}
	if diff := cmp.Diff(`test.tf: invalid content provided`, err.Error()); diff != "" {
		t.Fatalf("unexpected error: %s", diff)
	}
}

func TestDecoder_LoadFile_nilRootBody(t *testing.T) {
	d := NewDecoder()
	f := &hcl.File{
		Body: nil,
	}
	err := d.LoadFile("test.tf", f)
	if err == nil {
		t.Fatal("expected error for nil body")
	}
	if diff := cmp.Diff(`test.tf: file has no body`, err.Error()); diff != "" {
		t.Fatalf("unexpected error: %s", diff)
	}
}
