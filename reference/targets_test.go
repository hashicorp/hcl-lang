package reference

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestTargets_Match(t *testing.T) {
	testCases := []struct {
		name            string
		targets         Targets
		origin          Origin
		expectedTargets Targets
		expectedFound   bool
	}{
		{
			"no targets",
			Targets{},
			LocalOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "test"},
				},
			},
			Targets{},
			false,
		},
		{
			"single match",
			Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
				},
			},
			LocalOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "test"},
				},
				Constraints: OriginConstraints{{}},
			},
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
				},
			},
			true,
		},
		{
			"first of two matches",
			Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "data"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.Bool,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.Bool,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					ScopeId: lang.ScopeId("variable"),
				},
			},
			LocalOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "test"},
				},
				Constraints: OriginConstraints{
					{OfType: cty.Bool},
				},
			},
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.Bool,
				},
			},
			true,
		},
		{
			"match of unknown type",
			Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "data"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.Bool,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.DynamicPseudoType,
				},
			},
			LocalOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "foo"},
					lang.AttrStep{Name: "bar"},
				},
				Constraints: OriginConstraints{{}},
			},
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.DynamicPseudoType,
				},
			},
			true,
		},
		{
			"match of nested target",
			Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "data"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.Bool,
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.Object(map[string]cty.Type{
						"bar": cty.String,
					}),
					NestedTargets: Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "foo"},
								lang.AttrStep{Name: "bar"},
							},
							Type: cty.String,
						},
					},
				},
			},
			LocalOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "foo"},
					lang.AttrStep{Name: "bar"},
				},
				Constraints: OriginConstraints{
					{OfType: cty.String},
				},
			},
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
					Type: cty.String,
				},
			},
			true,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			refTarget, ok := tc.targets.Match(tc.origin.Address(), tc.origin.OriginConstraints())
			if !ok && tc.expectedFound {
				t.Fatalf("expected targetable to be found")
			}

			if diff := cmp.Diff(tc.expectedTargets, refTarget, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of reference target: %s", diff)
			}
		})
	}
}

func TestTargets_OutermostInFile(t *testing.T) {
	testCases := []struct {
		name            string
		targets         Targets
		filename        string
		expectedTargets Targets
	}{
		{
			"no targets",
			Targets{},
			"test.tf",
			Targets{},
		},
		{
			"mismatching filename",
			Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					RangePtr: &hcl.Range{
						Filename: "bar.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   2,
							Column: 1,
							Byte:   10,
						},
					},
				},
			},
			"test.tf",
			Targets{},
		},
		{
			"matching file",
			Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   2,
							Column: 1,
							Byte:   10,
						},
					},
					NestedTargets: Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "foo"},
								lang.AttrStep{Name: "bar"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.InitialPos,
								End: hcl.Pos{
									Line:   1,
									Column: 5,
									Byte:   4,
								},
							},
						},
					},
				},
			},
			"test.tf",
			Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   2,
							Column: 1,
							Byte:   10,
						},
					},
					NestedTargets: Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "foo"},
								lang.AttrStep{Name: "bar"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.InitialPos,
								End: hcl.Pos{
									Line:   1,
									Column: 5,
									Byte:   4,
								},
							},
						},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			targets := tc.targets.OutermostInFile(tc.filename)

			if diff := cmp.Diff(tc.expectedTargets, targets, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of targets: %s", diff)
			}
		})
	}
}

