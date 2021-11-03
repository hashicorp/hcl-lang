package decoder

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestCollectReferenceTargets_noSchema(t *testing.T) {
	d := testPathDecoder(t, &PathContext{})
	_, err := d.CollectReferenceTargets()
	if err == nil {
		t.Fatal("expected error when no schema is set")
	}

	noSchemaErr := &NoSchemaError{}
	if !errors.As(err, &noSchemaErr) {
		t.Fatalf("unexpected error: %#v, expected %#v", err, noSchemaErr)
	}
}

func TestReferenceTargetForOriginAtPos(t *testing.T) {
	dirPath := t.TempDir()

	testCases := []struct {
		name            string
		dirs            map[string]*PathContext
		path            lang.Path
		filename        string
		pos             hcl.Pos
		expectedTargets ReferenceTargets
		expectedErr     error
	}{
		{
			"no origins and no targets",
			map[string]*PathContext{
				dirPath: {
					ReferenceTargets: reference.Targets{},
					ReferenceOrigins: reference.Origins{},
				},
			},
			lang.Path{Path: dirPath},
			"test.tf",
			hcl.InitialPos,
			ReferenceTargets{},
			&reference.NoOriginFound{},
		},
		{
			"single matching origin targeting and no targets",
			map[string]*PathContext{
				dirPath: {
					ReferenceOrigins: reference.Origins{
						reference.LocalOrigin{
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
					ReferenceTargets: reference.Targets{},
				},
			},
			lang.Path{Path: dirPath},
			"test.tf",
			hcl.InitialPos,
			ReferenceTargets{},
			nil,
		},
		{
			"matching origin and target",
			map[string]*PathContext{
				dirPath: {
					ReferenceOrigins: reference.Origins{
						reference.LocalOrigin{
							Addr: lang.Address{
								lang.RootStep{Name: "one"},
							},
							Constraints: reference.OriginConstraints{
								reference.OriginConstraint{
									OfType: cty.Bool,
								},
							},
							Range: hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
								End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
							},
						},
					},
					ReferenceTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "one"},
							},
							Type: cty.Bool,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 2, Column: 1, Byte: 10},
								End:      hcl.Pos{Line: 2, Column: 4, Byte: 13},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "two"},
							},
							Type: cty.Bool,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 3, Column: 1, Byte: 16},
								End:      hcl.Pos{Line: 4, Column: 4, Byte: 19},
							},
						},
					},
				},
			},
			lang.Path{Path: dirPath},
			"test.tf",
			hcl.InitialPos,
			ReferenceTargets{
				&ReferenceTarget{
					OriginRange: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
					Path: lang.Path{Path: dirPath},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 10},
						End:      hcl.Pos{Line: 2, Column: 4, Byte: 13},
					},
				},
			},
			nil,
		},
		{
			"mismatching path origin",
			map[string]*PathContext{
				dirPath: {
					ReferenceOrigins: reference.Origins{
						reference.PathOrigin{
							Range: hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
								End:      hcl.Pos{Line: 1, Column: 6, Byte: 5},
							},
							TargetAddr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "foo"},
							},
							TargetPath: lang.Path{
								Path: filepath.Join(dirPath, "alpha"),
							},
							Constraints: reference.OriginConstraints{
								reference.OriginConstraint{
									OfScopeId: lang.ScopeId("variable"),
									OfType:    cty.String,
								},
							},
						},
					},
					ReferenceTargets: reference.Targets{},
				},
				filepath.Join(dirPath, "beta"): {
					ReferenceTargets: reference.Targets{
						reference.Target{
							Addr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "foo"},
							},
							ScopeId: lang.ScopeId("variable"),
							Type:    cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
								End:      hcl.Pos{Line: 3, Column: 2, Byte: 15},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
								End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
							},
							Name: "variable",
						},
					},
				},
			},
			lang.Path{Path: dirPath},
			"test.tf",
			hcl.InitialPos,
			ReferenceTargets{},
			nil,
		},
		{
			"matching path origin",
			map[string]*PathContext{
				dirPath: {
					ReferenceOrigins: reference.Origins{
						reference.PathOrigin{
							Range: hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
								End:      hcl.Pos{Line: 1, Column: 6, Byte: 5},
							},
							TargetAddr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "foo"},
							},
							TargetPath: lang.Path{
								Path: filepath.Join(dirPath, "beta"),
							},
							Constraints: reference.OriginConstraints{
								reference.OriginConstraint{
									OfScopeId: lang.ScopeId("variable"),
									OfType:    cty.String,
								},
							},
						},
					},
					ReferenceTargets: reference.Targets{},
				},
				filepath.Join(dirPath, "beta"): {
					ReferenceTargets: reference.Targets{
						reference.Target{
							Addr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "foo"},
							},
							ScopeId: lang.ScopeId("variable"),
							Type:    cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
								End:      hcl.Pos{Line: 3, Column: 2, Byte: 15},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
								End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
							},
							Name: "variable",
						},
					},
				},
			},
			lang.Path{Path: dirPath},
			"test.tf",
			hcl.InitialPos,
			ReferenceTargets{
				{
					OriginRange: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 6, Byte: 5},
					},
					Path: lang.Path{Path: filepath.Join(dirPath, "beta")},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 3, Column: 2, Byte: 15},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 10, Byte: 9},
					},
				},
			},
			nil,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			d := NewDecoder(&testPathReader{
				paths: tc.dirs,
			})

			targets, err := d.ReferenceTargetsForOriginAtPos(tc.path, tc.filename, tc.pos)
			if err != nil {
				if tc.expectedErr != nil && !errors.As(err, &tc.expectedErr) {
					t.Fatalf("unexpected error: %s\nexpected: %s\n",
						err, tc.expectedErr)
				} else if tc.expectedErr == nil {
					t.Fatal(err)
				}
			} else if tc.expectedErr != nil {
				t.Fatalf("expected error: %s", tc.expectedErr)
			}

			if diff := cmp.Diff(tc.expectedTargets, targets, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of reference targets: %s", diff)
			}
		})
	}
}
