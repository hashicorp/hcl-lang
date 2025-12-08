// Copyright IBM Corp. 2020, 2025
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestCollectRefOrigins_exprTuple_hcl(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefOrigins reference.Origins
	}{
		{
			"expression mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Reference{OfType: cty.Number},
							schema.Reference{OfType: cty.String},
						},
					},
				},
			},
			`attr = foo.bar
`,
			reference.Origins{},
		},
		{
			"no origins",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Reference{OfType: cty.Number},
							schema.Reference{OfType: cty.String},
						},
					},
				},
			},
			`attr = ["noot"]
`,
			reference.Origins{},
		},
		{
			"first origin",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Reference{OfType: cty.Number},
							schema.Reference{OfType: cty.String},
						},
					},
				},
			},
			`attr = [foo.bar]
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.Number},
					},
				},
			},
		},
		{
			"extra origin",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Reference{OfType: cty.Number},
							schema.Reference{OfType: cty.String},
						},
					},
				},
			},
			`attr = [foo, bar, baz]
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.Number},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 14, Byte: 13},
						End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
			},
		},
		{
			"multiple origins with skipped invalid expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Reference{OfType: cty.Number},
							schema.Reference{OfType: cty.String},
							schema.Reference{OfType: cty.Number},
						},
					},
				},
			},
			`attr = [foo, "noot", bar]
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.Number},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 22, Byte: 21},
						End:      hcl.Pos{Line: 1, Column: 25, Byte: 24},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.Number},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			bodySchema := &schema.BodySchema{
				Attributes: tc.attrSchema,
			}

			f, diags := hclsyntax.ParseConfig([]byte(tc.cfg), "test.hcl", hcl.InitialPos)
			if len(diags) > 0 {
				t.Error(diags)
			}
			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.hcl": f,
				},
			})

			origins, err := d.CollectReferenceOrigins()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedRefOrigins, origins, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("unexpected origins: %s", diff)
			}
		})
	}
}

func TestCollectRefOrigins_exprTuple_json(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefOrigins reference.Origins
	}{
		{
			"expression mismatch",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Reference{OfType: cty.Number},
							schema.Reference{OfType: cty.String},
						},
					},
				},
			},
			`{"attr": "foo.bar"}`,
			reference.Origins{},
		},
		{
			"no origins",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Reference{OfType: cty.Number},
							schema.Reference{OfType: cty.String},
						},
					},
				},
			},
			`{"attr": [42]}`,
			reference.Origins{},
		},
		{
			"first origin",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Reference{OfType: cty.Number},
							schema.Reference{OfType: cty.String},
						},
					},
				},
			},
			`{"attr": ["foo.bar"]}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 12, Byte: 11},
						End:      hcl.Pos{Line: 1, Column: 19, Byte: 18},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.Number},
					},
				},
			},
		},
		{
			"extra origin",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Reference{OfType: cty.Number},
							schema.Reference{OfType: cty.String},
						},
					},
				},
			},
			`{"attr": ["foo", "bar", "baz"]}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 12, Byte: 11},
						End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.Number},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 19, Byte: 18},
						End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
			},
		},
		{
			"multiple origins with skipped invalid expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Tuple{
						Elems: []schema.Constraint{
							schema.Reference{OfType: cty.Number},
							schema.Reference{OfType: cty.String},
							schema.Reference{OfType: cty.Number},
						},
					},
				},
			},
			`{"attr": ["foo", 42224, "bar"]}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 12, Byte: 11},
						End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.Number},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 26, Byte: 25},
						End:      hcl.Pos{Line: 1, Column: 29, Byte: 28},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.Number},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			bodySchema := &schema.BodySchema{
				Attributes: tc.attrSchema,
			}

			f, diags := json.ParseWithStartPos([]byte(tc.cfg), "test.hcl.json", hcl.InitialPos)
			if len(diags) > 0 {
				t.Error(diags)
			}
			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.hcl.json": f,
				},
			})

			origins, err := d.CollectReferenceOrigins()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedRefOrigins, origins, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("unexpected origins: %s", diff)
			}
		})
	}
}
