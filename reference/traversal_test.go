// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

func TestLegacyTraversalToOrigin(t *testing.T) {
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

			origin, err := LegacyTraversalToLocalOrigin(traversal, tc.traversalExprs)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedOrigin, origin, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("origin mismatch: %s", diff)
			}
		})
	}
}

func TestLegacyTraversalsToOrigin(t *testing.T) {
	testCases := []struct {
		testName        string
		rawTraversals   []string
		traversalExprs  schema.TraversalExprs
		allowSelfRefs   bool
		expectedOrigins Origins
	}{
		{
			"origin collection without self refs",
			[]string{"foo.bar", "self.bar"},
			schema.TraversalExprs{},
			false,
			Origins{
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
					},
				},
			},
		},
		{
			"origin collection with self refs",
			[]string{"foo.bar", "self.bar"},
			schema.TraversalExprs{},
			true,
			Origins{
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 8, Byte: 7},
					},
				},
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "self"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 9, Byte: 8},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			traversals := make([]hcl.Traversal, 0)
			for _, rawTraversal := range tc.rawTraversals {
				traversal, diags := hclsyntax.ParseTraversalAbs([]byte(rawTraversal), "test.tf", hcl.InitialPos)
				if len(diags) > 0 {
					t.Fatal(diags)
				}
				traversals = append(traversals, traversal)
			}

			origins := LegacyTraversalsToLocalOrigins(traversals, tc.traversalExprs, tc.allowSelfRefs)
			if diff := cmp.Diff(tc.expectedOrigins, origins, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("origin mismatch: %s", diff)
			}
		})
	}
}