func TestTargets_MatchWalk(t *testing.T) {
	testCases := []struct {
		name            string
		targets         Targets
		traversalConst  schema.TraversalExpr
		prefix          string
		expectedTargets Targets
	}{
		{
			"no targets",
			Targets{},
			schema.TraversalExpr{},
			"test",
			Targets{},
		},
		{
			"empty prefix and empty constraint",
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
				},
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
				},
			},
			schema.TraversalExpr{},
			"",
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
				},
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
				},
			},
		},
		{
			"prefix match empty constraint",
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
				},
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
				},
			},
			schema.TraversalExpr{},
			"var.f",
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
				},
			},
		},
		{
			"type only match",
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Type: cty.Bool,
				},
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.String,
				},
			},
			schema.TraversalExpr{
				OfType: cty.String,
			},
			"",
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.String,
				},
			},
		},
		{
			"scope only match",
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					ScopeId: lang.ScopeId("blue"),
				},
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					ScopeId: lang.ScopeId("green"),
				},
			},
			schema.TraversalExpr{
				OfScopeId: lang.ScopeId("green"),
			},
			"",
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					ScopeId: lang.ScopeId("green"),
				},
			},
		},
		{
			"type and scope match",
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type:    cty.Bool,
					ScopeId: lang.ScopeId("blue"),
				},
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type:    cty.Number,
					ScopeId: lang.ScopeId("green"),
				},
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type:    cty.Number,
					ScopeId: lang.ScopeId("red"),
				},
			},
			schema.TraversalExpr{
				OfType:    cty.Bool,
				OfScopeId: lang.ScopeId("blue"),
			},
			"",
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type:    cty.Bool,
					ScopeId: lang.ScopeId("blue"),
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			targets := make(Targets, 0)
			tc.targets.MatchWalk(tc.traversalConst, tc.prefix, func(t Target) error {
				targets = append(targets, t)
				return nil
			})
			if diff := cmp.Diff(tc.expectedTargets, targets, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of targets: %s", diff)
			}
		})
	}
}

func TestTargets_InnermostAtPos(t *testing.T) {
	testCases := []struct {
		name            string
		targets         Targets
		file            string
		pos             hcl.Pos
		expectedTargets Targets
		expectedFound   bool
	}{
		{
			"no targets",
			Targets{},
			"test.tf",
			hcl.InitialPos,
			Targets{},
			false,
		},
		{
			"target without range",
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
				},
			},
			"test.tf",
			hcl.InitialPos,
			Targets{},
			false,
		},
		{
			"top level match",
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   30,
						},
					},
				},
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
					RangePtr: &hcl.Range{
						Filename: "another.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   30,
						},
					},
				},
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "third"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   5,
							Column: 1,
							Byte:   32,
						},
						End: hcl.Pos{
							Line:   7,
							Column: 2,
							Byte:   60,
						},
					},
				},
			},
			"test.tf",
			hcl.Pos{
				Line:   6,
				Column: 1,
				Byte:   50,
			},
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "third"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   5,
							Column: 1,
							Byte:   32,
						},
						End: hcl.Pos{
							Line:   7,
							Column: 2,
							Byte:   60,
						},
					},
				},
			},
			true,
		},
		{
			"nested target match",
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   6,
							Column: 1,
							Byte:   60,
						},
					},
					NestedTargets: Targets{
						Target{
							Addr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "first"},
								lang.AttrStep{Name: "alpha"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 1,
									Byte:   10,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 11,
									Byte:   20,
								},
							},
						},
						Target{
							Addr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "first"},
								lang.AttrStep{Name: "beta"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 1,
									Byte:   21,
								},
								End: hcl.Pos{
									Line:   3,
									Column: 11,
									Byte:   31,
								},
							},
						},
					},
				},
			},
			"test.tf",
			hcl.Pos{
				Line:   3,
				Column: 5,
				Byte:   25,
			},
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
						lang.AttrStep{Name: "beta"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   3,
							Column: 1,
							Byte:   21,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 11,
							Byte:   31,
						},
					},
				},
			},
			true,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			targets, found := tc.targets.InnermostAtPos(tc.file, tc.pos)
			if !found && tc.expectedFound {
				t.Fatal("expected to find targets")
			}
			if diff := cmp.Diff(tc.expectedTargets, targets, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of targets: %s", diff)
			}
		})
	}
}
