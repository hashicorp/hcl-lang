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

func TestCollectReferenceTargets_hcl(t *testing.T) {
	testCases := []struct {
		name         string
		schema       *schema.BodySchema
		cfg          string
		expectedRefs reference.Targets
	}{
		{
			"root attribute as reference",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"testattr": {
						Address: &schema.AttributeAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "special"},
								schema.AttrNameStep{},
							},
							AsReference: true,
							ScopeId:     lang.ScopeId("specialthing"),
						},
						IsOptional: true,
						Expr:       schema.LiteralTypeOnly(cty.String),
					},
				},
			},
			`testattr = "example"
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "special"},
						lang.AttrStep{Name: "testattr"},
					},
					ScopeId: lang.ScopeId("specialthing"),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 21,
							Byte:   20,
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
							Column: 9,
							Byte:   8,
						},
					},
				},
			},
		},
		{
			"root attribute as string type",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"testattr": {
						Address: &schema.AttributeAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "special"},
								schema.AttrNameStep{},
							},
							AsExprType: true,
						},
						IsOptional: true,
						Expr:       schema.LiteralTypeOnly(cty.String),
					},
				},
			},
			`testattr = "example"
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "special"},
						lang.AttrStep{Name: "testattr"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 21,
							Byte:   20,
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
							Column: 9,
							Byte:   8,
						},
					},
				},
			},
		},
		{
			"root attribute as any type",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"testattr": {
						Address: &schema.AttributeAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "special"},
								schema.AttrNameStep{},
							},
							AsExprType: true,
						},
						IsOptional: true,
						Expr:       schema.LiteralTypeOnly(cty.DynamicPseudoType),
					},
				},
			},
			`testattr = "example"
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "special"},
						lang.AttrStep{Name: "testattr"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 21,
							Byte:   20,
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
							Column: 9,
							Byte:   8,
						},
					},
				},
			},
		},
		{
			"root attribute as object type",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"testattr": {
						Address: &schema.AttributeAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "special"},
								schema.AttrNameStep{},
							},
							AsExprType: true,
						},
						IsOptional: true,
						Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{
							"nestedattr": cty.String,
						})),
					},
				},
			},
			`testattr = {
	nestedattr = "test"
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "special"},
						lang.AttrStep{Name: "testattr"},
					},
					Type: cty.Object(map[string]cty.Type{
						"nestedattr": cty.String,
					}),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 2,
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
							Column: 9,
							Byte:   8,
						},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "special"},
								lang.AttrStep{Name: "testattr"},
								lang.AttrStep{Name: "nestedattr"},
							},
							Type: cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 2,
									Byte:   14,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 21,
									Byte:   33,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 2,
									Byte:   14,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 12,
									Byte:   24,
								},
							},
						},
					},
				},
			},
		},
		{
			"root attribute as map type",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"testattr": {
						Address: &schema.AttributeAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "special"},
								schema.AttrNameStep{},
							},
							AsExprType: true,
						},
						IsOptional: true,
						Expr:       schema.LiteralTypeOnly(cty.Map(cty.String)),
					},
				},
			},
			`testattr = {
	nestedattr = "test"
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "special"},
						lang.AttrStep{Name: "testattr"},
					},
					Type: cty.Map(cty.String),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 2,
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
							Column: 9,
							Byte:   8,
						},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "special"},
								lang.AttrStep{Name: "testattr"},
								lang.IndexStep{Key: cty.StringVal("nestedattr")},
							},
							Type: cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 2,
									Byte:   14,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 21,
									Byte:   33,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 2,
									Byte:   14,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 12,
									Byte:   24,
								},
							},
						},
					},
				},
			},
		},
		{
			"root attribute as list type",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"testattr": {
						Address: &schema.AttributeAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "special"},
								schema.AttrNameStep{},
							},
							AsExprType: true,
						},
						Expr:       schema.LiteralTypeOnly(cty.List(cty.String)),
						IsOptional: true,
					},
				},
			},
			`testattr = [ "example" ]
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "special"},
						lang.AttrStep{Name: "testattr"},
					},
					Type: cty.List(cty.String),
					DefRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 9,
							Byte:   8,
						},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 25,
							Byte:   24,
						},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "special"},
								lang.AttrStep{Name: "testattr"},
								lang.IndexStep{Key: cty.NumberIntVal(0)},
							},
							Type:        cty.String,
							DefRangePtr: nil,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 14,
									Byte:   13,
								},
								End: hcl.Pos{
									Line:   1,
									Column: 23,
									Byte:   22,
								},
							},
						},
					},
				},
			},
		},
		{
			"root attribute as list expression",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"testattr": {
						Address: &schema.AttributeAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "special"},
								schema.AttrNameStep{},
							},
							AsExprType: true,
						},
						Expr: schema.ExprConstraints{
							schema.ListExpr{
								Elem: schema.LiteralTypeOnly(cty.String),
							},
						},
						IsOptional: true,
					},
				},
			},
			`testattr = [ "example" ]
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "special"},
						lang.AttrStep{Name: "testattr"},
					},
					Type: cty.List(cty.String),
					DefRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 9,
							Byte:   8,
						},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 25,
							Byte:   24,
						},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "special"},
								lang.AttrStep{Name: "testattr"},
								lang.IndexStep{Key: cty.NumberIntVal(0)},
							},
							Type:        cty.String,
							DefRangePtr: nil,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 14,
									Byte:   13,
								},
								End: hcl.Pos{
									Line:   1,
									Column: 23,
									Byte:   22,
								},
							},
						},
					},
				},
			},
		},
		{
			"root attribute with undeclared type",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"testattr": {
						Address: &schema.AttributeAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "special"},
								schema.AttrNameStep{},
							},
							AsExprType: true,
						},
						IsOptional: true,
					},
				},
			},
			`testattr = "example"
