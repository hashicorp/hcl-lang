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

func TestReferenceTargetForOrigin(t *testing.T) {
	testCases := []struct {
		name              string
		refTargets        reference.Targets
		refOrigin         reference.Origin
		expectedRefTarget *reference.Target
	}{
		{
			"no targets",
			reference.Targets{},
			reference.Origin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "test"},
				},
			},
			nil,
		},
		{
			"single match",
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
				},
			},
			reference.Origin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "test"},
				},
				Constraints: reference.OriginConstraints{{}},
			},
			&reference.Target{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "test"},
				},
			},
		},
		{
			"first of two matches",
			reference.Targets{
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
			reference.Origin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "test"},
				},
				Constraints: reference.OriginConstraints{
					{OfType: cty.Bool},
				},
			},
			&reference.Target{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "test"},
				},
				Type: cty.Bool,
			},
		},
		{
			"match of unknown type",
			reference.Targets{
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
			reference.Origin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "foo"},
					lang.AttrStep{Name: "bar"},
				},
				Constraints: reference.OriginConstraints{{}},
			},
			&reference.Target{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "foo"},
				},
				Type: cty.DynamicPseudoType,
			},
		},
		{
			"match of nested target",
			reference.Targets{
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
					NestedTargets: reference.Targets{
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
			reference.Origin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "foo"},
					lang.AttrStep{Name: "bar"},
				},
				Constraints: reference.OriginConstraints{
					{OfType: cty.String},
				},
			},
			&reference.Target{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "foo"},
					lang.AttrStep{Name: "bar"},
				},
				Type: cty.String,
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			d := testPathDecoder(t, &PathContext{
				ReferenceTargets: tc.refTargets,
			})

			refTarget, err := d.ReferenceTargetForOrigin(tc.refOrigin)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedRefTarget, refTarget, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of reference target: %s", diff)
			}
		})
	}
}

func TestOutermostReferenceTargetsAtPos(t *testing.T) {
	testCases := []struct {
		name            string
		refTargets      reference.Targets
		filename        string
		pos             hcl.Pos
		expectedTargets reference.Targets
	}{
		{
			"no targets",
			reference.Targets{},
			"test.tf",
			hcl.InitialPos,
			reference.Targets{},
		},
		{
			"file mismatch",
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   35,
						},
					},
				},
			},
			"different.tf",
			hcl.Pos{
				Line:   1,
				Column: 3,
				Byte:   2,
			},
			reference.Targets{},
		},
		{
			"position mismatch",
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   35,
						},
					},
				},
			},
			"different.tf",
			hcl.Pos{
				Line:   5,
				Column: 1,
				Byte:   50,
			},
			reference.Targets{},
		},
		{
			"single matching target",
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   35,
						},
					},
				},
			},
			"test.tf",
			hcl.Pos{
				Line:   1,
				Column: 3,
				Byte:   2,
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   35,
						},
					},
				},
			},
		},
		{
			"two matching targets for the same position",
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   35,
						},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					ScopeId: lang.ScopeId("test"),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   35,
						},
					},
				},
			},
			"test.tf",
			hcl.Pos{
				Line:   1,
				Column: 3,
				Byte:   2,
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   35,
						},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					ScopeId: lang.ScopeId("test"),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   35,
						},
					},
				},
			},
		},
		{
			"nested target matches outermost",
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_instance"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.Object(map[string]cty.Type{
						"instance_type": cty.String,
					}),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   63,
						},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_instance"},
								lang.AttrStep{Name: "test"},
								lang.AttrStep{Name: "instance_type"},
							},
							Type: cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 3,
									Byte:   35,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 29,
									Byte:   61,
								},
							},
						},
					},
				},
			},
			"test.tf",
			hcl.Pos{
				Line:   2,
				Column: 4,
				Byte:   36,
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_instance"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.Object(map[string]cty.Type{
						"instance_type": cty.String,
					}),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   63,
						},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_instance"},
								lang.AttrStep{Name: "test"},
								lang.AttrStep{Name: "instance_type"},
							},
							Type: cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 3,
									Byte:   35,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 29,
									Byte:   61,
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
			d := testPathDecoder(t, &PathContext{
				ReferenceTargets: tc.refTargets,
			})

			refTargets, err := d.OutermostReferenceTargetsAtPos(tc.filename, tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedTargets, refTargets, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of reference targets: %s", diff)
			}
		})
	}
}

