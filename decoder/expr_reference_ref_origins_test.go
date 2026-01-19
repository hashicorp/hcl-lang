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

func TestCollectRefOrigins_exprReference_hcl(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefOrigins reference.Origins
	}{
		{
			"no traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
					},
					IsOptional: true,
				},
			},
			`attr = "foo"`,
			reference.Origins{},
		},
		{
			"wrapped traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
					},
					IsOptional: true,
				},
			},
			`attr = "${foo}"`,
			reference.Origins{},
		},
		{
			"traversal with string",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
					},
					IsOptional: true,
				},
			},
			`attr = "${foo}-bar"`,
			reference.Origins{},
		},
		{
			"simple traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
					},
					IsOptional: true,
				},
			},
			`attr = foo`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.String,
						},
					},
				},
			},
		},
		{
			"traversal with index steps",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
					},
					IsOptional: true,
				},
			},
			`attr = one.two["key"].attr[0]`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "one"},
						lang.AttrStep{Name: "two"},
						lang.IndexStep{Key: cty.StringVal("key")},
						lang.AttrStep{Name: "attr"},
						lang.IndexStep{Key: cty.NumberIntVal(0)},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 30, Byte: 29},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.String,
						},
					},
				},
			},
		},
		{
			"simple traversal - scope and type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType:    cty.String,
						OfScopeId: lang.ScopeId("foobar"),
					},
					IsOptional: true,
				},
			},
			`attr = foo`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType:    cty.String,
							OfScopeId: lang.ScopeId("foobar"),
						},
					},
				},
			},
		},
		{
			"string which happens to match address",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
					},
					IsOptional: true,
				},
			},
			`attr = "foo"`,
			reference.Origins{
				// This should only work in JSON
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			bodySchema := &schema.BodySchema{
				Attributes: tc.attrSchema,
			}

			f, diags := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			if len(diags) > 0 {
				t.Error(diags)
			}
			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
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

func TestCollectRefOrigins_exprReference_json(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefOrigins reference.Origins
	}{
		{
			"no traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
					},
					IsOptional: true,
				},
			},
			`{"attr": 422}`,
			reference.Origins{},
		},
		{
			"traversal with string",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
					},
					IsOptional: true,
				},
			},
			`{"attr": "${foo}-bar"}`,
			reference.Origins{},
		},
		{
			"simple traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
					},
					IsOptional: true,
				},
			},
			`{"attr": "${foo}"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 13, Byte: 12},
						End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.String,
						},
					},
				},
			},
		},
		{
			"traversal with numeric index steps",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
					},
					IsOptional: true,
				},
			},
			`{"attr": "${one.two[42].attr[0]}"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "one"},
						lang.AttrStep{Name: "two"},
						lang.IndexStep{Key: cty.NumberIntVal(42)},
						lang.AttrStep{Name: "attr"},
						lang.IndexStep{Key: cty.NumberIntVal(0)},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 13, Byte: 12},
						End:      hcl.Pos{Line: 1, Column: 32, Byte: 31},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.String,
						},
					},
				},
			},
		},
		{
			"traversal with string index steps",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
					},
					IsOptional: true,
				},
			},
			`{"attr": "${one.two[\"key\"].attr[\"foo\"]}"}`,
			reference.Origins{
				// HCL misreports traversals' range w/ string keys in JSON
				// See https://github.com/hashicorp/hcl/issues/598
			},
		},
		{ // Terraform uses this in most places where it expects references only
			"legacy style string",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
					},
					IsOptional: true,
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
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
						End:      hcl.Pos{Line: 1, Column: 18, Byte: 17},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.String,
						},
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

			f, diags := json.ParseWithStartPos([]byte(tc.cfg), "test.tf.json", hcl.InitialPos)
			if len(diags) > 0 {
				t.Error(diags)
			}
			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.tf.json": f,
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

func TestCollectRefOrigins_exprReference_self(t *testing.T) {
	bodySchema := &schema.BodySchema{
		Attributes: map[string]*schema.AttributeSchema{
			"attr": {
				Constraint: schema.Reference{
					OfType:    cty.String,
					OfScopeId: lang.ScopeId("foobar"),
				},
			},
		},
		Extensions: &schema.BodyExtensions{
			SelfRefs: true,
		},
	}

	f, diags := hclsyntax.ParseConfig([]byte(`attr = self`), "test.tf", hcl.InitialPos)
	if len(diags) > 0 {
		t.Error(diags)
	}
	d := testPathDecoder(t, &PathContext{
		Schema: bodySchema,
		Files: map[string]*hcl.File{
			"test.tf": f,
		},
	})

	origins, err := d.CollectReferenceOrigins()
	if err != nil {
		t.Fatal(err)
	}

	expectedRefOrigins := reference.Origins{
		reference.LocalOrigin{
			Addr: lang.Address{
				lang.RootStep{Name: "self"},
			},
			Range: hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
				End:      hcl.Pos{Line: 1, Column: 12, Byte: 11},
			},
			Constraints: reference.OriginConstraints{
				{
					OfType:    cty.String,
					OfScopeId: lang.ScopeId("foobar"),
				},
			},
		},
	}

	if diff := cmp.Diff(expectedRefOrigins, origins, ctydebug.CmpOptions); diff != "" {
		t.Fatalf("unexpected origins: %s", diff)
	}

}
