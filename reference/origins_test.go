package reference

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestOrigins_AtPos(t *testing.T) {
	testCases := []struct {
		name            string
		origins         Origins
		pos             hcl.Pos
		expectedOrigins Origins
		expectedFound   bool
	}{
		{
			"no origins",
			Origins{},
			hcl.InitialPos,
			Origins{},
			false,
		},
		{
			"single mismatching origin",
			Origins{
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "blah"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
					},
				},
			},
			hcl.Pos{
				Line:   1,
				Column: 3,
				Byte:   2,
			},
			Origins{},
			false,
		},
		{
			"single matching origin",
			Origins{
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "blah"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
					},
				},
			},
			hcl.Pos{
				Line:   1,
				Column: 9,
				Byte:   8,
			},
			Origins{
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "blah"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
					},
				},
			},
			true,
		},
		{
			"multiple origins - single match",
			Origins{
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
					},
				},
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
					},
					Range: hcl.Range{
						Filename: "differentfile.tf",
						Start:    hcl.Pos{Line: 2, Column: 8, Byte: 14},
						End:      hcl.Pos{Line: 2, Column: 12, Byte: 18},
					},
				},
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 8, Byte: 14},
						End:      hcl.Pos{Line: 2, Column: 12, Byte: 18},
					},
				},
			},
			hcl.Pos{
				Line:   2,
				Column: 9,
				Byte:   15,
			},
			Origins{
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 8, Byte: 14},
						End:      hcl.Pos{Line: 2, Column: 12, Byte: 18},
					},
				},
			},
			true,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			origins, ok := tc.origins.AtPos("test.tf", tc.pos)
			if !ok && tc.expectedFound {
				t.Fatal("expected origin to be found")
			}

			if diff := cmp.Diff(tc.expectedOrigins, origins, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatched origin: %s", diff)
			}
		})
	}
}

func TestOrigins_Match(t *testing.T) {
	testCases := []struct {
		name            string
		origins         Origins
		refTarget       Target
		expectedOrigins Origins
	}{
		{
			"no origins",
			Origins{},
			Target{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
				},
				Type: cty.String,
			},
			Origins{},
		},
		{
			"exact address match",
			Origins{
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
				},
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "secondstep"},
					},
					Constraints: OriginConstraints{
						{OfType: cty.String},
					},
				},
			},
			Target{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
					lang.AttrStep{Name: "secondstep"},
				},
				Type: cty.String,
			},
			Origins{
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "secondstep"},
					},
					Constraints: OriginConstraints{
						{OfType: cty.String},
					},
				},
			},
		},
		{
			"no match",
			Origins{
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
				},
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "secondstep"},
					},
				},
			},
			Target{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
					lang.AttrStep{Name: "different"},
				},
				Type: cty.String,
			},
			Origins{},
		},
		{
			"match of nested target - two matches",
			Origins{
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
				},
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Constraints: OriginConstraints{
						{OfType: cty.DynamicPseudoType},
					},
				},
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "second"},
					},
					Constraints: OriginConstraints{
						{OfType: cty.String},
					},
				},
			},
			Target{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
				},
				Type: cty.Object(map[string]cty.Type{
					"second": cty.String,
				}),
				NestedTargets: Targets{
					{
						Addr: lang.Address{
							lang.RootStep{Name: "test"},
							lang.AttrStep{Name: "second"},
						},
						Type: cty.String,
					},
				},
			},
			Origins{
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Constraints: OriginConstraints{
						{OfType: cty.DynamicPseudoType},
					},
				},
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "second"},
					},
					Constraints: OriginConstraints{
						{OfType: cty.String},
					},
				},
			},
		},
		{
			"loose match of target of unknown type",
			Origins{
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					Constraints: OriginConstraints{{}},
				},
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Constraints: OriginConstraints{{}},
				},
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "second"},
					},
					Constraints: OriginConstraints{{}},
				},
			},
			Target{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
				},
				Type: cty.DynamicPseudoType,
			},
			Origins{
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Constraints: OriginConstraints{{}},
				},
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "second"},
					},
					Constraints: OriginConstraints{{}},
				},
			},
		},
		{
			"mismatch of target nil type",
			Origins{
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Constraints: OriginConstraints{
						{OfScopeId: lang.ScopeId("test")},
					},
				},
			},
			Target{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
				},
				ScopeId: lang.ScopeId("test"),
				Type:    cty.String,
			},
			Origins{},
		},
		// JSON edge cases
		{
			"constraint-less origin mismatching scope-only target",
			Origins{
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "alpha"},
					},
					Constraints: nil,
				},
			},
			Target{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "alpha"},
				},
				ScopeId: "variable",
				Type:    cty.NilType,
			},
			Origins{},
		},
		{
			"constraint-less origin matching type-aware target",
			Origins{
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "beta"},
					},
					Constraints: nil,
				},
			},
			Target{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "beta"},
				},
				ScopeId: "variable",
				Type:    cty.DynamicPseudoType,
			},
			Origins{
				LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "beta"},
					},
					Constraints: nil,
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			// TODO: test PathOrigins
			origins := tc.origins.Match(tc.refTarget, lang.Path{Path: t.TempDir()})

			if diff := cmp.Diff(tc.expectedOrigins, origins, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatched reference origins: %s", diff)
			}
		})
	}
}
