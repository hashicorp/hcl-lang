// Copyright IBM Corp. 2020, 2025
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestHoverAtPos_exprReference(t *testing.T) {
	testCases := []struct {
		testName          string
		attrSchema        map[string]*schema.AttributeSchema
		refOrigins        reference.Origins
		refTargets        reference.Targets
		cfg               string
		pos               hcl.Pos
		expectedHoverData *lang.HoverData
	}{
		{
			"unknown origin",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "d"},
						lang.AttrStep{Name: "fx"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 13, Byte: 12},
						End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.String,
						},
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 13, Byte: 29},
					},
				},
			},
			`attr = l.ca+d.fx
foo = "noot"
`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			nil,
		},
		{
			"matching origin no target",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.String,
						},
					},
				},
			},
			reference.Targets{},
			`attr = local.foo
foo = "noot"
`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			nil,
		},
		{
			"matching origin and target",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
					},
				},
			},
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.String,
						},
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 13, Byte: 29},
					},
				},
			},
			`attr = local.foo
foo = "noot"
`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			&lang.HoverData{
				Content: lang.Markdown("`local.foo`\n_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
					End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
				},
			},
		},
		{
			"matching origin and target inside set",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Set{
						Elem: schema.OneOf{
							schema.Reference{OfScopeId: lang.ScopeId("one")},
							schema.Reference{OfScopeId: lang.ScopeId("two")},
							schema.Reference{OfScopeId: lang.ScopeId("three")},
						},
					},
				},
			},
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
						End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
					},
					Constraints: reference.OriginConstraints{
						{
							OfScopeId: lang.ScopeId("two"),
						},
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
					ScopeId: lang.ScopeId("two"),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 1, Byte: 19},
						End:      hcl.Pos{Line: 2, Column: 13, Byte: 31},
					},
				},
			},
			`attr = [ foo.bar ]
foo = "noot"
`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			&lang.HoverData{
				Content: lang.Markdown("`foo.bar` reference"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
					End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			bodySchema := &schema.BodySchema{
				Attributes: tc.attrSchema,
			}

			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
				ReferenceOrigins: tc.refOrigins,
				ReferenceTargets: tc.refTargets,
			})

			ctx := context.Background()
			hoverData, err := d.HoverAtPos(ctx, "test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedHoverData, hoverData); diff != "" {
				t.Fatalf("unexpected hover data: %s", diff)
			}
		})
	}
}
