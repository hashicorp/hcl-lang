// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reference

import (
	"context"
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
		origin          MatchableOrigin
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
		{
			"match of global nested target with local addrs set",
			Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_acmpca_certificate"},
						lang.AttrStep{Name: "foo"},
					},
					LocalAddr: lang.Address{
						lang.RootStep{Name: "self"},
					},
					Type: cty.Object(map[string]cty.Type{
						"signing_algorithm": cty.String,
					}),
					ScopeId: "resource",
					NestedTargets: Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_acmpca_certificate"},
								lang.AttrStep{Name: "foo"},
								lang.AttrStep{Name: "signing_algorithm"},
							},
							LocalAddr: lang.Address{
								lang.RootStep{Name: "self"},
								lang.AttrStep{Name: "signing_algorithm"},
							},
							Type: cty.String,
							TargetableFromRangePtr: &hcl.Range{
								Filename: "main.tf",
								Start:    hcl.Pos{Line: 26, Column: 41, Byte: 360},
								End:      hcl.Pos{Line: 41, Column: 2, Byte: 657},
							},
						},
					},
				},
			},
			LocalOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "aws_acmpca_certificate"},
					lang.AttrStep{Name: "foo"},
					lang.AttrStep{Name: "signing_algorithm"},
				},

				Constraints: OriginConstraints{
					{OfType: cty.DynamicPseudoType},
				},
				Range: hcl.Range{
					Filename: "main.tf",
					Start:    hcl.Pos{Line: 44, Column: 11, Byte: 684},
					End:      hcl.Pos{Line: 44, Column: 55, Byte: 728},
				},
			},
			Targets{
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_acmpca_certificate"},
						lang.AttrStep{Name: "foo"},
						lang.AttrStep{Name: "signing_algorithm"},
					},
					LocalAddr: lang.Address{
						lang.RootStep{Name: "self"},
						lang.AttrStep{Name: "signing_algorithm"},
					},
					Type: cty.String,
					TargetableFromRangePtr: &hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 26, Column: 41, Byte: 360},
						End:      hcl.Pos{Line: 41, Column: 2, Byte: 657},
					},
				},
			},
			true,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			refTarget, ok := tc.targets.Match(tc.origin)
			if !ok && tc.expectedFound {
				t.Fatalf("expected targetable to be found")
			}

			if diff := cmp.Diff(tc.expectedTargets, refTarget, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of reference target: %s", diff)
			}
		})
	}
}

