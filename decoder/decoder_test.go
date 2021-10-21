package decoder

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

type testPathReader struct {
	paths map[string]*PathContext
}

func (r *testPathReader) Paths(ctx context.Context) []lang.Path {
	paths := make([]lang.Path, len(r.paths))

	i := 0
	for path := range r.paths {
		paths[i] = lang.Path{Path: path}
		i++
	}

	return paths
}

func (r *testPathReader) PathContext(path lang.Path) (*PathContext, error) {
	return r.paths[path.Path], nil
}

func testPathDecoder(t *testing.T, pathCtx *PathContext) *PathDecoder {
	dirPath := t.TempDir()
	dirs := map[string]*PathContext{
		dirPath: pathCtx,
	}

	d := NewDecoder(&testPathReader{
		paths: dirs,
	})

	pathDecoder, err := d.Path(lang.Path{Path: dirPath})
	if err != nil {
		t.Fatal(err)
	}

	return pathDecoder
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
