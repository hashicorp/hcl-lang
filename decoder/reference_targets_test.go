package decoder

import (
	"errors"
	"fmt"
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
		name           string
		origins        reference.Origins
		targets        reference.Targets
		filename       string
		pos            hcl.Pos
		expectedTarget *ReferenceTarget
		expectedErr    error
	}{
		{
			"no origins and no targets",
			reference.Origins{},
			reference.Targets{},
			"test.tf",
			hcl.InitialPos,
			nil,
			&reference.NoOriginFound{},
		},
		{
			"single matching origin targeting and no targets",
			reference.Origins{
				{
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
			reference.Targets{},
			"test.tf",
			hcl.InitialPos,
			nil,
			&reference.NoTargetFound{},
		},
		{
			"matching origin and target",
			reference.Origins{
				{
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
			reference.Targets{
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
			"test.tf",
			hcl.InitialPos,
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
			nil,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			dirs := map[string]*PathContext{
				dirPath: {
					ReferenceTargets: tc.targets,
					ReferenceOrigins: tc.origins,
				},
			}

			d := NewDecoder(&testPathReader{
				paths: dirs,
			})

			pathDecoder, err := d.Path(lang.Path{Path: dirPath})
			if err != nil {
				t.Fatal(err)
			}

			origins, err := pathDecoder.ReferenceTargetForOriginAtPos(tc.filename, tc.pos)
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

			if diff := cmp.Diff(tc.expectedTarget, origins, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of reference origins: %s", diff)
			}
		})
	}
}
