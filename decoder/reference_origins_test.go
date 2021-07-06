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

func TestCollectReferenceOrigins(t *testing.T) {
	testCases := []struct {
		name            string
		cfg             string
		expectedOrigins lang.ReferenceOrigins
	}{
		{
			"no origins",
			`attribute = "foo-bar"`,
			lang.ReferenceOrigins{},
		},
		{
			"root attribute single step",
			`attr = onestep`,
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
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
							Column: 15,
							Byte:   14,
						},
					},
				},
			},
		},
		{
			"multiple root attributes single step",
			`attr1 = onestep
attr2 = anotherstep
attr3 = onestep`,
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 9,
							Byte:   8,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 16,
							Byte:   15,
						},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "anotherstep"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 9,
							Byte:   24,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 20,
							Byte:   35,
						},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   3,
							Column: 9,
							Byte:   44,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 16,
							Byte:   51,
						},
					},
				},
			},
		},
		{
			"root attribute multiple origins",
			`attr1 = "${onestep}-${onestep}-${another.foo.bar}"`,
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 12,
							Byte:   11,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 19,
							Byte:   18,
						},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 23,
							Byte:   22,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 30,
							Byte:   29,
						},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "another"},
						lang.AttrStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 34,
							Byte:   33,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 49,
							Byte:   48,
						},
					},
				},
			},
		},
		{
			"root attribute multi-step",
			`attr = one.two["key"].attr[0]`,
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "one"},
						lang.AttrStep{Name: "two"},
						lang.IndexStep{Key: cty.StringVal("key")},
						lang.AttrStep{Name: "attr"},
						lang.IndexStep{Key: cty.NumberIntVal(0)},
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
		},
		{
			"attribute in block",
			`myblock {
  attr = onestep
}
`,
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 10,
							Byte:   19,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 17,
							Byte:   26,
						},
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

			origins, err := d.CollectReferenceOrigins()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedOrigins, origins, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatched reference origins: %s", diff)
			}
		})
	}
}

func TestReferenceOriginsTargeting(t *testing.T) {
	testCases := []struct {
		name            string
		allOrigins      lang.ReferenceOrigins
		refTarget       lang.ReferenceTarget
		expectedOrigins lang.ReferenceOrigins
	}{
		{
			"no origins",
			lang.ReferenceOrigins{},
			lang.ReferenceTarget{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
				},
				Type: cty.String,
			},
			lang.ReferenceOrigins{},
		},
		{
			"exact address match",
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "secondstep"},
					},
				},
			},
			lang.ReferenceTarget{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
					lang.AttrStep{Name: "secondstep"},
				},
				Type: cty.String,
			},
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "secondstep"},
					},
				},
			},
		},
		{
			"no match",
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "secondstep"},
					},
				},
			},
			lang.ReferenceTarget{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
					lang.AttrStep{Name: "different"},
				},
				Type: cty.String,
			},
			lang.ReferenceOrigins{},
		},
		{
			"match of nested target - two matches",
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "second"},
					},
				},
			},
			lang.ReferenceTarget{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
				},
				Type: cty.Object(map[string]cty.Type{
					"second": cty.String,
				}),
				NestedTargets: lang.ReferenceTargets{
					{
						Addr: lang.Address{
							lang.RootStep{Name: "test"},
							lang.AttrStep{Name: "second"},
						},
						Type: cty.String,
					},
				},
			},
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "second"},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			d := NewDecoder()
			d.SetReferenceOriginReader(func() lang.ReferenceOrigins {
				return tc.allOrigins
			})
			origins, err := d.ReferenceOriginsTargeting(tc.refTarget)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedOrigins, origins, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatched reference origins: %s", diff)
			}
		})
	}
}