func TestTargets_Match_localRefs(t *testing.T) {
	testCases := []struct {
		name            string
		targets         Targets
		origin          MatchableOrigin
		expectedTargets Targets
		expectedMatch   bool
	}{
		{
			"no targets",
			Targets{},
			LocalOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "count"},
					lang.AttrStep{Name: "index"},
				},
			},
			Targets{},
			false,
		},
		{
			"local address",
			Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "count"},
						lang.AttrStep{Name: "index"},
					},
					Type: cty.Number,
				},
			},
			LocalOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "count"},
					lang.AttrStep{Name: "index"},
				},
			},
			Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "count"},
						lang.AttrStep{Name: "index"},
					},
					Type: cty.Number,
				},
			},
			true,
		},
		{
			"local address with Constraint",
			Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "count"},
						lang.AttrStep{Name: "index"},
					},
					Type: cty.Number,
				},
			},
			LocalOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "count"},
					lang.AttrStep{Name: "index"},
				},
				Constraints: OriginConstraints{
					{OfType: cty.Number},
				},
			},
			Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "count"},
						lang.AttrStep{Name: "index"},
					},
					Type: cty.Number,
				},
			},
			true,
		},
		{
			"local address with Type and TargetableFromRange positive",
			Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "count"},
						lang.AttrStep{Name: "index"},
					},
					Type: cty.Number,
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 5, Column: 1, Byte: 50},
					},
				},
			},
			LocalOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "count"},
					lang.AttrStep{Name: "index"},
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 1, Byte: 20},
					End:      hcl.Pos{Line: 2, Column: 10, Byte: 29},
				},
				Constraints: OriginConstraints{
					{OfType: cty.Number},
				},
			},
			Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "count"},
						lang.AttrStep{Name: "index"},
					},
					Type: cty.Number,
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 5, Column: 1, Byte: 50},
					},
				},
			},
			true,
		},
		{
			"local address with Type and TargetableFromRange negative",
			Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "count"},
						lang.AttrStep{Name: "index"},
					},
					Type: cty.Number,
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 5, Column: 1, Byte: 20},
					},
				},
			},
			LocalOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "count"},
					lang.AttrStep{Name: "index"},
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 2, Column: 1, Byte: 20},
					End:      hcl.Pos{Line: 2, Column: 10, Byte: 29},
				},
				Constraints: OriginConstraints{
					{OfType: cty.Number},
				},
			},
			Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "count"},
						lang.AttrStep{Name: "index"},
					},
					Type: cty.Number,
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 5, Column: 1, Byte: 20},
					},
				},
			},
			true,
		},
		{
			"local origin and global target",
			Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "module"},
						lang.AttrStep{Name: "localmodd"},
						lang.AttrStep{Name: "someattribute"},
					},
					Type: cty.DynamicPseudoType,
				},
			},
			LocalOrigin{
				Addr: lang.Address{
					lang.RootStep{Name: "self"},
					lang.AttrStep{Name: "attribute"},
				},
				Constraints: OriginConstraints{
					{
						OfScopeId: "",
						OfType:    cty.String,
					},
				},
			},
			Targets{},
			false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			refTarget, ok := tc.targets.Match(tc.origin)
			if !ok && tc.expectedMatch {
				t.Fatalf("expected target to be matched")
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

func TestTargets_LegacyMatchWalk(t *testing.T) {
	testCases := []struct {
		name             string
		targets          Targets
		traversalConst   schema.TraversalExpr
		prefix           string
		outermostBodyRng hcl.Range
		originRng        hcl.Range
		expectedTargets  Targets
	}{
		{
			"no targets",
			Targets{},
			schema.TraversalExpr{},
			"test",
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			},
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			},
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
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			},
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			},
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
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			},
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			},
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
				Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "another"},
					},
					Type: cty.List(cty.String),
				},
			},
			schema.TraversalExpr{
				OfType: cty.String,
			},
			"",
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			},
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			},
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
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			},
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			},
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
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			},
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			},
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
			ctx := context.Background()
			tc.targets.LegacyMatchWalk(ctx, tc.traversalConst, tc.prefix, tc.outermostBodyRng, tc.originRng, func(t Target) error {
				targets = append(targets, t)
				return nil
			})
			if diff := cmp.Diff(tc.expectedTargets, targets, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of targets: %s", diff)
			}
		})
	}
}

