package reference

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestTraversalToOrigin(t *testing.T) {
	testCases := []struct {
		rawTraversal   string
		traversalExprs schema.TraversalExprs
		expectedOrigin Origin
	}{
		{
			"one",
			schema.TraversalExprs{},
			LocalOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "one"},
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
				},
			},
		},
		{
			"first.second",
			schema.TraversalExprs{},
			LocalOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "first"},
					lang.AttrStep{Name: "second"},
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
				},
			},
		},
		{
			"foo[2]",
			schema.TraversalExprs{},
			LocalOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "foo"},
					lang.IndexStep{Key: cty.NumberIntVal(2)},
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:      hcl.Pos{Line: 1, Column: 7, Byte: 6},
				},
			},
		},
		{
			`foo["bar"]`,
			schema.TraversalExprs{},
			LocalOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "foo"},
					lang.IndexStep{Key: cty.StringVal("bar")},
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			traversal, diags := hclsyntax.ParseTraversalAbs([]byte(tc.rawTraversal), "test.tf", hcl.InitialPos)
			if len(diags) > 0 {
				t.Fatal(diags)
			}

			origin, err := TraversalToLocalOrigin(traversal, tc.traversalExprs)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedOrigin, origin, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("origin mismatch: %s", diff)
			}
		})
	}
}
