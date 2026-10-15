// Copyright IBM Corp. 2020, 2026
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
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestCollectReferenceOrigins_hcl_local(t *testing.T) {
	testCases := []struct {
		name            string
		schema          *schema.BodySchema
		cfg             string
		expectedOrigins reference.Origins
	}{
		{
			"no origins",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attribute": {
						Constraint: schema.LiteralType{Type: cty.String},
					},
				},
			},
			`attribute = "foo-bar"`,
			reference.Origins{},
		},
		{
			"root attribute single step",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Constraint: schema.Reference{},
					},
				},
			},
			`attr = onestep`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Constraints: reference.OriginConstraints{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 8,
							Byte:   7,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 15,
							Byte:   14,
						},
					},
				},
			},
		},
		{
			"multiple root attributes single step",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr1": {
						Constraint: schema.Reference{},
					},
					"attr2": {
						Constraint: schema.Reference{},
					},
					"attr3": {
						Constraint: schema.Reference{},
					},
				},
			},
			`attr1 = onestep
attr2 = anotherstep
attr3 = onestep`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Constraints: reference.OriginConstraints{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 9,
							Byte:   8,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 16,
							Byte:   15,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "anotherstep"},
					},
					Constraints: reference.OriginConstraints{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 9,
							Byte:   24,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 20,
							Byte:   35,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Constraints: reference.OriginConstraints{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   3,
							Column: 9,
							Byte:   44,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 16,
							Byte:   51,
						},
					},
				},
			},
		},
		{
			"root attribute multiple origins",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr1": {
						Constraint: schema.AnyExpression{},
					},
				},
			},
			`attr1 = "${onestep}-${onestep}-${another.foo.bar}"`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 12,
							Byte:   11,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 19,
							Byte:   18,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 23,
							Byte:   22,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 30,
							Byte:   29,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "another"},
						lang.AttrStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 34,
							Byte:   33,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 49,
							Byte:   48,
						},
					},
				},
			},
		},
		{
			"root attribute multi-step",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Constraint: schema.Reference{},
					},
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
					Constraints: reference.OriginConstraints{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 8,
							Byte:   7,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 30,
							Byte:   29,
						},
					},
				},
			},
		},
		{
			"attribute in block",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"myblock": {
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Constraint: schema.Reference{},
								},
							},
						},
					},
				},
			},
			`myblock {
  attr = onestep
}
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Constraints: reference.OriginConstraints{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 10,
							Byte:   19,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 17,
							Byte:   26,
						},
					},
				},
			},
		},
		{
			"any attribute in block",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"myblock": {
						Body: &schema.BodySchema{
							AnyAttribute: &schema.AttributeSchema{
								Constraint: schema.Reference{},
							},
						},
					},
				},
			},
			`myblock {
  attr = onestep
}
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Constraints: reference.OriginConstraints{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 10,
							Byte:   19,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 17,
							Byte:   26,
						},
					},
				},
			},
		},
		{
			"origins within block with matching dependent body",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"myblock": {
						Labels: []*schema.LabelSchema{
							{Name: "type", IsDepKey: true},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"static": {
									Constraint: schema.Reference{},
								},
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "special"},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"dep_attr": {
										Constraint: schema.Reference{},
									},
								},
							},
						},
					},
				},
			},
			`myblock "special" {
  static = var.first
  dep_attr = var.second
}
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Constraints: reference.OriginConstraints{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 12,
							Byte:   31,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 21,
							Byte:   40,
						},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
					Constraints: reference.OriginConstraints{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   3,
							Column: 14,
							Byte:   54,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 24,
							Byte:   64,
						},
					},
				},
			},
		},
		{
			"origins within block with mismatching dependent body",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"myblock": {
						Labels: []*schema.LabelSchema{
							{Name: "type", IsDepKey: true},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"static": {
									Constraint: schema.Reference{},
								},
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "special"},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"dep_attr": {
										Constraint: schema.Reference{},
									},
								},
							},
						},
					},
				},
			},
			`myblock "different" {
  static = var.first
  dep_attr = var.second
}
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Constraints: reference.OriginConstraints{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 12,
							Byte:   33,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 21,
							Byte:   42,
						},
					},
				},
			},
		},
		{
			"origin inside collection expressions",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"list": {
						Constraint: schema.List{
							Elem: schema.Reference{
								OfScopeId: lang.ScopeId("test"),
							},
						},
					},
					"set": {
						Constraint: schema.Set{
							Elem: schema.Reference{
								OfScopeId: lang.ScopeId("test"),
							},
						},
					},
					"tuple": {
						Constraint: schema.Tuple{
							Elems: []schema.Constraint{
								schema.Reference{
									OfScopeId: lang.ScopeId("test"),
								},
							},
						},
					},
				},
			},
			`list = [ var.first ]
set = [ var.second ]
tuple = [ var.third ]
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 10,
							Byte:   9,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 19,
							Byte:   18,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfScopeId: lang.ScopeId("test")},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 9,
							Byte:   29,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 19,
							Byte:   39,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfScopeId: lang.ScopeId("test")},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "third"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   3,
							Column: 11,
							Byte:   52,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 20,
							Byte:   61,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfScopeId: lang.ScopeId("test")},
					},
				},
			},
		},
		{
			"origin inside object expression",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"obj": {
						Constraint: schema.Object{
							Attributes: schema.ObjectAttributes{
								"attr": &schema.AttributeSchema{
									Constraint: schema.Reference{
										OfScopeId: lang.ScopeId("test"),
									},
								},
							},
						},
					},
				},
			},
			`obj = {
  attr = var.first
}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 10,
							Byte:   17,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 19,
							Byte:   26,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfScopeId: lang.ScopeId("test")},
					},
				},
			},
		},
		{
			"origin inside map expression",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"map": {
						Constraint: schema.Map{
							Elem: schema.Reference{
								OfScopeId: lang.ScopeId("test"),
							},
						},
					},
				},
			},
			`map = {
  key = var.first
}`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 9,
							Byte:   16,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 18,
							Byte:   25,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfScopeId: lang.ScopeId("test")},
					},
				},
			},
		},
		{
			"origin inside object and map expression with multiple matches",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"map": {
						Constraint: schema.OneOf{
							schema.Reference{OfType: cty.Map(cty.String)},
							schema.Map{
								Elem: schema.Reference{OfType: cty.String},
							},
						},
					},
					"obj": {
						Constraint: schema.OneOf{
							schema.Reference{OfType: cty.Object(map[string]cty.Type{
								"foo": cty.String,
							})},
							schema.Object{
								Attributes: schema.ObjectAttributes{
									"foo": &schema.AttributeSchema{
										Constraint: schema.Reference{OfType: cty.String},
									},
								},
							},
						},
					},
				},
			},
			`map = {
  bar = var.one
}
obj = {
  foo = var.two
}
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "one"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 9,
							Byte:   16,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 16,
							Byte:   23,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "two"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   5,
							Column: 9,
							Byte:   42,
						},
						End: hcl.Pos{
							Line:   5,
							Column: 16,
							Byte:   49,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
			},
		},
		{
			"origin inside list, set and tuple expression with multiple matches",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"list": {
						Constraint: schema.OneOf{
							schema.Reference{OfType: cty.List(cty.String)},
							schema.List{
								Elem: schema.Reference{OfType: cty.String},
							},
						},
					},
					"set": {
						Constraint: schema.OneOf{
							schema.Reference{OfType: cty.Set(cty.String)},
							schema.Set{
								Elem: schema.Reference{OfType: cty.String},
							},
						},
					},
					"tup": {
						Constraint: schema.OneOf{
							schema.Reference{OfType: cty.Tuple([]cty.Type{cty.String})},
							schema.Tuple{
								Elems: []schema.Constraint{
									schema.Reference{OfType: cty.String},
								},
							},
						},
					},
				},
			},
			`list = [ var.one ]