func TestTargets_LegacyMatchWalk_localRefs(t *testing.T) {
	testCases := []struct {
		name             string
		targets          Targets
		traversalConst   schema.TraversalExpr
		prefix           string
		outermostBodyRng hcl.Range
		originRng        hcl.Range
		activeSelfRefs   bool
		expectedTargets  Targets
	}{
		{
			"no targets",
			Targets{},
			schema.TraversalExpr{},
			"test",
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			},
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			},
			false,
			Targets{},
		},
		{
			"targets with local address only",
			Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "count"},
						lang.AttrStep{Name: "index"},
					},
				},
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
				},
			},
			schema.TraversalExpr{},
			"co",
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			},
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			},
			false,
			Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "count"},
						lang.AttrStep{Name: "index"},
					},
				},
			},
		},
		{
			"targets with mixed address and same block",
			Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "foobar"},
					},
					Addr: lang.Address{
						lang.RootStep{Name: "abs"},
						lang.AttrStep{Name: "foobar"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 5, Column: 1, Byte: 25},
						End:      hcl.Pos{Line: 5, Column: 10, Byte: 35},
					},
				},
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "boo"},
					},
					Addr: lang.Address{
						lang.RootStep{Name: "abs"},
						lang.AttrStep{Name: "boo"},
					},
				},
			},
			schema.TraversalExpr{},
			"local",
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:      hcl.Pos{Line: 10, Column: 1, Byte: 50},
			},
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			},
			false,
			Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "foobar"},
					},
					Addr: lang.Address{
						lang.RootStep{Name: "abs"},
						lang.AttrStep{Name: "foobar"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 5, Column: 1, Byte: 25},
						End:      hcl.Pos{Line: 5, Column: 10, Byte: 35},
					},
				},
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "boo"},
					},
					Addr: lang.Address{
						lang.RootStep{Name: "abs"},
						lang.AttrStep{Name: "boo"},
					},
				},
			},
		},
		{
			"targets matching only the local block",
			Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "count"},
						lang.AttrStep{Name: "index"},
					},
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 10, Column: 1, Byte: 50},
					},
				},
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "count"},
						lang.AttrStep{Name: "index"},
					},
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 12, Column: 1, Byte: 52},
						End:      hcl.Pos{Line: 20, Column: 1, Byte: 80},
					},
				},
			},
			schema.TraversalExpr{},
			"co",
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			},
			hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 5, Column: 1, Byte: 25},
				End:      hcl.Pos{Line: 5, Column: 10, Byte: 35},
			},
			false,
			Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "count"},
						lang.AttrStep{Name: "index"},
					},
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 10, Column: 1, Byte: 50},
					},
				},
			},
		},
		{
			// ensure that e.g. count = count.index is mismatched, e.g. in
			// resource "aws_alb" "test" {
			//   count = count.index
			// }
			"target pointing to the same attribute as origin",
			Targets{
				{
					LocalAddr: lang.Address{
						lang.RootStep{Name: "count"},
						lang.AttrStep{Name: "index"},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 30},
						End:      hcl.Pos{Line: 2, Column: 8, Byte: 35},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 3, Byte: 30},
						End:      hcl.Pos{Line: 2, Column: 10, Byte: 37},
					},
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
						End:      hcl.Pos{Line: 3, Column: 1, Byte: 39},
					},
				},
			},
			schema.TraversalExpr{},
			"",
			hcl.Range{ // outermost body range
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
				End:      hcl.Pos{Line: 3, Column: 1, Byte: 39},
			},
			hcl.Range{ // origin range
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 2, Column: 11, Byte: 38},
				End:      hcl.Pos{Line: 2, Column: 11, Byte: 38},
			},
			false,
			Targets{},
		},
		{
			// ensure that e.g. foo = self is matched, e.g. in
			// resource "aws_alb" "test" {
			//   foo = self
			// }
			"target pointing to outside body of an origin",
			Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_alb"},
						lang.AttrStep{Name: "test"},
					},
					LocalAddr: lang.Address{
						lang.RootStep{Name: "self"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
						End:      hcl.Pos{Line: 3, Column: 1, Byte: 37},
					},
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
						End:      hcl.Pos{Line: 3, Column: 1, Byte: 37},
					},
					NestedTargets: Targets{
						{
							LocalAddr: lang.Address{
								lang.RootStep{Name: "self"},
								lang.AttrStep{Name: "bar"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
								End:      hcl.Pos{Line: 3, Column: 1, Byte: 37},
							},
							TargetableFromRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
								End:      hcl.Pos{Line: 3, Column: 1, Byte: 37},
							},
						},
					},
				},
			},
			schema.TraversalExpr{},
			"",
			hcl.Range{ // outermost body range
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
				End:      hcl.Pos{Line: 3, Column: 1, Byte: 37},
			},
			hcl.Range{ // origin range
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 2, Column: 9, Byte: 36},
				End:      hcl.Pos{Line: 2, Column: 9, Byte: 36},
			},
			true,
			Targets{},
		},
		{
			"self only matches when enabled",
			Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_alb"},
						lang.AttrStep{Name: "test"},
					},
					LocalAddr: lang.Address{
						lang.RootStep{Name: "self"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
						End:      hcl.Pos{Line: 4, Column: 1, Byte: 37},
					},
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
						End:      hcl.Pos{Line: 4, Column: 1, Byte: 37},
					},
					NestedTargets: Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_alb"},
								lang.AttrStep{Name: "test"},
								lang.AttrStep{Name: "bar"},
							},
							LocalAddr: lang.Address{
								lang.RootStep{Name: "self"},
								lang.AttrStep{Name: "bar"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 2, Column: 1, Byte: 30},
								End:      hcl.Pos{Line: 2, Column: 20, Byte: 35},
							},
							TargetableFromRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
								End:      hcl.Pos{Line: 4, Column: 1, Byte: 37},
							},
						},
					},
				},
			},
			schema.TraversalExpr{},
			"",
			hcl.Range{ // outermost body range
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
				End:      hcl.Pos{Line: 3, Column: 1, Byte: 37},
			},
			hcl.Range{ // origin range
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 3, Column: 2, Byte: 36},
				End:      hcl.Pos{Line: 3, Column: 2, Byte: 36},
			},
			true,
			Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_alb"},
						lang.AttrStep{Name: "test"},
					},
					LocalAddr: lang.Address{
						lang.RootStep{Name: "self"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
						End:      hcl.Pos{Line: 4, Column: 1, Byte: 37},
					},
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
						End:      hcl.Pos{Line: 4, Column: 1, Byte: 37},
					},
					NestedTargets: Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_alb"},
								lang.AttrStep{Name: "test"},
								lang.AttrStep{Name: "bar"},
							},
							LocalAddr: lang.Address{
								lang.RootStep{Name: "self"},
								lang.AttrStep{Name: "bar"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 2, Column: 1, Byte: 30},
								End:      hcl.Pos{Line: 2, Column: 20, Byte: 35},
							},
							TargetableFromRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
								End:      hcl.Pos{Line: 4, Column: 1, Byte: 37},
							},
						},
					},
				},
			},
		},
		{
			"self doesn't match when disabled",
			Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_alb"},
						lang.AttrStep{Name: "test"},
					},
					LocalAddr: lang.Address{
						lang.RootStep{Name: "self"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
						End:      hcl.Pos{Line: 4, Column: 1, Byte: 37},
					},
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
						End:      hcl.Pos{Line: 4, Column: 1, Byte: 37},
					},
					NestedTargets: Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_alb"},
								lang.AttrStep{Name: "test"},
								lang.AttrStep{Name: "bar"},
							},
							LocalAddr: lang.Address{
								lang.RootStep{Name: "self"},
								lang.AttrStep{Name: "bar"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 2, Column: 1, Byte: 30},
								End:      hcl.Pos{Line: 2, Column: 20, Byte: 35},
							},
							TargetableFromRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
								End:      hcl.Pos{Line: 4, Column: 1, Byte: 37},
							},
						},
					},
				},
			},
			schema.TraversalExpr{},
			"",
			hcl.Range{ // outermost body range
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 28, Byte: 27},
				End:      hcl.Pos{Line: 3, Column: 1, Byte: 37},
			},
			hcl.Range{ // origin range
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 3, Column: 2, Byte: 36},
				End:      hcl.Pos{Line: 3, Column: 2, Byte: 36},
			},
			false,
			Targets{},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			targets := make(Targets, 0)
			ctx := context.Background()
			if tc.activeSelfRefs {
				ctx = schema.WithActiveSelfRefs(ctx)
			}
			tc.targets.LegacyMatchWalk(ctx, tc.traversalConst, tc.prefix, tc.outermostBodyRng, tc.originRng, func(t Target) error {
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
