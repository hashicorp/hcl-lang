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

func TestCollectRefOrigins_exprAny_references_hcl(t *testing.T) {
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
					Constraint: schema.AnyExpression{
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
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
					IsOptional: true,
				},
			},
			`attr = "${foo}"`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
						End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
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
			"traversal with string",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
					IsOptional: true,
				},
			},
			`attr = "${foo}-bar"`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
						End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
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
			"simple traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
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
					Constraint: schema.AnyExpression{
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
			"string which happens to match address",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
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

func TestCollectRefOrigins_exprAny_references_json(t *testing.T) {
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
					Constraint: schema.AnyExpression{
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
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
					IsOptional: true,
				},
			},
			`{"attr": "${foo}-bar"}`,
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
							OfType: cty.DynamicPseudoType,
						},
					},
				},
			},
		},
		{
			"simple traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
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
							OfType: cty.DynamicPseudoType,
						},
					},
				},
			},
		},
		{
			"traversal with numeric index steps",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
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
							OfType: cty.DynamicPseudoType,
						},
					},
				},
			},
		},
		{
			"traversal with string index steps",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
					IsOptional: true,
				},
			},
			`{"attr": "${one.two[\"key\"].attr[\"foo\"]}"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "one"},
						lang.AttrStep{Name: "two"},
						lang.IndexStep{Key: cty.StringVal("key")},
						lang.AttrStep{Name: "attr"},
						lang.IndexStep{Key: cty.StringVal("foo")},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						// HCL misreports traversals' range w/ string keys in JSON
						// See https://github.com/hashicorp/hcl/issues/598
						Start: hcl.Pos{Line: 1, Column: 13, Byte: 12},
						End:   hcl.Pos{Line: 1, Column: 39, Byte: 38 /* 42 */},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
						},
					},
				},
			},
		},
		{ // Terraform uses this in most places where it expects references only
			"legacy style string",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
					IsOptional: true,
				},
			},
			`{"attr": "foo.bar"}`,
			reference.Origins{},
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

func TestCollectRefOrigins_exprAny_functions_hcl(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefOrigins reference.Origins
	}{
		{
			"unknown function",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = unknown(var.foo)
`,
			reference.Origins{},
		},
		{
			"known function parameter",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = lower(var.foo)
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 14, Byte: 13},
						End:      hcl.Pos{Line: 1, Column: 21, Byte: 20},
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
			"known function variadic parameter",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = join(",", [var.foo])
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 19, Byte: 18},
						End:      hcl.Pos{Line: 1, Column: 26, Byte: 25},
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
			"too many arguments",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = lower("FOO", var.foo)
`,
			reference.Origins{},
		}, {
			"(unsupported) expression in function parameter",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = lower("${var.foo}")
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 17, Byte: 16},
						End:      hcl.Pos{Line: 1, Column: 24, Byte: 23},
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

			f, diags := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			if len(diags) > 0 {
				t.Error(diags)
			}
			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
				Functions: testFunctionSignatures(),
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

func TestCollectRefOrigins_exprAny_functions_json(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefOrigins reference.Origins
	}{
		{
			"unknown function",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`{"attr": "${unknown(var.foo)}"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 21, Byte: 20},
						End:      hcl.Pos{Line: 1, Column: 28, Byte: 27},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
						},
					},
				},
			},
		},
		{
			"known function parameter",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`{"attr": "${lower(var.foo)}"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 19, Byte: 18},
						End:      hcl.Pos{Line: 1, Column: 26, Byte: 25},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
						},
					},
				},
			},
		},
		{
			"known function variadic parameter",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`{"attr": "${join(\",\", [var.foo])}"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 24, Byte: 23},
						End:      hcl.Pos{Line: 1, Column: 31, Byte: 30},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
						},
					},
				},
			},
		},
		{
			"too many arguments",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`{"attr": "${lower(\"FOO\", var.foo)}"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 26, Byte: 25},
						End:      hcl.Pos{Line: 1, Column: 33, Byte: 32},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
						},
					},
				},
			},
		}, {
			"(unsupported) expression in function parameter",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`{"attr": "${lower(\"${var.foo}\")}"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 22, Byte: 21},
						End:      hcl.Pos{Line: 1, Column: 29, Byte: 28},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
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
				Functions: testFunctionSignatures(),
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

func TestCollectRefOrigins_exprAny_operators_hcl(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefOrigins reference.Origins
	}{
		{
			"binary operator",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Number,
					},
				},
			},
			`attr = var.foo + var.bar
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 15, Byte: 14},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.Number,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 18, Byte: 17},
						End:      hcl.Pos{Line: 1, Column: 25, Byte: 24},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.Number,
						},
					},
				},
			},
		},
		{
			"binary operator mismatching constraint",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
					},
				},
			},
			`attr = var.foo + var.bar
