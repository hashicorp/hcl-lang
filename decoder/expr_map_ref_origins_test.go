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

func TestCollectRefOrigins_exprMap_hcl(t *testing.T) {
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
					Constraint: schema.Map{
						Elem: schema.Reference{OfType: cty.Number},
					},
				},
			},
			`attr = [foobar]
`,
			reference.Origins{},
		},
		{
			"no origins",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Reference{OfType: cty.Number},
					},
				},
			},
			`attr = {
  foo = "noot"
  bar = "noot"
}
`,
			reference.Origins{},
		},
		{
			"one origin",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Reference{OfType: cty.Number},
					},
				},
			},
			`attr = {
  foo = foo.bar
  bar = "noot"
}
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 2, Column: 9, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 16, Byte: 24},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.Number},
					},
				},
			},
		},
		{
			"one origin with invalid key expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Reference{OfType: cty.Number},
					},
				},
			},
			`attr = {
  422 = bar
  foo = foo.bar
  bar = "noot"
}
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 2, Column: 9, Byte: 17},
						End:      hcl.Pos{Line: 2, Column: 12, Byte: 20},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.Number},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.hcl",
						Start:    hcl.Pos{Line: 3, Column: 9, Byte: 29},
						End:      hcl.Pos{Line: 3, Column: 16, Byte: 36},
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

func TestCollectRefOrigins_exprMap_json(t *testing.T) {
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
					Constraint: schema.Map{
						Elem: schema.Reference{OfType: cty.Number},
					},
				},
			},
			`{"attr": ["foobar"]}`,
			reference.Origins{},
		},
		{
			"no origins",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Reference{OfType: cty.Number},
					},
				},
			},
			`{"attr": {
  "foo": 42,
  "bar": true
}}`,
			reference.Origins{},
		},
		{
			"one origin",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Map{
						Elem: schema.Reference{OfType: cty.Number},
					},
				},
			},
			`{"attr": {
  "foo": "foo.bar",
  "bar": 42
}}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.hcl.json",
						Start:    hcl.Pos{Line: 2, Column: 11, Byte: 21},
						End:      hcl.Pos{Line: 2, Column: 18, Byte: 28},
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