func TestInnermostReferenceTargetsAtPos(t *testing.T) {
	testCases := []struct {
		name            string
		refTargets      reference.Targets
		filename        string
		pos             hcl.Pos
		expectedTargets reference.Targets
	}{
		{
			"no targets",
			reference.Targets{},
			"test.tf",
			hcl.InitialPos,
			nil,
		},
		{
			"file mismatch",
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   35,
						},
					},
				},
			},
			"different.tf",
			hcl.Pos{
				Line:   1,
				Column: 3,
				Byte:   2,
			},
			nil,
		},
		{
			"position mismatch",
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   35,
						},
					},
				},
			},
			"different.tf",
			hcl.Pos{
				Line:   5,
				Column: 1,
				Byte:   50,
			},
			nil,
		},
		{
			"single target match",
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   35,
						},
					},
				},
			},
			"test.tf",
			hcl.Pos{
				Line:   1,
				Column: 3,
				Byte:   2,
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   35,
						},
					},
				},
			},
		},
		{
			"multiple targets matching at the same position",
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   35,
						},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					ScopeId: lang.ScopeId("test"),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   35,
						},
					},
				},
			},
			"test.tf",
			hcl.Pos{
				Line:   1,
				Column: 3,
				Byte:   2,
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   35,
						},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "test"},
					},
					ScopeId: lang.ScopeId("test"),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   35,
						},
					},
				},
			},
		},
		{
			"nested target matches innermost",
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_instance"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.Object(map[string]cty.Type{
						"instance_type": cty.String,
					}),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   63,
						},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_instance"},
								lang.AttrStep{Name: "test"},
								lang.AttrStep{Name: "instance_type"},
							},
							Type: cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 3,
									Byte:   35,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 29,
									Byte:   61,
								},
							},
						},
					},
				},
			},
			"test.tf",
			hcl.Pos{
				Line:   2,
				Column: 4,
				Byte:   36,
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_instance"},
						lang.AttrStep{Name: "test"},
						lang.AttrStep{Name: "instance_type"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   35,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 29,
							Byte:   61,
						},
					},
				},
			},
		},
		{
			"matching nested targets with position at block definition",
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "module"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.Object(map[string]cty.Type{
						"instance_id": cty.String,
					}),
					DefRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   1,
							Column: 20,
							Byte:   21,
						},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   63,
						},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "module"},
								lang.AttrStep{Name: "test"},
								lang.AttrStep{Name: "instance_id"},
							},
							Type: cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.InitialPos,
								End: hcl.Pos{
									Line:   3,
									Column: 2,
									Byte:   63,
								},
							},
						},
					},
				},
			},
			"test.tf",
			hcl.Pos{
				Line:   1,
				Column: 15,
				Byte:   16,
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "module"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.Object(map[string]cty.Type{
						"instance_id": cty.String,
					}),
					DefRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   1,
							Column: 20,
							Byte:   21,
						},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.InitialPos,
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   63,
						},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "module"},
								lang.AttrStep{Name: "test"},
								lang.AttrStep{Name: "instance_id"},
							},
							Type: cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start:    hcl.InitialPos,
								End: hcl.Pos{
									Line:   3,
									Column: 2,
									Byte:   63,
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
			d := testPathDecoder(t, &PathContext{
				ReferenceTargets: tc.refTargets,
			})

			refTargets, err := d.InnermostReferenceTargetsAtPos(tc.filename, tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedTargets, refTargets, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of reference targets: %s", diff)
			}
		})
	}
}

func TestReferenceTargetsInFile(t *testing.T) {
	testCases := []struct {
		name            string
		refTargets      reference.Targets
		filename        string
		expectedTargets reference.Targets
	}{
		{
			"no targets",
			reference.Targets{},
			"test.tf",
			reference.Targets{},
		},
		{
			"mismatching filename",
			reference.Targets{
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
			reference.Targets{},
		},
		{
			"matching file",
			reference.Targets{
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
					NestedTargets: reference.Targets{
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
			reference.Targets{
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
					NestedTargets: reference.Targets{
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
			d := testPathDecoder(t, &PathContext{
				ReferenceTargets: tc.refTargets,
			})

			targets, err := d.ReferenceTargetsInFile(tc.filename)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedTargets, targets, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of reference targets: %s", diff)
			}
		})
	}
}
