package decoder

import (
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

func TestReferenceOriginsTargetingPos(t *testing.T) {
	dirPath := t.TempDir()

	testCases := []struct {
		name            string
		dirs            map[string]*PathContext
		path            lang.Path
		filename        string
		pos             hcl.Pos
		expectedOrigins ReferenceOrigins
	}{
		{
			"no targets and no origins",
			map[string]*PathContext{
				dirPath: {
					ReferenceOrigins: reference.Origins{},
					ReferenceTargets: reference.Targets{},
				},
			},
			lang.Path{Path: dirPath},
			"test.tf",
			hcl.InitialPos,
			ReferenceOrigins{},
		},
		{
			"target file mismatch",
			map[string]*PathContext{
				dirPath: {
					ReferenceOrigins: reference.Origins{},
					ReferenceTargets: reference.Targets{
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
			},
			lang.Path{Path: dirPath},
			"different.tf",
			hcl.Pos{
				Line:   1,
				Column: 3,
				Byte:   2,
			},
			ReferenceOrigins{},
		},
		{
			"target position mismatch",
			map[string]*PathContext{
				dirPath: {
					ReferenceOrigins: reference.Origins{},
					ReferenceTargets: reference.Targets{
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
			},
			lang.Path{Path: dirPath},
			"different.tf",
			hcl.Pos{
				Line:   5,
				Column: 1,
				Byte:   50,
			},
			ReferenceOrigins{},
		},
		{
			"single target match",
			map[string]*PathContext{
				dirPath: {
					ReferenceOrigins: reference.Origins{
						reference.LocalOrigin{
							Addr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "test"},
							},
							Range: hcl.Range{
								Filename: "another.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 5,
									Byte:   4,
								},
								End: hcl.Pos{
									Line:   1,
									Column: 13,
									Byte:   12,
								},
							},
							Constraints: reference.OriginConstraints{
								{
									OfType: cty.String,
								},
							},
						},
					},
					ReferenceTargets: reference.Targets{
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
			},
			lang.Path{Path: dirPath},
			"test.tf",
			hcl.Pos{
				Line:   1,
				Column: 3,
				Byte:   2,
			},
			ReferenceOrigins{
				{
					Path: lang.Path{Path: dirPath},
					Range: hcl.Range{
						Filename: "another.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 5,
							Byte:   4,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 13,
							Byte:   12,
						},
					},
				},
			},
		},
		{
			"multiple targets matching at the same position",
			map[string]*PathContext{
				dirPath: {
					ReferenceOrigins: reference.Origins{
						reference.LocalOrigin{
							Addr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "test"},
							},
							Range: hcl.Range{
								Filename: "another.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 5,
									Byte:   4,
								},
								End: hcl.Pos{
									Line:   1,
									Column: 13,
									Byte:   12,
								},
							},
							Constraints: reference.OriginConstraints{
								{
									OfType: cty.String,
								},
							},
						},
					},
					ReferenceTargets: reference.Targets{
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
			},
			lang.Path{Path: dirPath},
			"test.tf",
			hcl.Pos{
				Line:   1,
				Column: 3,
				Byte:   2,
			},
			ReferenceOrigins{
				{
					Path: lang.Path{Path: dirPath},
					Range: hcl.Range{
						Filename: "another.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 5,
							Byte:   4,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 13,
							Byte:   12,
						},
					},
				},
			},
		},
		{
			"nested target matches innermost",
			map[string]*PathContext{
				dirPath: {
					ReferenceTargets: reference.Targets{
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
					ReferenceOrigins: reference.Origins{
						reference.LocalOrigin{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_instance"},
								lang.AttrStep{Name: "test"},
							},
							Range: hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 5,
									Byte:   4,
								},
								End: hcl.Pos{
									Line:   1,
									Column: 13,
									Byte:   12,
								},
							},
							Constraints: reference.OriginConstraints{
								{
									OfType: cty.String,
								},
							},
						},
						reference.LocalOrigin{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_instance"},
								lang.AttrStep{Name: "test"},
								lang.AttrStep{Name: "instance_type"},
							},
							Range: hcl.Range{
								Filename: "another.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 5,
									Byte:   4,
								},
								End: hcl.Pos{
									Line:   1,
									Column: 13,
									Byte:   12,
								},
							},
							Constraints: reference.OriginConstraints{
								{
									OfType: cty.String,
								},
							},
						},
					},
				},
			},
			lang.Path{Path: dirPath},
			"test.tf",
			hcl.Pos{
				Line:   2,
				Column: 4,
				Byte:   36,
			},
			ReferenceOrigins{
				{
					Path: lang.Path{Path: dirPath},
					Range: hcl.Range{
						Filename: "another.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 5,
							Byte:   4,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 13,
							Byte:   12,
						},
					},
				},
			},
		},
		{
			"matching nested targets with position at block definition",
			map[string]*PathContext{
				dirPath: {
					ReferenceTargets: reference.Targets{
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
					ReferenceOrigins: reference.Origins{
						reference.LocalOrigin{
							Addr: lang.Address{
								lang.RootStep{Name: "module"},
								lang.AttrStep{Name: "test"},
							},
							Range: hcl.Range{
								Filename: "first.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 5,
									Byte:   4,
								},
								End: hcl.Pos{
									Line:   1,
									Column: 13,
									Byte:   12,
								},
							},
							Constraints: reference.OriginConstraints{
								{
									OfType: cty.DynamicPseudoType,
								},
							},
						},
						reference.LocalOrigin{
							Addr: lang.Address{
								lang.RootStep{Name: "module"},
								lang.AttrStep{Name: "test"},
								lang.AttrStep{Name: "instance_id"},
							},
							Range: hcl.Range{
								Filename: "second.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 5,
									Byte:   4,
								},
								End: hcl.Pos{
									Line:   1,
									Column: 13,
									Byte:   12,
								},
							},
							Constraints: reference.OriginConstraints{
								{
									OfType: cty.String,
								},
							},
						},
					},
				},
			},
			lang.Path{Path: dirPath},
			"test.tf",
			hcl.Pos{
				Line:   1,
				Column: 15,
				Byte:   16,
			},
			ReferenceOrigins{
				{
					Path: lang.Path{Path: dirPath},
					Range: hcl.Range{
						Filename: "first.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 5,
							Byte:   4,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 13,
							Byte:   12,
						},
					},
				},
				{
					Path: lang.Path{Path: dirPath},
					Range: hcl.Range{
						Filename: "second.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 5,
							Byte:   4,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 13,
							Byte:   12,
						},
					},
				},
			},
		},
		{
			"matching path origin",
			map[string]*PathContext{
				filepath.Join(dirPath, "alpha"): {
					ReferenceTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "foo"},
							},
							ScopeId: lang.ScopeId("variable"),
							Type:    cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 1,
									Byte:   0,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 1,
									Byte:   35,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 1,
									Byte:   0,
								},
								End: hcl.Pos{
									Line:   1,
									Column: 17,
									Byte:   16,
								},
							},
						},
					},
				},
				filepath.Join(dirPath, "beta"): {
					ReferenceOrigins: reference.Origins{
						reference.PathOrigin{
							Range: hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
								End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
							},
							TargetAddr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "foo"},
							},
							TargetPath: lang.Path{
								Path: filepath.Join(dirPath, "alpha"),
							},
							Constraints: reference.OriginConstraints{},
						},
					},
				},
				filepath.Join(dirPath, "charlie"): {
					ReferenceOrigins: reference.Origins{
						reference.PathOrigin{
							Range: hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
								End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
							},
							TargetAddr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "bar"},
							},
							TargetPath: lang.Path{
								Path: filepath.Join(dirPath, "beta"),
							},
							Constraints: reference.OriginConstraints{},
						},
					},
				},
			},
			lang.Path{Path: filepath.Join(dirPath, "alpha")},
			"test.tf",
			hcl.Pos{
				Line:   1,
				Column: 15,
				Byte:   16,
			},
			ReferenceOrigins{
				{
					Path: lang.Path{Path: filepath.Join(dirPath, "beta")},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
						End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
					},
				},
			},
		},
		{
			"matching local and path origin",
			map[string]*PathContext{
				filepath.Join(dirPath, "alpha"): {
					ReferenceTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "foo"},
							},
							ScopeId: lang.ScopeId("variable"),
							Type:    cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 1,
									Byte:   0,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 1,
									Byte:   35,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 1,
									Byte:   0,
								},
								End: hcl.Pos{
									Line:   1,
									Column: 17,
									Byte:   16,
								},
							},
						},
					},
					ReferenceOrigins: reference.Origins{
						reference.LocalOrigin{
							Range: hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 6, Column: 20, Byte: 45},
								End:      hcl.Pos{Line: 6, Column: 27, Byte: 52},
							},
							Addr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "foo"},
							},
							Constraints: reference.OriginConstraints{},
						},
					},
				},
				filepath.Join(dirPath, "beta"): {
					ReferenceOrigins: reference.Origins{
						reference.PathOrigin{
							Range: hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
								End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
							},
							TargetAddr: lang.Address{
								lang.RootStep{Name: "var"},
								lang.AttrStep{Name: "foo"},
							},
							TargetPath: lang.Path{
								Path: filepath.Join(dirPath, "alpha"),
							},
							Constraints: reference.OriginConstraints{},
						},
					},
				},
			},
			lang.Path{Path: filepath.Join(dirPath, "alpha")},
			"test.tf",
			hcl.Pos{
				Line:   1,
				Column: 15,
				Byte:   16,
			},
			ReferenceOrigins{
				{
					Path: lang.Path{Path: filepath.Join(dirPath, "alpha")},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 6, Column: 20, Byte: 45},
						End:      hcl.Pos{Line: 6, Column: 27, Byte: 52},
					},
				},
				{
					Path: lang.Path{Path: filepath.Join(dirPath, "beta")},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
						End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			d := NewDecoder(&testPathReader{
				paths: tc.dirs,
			})
			origins := d.ReferenceOriginsTargetingPos(tc.path, tc.filename, tc.pos)

			if diff := cmp.Diff(tc.expectedOrigins, origins, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of reference origins: %s", diff)
			}
		})
	}
}
