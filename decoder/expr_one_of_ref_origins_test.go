// Copyright IBM Corp. 2026
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

func TestCollectRefOrigins_exprOneOf_hcl(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefOrigins reference.Origins
	}{
		{
			"no origins",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.OneOf{
						schema.Reference{OfType: cty.String},
						schema.Reference{OfType: cty.Number},
					},
				},
			},
			`attr = "noot"
`,
			reference.Origins{},
		},
		{
			"one matching origin",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.OneOf{
						schema.Reference{OfType: cty.String},
						schema.Reference{OfType: cty.Number},
					},
				},
			},
			`attr = foo.bar
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
						{OfType: cty.Number},
					},
				},
			},
		},
		{
			"multiple origins",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.OneOf{
						schema.Reference{OfType: cty.Number},
						schema.Reference{OfType: cty.Bool},
					},
				},
			},
			`attr = "${foo}-${bar}"
`,
			reference.Origins{},
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

func TestCollectRefOrigins_exprOneOf_json(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefOrigins reference.Origins
	}{
		{
			"no origins",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.OneOf{
						schema.Reference{OfType: cty.String},
						schema.Reference{OfType: cty.Number},
					},
				},
			},
			`{"attr": "42"}`,
			reference.Origins{},
		},
		{
			"one matching origin interpolated",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.OneOf{
						schema.Reference{OfType: cty.String},
						schema.Reference{OfType: cty.Number},
					},
				},
			},
			`{"attr": "${foo.bar}"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 13, Byte: 12},
						End:      hcl.Pos{Line: 1, Column: 20, Byte: 19},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
						{OfType: cty.Number},
					},
				},
			},
		},
		{
			"one matching origin plaintext",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.OneOf{
						schema.Reference{OfType: cty.String},
						schema.Reference{OfType: cty.Number},
					},
				},
			},
			`{"attr": "foo.bar"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
						End:      hcl.Pos{Line: 1, Column: 18, Byte: 17},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
						{OfType: cty.Number},
					},
				},
			},
		},
		{
			"multiple origins",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.OneOf{
						schema.Reference{OfType: cty.Number},
						schema.Reference{OfType: cty.Bool},
					},
				},
			},
			`{"attr": "${foo}-${bar}"}`,
			reference.Origins{},
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