set = [ var.two ]
tup = [ var.three ]
`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "one"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 10,
							Byte:   9,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 17,
							Byte:   16,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "two"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 9,
							Byte:   27,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 16,
							Byte:   34,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "three"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   3,
							Column: 9,
							Byte:   45,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 18,
							Byte:   54,
						},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
			},
		},
		{
			"schema with implied origin",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Constraint: schema.Reference{},
					},
				},
				ImpliedOrigins: schema.ImpliedOrigins{
					{
						OriginAddress: lang.Address{
							lang.RootStep{Name: "module"},
							lang.AttrStep{Name: "refname"},
							lang.AttrStep{Name: "outname"},
						},
						TargetAddress: lang.Address{
							lang.RootStep{Name: "output"},
							lang.AttrStep{Name: "outname"},
						},
						Path:        lang.Path{Path: "./local", LanguageID: "terraform"},
						Constraints: schema.Constraints{ScopeId: "output"},
					},
				},
			},
			`attr = module.refname.outname`,
			reference.Origins{
				reference.LocalOrigin{
					Addr: lang.Address{
						lang.RootStep{Name: "module"},
						lang.AttrStep{Name: "refname"},
						lang.AttrStep{Name: "outname"},
					},
					Constraints: reference.OriginConstraints{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 8,
							Byte:   7,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 30,
							Byte:   29,
						},
					},
				},
				reference.PathOrigin{
					TargetAddr: lang.Address{
						lang.RootStep{Name: "output"},
						lang.AttrStep{Name: "outname"},
					},
					TargetPath: lang.Path{Path: "./local", LanguageID: "terraform"},
					Constraints: reference.OriginConstraints{{
						OfScopeId: "output",
					}},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 8,
							Byte:   7,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 30,
							Byte:   29,
						},
					},
				},
			},
		},
		{
			"cyclical origin referring back to the attribute",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Address: &schema.AttributeAddrSchema{
							Steps: schema.Address{
								schema.StaticStep{Name: "root"},
								schema.AttrNameStep{},
							},
						},
						IsOptional: true,
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{OfType: cty.String},
							schema.LiteralTypeExpr{Type: cty.String},
						},
					},
				},
			},
			`attr = root.attr`,
			reference.Origins{},
		},
		{
			"cyclical origin referring back to the attribute with implied address",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"blk": {
						Address: &schema.BlockAddrSchema{
							Steps: schema.Address{
								schema.StaticStep{Name: "blk"},
							},
							BodyAsData: true,
							InferBody:  true,
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									IsOptional: true,
									Expr: schema.ExprConstraints{
										schema.TraversalExpr{OfType: cty.String},
										schema.LiteralTypeExpr{Type: cty.String},
									},
								},
							},
						},
					},
				},
			},
			`blk {
  attr = blk.attr
}
`,
			reference.Origins{},
		},
		{
			"cyclical origin referring back to the block with implied address",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"blk": {
						Address: &schema.BlockAddrSchema{
							Steps: schema.Address{
								schema.StaticStep{Name: "blk"},
							},
							BodyAsData: true,
							InferBody:  true,
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									IsOptional: true,
									Expr: schema.ExprConstraints{
										schema.TraversalExpr{OfType: cty.String},
										schema.LiteralTypeExpr{Type: cty.String},
									},
								},
							},
						},
					},
				},
			},
			`blk {
  attr = blk
}
`,
			reference.Origins{},
		},
		{
			"cyclical origin referring to another attribute in the same the block",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"blk": {
						Address: &schema.BlockAddrSchema{
							Steps: schema.Address{
								schema.StaticStep{Name: "blk"},
							},
							BodyAsData: true,
							InferBody:  true,
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									IsOptional: true,
									Expr: schema.ExprConstraints{
										schema.TraversalExpr{OfType: cty.String},
										schema.LiteralTypeExpr{Type: cty.String},
									},
								},
							},
						},
					},
				},
			},
			`blk {
  foo = "test"
  attr = blk.foo
}
`,
			reference.Origins{},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%2d-%s", i, tc.name), func(t *testing.T) {
			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)

			d := testPathDecoder(t, &PathContext{
				Schema: tc.schema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			})

			origins, err := d.CollectReferenceOrigins()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedOrigins, origins, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatched reference origins: %s", diff)
			}
		})
	}
}

func TestCollectReferenceOrigins_hcl_path(t *testing.T) {
	testCases := []struct {
		name            string
		schema          *schema.BodySchema
		cfg             string
		expectedOrigins reference.Origins
	}{
		{
			"attribute with path target",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Constraint: schema.OneOf{
							schema.Reference{OfType: cty.String},
							schema.LiteralType{Type: cty.String},
						},
						OriginForTarget: &schema.PathTarget{
							Address: schema.Address{
								schema.StaticStep{Name: "var"},
								schema.AttrNameStep{},
							},
							Path: lang.Path{
								Path:       "another-path",
								LanguageID: "terraform",
							},
						},
					},
				},
			},
			`attr = "test"`,
			reference.Origins{
				reference.PathOrigin{
					TargetAddr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "attr"},
					},
					TargetPath: lang.Path{
						Path:       "another-path",
						LanguageID: "terraform",
					},
					Constraints: reference.OriginConstraints{{}},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 5,
							Byte:   4,
						},
					},
				},
			},
		},
		{
			"dependent attribute with path target",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"module": {
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"source": {
									Constraint: schema.LiteralType{Type: cty.String},
									IsDepKey:   true,
								},
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Attributes: []schema.AttributeDependent{
									{
										Name: "source",
										Expr: schema.ExpressionValue{
											Static: cty.StringVal("./submodule"),
										},
									},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"attr": {
										Constraint: schema.OneOf{
											schema.Reference{OfType: cty.String},
											schema.LiteralType{Type: cty.String},
										},
										OriginForTarget: &schema.PathTarget{
											Address: schema.Address{
												schema.StaticStep{Name: "var"},
												schema.AttrNameStep{},
											},
											Path: lang.Path{
												Path:       "./submodule",
												LanguageID: "terraform",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			`module "test" {
  source = "./submodule"
  attr = "test"
}`,
			reference.Origins{
				reference.PathOrigin{
					TargetAddr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "attr"},
					},
					TargetPath: lang.Path{
						Path:       "./submodule",
						LanguageID: "terraform",
					},
					Constraints: reference.OriginConstraints{{}},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   3,
							Column: 3,
							Byte:   43,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 7,
							Byte:   47,
						},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d/%s", i, tc.name), func(t *testing.T) {
			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)

			d := testPathDecoder(t, &PathContext{
				Schema: tc.schema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			})

			origins, err := d.CollectReferenceOrigins()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedOrigins, origins, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatched reference origins: %s", diff)
			}
		})
	}
}