`,
			reference.Origins{},
		},
		{
			"unary operator",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
					},
				},
			},
			`attr = !var.foo
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.Bool,
						},
					},
				},
			},
		},
		{
			"unary operator mismatching constraint",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
					},
				},
			},
			`attr = !var.foo
`,
			reference.Origins{},
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

func TestCollectRefOrigins_exprAny_parenthesis_hcl(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefOrigins reference.Origins
	}{
		{
			"basic",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Number,
					},
				},
			},
			`attr = (var.foo + var.bar)
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 9, Byte: 8},
						End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.Number,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 19, Byte: 18},
						End:      hcl.Pos{Line: 1, Column: 26, Byte: 25},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.Number,
						},
					},
				},
			},
		},
		{
			"reference as map key",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
					},
				},
			},
			`attr = {
  (var.foo) = "foo"
}
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 4, Byte: 12},
						End:      hcl.Pos{Line: 2, Column: 11, Byte: 19},
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
			"reference as object attribute name",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Object(map[string]cty.Type{
							"bar": cty.String,
						}),
					},
				},
			},
			`attr = {
  (var.foo) = "foo"
}
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2, Column: 4, Byte: 12},
						End:      hcl.Pos{Line: 2, Column: 11, Byte: 19},
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

func TestCollectRefOrigins_exprAny_forExpr_hcl(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefOrigins reference.Origins
	}{
		{
			"list",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
					},
				},
			},
			`attr = [for key, val in var.coll: val]
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "coll"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 25, Byte: 24},
						End:      hcl.Pos{Line: 1, Column: 33, Byte: 32},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.List(cty.DynamicPseudoType)},
						{OfType: cty.Set(cty.DynamicPseudoType)},
						{OfType: cty.EmptyTuple},
						{OfType: cty.Map(cty.DynamicPseudoType)},
						{OfType: cty.EmptyObject},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "val"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 35, Byte: 34},
						End:      hcl.Pos{Line: 1, Column: 38, Byte: 37},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
			},
		},
		{
			"set",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Set(cty.String),
					},
				},
			},
			`attr = [for key, val in var.coll: val]
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "coll"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 25, Byte: 24},
						End:      hcl.Pos{Line: 1, Column: 33, Byte: 32},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.List(cty.DynamicPseudoType)},
						{OfType: cty.Set(cty.DynamicPseudoType)},
						{OfType: cty.EmptyTuple},
						{OfType: cty.Map(cty.DynamicPseudoType)},
						{OfType: cty.EmptyObject},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "val"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 35, Byte: 34},
						End:      hcl.Pos{Line: 1, Column: 38, Byte: 37},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
			},
		},
		{
			"tuple",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.EmptyTuple,
					},
				},
			},
			`attr = [for key, val in var.coll: val]
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "coll"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 25, Byte: 24},
						End:      hcl.Pos{Line: 1, Column: 33, Byte: 32},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.List(cty.DynamicPseudoType)},
						{OfType: cty.Set(cty.DynamicPseudoType)},
						{OfType: cty.EmptyTuple},
						{OfType: cty.Map(cty.DynamicPseudoType)},
						{OfType: cty.EmptyObject},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "val"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 35, Byte: 34},
						End:      hcl.Pos{Line: 1, Column: 38, Byte: 37},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.DynamicPseudoType},
					},
				},
			},
		},
		{
			"map",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Map(cty.String),
					},
				},
			},
			`attr = {for key, val in var.coll: key => val}
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "coll"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 25, Byte: 24},
						End:      hcl.Pos{Line: 1, Column: 33, Byte: 32},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.List(cty.DynamicPseudoType)},
						{OfType: cty.Set(cty.DynamicPseudoType)},
						{OfType: cty.EmptyTuple},
						{OfType: cty.Map(cty.DynamicPseudoType)},
						{OfType: cty.EmptyObject},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "key"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 35, Byte: 34},
						End:      hcl.Pos{Line: 1, Column: 38, Byte: 37},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "val"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 42, Byte: 41},
						End:      hcl.Pos{Line: 1, Column: 45, Byte: 44},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
			},
		},
		{
			"object",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.EmptyObject,
					},
				},
			},
			`attr = {for key, val in var.coll: key => val}
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "coll"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 25, Byte: 24},
						End:      hcl.Pos{Line: 1, Column: 33, Byte: 32},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.List(cty.DynamicPseudoType)},
						{OfType: cty.Set(cty.DynamicPseudoType)},
						{OfType: cty.EmptyTuple},
						{OfType: cty.Map(cty.DynamicPseudoType)},
						{OfType: cty.EmptyObject},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "key"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 35, Byte: 34},
						End:      hcl.Pos{Line: 1, Column: 38, Byte: 37},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "val"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 42, Byte: 41},
						End:      hcl.Pos{Line: 1, Column: 45, Byte: 44},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.DynamicPseudoType},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%2d-%s", i, tc.testName), func(t *testing.T) {
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

