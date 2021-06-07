package decoder

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
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

func TestTraversalToAddress(t *testing.T) {
	testCases := []struct {
		rawTraversal string
		expectedAddr lang.Address
	}{
		{
			"one",
			lang.Address{
				lang.RootStep{Name: "one"},
			},
		},
		{
			"first.second",
			lang.Address{
				lang.RootStep{Name: "first"},
				lang.AttrStep{Name: "second"},
			},
		},
		{
			"foo[2]",
			lang.Address{
				lang.RootStep{Name: "foo"},
				lang.IndexStep{Key: cty.NumberIntVal(2)},
			},
		},
		{
			`foo["bar"]`,
			lang.Address{
				lang.RootStep{Name: "foo"},
				lang.IndexStep{Key: cty.StringVal("bar")},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			traversal, diags := hclsyntax.ParseTraversalAbs([]byte(tc.rawTraversal), "test.tf", hcl.InitialPos)
			if len(diags) > 0 {
				t.Fatal(diags)
			}

			addr, err := lang.TraversalToAddress(traversal)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedAddr, addr, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("address mismatch: %s", diff)
			}
		})
	}
}
