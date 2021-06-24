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

func TestReferenceOriginAtPos(t *testing.T) {
	testCases := []struct {
		name           string
		cfg            string
		pos            hcl.Pos
		expectedOrigin *lang.ReferenceOrigin
	}{
		{
			"empty config",
			``,
			hcl.InitialPos,
			nil,
		},
		{
			"single-step traversal in root attribute",
			`attr = blah`,
			hcl.Pos{
				Line:   1,
				Column: 9,
				Byte:   8,
			},
			&lang.ReferenceOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "blah"},
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 8,
						Byte:   7,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 12,
						Byte:   11,
					},
				},
			},
		},
		{
			"string literal in root attribute",
			`attr = "blah"`,
			hcl.Pos{
				Line:   1,
				Column: 9,
				Byte:   8,
			},
			nil,
		},
		{
			"multi-step traversal in root attribute",
			`attr = var.myobj.attr.foo.bar`,
			hcl.Pos{
				Line:   1,
				Column: 9,
				Byte:   8,
			},
			&lang.ReferenceOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "myobj"},
					lang.AttrStep{Name: "attr"},
					lang.AttrStep{Name: "foo"},
					lang.AttrStep{Name: "bar"},
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 8,
						Byte:   7,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 30,
						Byte:   29,
					},
				},
			},
		},
		{
			"multi-step traversal with map index step in root attribute",
			`attr = var.myobj.mapattr["key"]`,
			hcl.Pos{
				Line:   1,
				Column: 9,
				Byte:   8,
			},
			&lang.ReferenceOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "myobj"},
					lang.AttrStep{Name: "mapattr"},
					lang.IndexStep{Key: cty.StringVal("key")},
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 8,
						Byte:   7,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 32,
						Byte:   31,
					},
				},
			},
		},
		{
			"multi-step traversal with list index step in root attribute",
			`attr = var.myobj.listattr[4]`,
			hcl.Pos{
				Line:   1,
				Column: 9,
				Byte:   8,
			},
			&lang.ReferenceOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "myobj"},
					lang.AttrStep{Name: "listattr"},
					lang.IndexStep{Key: cty.NumberIntVal(4)},
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 8,
						Byte:   7,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 29,
						Byte:   28,
					},
				},
			},
		},
		{
			"multi-step traversal in block body",
			`customblock "foo" {
  attr = var.myobj.listattr[4]
}
`,
			hcl.Pos{
				Line:   2,
				Column: 11,
				Byte:   30,
			},
			&lang.ReferenceOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "myobj"},
					lang.AttrStep{Name: "listattr"},
					lang.IndexStep{Key: cty.NumberIntVal(4)},
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   2,
						Column: 10,
						Byte:   29,
					},
					End: hcl.Pos{
						Line:   2,
						Column: 31,
						Byte:   50,
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			d := NewDecoder()

			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			err := d.LoadFile("test.tf", f)
			if err != nil {
				t.Fatal(err)
			}

			refOrigin, err := d.ReferenceOriginAtPos("test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedOrigin, refOrigin, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatched reference origin: %s", diff)
			}
		})
	}
}