func TestCollectRefOrigins_exprAny_forExpr_forEach(t *testing.T) {
	bodySchema := &schema.BodySchema{
		Blocks: map[string]*schema.BlockSchema{
			"resource": {
				Labels: []*schema.LabelSchema{
					{Name: "type"}, {Name: "name"},
				},
				Body: &schema.BodySchema{
					Extensions: &schema.BodyExtensions{
						ForEach: true,
					},
				},
			},
		},
	}

	cfg := `resource "aws_instance" "foo" {
  for_each = { for i in var.coll : i.name => i }
}`

	f, diags := hclsyntax.ParseConfig([]byte(cfg), "test.tf", hcl.InitialPos)
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
				lang.RootStep{Name: "var"},
				lang.AttrStep{Name: "coll"},
			},
			Range: hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 2, Column: 25, Byte: 56},
				End:      hcl.Pos{Line: 2, Column: 33, Byte: 64},
			},
			Constraints: reference.OriginConstraints{
				{OfType: cty.List(cty.DynamicPseudoType)},
				{OfType: cty.Set(cty.DynamicPseudoType)},
				{OfType: cty.EmptyTuple},
				{OfType: cty.Map(cty.DynamicPseudoType)},
				{OfType: cty.EmptyObject},
				{OfType: cty.List(cty.DynamicPseudoType)},
				{OfType: cty.Set(cty.DynamicPseudoType)},
				{OfType: cty.EmptyTuple},
				{OfType: cty.Map(cty.DynamicPseudoType)},
				{OfType: cty.EmptyObject},
				{OfType: cty.List(cty.DynamicPseudoType)},
				{OfType: cty.Set(cty.DynamicPseudoType)},
				{OfType: cty.EmptyTuple},
				{OfType: cty.Map(cty.DynamicPseudoType)},
				{OfType: cty.EmptyObject},
			},
		},
		reference.LocalOrigin{
			Addr: lang.Address{
				lang.RootStep{Name: "i"},
				lang.AttrStep{Name: "name"},
			},
			Range: hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 2, Column: 36, Byte: 67},
				End:      hcl.Pos{Line: 2, Column: 42, Byte: 73},
			},
			Constraints: reference.OriginConstraints{
				{OfType: cty.String},
				{OfType: cty.String},
				{OfType: cty.String},
			},
		},
		reference.LocalOrigin{
			Addr: lang.Address{
				lang.RootStep{Name: "i"},
			},
			Range: hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 2, Column: 46, Byte: 77},
				End:      hcl.Pos{Line: 2, Column: 47, Byte: 78},
			},
			Constraints: reference.OriginConstraints{
				{OfType: cty.DynamicPseudoType},
				{OfType: cty.String},
				{OfType: cty.DynamicPseudoType},
			},
		},
	}

	if diff := cmp.Diff(expectedRefOrigins, origins, ctydebug.CmpOptions); diff != "" {
		t.Fatalf("unexpected origins: %s", diff)
	}
}

func TestCollectRefOrigins_exprAny_operators_json(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefOrigins reference.Origins
	}{
		{
			"binary operator",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Number,
					},
				},
			},
			`{"attr": "${var.foo + var.bar}"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 13, Byte: 12},
						End:      hcl.Pos{Line: 1, Column: 20, Byte: 19},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
						End:      hcl.Pos{Line: 1, Column: 30, Byte: 29},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
						},
					},
				},
			},
		},
		{
			"binary operator mismatching constraint",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
					},
				},
			},
			`{"attr": "${var.foo + var.bar}"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 13, Byte: 12},
						End:      hcl.Pos{Line: 1, Column: 20, Byte: 19},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 23, Byte: 22},
						End:      hcl.Pos{Line: 1, Column: 30, Byte: 29},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
						},
					},
				},
			},
		},
		{
			"unary operator",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.Bool,
					},
				},
			},
			`{"attr": "${!var.foo}"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 14, Byte: 13},
						End:      hcl.Pos{Line: 1, Column: 21, Byte: 20},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
						},
					},
				},
			},
		},
		{
			"unary operator mismatching constraint",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.List(cty.String),
					},
				},
			},
			`{"attr": "${!var.foo}"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 14, Byte: 13},
						End:      hcl.Pos{Line: 1, Column: 21, Byte: 20},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
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

