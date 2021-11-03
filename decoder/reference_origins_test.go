package decoder

import (
	"fmt"
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
		targets         reference.Targets
		origins         reference.Origins
		filename        string
		pos             hcl.Pos
		expectedOrigins ReferenceOrigins
	}{
		{
			"no targets and no origins",
			reference.Targets{},
			reference.Origins{},
			"test.tf",
			hcl.InitialPos,
			ReferenceOrigins{},
		},
		{
			"target file mismatch",
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
			reference.Origins{},
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
			reference.Origins{},
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
			reference.Origins{
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
			reference.Origins{
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
			reference.Origins{
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
			reference.Origins{
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

			origins := pathDecoder.ReferenceOriginsTargetingPos(tc.filename, tc.pos)

			if diff := cmp.Diff(tc.expectedOrigins, origins, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of reference origins: %s", diff)
			}
		})
	}
}