`,
			reference.Targets{},
		},
		{
			"block with mismatching steps",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type"},
							{Name: "name"},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
								schema.LabelStep{Index: 1},
							},
							AsReference: true,
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Expr:       schema.LiteralTypeOnly(cty.Number),
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`resource "blah" {
	attr = 3
}
`,
			reference.Targets{},
		},
		{
			"block as ref only",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type"},
							{Name: "name"},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
								schema.LabelStep{Index: 1},
							},
							AsReference: true,
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Expr:       schema.LiteralTypeOnly(cty.Number),
									IsOptional: true,
								},
								"name": {
									Expr:       schema.LiteralTypeOnly(cty.String),
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`resource "blah" "test" {
	attr = 3
	name = "lorem ipsum"
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "blah"},
						lang.AttrStep{Name: "test"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 2,
							Byte:   58,
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
							Column: 23,
							Byte:   22,
						},
					},
				},
			},
		},
		{
			"block as data - single object",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type"},
							{Name: "name"},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
								schema.LabelStep{Index: 1},
							},
							BodyAsData: true,
						},
						Type: schema.BlockTypeObject,
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Expr:       schema.LiteralTypeOnly(cty.Number),
									IsOptional: true,
								},
								"name": {
									Expr:       schema.LiteralTypeOnly(cty.String),
									IsOptional: true,
								},
								"map_attr": {
									Expr: schema.ExprConstraints{
										schema.MapExpr{Elem: schema.LiteralTypeOnly(cty.String)},
									},
									IsOptional: true,
								},
								"list_attr": {
									Expr: schema.ExprConstraints{
										schema.ListExpr{Elem: schema.LiteralTypeOnly(cty.String)},
									},
									IsOptional: true,
								},
								"set_attr": {
									Expr: schema.ExprConstraints{
										schema.SetExpr{Elem: schema.LiteralTypeOnly(cty.String)},
									},
									IsOptional: true,
								},
								"tuple_attr": {
									Expr: schema.ExprConstraints{
										schema.TupleExpr{Elems: []schema.ExprConstraints{
											schema.LiteralTypeOnly(cty.String),
											schema.LiteralTypeOnly(cty.Number),
										}},
									},
									IsOptional: true,
								},
								"obj_attr": {
									Expr: schema.ExprConstraints{
										schema.ObjectExpr{
											Attributes: schema.ObjectExprAttributes{
												"example": &schema.AttributeSchema{
													Expr: schema.LiteralTypeOnly(cty.String),
												},
											},
										},
									},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`resource "blah" "test" {
	attr = 3
	name = "lorem ipsum"
	map_attr = {
		one = "hello"
		two = "world"
	}
	list_attr = [ "one", "two" ]
	set_attr = [ "one", "two" ]
	tuple_attr = [ "one", 42 ]
	obj_attr = {
		example = "blah"
	}
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "blah"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.Object(map[string]cty.Type{
						"attr":       cty.Number,
						"name":       cty.String,
						"map_attr":   cty.Map(cty.String),
						"list_attr":  cty.List(cty.String),
						"set_attr":   cty.Set(cty.String),
						"tuple_attr": cty.Tuple([]cty.Type{cty.String, cty.Number}),
						"obj_attr": cty.Object(map[string]cty.Type{
							"example": cty.String,
						}),
					}),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   14,
							Column: 2,
							Byte:   230,
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
							Column: 23,
							Byte:   22,
						},
					},
				},
			},
		},
		{
			"block as data - list of objects",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type"},
							{Name: "name"},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
								schema.LabelStep{Index: 1},
							},
							BodyAsData: true,
						},
						Type: schema.BlockTypeList,
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Expr:       schema.LiteralTypeOnly(cty.Number),
									IsOptional: true,
								},
								"name": {
									Expr:       schema.LiteralTypeOnly(cty.String),
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`resource "blah" "test" {
	attr = 3
	name = "lorem ipsum"
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "blah"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.List(cty.Object(map[string]cty.Type{
						"attr": cty.Number,
						"name": cty.String,
					})),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 2,
							Byte:   58,
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
							Column: 23,
							Byte:   22,
						},
					},
				},
			},
		},
		{
			"block as data - set of objects",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Labels: []*schema.LabelSchema{
							{Name: "type"},
							{Name: "name"},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
								schema.LabelStep{Index: 1},
							},
							BodyAsData: true,
						},
						Type: schema.BlockTypeSet,
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Expr:       schema.LiteralTypeOnly(cty.Number),
									IsOptional: true,
								},
								"name": {
									Expr:       schema.LiteralTypeOnly(cty.String),
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`resource "blah" "test" {
	attr = 3
	name = "lorem ipsum"
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "blah"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.Set(cty.Object(map[string]cty.Type{
						"attr": cty.Number,
						"name": cty.String,
					})),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 2,
							Byte:   58,
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
							Column: 23,
							Byte:   22,
						},
					},
				},
			},
		},
		{
			"block as data - map",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"listener": {
						Labels: []*schema.LabelSchema{
							{Name: "name"},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
							},
							BodyAsData: true,
						},
						Type: schema.BlockTypeMap,
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"source_port": {
									Expr:       schema.LiteralTypeOnly(cty.Number),
									IsOptional: true,
								},
								"protocol": {
									Expr:       schema.LiteralTypeOnly(cty.String),
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`listener "http" {
	source_port = 80
	protocol = "tcp"
}
listener "https" {
	source_port = 443
	protocol = "tcp"
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "http"},
					},
					Type: cty.Map(cty.Object(map[string]cty.Type{
						"source_port": cty.Number,
						"protocol":    cty.String,
					})),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 2,
							Byte:   55,
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
							Column: 16,
							Byte:   15,
						},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "https"},
					},
					Type: cty.Map(cty.Object(map[string]cty.Type{
						"source_port": cty.Number,
						"protocol":    cty.String,
					})),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   5,
							Column: 1,
							Byte:   56,
						},
						End: hcl.Pos{
							Line:   8,
							Column: 2,
							Byte:   113,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   5,
							Column: 1,
							Byte:   56,
						},
						End: hcl.Pos{
							Line:   5,
							Column: 17,
							Byte:   72,
						},
					},
				},
			},
		},
		{
			"block with inferred body as data",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"provider": {
						Labels: []*schema.LabelSchema{
							{Name: "name"},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
							},
							BodyAsData: true,
							InferBody:  true,
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Expr:       schema.LiteralTypeOnly(cty.Number),
									IsOptional: true,
								},
								"name": {
									Expr:       schema.LiteralTypeOnly(cty.String),
									IsOptional: true,
								},
								"map_attr": {
									Expr: schema.ExprConstraints{
										schema.MapExpr{Elem: schema.LiteralTypeOnly(cty.String)},
									},
									IsOptional: true,
								},
								"list_attr": {
									Expr: schema.ExprConstraints{
										schema.ListExpr{Elem: schema.LiteralTypeOnly(cty.String)},
									},
									IsOptional: true,
								},
								"obj_attr": {
									Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{
										"nestedattr": cty.String,
									})),
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`provider "aws" {
  attr = 42
  name = "hello world"
  map_attr = {
    one = "hello"
    two = "world"
  }
  list_attr = [ "one", "two" ]
  obj_attr = {
    nestedattr = "foo"
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws"},
					},
					Type: cty.Object(map[string]cty.Type{
						"attr":      cty.Number,
						"name":      cty.String,
						"map_attr":  cty.Map(cty.String),
						"list_attr": cty.List(cty.String),
						"obj_attr": cty.Object(map[string]cty.Type{
							"nestedattr": cty.String,
						}),
					}),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   12,
							Column: 2,
							Byte:   181,
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
							Column: 15,
							Byte:   14,
						},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws"},
								lang.AttrStep{Name: "attr"},
							},
							Type: cty.Number,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 3,
									Byte:   19,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 12,
									Byte:   28,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 3,
									Byte:   19,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 7,
									Byte:   23,
								},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws"},
								lang.AttrStep{Name: "list_attr"},
							},
							Type: cty.List(cty.String),
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   8,
									Column: 3,
									Byte:   109,
								},
								End: hcl.Pos{
									Line:   8,
									Column: 31,
									Byte:   137,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   8,
									Column: 3,
									Byte:   109,
								},
								End: hcl.Pos{
									Line:   8,
									Column: 12,
									Byte:   118,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws"},
										lang.AttrStep{Name: "list_attr"},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
									},
									Type: cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   8,
											Column: 17,
											Byte:   123,
										},
										End: hcl.Pos{
											Line:   8,
											Column: 22,
											Byte:   128,
										},
									},
								},
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws"},
										lang.AttrStep{Name: "list_attr"},
										lang.IndexStep{Key: cty.NumberIntVal(1)},
									},
									Type: cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   8,
											Column: 24,
											Byte:   130,
										},
										End: hcl.Pos{
											Line:   8,
											Column: 29,
											Byte:   135,
										},
									},
								},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws"},
								lang.AttrStep{Name: "map_attr"},
							},
							Type: cty.Map(cty.String),
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   4,
									Column: 3,
									Byte:   54,
								},
								End: hcl.Pos{
									Line:   7,
									Column: 4,
									Byte:   106,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   4,
									Column: 3,
									Byte:   54,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 11,
									Byte:   62,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws"},
										lang.AttrStep{Name: "map_attr"},
										lang.IndexStep{Key: cty.StringVal("one")},
									},
									Type: cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   5,
											Column: 5,
											Byte:   71,
										},
										End: hcl.Pos{
											Line:   5,
											Column: 18,
											Byte:   84,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   5,
											Column: 5,
											Byte:   71,
										},
										End: hcl.Pos{
											Line:   5,
											Column: 8,
											Byte:   74,
										},
									},
								},
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws"},
										lang.AttrStep{Name: "map_attr"},
										lang.IndexStep{Key: cty.StringVal("two")},
									},
									Type: cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   6,
											Column: 5,
											Byte:   89,
										},
										End: hcl.Pos{
											Line:   6,
											Column: 18,
											Byte:   102,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   6,
											Column: 5,
											Byte:   89,
										},
										End: hcl.Pos{
											Line:   6,
											Column: 8,
											Byte:   92,
										},
									},
								},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws"},
								lang.AttrStep{Name: "name"},
							},
							Type: cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 3,
									Byte:   31,
								},
								End: hcl.Pos{
									Line:   3,
									Column: 23,
									Byte:   51,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 3,
									Byte:   31,
								},
								End: hcl.Pos{
									Line:   3,
									Column: 7,
									Byte:   35,
								},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws"},
								lang.AttrStep{Name: "obj_attr"},
							},
							Type: cty.Object(map[string]cty.Type{
								"nestedattr": cty.String,
							}),
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   9,
									Column: 3,
									Byte:   140,
								},
								End: hcl.Pos{
									Line:   11,
									Column: 4,
									Byte:   179,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   9,
									Column: 3,
									Byte:   140,
								},
								End: hcl.Pos{
									Line:   9,
									Column: 11,
									Byte:   148,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws"},
										lang.AttrStep{Name: "obj_attr"},
										lang.AttrStep{Name: "nestedattr"},
									},
									Type: cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   10,
											Column: 5,
											Byte:   157,
										},
										End: hcl.Pos{
											Line:   10,
											Column: 23,
											Byte:   175,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   10,
											Column: 5,
											Byte:   157,
										},
										End: hcl.Pos{
											Line:   10,
											Column: 15,
											Byte:   167,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"block with dependent body as data",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"provider": {
						Labels: []*schema.LabelSchema{
							{Name: "name", IsDepKey: true},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
							},
							DependentBodyAsData: true,
						},
						Type: schema.BlockTypeObject,
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "aws"},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"attr": {
										Expr:       schema.LiteralTypeOnly(cty.Number),
										IsOptional: true,
									},
									"name": {
										Expr:       schema.LiteralTypeOnly(cty.String),
										IsOptional: true,
									},
									"attr_list": {
										Expr:       schema.LiteralTypeOnly(cty.List(cty.String)),
										IsOptional: true,
									},
									"attr_map": {
										Expr:       schema.LiteralTypeOnly(cty.Map(cty.String)),
										IsOptional: true,
									},
									"obj": {
										Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{
											"nestedattr": cty.String,
										})),
										IsOptional: true,
									},
								},
							},
						},
					},
				},
			},
			`provider "aws" {
  attr = 42
  name = "hello world"
  attr_list = ["one", "two"]
  attr_map = {
    foo = "bar"
  }
  obj = {
    nestedattr = "test"
  }
}
provider "test" {
  attr = 42
  name = "hello world"
  attr_list = ["one", "two"]
  attr_map = {
    foo = "bar"
  }
  obj = {
    nestedattr = "test"
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws"},
					},
					Type: cty.Object(map[string]cty.Type{
						"attr":      cty.Number,
						"name":      cty.String,
						"attr_list": cty.List(cty.String),
						"attr_map":  cty.Map(cty.String),
						"obj": cty.Object(map[string]cty.Type{
							"nestedattr": cty.String,
						}),
					}),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   11,
							Column: 2,
							Byte:   155,
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
							Column: 15,
							Byte:   14,
						},
					},
				},
			},
		},
		{
			"block with inferred body data",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"provider": {
						Labels: []*schema.LabelSchema{
							{Name: "name", IsDepKey: true},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
							},
							DependentBodyAsData: true,
							InferDependentBody:  true,
						},
						Type: schema.BlockTypeObject,
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "aws"},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"attr": {
										Expr:       schema.LiteralTypeOnly(cty.Number),
										IsOptional: true,
									},
									"name": {
										Expr:       schema.LiteralTypeOnly(cty.String),
										IsOptional: true,
									},
									"attr_list": {
										Expr:       schema.LiteralTypeOnly(cty.List(cty.String)),
										IsOptional: true,
									},
									"attr_map": {
										Expr:       schema.LiteralTypeOnly(cty.Map(cty.String)),
										IsOptional: true,
									},
									"obj": {
										Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{
											"nestedattr": cty.String,
										})),
										IsOptional: true,
									},
								},
							},
						},
					},
				},
			},
			`provider "aws" {
  attr = 42
  name = "hello world"
  attr_list = ["one", "two"]
  attr_map = {
    foo = "bar"
  }
  obj = {
    nestedattr = "test"
  }
}
provider "test" {
  attr = 42
  name = "hello world"
  attr_list = ["one", "two"]
  attr_map = {
    foo = "bar"
  }
  obj = {
    nestedattr = "test"
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws"},
					},
					Type: cty.Object(map[string]cty.Type{
						"attr":      cty.Number,
						"name":      cty.String,
						"attr_list": cty.List(cty.String),
						"attr_map":  cty.Map(cty.String),
						"obj": cty.Object(map[string]cty.Type{
							"nestedattr": cty.String,
						}),
					}),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   11,
							Column: 2,
							Byte:   155,
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
							Column: 15,
							Byte:   14,
						},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws"},
								lang.AttrStep{Name: "attr"},
							},
							Type: cty.Number,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 3,
									Byte:   19,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 12,
									Byte:   28,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 3,
									Byte:   19,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 7,
									Byte:   23,
								},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws"},
								lang.AttrStep{Name: "attr_list"},
							},
							Type: cty.List(cty.String),
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   4,
									Column: 3,
									Byte:   54,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 29,
									Byte:   80,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   4,
									Column: 3,
									Byte:   54,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 12,
									Byte:   63,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws"},
										lang.AttrStep{Name: "attr_list"},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
									},
									Type: cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   4,
											Column: 16,
											Byte:   67,
										},
										End: hcl.Pos{
											Line:   4,
											Column: 21,
											Byte:   72,
										},
									},
								},
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws"},
										lang.AttrStep{Name: "attr_list"},
										lang.IndexStep{Key: cty.NumberIntVal(1)},
									},
									Type: cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   4,
											Column: 23,
											Byte:   74,
										},
										End: hcl.Pos{
											Line:   4,
											Column: 28,
											Byte:   79,
										},
									},
								},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws"},
								lang.AttrStep{Name: "attr_map"},
							},
							Type: cty.Map(cty.String),
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   5,
									Column: 3,
									Byte:   83,
								},
								End: hcl.Pos{
									Line:   7,
									Column: 4,
									Byte:   115,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   5,
									Column: 3,
									Byte:   83,
								},
								End: hcl.Pos{
									Line:   5,
									Column: 11,
									Byte:   91,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws"},
										lang.AttrStep{Name: "attr_map"},
										lang.IndexStep{Key: cty.StringVal("foo")},
									},
									Type: cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   6,
											Column: 5,
											Byte:   100,
										},
										End: hcl.Pos{
											Line:   6,
											Column: 16,
											Byte:   111,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   6,
											Column: 5,
											Byte:   100,
										},
										End: hcl.Pos{
											Line:   6,
											Column: 8,
											Byte:   103,
										},
									},
								},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws"},
								lang.AttrStep{Name: "name"},
							},
							Type: cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 3,
									Byte:   31,
								},
								End: hcl.Pos{
									Line:   3,
									Column: 23,
									Byte:   51,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 3,
									Byte:   31,
								},
								End: hcl.Pos{
									Line:   3,
									Column: 7,
									Byte:   35,
								},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws"},
								lang.AttrStep{Name: "obj"},
							},
							Type: cty.Object(map[string]cty.Type{
								"nestedattr": cty.String,
							}),
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   8,
									Column: 3,
									Byte:   118,
								},
								End: hcl.Pos{
									Line:   10,
									Column: 4,
									Byte:   153,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   8,
									Column: 3,
									Byte:   118,
								},
								End: hcl.Pos{
									Line:   8,
									Column: 6,
									Byte:   121,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws"},
										lang.AttrStep{Name: "obj"},
										lang.AttrStep{Name: "nestedattr"},
									},
									Type: cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   9,
											Column: 5,
											Byte:   130,
										},
										End: hcl.Pos{
											Line:   9,
											Column: 24,
											Byte:   149,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   9,
											Column: 5,
											Byte:   130,
										},
										End: hcl.Pos{
											Line:   9,
											Column: 15,
											Byte:   140,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"nested single object block with inferred body data",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"rootblock": {
						Labels: []*schema.LabelSchema{
							{Name: "name", IsDepKey: true},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "root"},
								schema.LabelStep{Index: 0},
							},
							BodyAsData: true,
							InferBody:  true,
						},
						Type: schema.BlockTypeObject,
						Body: &schema.BodySchema{
							Blocks: map[string]*schema.BlockSchema{
								"objblock": {
									Type: schema.BlockTypeObject,
									Body: &schema.BodySchema{
										Attributes: map[string]*schema.AttributeSchema{
											"protocol": {
												Expr:       schema.LiteralTypeOnly(cty.String),
												IsOptional: true,
											},
											"port": {
												Expr:       schema.LiteralTypeOnly(cty.Number),
												IsOptional: true,
											},
										},
									},
								},
							},
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Expr:       schema.LiteralTypeOnly(cty.Number),
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`rootblock "aws" {
  attr = 42
  objblock {
    port = 80
    protocol = "tcp"
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "root"},
						lang.AttrStep{Name: "aws"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   7,
							Column: 2,
							Byte:   83,
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
							Column: 16,
							Byte:   15,
						},
					},
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.Number,
						"objblock": cty.Object(map[string]cty.Type{
							"port":     cty.Number,
							"protocol": cty.String,
						}),
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "root"},
								lang.AttrStep{Name: "aws"},
								lang.AttrStep{Name: "attr"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 3,
									Byte:   20,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 12,
									Byte:   29,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 3,
									Byte:   20,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 7,
									Byte:   24,
								},
							},
							Type: cty.Number,
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "root"},
								lang.AttrStep{Name: "aws"},
								lang.AttrStep{Name: "objblock"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 3,
									Byte:   32,
								},
								End: hcl.Pos{
									Line:   6,
									Column: 4,
									Byte:   81,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 3,
									Byte:   32,
								},
								End: hcl.Pos{
									Line:   3,
									Column: 11,
									Byte:   40,
								},
							},
							Type: cty.Object(map[string]cty.Type{
								"port":     cty.Number,
								"protocol": cty.String,
							}),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "root"},
										lang.AttrStep{Name: "aws"},
										lang.AttrStep{Name: "objblock"},
										lang.AttrStep{Name: "port"},
									},
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   4,
											Column: 5,
											Byte:   47,
										},
										End: hcl.Pos{
											Line:   4,
											Column: 14,
											Byte:   56,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   4,
											Column: 5,
											Byte:   47,
										},
										End: hcl.Pos{
											Line:   4,
											Column: 9,
											Byte:   51,
										},
									},
									Type: cty.Number,
								},
								{
									Addr: lang.Address{
										lang.RootStep{Name: "root"},
										lang.AttrStep{Name: "aws"},
										lang.AttrStep{Name: "objblock"},
										lang.AttrStep{Name: "protocol"},
									},
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   5,
											Column: 5,
											Byte:   61,
										},
										End: hcl.Pos{
											Line:   5,
											Column: 21,
											Byte:   77,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   5,
											Column: 5,
											Byte:   61,
										},
										End: hcl.Pos{
											Line:   5,
											Column: 13,
											Byte:   69,
										},
									},
									Type: cty.String,
								},
							},
						},
					},
				},
			},
		},
		{
			"nested list block with inferred body data",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"rootblock": {
						Labels: []*schema.LabelSchema{
							{Name: "name", IsDepKey: true},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "root"},
								schema.LabelStep{Index: 0},
							},
							BodyAsData: true,
							InferBody:  true,
						},
						Type: schema.BlockTypeObject,
						Body: &schema.BodySchema{
							Blocks: map[string]*schema.BlockSchema{
								"listblock": {
									Type: schema.BlockTypeList,
									Body: &schema.BodySchema{
										Attributes: map[string]*schema.AttributeSchema{
											"protocol": {
												Expr:       schema.LiteralTypeOnly(cty.String),
												IsOptional: true,
											},
											"port": {
												Expr:       schema.LiteralTypeOnly(cty.Number),
												IsOptional: true,
											},
										},
									},
								},
							},
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Expr:       schema.LiteralTypeOnly(cty.Number),
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`rootblock "aws" {
  attr = 42
  listblock {
    port = 80
    protocol = "tcp"
  }
  listblock {
    port = 443
    protocol = "tcp"
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "root"},
						lang.AttrStep{Name: "aws"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   11,
							Column: 2,
							Byte:   138,
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
							Column: 16,
							Byte:   15,
						},
					},
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.Number,
						"listblock": cty.List(cty.Object(map[string]cty.Type{
							"port":     cty.Number,
							"protocol": cty.String,
						})),
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "root"},
								lang.AttrStep{Name: "aws"},
								lang.AttrStep{Name: "attr"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 3,
									Byte:   20,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 12,
									Byte:   29,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 3,
									Byte:   20,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 7,
									Byte:   24,
								},
							},
							Type: cty.Number,
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "root"},
								lang.AttrStep{Name: "aws"},
								lang.AttrStep{Name: "listblock"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 3,
									Byte:   32,
								},
								End: hcl.Pos{
									Line:   10,
									Column: 4,
									Byte:   136,
								},
							},
							DefRangePtr: nil,
							Type: cty.List(cty.Object(map[string]cty.Type{
								"port":     cty.Number,
								"protocol": cty.String,
							})),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "root"},
										lang.AttrStep{Name: "aws"},
										lang.AttrStep{Name: "listblock"},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
									},
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   3,
											Column: 3,
											Byte:   32,
										},
										End: hcl.Pos{
											Line:   10,
											Column: 4,
											Byte:   136,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   3,
											Column: 3,
											Byte:   32,
										},
										End: hcl.Pos{
											Line:   3,
											Column: 12,
											Byte:   41,
										},
									},
									Type: cty.Object(map[string]cty.Type{
										"port":     cty.Number,
										"protocol": cty.String,
									}),
									NestedTargets: reference.Targets{
										{
											Addr: lang.Address{
												lang.RootStep{Name: "root"},
												lang.AttrStep{Name: "aws"},
												lang.AttrStep{Name: "listblock"},
												lang.IndexStep{Key: cty.NumberIntVal(0)},
												lang.AttrStep{Name: "port"},
											},
											RangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   4,
													Column: 5,
													Byte:   48,
												},
												End: hcl.Pos{
													Line:   4,
													Column: 14,
													Byte:   57,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   4,
													Column: 5,
													Byte:   48,
												},
												End: hcl.Pos{
													Line:   4,
													Column: 9,
													Byte:   52,
												},
											},
											Type: cty.Number,
										},
										{
											Addr: lang.Address{
												lang.RootStep{Name: "root"},
												lang.AttrStep{Name: "aws"},
												lang.AttrStep{Name: "listblock"},
												lang.IndexStep{Key: cty.NumberIntVal(0)},
												lang.AttrStep{Name: "protocol"},
											},
											RangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   5,
													Column: 5,
													Byte:   62,
												},
												End: hcl.Pos{
													Line:   5,
													Column: 21,
													Byte:   78,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   5,
													Column: 5,
													Byte:   62,
												},
												End: hcl.Pos{
													Line:   5,
													Column: 13,
													Byte:   70,
												},
											},
											Type: cty.String,
										},
									},
								},
								{
									Addr: lang.Address{
										lang.RootStep{Name: "root"},
										lang.AttrStep{Name: "aws"},
										lang.AttrStep{Name: "listblock"},
										lang.IndexStep{Key: cty.NumberIntVal(1)},
									},
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   7,
											Column: 3,
											Byte:   85,
										},
										End: hcl.Pos{
											Line:   10,
											Column: 4,
											Byte:   136,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   7,
											Column: 3,
											Byte:   85,
										},
										End: hcl.Pos{
											Line:   7,
											Column: 12,
											Byte:   94,
										},
									},
									Type: cty.Object(map[string]cty.Type{
										"port":     cty.Number,
										"protocol": cty.String,
									}),
									NestedTargets: reference.Targets{
										{
											Addr: lang.Address{
												lang.RootStep{Name: "root"},
												lang.AttrStep{Name: "aws"},
												lang.AttrStep{Name: "listblock"},
												lang.IndexStep{Key: cty.NumberIntVal(1)},
												lang.AttrStep{Name: "port"},
											},
											RangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   8,
													Column: 5,
													Byte:   101,
												},
												End: hcl.Pos{
													Line:   8,
													Column: 15,
													Byte:   111,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   8,
													Column: 5,
													Byte:   101,
												},
												End: hcl.Pos{
													Line:   8,
													Column: 9,
													Byte:   105,
												},
											},
											Type: cty.Number,
										},
										{
											Addr: lang.Address{
												lang.RootStep{Name: "root"},
												lang.AttrStep{Name: "aws"},
												lang.AttrStep{Name: "listblock"},
												lang.IndexStep{Key: cty.NumberIntVal(1)},
												lang.AttrStep{Name: "protocol"},
											},
											RangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   9,
													Column: 5,
													Byte:   116,
												},
												End: hcl.Pos{
													Line:   9,
													Column: 21,
													Byte:   132,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   9,
													Column: 5,
													Byte:   116,
												},
												End: hcl.Pos{
													Line:   9,
													Column: 13,
													Byte:   124,
												},
											},
											Type: cty.String,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"nested set block with inferred body data",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"rootblock": {
						Labels: []*schema.LabelSchema{
							{Name: "name", IsDepKey: true},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "root"},
								schema.LabelStep{Index: 0},
							},
							BodyAsData: true,
							InferBody:  true,
						},
						Type: schema.BlockTypeObject,
						Body: &schema.BodySchema{
							Blocks: map[string]*schema.BlockSchema{
								"setblock": {
									Type: schema.BlockTypeSet,
									Body: &schema.BodySchema{
										Attributes: map[string]*schema.AttributeSchema{
											"protocol": {
												Expr:       schema.LiteralTypeOnly(cty.String),
												IsOptional: true,
											},
											"port": {
												Expr:       schema.LiteralTypeOnly(cty.Number),
												IsOptional: true,
											},
										},
									},
								},
							},
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Expr:       schema.LiteralTypeOnly(cty.Number),
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`rootblock "aws" {
  attr = 42
  setblock {
    port = 80
    protocol = "tcp"
  }
  setblock {
    port = 443
    protocol = "tcp"
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "root"},
						lang.AttrStep{Name: "aws"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   11,
							Column: 2,
							Byte:   136,
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
							Column: 16,
							Byte:   15,
						},
					},
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.Number,
						"setblock": cty.Set(cty.Object(map[string]cty.Type{
							"port":     cty.Number,
							"protocol": cty.String,
						})),
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "root"},
								lang.AttrStep{Name: "aws"},
								lang.AttrStep{Name: "attr"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 3,
									Byte:   20,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 12,
									Byte:   29,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 3,
									Byte:   20,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 7,
									Byte:   24,
								},
							},
							Type: cty.Number,
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "root"},
								lang.AttrStep{Name: "aws"},
								lang.AttrStep{Name: "setblock"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 3,
									Byte:   32,
								},
								End: hcl.Pos{
									Line:   10,
									Column: 4,
									Byte:   134,
								},
							},
							Type: cty.Set(cty.Object(map[string]cty.Type{
								"port":     cty.Number,
								"protocol": cty.String,
							})),
						},
					},
				},
			},
		},
		{
			"separated nested list blocks with inferred body data",
			// This is to verify that Range (correctly) points to the 1st block
			// when block instances are non-consecutive
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"rootblock": {
						Labels: []*schema.LabelSchema{
							{Name: "name", IsDepKey: true},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "root"},
								schema.LabelStep{Index: 0},
							},
							BodyAsData: true,
							InferBody:  true,
						},
						Type: schema.BlockTypeObject,
						Body: &schema.BodySchema{
							Blocks: map[string]*schema.BlockSchema{
								"listblock": {
									Type: schema.BlockTypeList,
									Body: &schema.BodySchema{
										Attributes: map[string]*schema.AttributeSchema{
											"protocol": {
												Expr:       schema.LiteralTypeOnly(cty.String),
												IsOptional: true,
											},
											"port": {
												Expr:       schema.LiteralTypeOnly(cty.Number),
												IsOptional: true,
											},
										},
									},
								},
							},
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Expr:       schema.LiteralTypeOnly(cty.Number),
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`rootblock "aws" {
  listblock {
    port = 80
    protocol = "tcp"
  }
  attr = 42
  listblock {
    port = 443
    protocol = "tcp"
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "root"},
						lang.AttrStep{Name: "aws"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   11,
							Column: 2,
							Byte:   138,
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
							Column: 16,
							Byte:   15,
						},
					},
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.Number,
						"listblock": cty.List(cty.Object(map[string]cty.Type{
							"port":     cty.Number,
							"protocol": cty.String,
						})),
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "root"},
								lang.AttrStep{Name: "aws"},
								lang.AttrStep{Name: "attr"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   6,
									Column: 3,
									Byte:   73,
								},
								End: hcl.Pos{
									Line:   6,
									Column: 12,
									Byte:   82,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   6,
									Column: 3,
									Byte:   73,
								},
								End: hcl.Pos{
									Line:   6,
									Column: 7,
									Byte:   77,
								},
							},
							Type: cty.Number,
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "root"},
								lang.AttrStep{Name: "aws"},
								lang.AttrStep{Name: "listblock"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 3,
									Byte:   20,
								},
								End: hcl.Pos{
									Line:   5,
									Column: 4,
									Byte:   70,
								},
							},
							DefRangePtr: nil,
							Type: cty.List(cty.Object(map[string]cty.Type{
								"port":     cty.Number,
								"protocol": cty.String,
							})),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "root"},
										lang.AttrStep{Name: "aws"},
										lang.AttrStep{Name: "listblock"},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
									},
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   2,
											Column: 3,
											Byte:   20,
										},
										End: hcl.Pos{
											Line:   5,
											Column: 4,
											Byte:   70,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   2,
											Column: 3,
											Byte:   20,
										},
										End: hcl.Pos{
											Line:   2,
											Column: 12,
											Byte:   29,
										},
									},
									Type: cty.Object(map[string]cty.Type{
										"port":     cty.Number,
										"protocol": cty.String,
									}),
									NestedTargets: reference.Targets{
										{
											Addr: lang.Address{
												lang.RootStep{Name: "root"},
												lang.AttrStep{Name: "aws"},
												lang.AttrStep{Name: "listblock"},
												lang.IndexStep{Key: cty.NumberIntVal(0)},
												lang.AttrStep{Name: "port"},
											},
											RangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   3,
													Column: 5,
													Byte:   36,
												},
												End: hcl.Pos{
													Line:   3,
													Column: 14,
													Byte:   45,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   3,
													Column: 5,
													Byte:   36,
												},
												End: hcl.Pos{
													Line:   3,
													Column: 9,
													Byte:   40,
												},
											},
											Type: cty.Number,
										},
										{
											Addr: lang.Address{
												lang.RootStep{Name: "root"},
												lang.AttrStep{Name: "aws"},
												lang.AttrStep{Name: "listblock"},
												lang.IndexStep{Key: cty.NumberIntVal(0)},
												lang.AttrStep{Name: "protocol"},
											},
											RangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   4,
													Column: 5,
													Byte:   50,
												},
												End: hcl.Pos{
													Line:   4,
													Column: 21,
													Byte:   66,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   4,
													Column: 5,
													Byte:   50,
												},
												End: hcl.Pos{
													Line:   4,
													Column: 13,
													Byte:   58,
												},
											},
											Type: cty.String,
										},
									},
								},
								{
									Addr: lang.Address{
										lang.RootStep{Name: "root"},
										lang.AttrStep{Name: "aws"},
										lang.AttrStep{Name: "listblock"},
										lang.IndexStep{Key: cty.NumberIntVal(1)},
									},
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   7,
											Column: 3,
											Byte:   85,
										},
										End: hcl.Pos{
											Line:   10,
											Column: 4,
											Byte:   136,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   7,
											Column: 3,
											Byte:   85,
										},
										End: hcl.Pos{
											Line:   7,
											Column: 12,
											Byte:   94,
										},
									},
									Type: cty.Object(map[string]cty.Type{
										"port":     cty.Number,
										"protocol": cty.String,
									}),
									NestedTargets: reference.Targets{
										{
											Addr: lang.Address{
												lang.RootStep{Name: "root"},
												lang.AttrStep{Name: "aws"},
												lang.AttrStep{Name: "listblock"},
												lang.IndexStep{Key: cty.NumberIntVal(1)},
												lang.AttrStep{Name: "port"},
											},
											RangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   8,
													Column: 5,
													Byte:   101,
												},
												End: hcl.Pos{
													Line:   8,
													Column: 15,
													Byte:   111,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   8,
													Column: 5,
													Byte:   101,
												},
												End: hcl.Pos{
													Line:   8,
													Column: 9,
													Byte:   105,
												},
											},
											Type: cty.Number,
										},
										{
											Addr: lang.Address{
												lang.RootStep{Name: "root"},
												lang.AttrStep{Name: "aws"},
												lang.AttrStep{Name: "listblock"},
												lang.IndexStep{Key: cty.NumberIntVal(1)},
												lang.AttrStep{Name: "protocol"},
											},
											RangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   9,
													Column: 5,
													Byte:   116,
												},
												End: hcl.Pos{
													Line:   9,
													Column: 21,
													Byte:   132,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   9,
													Column: 5,
													Byte:   116,
												},
												End: hcl.Pos{
													Line:   9,
													Column: 13,
													Byte:   124,
												},
											},
											Type: cty.String,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"nested map blocks with inferred body data",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"load_balancer": {
						Labels: []*schema.LabelSchema{
							{Name: "name", IsDepKey: true},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "lb"},
								schema.LabelStep{Index: 0},
							},
							BodyAsData: true,
							InferBody:  true,
						},
						Type: schema.BlockTypeObject,
						Body: &schema.BodySchema{
							Blocks: map[string]*schema.BlockSchema{
								"listener": {
									Labels: []*schema.LabelSchema{
										{Name: "name"},
									},
									Type: schema.BlockTypeMap,
									Body: &schema.BodySchema{
										Attributes: map[string]*schema.AttributeSchema{
											"protocol": {
												Expr:       schema.LiteralTypeOnly(cty.String),
												IsOptional: true,
											},
											"port": {
												Expr:       schema.LiteralTypeOnly(cty.Number),
												IsOptional: true,
											},
										},
									},
								},
							},
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Expr:       schema.LiteralTypeOnly(cty.Number),
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`load_balancer "aws" {
  attr = 42
  listener "http" {
    port = 80
    protocol = "tcp"
  }
  listener "https" {
    port = 443
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "lb"},
						lang.AttrStep{Name: "aws"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   10,
							Column: 2,
							Byte:   134,
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
							Column: 20,
							Byte:   19,
						},
					},
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.Number,
						"listener": cty.Map(cty.Object(map[string]cty.Type{
							"port":     cty.Number,
							"protocol": cty.String,
						})),
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "lb"},
								lang.AttrStep{Name: "aws"},
								lang.AttrStep{Name: "attr"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 3,
									Byte:   24,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 12,
									Byte:   33,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 3,
									Byte:   24,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 7,
									Byte:   28,
								},
							},
							Type: cty.Number,
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "lb"},
								lang.AttrStep{Name: "aws"},
								lang.AttrStep{Name: "listener"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 3,
									Byte:   36,
								},
								End: hcl.Pos{
									Line:   9,
									Column: 4,
									Byte:   132,
								},
							},
							DefRangePtr: nil,
							Type: cty.Map(cty.Object(map[string]cty.Type{
								"port":     cty.Number,
								"protocol": cty.String,
							})),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "lb"},
										lang.AttrStep{Name: "aws"},
										lang.AttrStep{Name: "listener"},
										lang.IndexStep{Key: cty.StringVal("http")},
									},
									Type: cty.Object(map[string]cty.Type{
										"port":     cty.Number,
										"protocol": cty.String,
									}),
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   3,
											Column: 3,
											Byte:   36,
										},
										End: hcl.Pos{
											Line:   9,
											Column: 4,
											Byte:   132,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   3,
											Column: 3,
											Byte:   36,
										},
										End: hcl.Pos{
											Line:   3,
											Column: 18,
											Byte:   51,
										},
									},
									NestedTargets: reference.Targets{
										{
											Addr: lang.Address{
												lang.RootStep{Name: "lb"},
												lang.AttrStep{Name: "aws"},
												lang.AttrStep{Name: "listener"},
												lang.IndexStep{Key: cty.StringVal("http")},
												lang.AttrStep{Name: "port"},
											},
											RangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   4,
													Column: 5,
													Byte:   58,
												},
												End: hcl.Pos{
													Line:   4,
													Column: 14,
													Byte:   67,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   4,
													Column: 5,
													Byte:   58,
												},
												End: hcl.Pos{
													Line:   4,
													Column: 9,
													Byte:   62,
												},
											},
											Type: cty.Number,
										},
										{
											Addr: lang.Address{
												lang.RootStep{Name: "lb"},
												lang.AttrStep{Name: "aws"},
												lang.AttrStep{Name: "listener"},
												lang.IndexStep{Key: cty.StringVal("http")},
												lang.AttrStep{Name: "protocol"},
											},
											RangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   5,
													Column: 5,
													Byte:   72,
												},
												End: hcl.Pos{
													Line:   5,
													Column: 21,
													Byte:   88,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   5,
													Column: 5,
													Byte:   72,
												},
												End: hcl.Pos{
													Line:   5,
													Column: 13,
													Byte:   80,
												},
											},
											Type: cty.String,
										},
									},
								},
								{
									Addr: lang.Address{
										lang.RootStep{Name: "lb"},
										lang.AttrStep{Name: "aws"},
										lang.AttrStep{Name: "listener"},
										lang.IndexStep{Key: cty.StringVal("https")},
									},
									Type: cty.Object(map[string]cty.Type{
										"port":     cty.Number,
										"protocol": cty.String,
									}),
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   7,
											Column: 3,
											Byte:   95,
										},
										End: hcl.Pos{
											Line:   9,
											Column: 4,
											Byte:   132,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   7,
											Column: 3,
											Byte:   95,
										},
										End: hcl.Pos{
											Line:   7,
											Column: 19,
											Byte:   111,
										},
									},
									NestedTargets: reference.Targets{
										{
											Addr: lang.Address{
												lang.RootStep{Name: "lb"},
												lang.AttrStep{Name: "aws"},
												lang.AttrStep{Name: "listener"},
												lang.IndexStep{Key: cty.StringVal("https")},
												lang.AttrStep{Name: "port"},
											},
											RangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   8,
													Column: 5,
													Byte:   118,
												},
												End: hcl.Pos{
													Line:   8,
													Column: 15,
													Byte:   128,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   8,
													Column: 5,
													Byte:   118,
												},
												End: hcl.Pos{
													Line:   8,
													Column: 9,
													Byte:   122,
												},
											},
											Type: cty.Number,
										},
										{
											Addr: lang.Address{
												lang.RootStep{Name: "lb"},
												lang.AttrStep{Name: "aws"},
												lang.AttrStep{Name: "listener"},
												lang.IndexStep{Key: cty.StringVal("https")},
												lang.AttrStep{Name: "protocol"},
											},
											RangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   7,
													Column: 20,
													Byte:   112,
												},
												End: hcl.Pos{
													Line:   7,
													Column: 20,
													Byte:   112,
												},
											},
											Type: cty.String,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"traversal reference",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"testattr": {
						IsOptional: true,
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{
								Address: &schema.TraversalAddrSchema{
									ScopeId: lang.ScopeId("specialthing"),
								},
							},
						},
					},
				},
			},
			`testattr = special.test
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "special"},
						lang.AttrStep{Name: "test"},
					},
					ScopeId: lang.ScopeId("specialthing"),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 12,
							Byte:   11,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 24,
							Byte:   23,
						},
					},
					DefRangePtr: nil,
				},
			},
		},
		{
			"block with attribute value in address",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"provider": {
						Labels: []*schema.LabelSchema{
							{Name: "name"},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
								schema.AttrValueStep{Name: "alias"},
							},
							ScopeId:     lang.ScopeId("provider"),
							AsReference: true,
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"alias": {
									IsOptional: true,
									Expr:       schema.LiteralTypeOnly(cty.String),
								},
							},
						},
					},
				},
			},
			`provider "aws" {
  alias = "euwest"
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws"},
						lang.AttrStep{Name: "euwest"},
					},
					ScopeId: lang.ScopeId("provider"),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   37,
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
							Column: 15,
							Byte:   14,
						},
					},
				},
			},
		},
		{
			"block as data type per attribute - undeclared",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"variable": {
						Labels: []*schema.LabelSchema{
							{Name: "name"},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
							},
							AsTypeOf: &schema.BlockAsTypeOf{},
						},
						Type: schema.BlockTypeObject,
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{},
						},
					},
				},
			},
			`variable "test" {
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Type: cty.DynamicPseudoType,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 2,
							Byte:   19,
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
							Column: 16,
							Byte:   15,
						},
					},
				},
			},
		},
		{
			"block as data type per attribute - type only",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"variable": {
						Labels: []*schema.LabelSchema{
							{Name: "name"},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
							},
							AsTypeOf: &schema.BlockAsTypeOf{
								AttributeExpr: "type",
							},
						},
						Type: schema.BlockTypeObject,
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"type": {
									IsOptional: true,
									Expr: schema.ExprConstraints{
										schema.TypeDeclarationExpr{},
									},
								},
							},
						},
					},
				},
			},
			`variable "test" {
  type = map(string)
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Type: cty.Map(cty.String),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   40,
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
							Column: 16,
							Byte:   15,
						},
					},
				},
			},
		},
		{
			"block as data type per attribute - default string",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"variable": {
						Labels: []*schema.LabelSchema{
							{Name: "name"},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
							},
							AsTypeOf: &schema.BlockAsTypeOf{
								AttributeValue: "default",
							},
						},
						Type: schema.BlockTypeObject,
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"type": {
									IsOptional: true,
									Expr: schema.ExprConstraints{
										schema.TypeDeclarationExpr{},
									},
								},
								"default": {
									IsOptional: true,
									Expr:       schema.LiteralTypeOnly(cty.DynamicPseudoType),
								},
							},
						},
					},
				},
			},
			`variable "test" {
  default = "something"
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   43,
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
							Column: 16,
							Byte:   15,
						},
					},
				},
			},
		},
		{
			"block as data type per attribute - default tuple constant",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"variable": {
						Labels: []*schema.LabelSchema{
							{Name: "name"},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
							},
							AsTypeOf: &schema.BlockAsTypeOf{
								AttributeExpr:  "type",
								AttributeValue: "default",
							},
						},
						Type: schema.BlockTypeObject,
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"type": {
									IsOptional: true,
									Expr: schema.ExprConstraints{
										schema.TypeDeclarationExpr{},
									},
								},
								"default": {
									IsOptional: true,
									Expr:       schema.LiteralTypeOnly(cty.DynamicPseudoType),
								},
							},
						},
					},
				},
			},
			`variable "test" {
  default = ["something"]
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Type: cty.Tuple([]cty.Type{cty.String}),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   45,
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
							Column: 16,
							Byte:   15,
						},
					},
				},
			},
		},
		{
			"block as data type per attribute - default list of any",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"variable": {
						Labels: []*schema.LabelSchema{
							{Name: "name"},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
							},
							AsTypeOf: &schema.BlockAsTypeOf{
								AttributeExpr:  "type",
								AttributeValue: "default",
							},
						},
						Type: schema.BlockTypeObject,
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"type": {
									IsOptional: true,
									Expr: schema.ExprConstraints{
										schema.TypeDeclarationExpr{},
									},
								},
								"default": {
									IsOptional: true,
									Expr:       schema.LiteralTypeOnly(cty.DynamicPseudoType),
								},
							},
						},
					},
				},
			},
			`variable "test" {
  type = list(any)
  default = [
    "one"
  ]
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Type: cty.List(cty.String),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   6,
							Column: 2,
							Byte:   66,
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
							Column: 16,
							Byte:   15,
						},
					},
				},
			},
		},
		{
			"block as data type per attribute - both type and default",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"variable": {
						Labels: []*schema.LabelSchema{
							{Name: "name"},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.LabelStep{Index: 0},
							},
							AsTypeOf: &schema.BlockAsTypeOf{
								AttributeValue: "default",
							},
						},
						Type: schema.BlockTypeObject,
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"type": {
									IsOptional: true,
									Expr: schema.ExprConstraints{
										schema.TypeDeclarationExpr{},
									},
								},
								"default": {
									IsOptional: true,
									Expr:       schema.LiteralTypeOnly(cty.DynamicPseudoType),
								},
							},
						},
					},
				},
			},
			`variable "test" {
  type = any
  default = "something"
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 2,
							Byte:   56,
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
							Column: 16,
							Byte:   15,
						},
					},
				},
			},
		},
		{
			"additional targetables",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"module": {
						Labels: []*schema.LabelSchema{
							{Name: "name"},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "module"},
								schema.LabelStep{Index: 0},
							},
							AsReference: true,
						},
						Type: schema.BlockTypeObject,
						Body: &schema.BodySchema{
							TargetableAs: []*schema.Targetable{
								{
									Address: lang.Address{
										lang.RootStep{Name: "module"},
										lang.AttrStep{Name: "xyz"},
										lang.AttrStep{Name: "test"},
									},
									AsType: cty.String,
								},
							},
						},
					},
				},
			},
			`module "test" {
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "module"},
						lang.AttrStep{Name: "test"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 2,
							Byte:   17,
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
							Column: 14,
							Byte:   13,
						},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "module"},
						lang.AttrStep{Name: "xyz"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 2,
							Byte:   17,
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
							Column: 14,
							Byte:   13,
						},
					},
				},
			},
		},
		{
			"block with dependent body",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"module": {
						Labels: []*schema.LabelSchema{
							{Name: "name", IsDepKey: true},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "module"},
								schema.LabelStep{Index: 0},
							},
							DependentBodyAsData: true,
							InferDependentBody:  true,
						},
						Type: schema.BlockTypeObject,
						Body: &schema.BodySchema{},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{Index: 0, Value: "test"},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"attr": {
										Expr:       schema.LiteralTypeOnly(cty.String),
										IsOptional: true,
									},
								},
							},
						},
					},
				},
			},
			`module "test" {
  attr = "foo"
}
module "different" {
  attr = "foo"
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "module"},
						lang.AttrStep{Name: "test"},
					},
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.String,
					}),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 2,
							Byte:   32,
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
							Column: 14,
							Byte:   13,
						},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "module"},
								lang.AttrStep{Name: "test"},
								lang.AttrStep{Name: "attr"},
							},
							Type: cty.String,
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 3,
									Byte:   18,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 15,
									Byte:   30,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   2,
									Column: 3,
									Byte:   18,
								},
								End: hcl.Pos{
									Line:   2,
									Column: 7,
									Byte:   22,
								},
							},
						},
					},
				},
			},
		},
		{
			// repro case for https://github.com/hashicorp/terraform-ls/issues/573
			"nested complex objects",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"locals": {
						Body: &schema.BodySchema{
							AnyAttribute: &schema.AttributeSchema{
								Address: &schema.AttributeAddrSchema{
									Steps: []schema.AddrStep{
										schema.StaticStep{Name: "local"},
										schema.AttrNameStep{},
									},
									ScopeId:    lang.ScopeId("local"),
									AsExprType: true,
								},
								Expr: schema.ExprConstraints{
									schema.TraversalExpr{OfType: cty.DynamicPseudoType},
									schema.LiteralTypeExpr{Type: cty.DynamicPseudoType},
								},
							},
						},
					},
				},
			},
			`locals {
  top_obj = {
    first = {
      attr = "val"
    }
    second = {
      attr = "val"
    }
    third = {
      attr = "val"
    }
    fourth = {
      attr = "val"
    }
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "local"},
						lang.AttrStep{Name: "top_obj"},
					},
					Type: cty.Object(map[string]cty.Type{
						"first": cty.Object(map[string]cty.Type{
							"attr": cty.String,
						}),
						"second": cty.Object(map[string]cty.Type{
							"attr": cty.String,
						}),
						"third": cty.Object(map[string]cty.Type{
							"attr": cty.String,
						}),
						"fourth": cty.Object(map[string]cty.Type{
							"attr": cty.String,
						}),
					}),
					ScopeId: lang.ScopeId("local"),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   11,
						},
						End: hcl.Pos{
							Line:   15,
							Column: 4,
							Byte:   184,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   11,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 10,
							Byte:   18,
						},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "local"},
								lang.AttrStep{Name: "top_obj"},
								lang.AttrStep{Name: "first"},
							},
							Type: cty.Object(map[string]cty.Type{
								"attr": cty.String,
							}),
							ScopeId: lang.ScopeId("local"),
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 5,
									Byte:   27,
								},
								End: hcl.Pos{
									Line:   5,
									Column: 6,
									Byte:   61,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 5,
									Byte:   27,
								},
								End: hcl.Pos{
									Line:   3,
									Column: 10,
									Byte:   32,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "local"},
										lang.AttrStep{Name: "top_obj"},
										lang.AttrStep{Name: "first"},
										lang.AttrStep{Name: "attr"},
									},
									Type:    cty.String,
									ScopeId: lang.ScopeId("local"),
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   4,
											Column: 7,
											Byte:   43,
										},
										End: hcl.Pos{
											Line:   4,
											Column: 19,
											Byte:   55,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   4,
											Column: 7,
											Byte:   43,
										},
										End: hcl.Pos{
											Line:   4,
											Column: 11,
											Byte:   47,
										},
									},
								},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "local"},
								lang.AttrStep{Name: "top_obj"},
								lang.AttrStep{Name: "second"},
							},
							Type: cty.Object(map[string]cty.Type{
								"attr": cty.String,
							}),
							ScopeId: lang.ScopeId("local"),
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   6,
									Column: 5,
									Byte:   66,
								},
								End: hcl.Pos{
									Line:   8,
									Column: 6,
									Byte:   101,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   6,
									Column: 5,
									Byte:   66,
								},
								End: hcl.Pos{
									Line:   6,
									Column: 11,
									Byte:   72,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "local"},
										lang.AttrStep{Name: "top_obj"},
										lang.AttrStep{Name: "second"},
										lang.AttrStep{Name: "attr"},
									},
									Type:    cty.String,
									ScopeId: lang.ScopeId("local"),
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   7,
											Column: 7,
											Byte:   83,
										},
										End: hcl.Pos{
											Line:   7,
											Column: 19,
											Byte:   95,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   7,
											Column: 7,
											Byte:   83,
										},
										End: hcl.Pos{
											Line:   7,
											Column: 11,
											Byte:   87,
										},
									},
								},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "local"},
								lang.AttrStep{Name: "top_obj"},
								lang.AttrStep{Name: "third"},
							},
							Type: cty.Object(map[string]cty.Type{
								"attr": cty.String,
							}),
							ScopeId: lang.ScopeId("local"),
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   9,
									Column: 5,
									Byte:   106,
								},
								End: hcl.Pos{
									Line:   11,
									Column: 6,
									Byte:   140,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   9,
									Column: 5,
									Byte:   106,
								},
								End: hcl.Pos{
									Line:   9,
									Column: 10,
									Byte:   111,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "local"},
										lang.AttrStep{Name: "top_obj"},
										lang.AttrStep{Name: "third"},
										lang.AttrStep{Name: "attr"},
									},
									Type:    cty.String,
									ScopeId: lang.ScopeId("local"),
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   10,
											Column: 7,
											Byte:   122,
										},
										End: hcl.Pos{
											Line:   10,
											Column: 19,
											Byte:   134,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   10,
											Column: 7,
											Byte:   122,
										},
										End: hcl.Pos{
											Line:   10,
											Column: 11,
											Byte:   126,
										},
									},
								},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "local"},
								lang.AttrStep{Name: "top_obj"},
								lang.AttrStep{Name: "fourth"},
							},
							Type: cty.Object(map[string]cty.Type{
								"attr": cty.String,
							}),
							ScopeId: lang.ScopeId("local"),
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   12,
									Column: 5,
									Byte:   145,
								},
								End: hcl.Pos{
									Line:   14,
									Column: 6,
									Byte:   180,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   12,
									Column: 5,
									Byte:   145,
								},
								End: hcl.Pos{
									Line:   12,
									Column: 11,
									Byte:   151,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "local"},
										lang.AttrStep{Name: "top_obj"},
										lang.AttrStep{Name: "fourth"},
										lang.AttrStep{Name: "attr"},
									},
									Type:    cty.String,
									ScopeId: lang.ScopeId("local"),
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   13,
											Column: 7,
											Byte:   162,
										},
										End: hcl.Pos{
											Line:   13,
											Column: 19,
											Byte:   174,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   13,
											Column: 7,
											Byte:   162,
										},
										End: hcl.Pos{
											Line:   13,
											Column: 11,
											Byte:   166,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"block with missing label",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"output": {
						Labels: []*schema.LabelSchema{
							{Name: "name", IsDepKey: true},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "output"},
								schema.LabelStep{Index: 0},
							},
						},
						Body: &schema.BodySchema{},
					},
				},
			},
			`output {
}
`,
			reference.Targets{},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)

			d := testPathDecoder(t, &PathContext{
				Schema: tc.schema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			})

			refs, err := d.CollectReferenceTargets()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedRefs, refs, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of references: %s", diff)
			}
		})
	}
}