func TestCollectRefOrigins_exprAny_templates_hcl(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefOrigins reference.Origins
	}{
		{
			"template expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = "${var.foo}_${var.bar}"
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
						End:      hcl.Pos{Line: 1, Column: 18, Byte: 17},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.String,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 22, Byte: 21},
						End:      hcl.Pos{Line: 1, Column: 29, Byte: 28},
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

func TestCollectRefOrigins_exprAny_templates_json(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefOrigins reference.Origins
	}{
		{
			"template expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`{"attr": "${var.foo}_${var.bar}"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 13, Byte: 12},
						End:      hcl.Pos{Line: 1, Column: 20, Byte: 19},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 24, Byte: 23},
						End:      hcl.Pos{Line: 1, Column: 31, Byte: 30},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
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

func TestCollectRefOrigins_exprAny_conditional_hcl(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefOrigins reference.Origins
	}{
		{
			"simple conditional",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = foo ? bar : baz
`,
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
							OfType: cty.Bool,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 14, Byte: 13},
						End:      hcl.Pos{Line: 1, Column: 17, Byte: 16},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.String,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "baz"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 20, Byte: 19},
						End:      hcl.Pos{Line: 1, Column: 23, Byte: 22},
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
			"conditional in template",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = "x-${foo ? bar : baz}"
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 13, Byte: 12},
						End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.Bool,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 19, Byte: 18},
						End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.String,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "baz"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 25, Byte: 24},
						End:      hcl.Pos{Line: 1, Column: 28, Byte: 27},
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
			"conditional as directive in template",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = "x-%{if foo}${bar}%{else}${baz}%{endif}"
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 16, Byte: 15},
						End:      hcl.Pos{Line: 1, Column: 19, Byte: 18},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.Bool,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 22, Byte: 21},
						End:      hcl.Pos{Line: 1, Column: 25, Byte: 24},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.String,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "baz"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 35, Byte: 34},
						End:      hcl.Pos{Line: 1, Column: 38, Byte: 37},
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
			"conditional as short directive in template",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`attr = "x-%{if foo}${bar}%{endif}"
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 16, Byte: 15},
						End:      hcl.Pos{Line: 1, Column: 19, Byte: 18},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.Bool,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 22, Byte: 21},
						End:      hcl.Pos{Line: 1, Column: 25, Byte: 24},
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

func TestCollectRefOrigins_exprAny_conditional_json(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefOrigins reference.Origins
	}{
		{
			"simple conditional",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`{"attr": "${foo ? bar : baz}"}`,
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
							OfType: cty.DynamicPseudoType,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 19, Byte: 18},
						End:      hcl.Pos{Line: 1, Column: 22, Byte: 21},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "baz"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 25, Byte: 24},
						End:      hcl.Pos{Line: 1, Column: 28, Byte: 27},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
						},
					},
				},
			},
		},
		{
			"conditional as directive",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`{"attr": "x-%{if foo}${bar}%{else}${baz}%{endif}"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 18, Byte: 17},
						End:      hcl.Pos{Line: 1, Column: 21, Byte: 20},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 24, Byte: 23},
						End:      hcl.Pos{Line: 1, Column: 27, Byte: 26},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "baz"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 37, Byte: 36},
						End:      hcl.Pos{Line: 1, Column: 40, Byte: 39},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
						},
					},
				},
			},
		},
		{
			"conditional as short directive",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.AnyExpression{
						OfType: cty.String,
					},
				},
			},
			`{"attr": "x-%{if foo}${bar}%{endif}"}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 18, Byte: 17},
						End:      hcl.Pos{Line: 1, Column: 21, Byte: 20},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "bar"},
					},
					Range: hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 24, Byte: 23},
						End:      hcl.Pos{Line: 1, Column: 27, Byte: 26},
					},
					Constraints: reference.OriginConstraints{
						{
							OfType: cty.DynamicPseudoType,
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
