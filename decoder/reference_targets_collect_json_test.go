// Copyright (c) HashiCorp, Inc.
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
	"github.com/hashicorp/hcl/v2/json"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestCollectReferenceTargets_json(t *testing.T) {
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
						Constraint: schema.AnyExpression{OfType: cty.String},
					},
				},
			},
			`{"testattr": "${example}"}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "special"},
						lang.AttrStep{Name: "testattr"},
					},
					ScopeId: lang.ScopeId("specialthing"),
					RangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   1,
							Column: 2,
							Byte:   1,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 26,
							Byte:   25,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   1,
							Column: 2,
							Byte:   1,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 12,
							Byte:   11,
						},
					},
					NestedTargets: reference.Targets{},
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
						Constraint: schema.LiteralType{Type: cty.String},
					},
				},
			},
			`{"testattr": "example"}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "special"},
						lang.AttrStep{Name: "testattr"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   1,
							Column: 2,
							Byte:   1,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 23,
							Byte:   22,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   1,
							Column: 2,
							Byte:   1,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 12,
							Byte:   11,
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
						Constraint: schema.LiteralType{Type: cty.DynamicPseudoType},
					},
				},
			},
			`{"testattr": "example"}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "special"},
						lang.AttrStep{Name: "testattr"},
					},
					Type: cty.String,
					RangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   1,
							Column: 2,
							Byte:   1,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 23,
							Byte:   22,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   1,
							Column: 2,
							Byte:   1,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 12,
							Byte:   11,
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
						Constraint: schema.LiteralType{
							Type: cty.Object(map[string]cty.Type{
								"nestedattr": cty.String,
							}),
						},
					},
				},
			},
			`{
  "testattr": {
    "nestedattr": "test"
  }
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
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   4,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 4,
							Byte:   46,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   4,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 13,
							Byte:   14,
						},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "special"},
								lang.AttrStep{Name: "testattr"},
								lang.AttrStep{Name: "nestedattr"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start:    hcl.Pos{Line: 3, Column: 5, Byte: 22},
								End:      hcl.Pos{Line: 3, Column: 25, Byte: 42},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start:    hcl.Pos{Line: 3, Column: 5, Byte: 22},
								End:      hcl.Pos{Line: 3, Column: 17, Byte: 34},
							},
							Type: cty.String,
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
						Constraint: schema.LiteralType{Type: cty.Map(cty.String)},
					},
				},
			},
			`{
  "testattr": {
    "nestedattr": "test"
  }
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
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   4,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 4,
							Byte:   46,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   4,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 13,
							Byte:   14,
						},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "special"},
								lang.AttrStep{Name: "testattr"},
								lang.IndexStep{Key: cty.StringVal("nestedattr")},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start:    hcl.Pos{Line: 3, Column: 5, Byte: 22},
								End:      hcl.Pos{Line: 3, Column: 25, Byte: 42},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start:    hcl.Pos{Line: 3, Column: 5, Byte: 22},
								End:      hcl.Pos{Line: 3, Column: 17, Byte: 34},
							},
							Type: cty.String,
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
						Constraint: schema.LiteralType{Type: cty.List(cty.String)},
						IsOptional: true,
					},
				},
			},
			`{
  "testattr": [ "example" ]
}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "special"},
						lang.AttrStep{Name: "testattr"},
					},
					Type: cty.List(cty.String),
					RangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   4,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 28,
							Byte:   29,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   4,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 13,
							Byte:   14,
						},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "special"},
								lang.AttrStep{Name: "testattr"},
								lang.IndexStep{Key: cty.NumberIntVal(0)},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start:    hcl.Pos{Line: 2, Column: 17, Byte: 18},
								End:      hcl.Pos{Line: 2, Column: 26, Byte: 27},
							},
							Type: cty.String,
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
						Constraint: schema.List{
							Elem: schema.LiteralType{Type: cty.String},
						},
						IsOptional: true,
					},
				},
			},
			`{
  "testattr": [ "example" ]
}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "special"},
						lang.AttrStep{Name: "testattr"},
					},
					Type: cty.List(cty.String),
					RangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   4,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 28,
							Byte:   29,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   4,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 13,
							Byte:   14,
						},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "special"},
								lang.AttrStep{Name: "testattr"},
								lang.IndexStep{Key: cty.NumberIntVal(0)},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start:    hcl.Pos{Line: 2, Column: 17, Byte: 18},
								End:      hcl.Pos{Line: 2, Column: 26, Byte: 27},
							},
							Type: cty.String,
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
			`{"testattr": "example"}`,
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
									Constraint: schema.LiteralType{Type: cty.Number},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`{
  "resource": {
  	"blah": {
      "attr": 3
    }
  }
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
									Constraint: schema.LiteralType{Type: cty.Number},
									IsOptional: true,
								},
								"name": {
									Constraint: schema.LiteralType{Type: cty.String},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`{
  "resource": {
    "blah": {
      "test": {
        "attr": 3,
        "name": "lorem ipsum"
      }
    }
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "blah"},
						lang.AttrStep{Name: "test"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   4,
							Column: 15,
							Byte:   46,
						},
						End: hcl.Pos{
							Line:   7,
							Column: 8,
							Byte:   104,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   4,
							Column: 15,
							Byte:   46,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 16,
							Byte:   47,
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
									Constraint: schema.LiteralType{Type: cty.Number},
									IsOptional: true,
								},
								"name": {
									Constraint: schema.LiteralType{Type: cty.String},
									IsOptional: true,
								},
								"map_attr": {
									Constraint: schema.Map{Elem: schema.LiteralType{Type: cty.String}},
									IsOptional: true,
								},
								"list_attr": {
									Constraint: schema.List{Elem: schema.LiteralType{Type: cty.String}},
									IsOptional: true,
								},
								"set_attr": {
									Constraint: schema.Set{Elem: schema.LiteralType{Type: cty.String}},
									IsOptional: true,
								},
								"tuple_attr": {
									Constraint: schema.Tuple{
										Elems: []schema.Constraint{
											schema.LiteralType{Type: cty.String},
											schema.LiteralType{Type: cty.Number},
										},
									},
									IsOptional: true,
								},
								"obj_attr": {
									Constraint: schema.Object{
										Attributes: schema.ObjectAttributes{
											"example": &schema.AttributeSchema{
												Constraint: schema.LiteralType{Type: cty.String},
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
			`{
  "resource": {
    "blah": {
      "test": {
        "attr": 3,
        "name": "lorem ipsum",
        "map_attr": {
          "one": "hello",
          "two": "world"
        },
        "list_attr": [ "one", "two" ],
        "set_attr": [ "one", "two" ],
        "tuple_attr": [ "one", 42 ],
        "obj_attr": {
          "example": "blah"
        }
      }
    }
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
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   4,
							Column: 15,
							Byte:   46,
						},
						End: hcl.Pos{
							Line:   17,
							Column: 8,
							Byte:   363,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   4,
							Column: 15,
							Byte:   46,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 16,
							Byte:   47,
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
									Constraint: schema.LiteralType{Type: cty.Number},
									IsOptional: true,
								},
								"name": {
									Constraint: schema.LiteralType{Type: cty.String},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`{
  "resource": {
    "blah": {
      "test": {
        "attr": 3,
        "name": "lorem ipsum"
      }
    }
  }
}`,
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
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   4,
							Column: 15,
							Byte:   46,
						},
						End: hcl.Pos{
							Line:   7,
							Column: 8,
							Byte:   104,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   4,
							Column: 15,
							Byte:   46,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 16,
							Byte:   47,
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
									Constraint: schema.LiteralType{Type: cty.Number},
									IsOptional: true,
								},
								"name": {
									Constraint: schema.LiteralType{Type: cty.String},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`{
  "resource": {
    "blah": {
      "test": {
        "attr": 3,
        "name": "lorem ipsum"
      }
    }
  }
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
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   4,
							Column: 15,
							Byte:   46,
						},
						End: hcl.Pos{
							Line:   7,
							Column: 8,
							Byte:   104,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   4,
							Column: 15,
							Byte:   46,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 16,
							Byte:   47,
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
									Constraint: schema.LiteralType{Type: cty.Number},
									IsOptional: true,
								},
								"protocol": {
									Constraint: schema.LiteralType{Type: cty.String},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`{
  "listener": {
    "http": {
      "source_port": 80,
      "protocol": "tcp"
    },
    "https": {
      "source_port": 443,
      "protocol": "tcp"
    }
  }
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
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   30,
						},
						End: hcl.Pos{
							Line:   6,
							Column: 6,
							Byte:   86,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   30,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 14,
							Byte:   31,
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
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   7,
							Column: 14,
							Byte:   101,
						},
						End: hcl.Pos{
							Line:   10,
							Column: 6,
							Byte:   158,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   7,
							Column: 14,
							Byte:   101,
						},
						End: hcl.Pos{
							Line:   7,
							Column: 15,
							Byte:   102,
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
									Constraint: schema.LiteralType{Type: cty.Number},
									IsOptional: true,
								},
								"name": {
									Constraint: schema.LiteralType{Type: cty.String},
									IsOptional: true,
								},
								"map_attr": {
									Constraint: schema.Map{
										Elem: schema.LiteralType{Type: cty.String},
									},
									IsOptional: true,
								},
								"list_attr": {
									Constraint: schema.List{
										Elem: schema.LiteralType{Type: cty.String},
									},
									IsOptional: true,
								},
								"obj_attr": {
									Constraint: schema.LiteralType{
										Type: cty.Object(map[string]cty.Type{
											"nestedattr": cty.String,
										}),
									},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`{
  "provider": {
    "aws": {
      "attr": 42,
      "name": "hello world",
      "map_attr": {
        "one": "hello",
        "two": "world"
      },
      "list_attr": [ "one", "two" ],
      "obj_attr": {
        "nestedattr": "foo"
      }
    }
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
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   29,
						},
						End: hcl.Pos{
							Line:   14,
							Column: 6,
							Byte:   252,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   29,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   30,
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
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   4,
									Column: 7,
									Byte:   37,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 17,
									Byte:   47,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   4,
									Column: 7,
									Byte:   37,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 13,
									Byte:   43,
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
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   10,
									Column: 7,
									Byte:   160,
								},
								End: hcl.Pos{
									Line:   10,
									Column: 36,
									Byte:   189,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   10,
									Column: 7,
									Byte:   160,
								},
								End: hcl.Pos{
									Line:   10,
									Column: 18,
									Byte:   171,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws"},
										lang.AttrStep{Name: "list_attr"},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
									},
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 10, Column: 22, Byte: 175},
										End:      hcl.Pos{Line: 10, Column: 27, Byte: 180},
									},
									Type: cty.String,
								},
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws"},
										lang.AttrStep{Name: "list_attr"},
										lang.IndexStep{Key: cty.NumberIntVal(1)},
									},
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 10, Column: 29, Byte: 182},
										End:      hcl.Pos{Line: 10, Column: 34, Byte: 187},
									},
									Type: cty.String,
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
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   6,
									Column: 7,
									Byte:   84,
								},
								End: hcl.Pos{
									Line:   9,
									Column: 8,
									Byte:   152,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   6,
									Column: 7,
									Byte:   84,
								},
								End: hcl.Pos{
									Line:   6,
									Column: 17,
									Byte:   94,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws"},
										lang.AttrStep{Name: "map_attr"},
										lang.IndexStep{Key: cty.StringVal("one")},
									},
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 7, Column: 9, Byte: 106},
										End:      hcl.Pos{Line: 7, Column: 23, Byte: 120},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 7, Column: 9, Byte: 106},
										End:      hcl.Pos{Line: 7, Column: 14, Byte: 111},
									},
									Type: cty.String,
								},
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws"},
										lang.AttrStep{Name: "map_attr"},
										lang.IndexStep{Key: cty.StringVal("two")},
									},
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 8, Column: 9, Byte: 130},
										End:      hcl.Pos{Line: 8, Column: 23, Byte: 144},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 8, Column: 9, Byte: 130},
										End:      hcl.Pos{Line: 8, Column: 14, Byte: 135},
									},
									Type: cty.String,
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
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   5,
									Column: 7,
									Byte:   55,
								},
								End: hcl.Pos{
									Line:   5,
									Column: 28,
									Byte:   76,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   5,
									Column: 7,
									Byte:   55,
								},
								End: hcl.Pos{
									Line:   5,
									Column: 13,
									Byte:   61,
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
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   11,
									Column: 7,
									Byte:   197,
								},
								End: hcl.Pos{
									Line:   13,
									Column: 8,
									Byte:   246,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   11,
									Column: 7,
									Byte:   197,
								},
								End: hcl.Pos{
									Line:   11,
									Column: 17,
									Byte:   207,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws"},
										lang.AttrStep{Name: "obj_attr"},
										lang.AttrStep{Name: "nestedattr"},
									},
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 12, Column: 9, Byte: 219},
										End:      hcl.Pos{Line: 12, Column: 28, Byte: 238},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 12, Column: 9, Byte: 219},
										End:      hcl.Pos{Line: 12, Column: 21, Byte: 231},
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
										Constraint: schema.LiteralType{Type: cty.Number},
										IsOptional: true,
									},
									"name": {
										Constraint: schema.LiteralType{Type: cty.String},
										IsOptional: true,
									},
									"attr_list": {
										Constraint: schema.LiteralType{Type: cty.List(cty.String)},
										IsOptional: true,
									},
									"attr_map": {
										Constraint: schema.LiteralType{Type: cty.Map(cty.String)},
										IsOptional: true,
									},
									"obj": {
										Constraint: schema.LiteralType{Type: cty.Object(map[string]cty.Type{
											"nestedattr": cty.String,
										})},
										IsOptional: true,
									},
								},
							},
						},
					},
				},
			},
			`{
  "provider": {
    "aws": {
      "attr": 42,
      "name": "hello world",
      "attr_list": ["one", "two"],
      "attr_map": {
        "foo": "bar"
      },
      "obj": {
        "nestedattr": "test"
      }
    },
    "test": {
      "attr": 42,
      "name": "hello world",
      "attr_list": ["one", "two"],
      "attr_map": {
        "foo": "bar"
      },
      "obj": {
        "nestedattr": "test"
      }
    }
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
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   29,
						},
						End: hcl.Pos{
							Line:   13,
							Column: 6,
							Byte:   220,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   29,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   30,
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
										Constraint: schema.LiteralType{Type: cty.Number},
										IsOptional: true,
									},
									"name": {
										Constraint: schema.LiteralType{Type: cty.String},
										IsOptional: true,
									},
									"attr_list": {
										Constraint: schema.LiteralType{Type: cty.List(cty.String)},
										IsOptional: true,
									},
									"attr_map": {
										Constraint: schema.LiteralType{Type: cty.Map(cty.String)},
										IsOptional: true,
									},
									"obj": {
										Constraint: schema.LiteralType{Type: cty.Object(map[string]cty.Type{
											"nestedattr": cty.String,
										})},
										IsOptional: true,
									},
								},
							},
						},
					},
				},
			},
			`{
  "provider": {
    "aws": {
      "attr": 42,
      "name": "hello world",
      "attr_list": ["one", "two"],
      "attr_map": {
        "foo": "bar"
      },
      "obj": {
        "nestedattr": "test"
      }
    },
    "test": {
      "attr": 42,
      "name": "hello world",
      "attr_list": ["one", "two"],
      "attr_map": {
        "foo": "bar"
      },
      "obj": {
        "nestedattr": "test"
      }
    }
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws"},
					},
					LocalAddr: lang.Address{},
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
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   29,
						},
						End: hcl.Pos{
							Line:   13,
							Column: 6,
							Byte:   220,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   29,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   30,
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
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   4,
									Column: 7,
									Byte:   37,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 17,
									Byte:   47,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   4,
									Column: 7,
									Byte:   37,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 13,
									Byte:   43,
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
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   6,
									Column: 7,
									Byte:   84,
								},
								End: hcl.Pos{
									Line:   6,
									Column: 34,
									Byte:   111,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   6,
									Column: 7,
									Byte:   84,
								},
								End: hcl.Pos{
									Line:   6,
									Column: 18,
									Byte:   95,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws"},
										lang.AttrStep{Name: "attr_list"},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
									},
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 6, Column: 21, Byte: 98},
										End:      hcl.Pos{Line: 6, Column: 26, Byte: 103},
									},
									Type: cty.String,
								},
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws"},
										lang.AttrStep{Name: "attr_list"},
										lang.IndexStep{Key: cty.NumberIntVal(1)},
									},
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 6, Column: 28, Byte: 105},
										End:      hcl.Pos{Line: 6, Column: 33, Byte: 110},
									},
									Type: cty.String,
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
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   7,
									Column: 7,
									Byte:   119,
								},
								End: hcl.Pos{
									Line:   9,
									Column: 8,
									Byte:   161,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   7,
									Column: 7,
									Byte:   119,
								},
								End: hcl.Pos{
									Line:   7,
									Column: 17,
									Byte:   129,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws"},
										lang.AttrStep{Name: "attr_map"},
										lang.IndexStep{Key: cty.StringVal("foo")},
									},
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 8, Column: 9, Byte: 141},
										End:      hcl.Pos{Line: 8, Column: 21, Byte: 153},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 8, Column: 9, Byte: 141},
										End:      hcl.Pos{Line: 8, Column: 14, Byte: 146},
									},
									Type: cty.String,
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
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   5,
									Column: 7,
									Byte:   55,
								},
								End: hcl.Pos{
									Line:   5,
									Column: 28,
									Byte:   76,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   5,
									Column: 7,
									Byte:   55,
								},
								End: hcl.Pos{
									Line:   5,
									Column: 13,
									Byte:   61,
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
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   10,
									Column: 7,
									Byte:   169,
								},
								End: hcl.Pos{
									Line:   12,
									Column: 8,
									Byte:   214,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   10,
									Column: 7,
									Byte:   169,
								},
								End: hcl.Pos{
									Line:   10,
									Column: 12,
									Byte:   174,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws"},
										lang.AttrStep{Name: "obj"},
										lang.AttrStep{Name: "nestedattr"},
									},
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 11, Column: 9, Byte: 186},
										End:      hcl.Pos{Line: 11, Column: 29, Byte: 206},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 11, Column: 9, Byte: 186},
										End:      hcl.Pos{Line: 11, Column: 21, Byte: 198},
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
												Constraint: schema.LiteralType{Type: cty.String},
												IsOptional: true,
											},
											"port": {
												Constraint: schema.LiteralType{Type: cty.Number},
												IsOptional: true,
											},
										},
									},
								},
							},
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Constraint: schema.LiteralType{Type: cty.Number},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`{
  "rootblock": {
  	"aws": {
      "attr": 42,
      "objblock": {
        "port": 80,
        "protocol": "tcp"
      }
    }
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
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   29,
						},
						End: hcl.Pos{
							Line:   9,
							Column: 6,
							Byte:   128,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   29,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   30,
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
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   4,
									Column: 7,
									Byte:   37,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 17,
									Byte:   47,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   4,
									Column: 7,
									Byte:   37,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 13,
									Byte:   43,
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
							LocalAddr: lang.Address{},
							RangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   5,
									Column: 19,
									Byte:   67,
								},
								End: hcl.Pos{
									Line:   8,
									Column: 8,
									Byte:   122,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   5,
									Column: 19,
									Byte:   67,
								},
								End: hcl.Pos{
									Line:   5,
									Column: 20,
									Byte:   68,
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
										Filename: "test.tf.json",
										Start: hcl.Pos{
											Line:   6,
											Column: 9,
											Byte:   77,
										},
										End: hcl.Pos{
											Line:   6,
											Column: 19,
											Byte:   87,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start: hcl.Pos{
											Line:   6,
											Column: 9,
											Byte:   77,
										},
										End: hcl.Pos{
											Line:   6,
											Column: 15,
											Byte:   83,
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
										Filename: "test.tf.json",
										Start: hcl.Pos{
											Line:   7,
											Column: 9,
											Byte:   97,
										},
										End: hcl.Pos{
											Line:   7,
											Column: 26,
											Byte:   114,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start: hcl.Pos{
											Line:   7,
											Column: 9,
											Byte:   97,
										},
										End: hcl.Pos{
											Line:   7,
											Column: 19,
											Byte:   107,
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
												Constraint: schema.LiteralType{Type: cty.String},
												IsOptional: true,
											},
											"port": {
												Constraint: schema.LiteralType{Type: cty.Number},
												IsOptional: true,
											},
										},
									},
								},
							},
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Constraint: schema.LiteralType{Type: cty.Number},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`{
  "rootblock": {
  	"aws": {
      "attr": 42,
      "listblock": {
        "port": 80,
        "protocol": "tcp"
      },
      "listblock": {
        "port": 443,
        "protocol": "tcp"
      }
    }
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
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   29,
						},
						End: hcl.Pos{
							Line:   13,
							Column: 6,
							Byte:   206,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   29,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   30,
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
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   4,
									Column: 7,
									Byte:   37,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 17,
									Byte:   47,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   4,
									Column: 7,
									Byte:   37,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 13,
									Byte:   43,
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
							LocalAddr: lang.Address{},
							RangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   5,
									Column: 20,
									Byte:   68,
								},
								End: hcl.Pos{
									Line:   8,
									Column: 8,
									Byte:   123,
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
									LocalAddr: lang.Address{},
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start: hcl.Pos{
											Line:   5,
											Column: 20,
											Byte:   68,
										},
										End: hcl.Pos{
											Line:   8,
											Column: 8,
											Byte:   123,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start: hcl.Pos{
											Line:   5,
											Column: 20,
											Byte:   68,
										},
										End: hcl.Pos{
											Line:   5,
											Column: 21,
											Byte:   69,
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
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   6,
													Column: 9,
													Byte:   78,
												},
												End: hcl.Pos{
													Line:   6,
													Column: 19,
													Byte:   88,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   6,
													Column: 9,
													Byte:   78,
												},
												End: hcl.Pos{
													Line:   6,
													Column: 15,
													Byte:   84,
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
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   7,
													Column: 9,
													Byte:   98,
												},
												End: hcl.Pos{
													Line:   7,
													Column: 26,
													Byte:   115,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   7,
													Column: 9,
													Byte:   98,
												},
												End: hcl.Pos{
													Line:   7,
													Column: 19,
													Byte:   108,
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
									LocalAddr: lang.Address{},
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start: hcl.Pos{
											Line:   9,
											Column: 20,
											Byte:   144,
										},
										End: hcl.Pos{
											Line:   12,
											Column: 8,
											Byte:   200,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start: hcl.Pos{
											Line:   9,
											Column: 20,
											Byte:   144,
										},
										End: hcl.Pos{
											Line:   9,
											Column: 21,
											Byte:   145,
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
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   10,
													Column: 9,
													Byte:   154,
												},
												End: hcl.Pos{
													Line:   10,
													Column: 20,
													Byte:   165,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   10,
													Column: 9,
													Byte:   154,
												},
												End: hcl.Pos{
													Line:   10,
													Column: 15,
													Byte:   160,
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
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   11,
													Column: 9,
													Byte:   175,
												},
												End: hcl.Pos{
													Line:   11,
													Column: 26,
													Byte:   192,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   11,
													Column: 9,
													Byte:   175,
												},
												End: hcl.Pos{
													Line:   11,
													Column: 19,
													Byte:   185,
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
												Constraint: schema.LiteralType{Type: cty.String},
												IsOptional: true,
											},
											"port": {
												Constraint: schema.LiteralType{Type: cty.Number},
												IsOptional: true,
											},
										},
									},
								},
							},
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Constraint: schema.LiteralType{Type: cty.Number},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`{
  "rootblock": {
  	"aws": {
      "attr": 42,
      "setblock": {
        "port": 80,
        "protocol": "tcp"
      },
      "setblock": {
        "port": 443,
        "protocol": "tcp"
      }
    }
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
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   29,
						},
						End: hcl.Pos{
							Line:   13,
							Column: 6,
							Byte:   204,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   29,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   30,
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
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   4,
									Column: 7,
									Byte:   37,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 17,
									Byte:   47,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   4,
									Column: 7,
									Byte:   37,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 13,
									Byte:   43,
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
							LocalAddr: lang.Address{},
							RangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   5,
									Column: 19,
									Byte:   67,
								},
								End: hcl.Pos{
									Line:   8,
									Column: 8,
									Byte:   122,
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
												Constraint: schema.LiteralType{Type: cty.String},
												IsOptional: true,
											},
											"port": {
												Constraint: schema.LiteralType{Type: cty.Number},
												IsOptional: true,
											},
										},
									},
								},
							},
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Constraint: schema.LiteralType{Type: cty.Number},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`{
  "rootblock": {
    "aws": {
      "listblock": {
        "port": 80,
        "protocol": "tcp"
      },
      "attr": 42,
      "listblock": {
        "port": 443,
        "protocol": "tcp"
      }
    }
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
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   30,
						},
						End: hcl.Pos{
							Line:   13,
							Column: 6,
							Byte:   207,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   30,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   31,
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
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   8,
									Column: 7,
									Byte:   114,
								},
								End: hcl.Pos{
									Line:   8,
									Column: 17,
									Byte:   124,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   8,
									Column: 7,
									Byte:   114,
								},
								End: hcl.Pos{
									Line:   8,
									Column: 13,
									Byte:   120,
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
							LocalAddr: lang.Address{},
							RangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   4,
									Column: 20,
									Byte:   51,
								},
								End: hcl.Pos{
									Line:   7,
									Column: 8,
									Byte:   106,
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
									LocalAddr: lang.Address{},
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start: hcl.Pos{
											Line:   4,
											Column: 20,
											Byte:   51,
										},
										End: hcl.Pos{
											Line:   7,
											Column: 8,
											Byte:   106,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start: hcl.Pos{
											Line:   4,
											Column: 20,
											Byte:   51,
										},
										End: hcl.Pos{
											Line:   4,
											Column: 21,
											Byte:   52,
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
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   5,
													Column: 9,
													Byte:   61,
												},
												End: hcl.Pos{
													Line:   5,
													Column: 19,
													Byte:   71,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   5,
													Column: 9,
													Byte:   61,
												},
												End: hcl.Pos{
													Line:   5,
													Column: 15,
													Byte:   67,
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
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   6,
													Column: 9,
													Byte:   81,
												},
												End: hcl.Pos{
													Line:   6,
													Column: 26,
													Byte:   98,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   6,
													Column: 9,
													Byte:   81,
												},
												End: hcl.Pos{
													Line:   6,
													Column: 19,
													Byte:   91,
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
									LocalAddr: lang.Address{},
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start: hcl.Pos{
											Line:   9,
											Column: 20,
											Byte:   145,
										},
										End: hcl.Pos{
											Line:   12,
											Column: 8,
											Byte:   201,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start: hcl.Pos{
											Line:   9,
											Column: 20,
											Byte:   145,
										},
										End: hcl.Pos{
											Line:   9,
											Column: 21,
											Byte:   146,
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
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   10,
													Column: 9,
													Byte:   155,
												},
												End: hcl.Pos{
													Line:   10,
													Column: 20,
													Byte:   166,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   10,
													Column: 9,
													Byte:   155,
												},
												End: hcl.Pos{
													Line:   10,
													Column: 15,
													Byte:   161,
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
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   11,
													Column: 9,
													Byte:   176,
												},
												End: hcl.Pos{
													Line:   11,
													Column: 26,
													Byte:   193,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   11,
													Column: 9,
													Byte:   176,
												},
												End: hcl.Pos{
													Line:   11,
													Column: 19,
													Byte:   186,
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
												Constraint: schema.LiteralType{Type: cty.String},
												IsOptional: true,
											},
											"port": {
												Constraint: schema.LiteralType{Type: cty.Number},
												IsOptional: true,
											},
										},
									},
								},
							},
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									Constraint: schema.LiteralType{Type: cty.Number},
									IsOptional: true,
								},
							},
						},
					},
				},
			},
			`{
  "load_balancer": {
    "aws": {
      "attr": 42,
      "listener": {
        "http": {
          "port": 80,
          "protocol": "tcp"
        },
        "https": {
          "port": 443
        }
      }
    }
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
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   34,
						},
						End: hcl.Pos{
							Line:   14,
							Column: 6,
							Byte:   217,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   34,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   35,
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
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   4,
									Column: 7,
									Byte:   42,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 17,
									Byte:   52,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   4,
									Column: 7,
									Byte:   42,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 13,
									Byte:   48,
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
							LocalAddr: lang.Address{},
							RangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   6,
									Column: 17,
									Byte:   90,
								},
								End: hcl.Pos{
									Line:   9,
									Column: 10,
									Byte:   151,
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
									LocalAddr: lang.Address{},
									Type: cty.Object(map[string]cty.Type{
										"port":     cty.Number,
										"protocol": cty.String,
									}),
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start: hcl.Pos{
											Line:   6,
											Column: 17,
											Byte:   90,
										},
										End: hcl.Pos{
											Line:   9,
											Column: 10,
											Byte:   151,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start: hcl.Pos{
											Line:   6,
											Column: 17,
											Byte:   90,
										},
										End: hcl.Pos{
											Line:   6,
											Column: 18,
											Byte:   91,
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
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   7,
													Column: 11,
													Byte:   102,
												},
												End: hcl.Pos{
													Line:   7,
													Column: 21,
													Byte:   112,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   7,
													Column: 11,
													Byte:   102,
												},
												End: hcl.Pos{
													Line:   7,
													Column: 17,
													Byte:   108,
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
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   8,
													Column: 11,
													Byte:   124,
												},
												End: hcl.Pos{
													Line:   8,
													Column: 28,
													Byte:   141,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   8,
													Column: 11,
													Byte:   124,
												},
												End: hcl.Pos{
													Line:   8,
													Column: 21,
													Byte:   134,
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
									LocalAddr: lang.Address{},
									Type: cty.Object(map[string]cty.Type{
										"port":     cty.Number,
										"protocol": cty.String,
									}),
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start: hcl.Pos{
											Line:   10,
											Column: 18,
											Byte:   170,
										},
										End: hcl.Pos{
											Line:   12,
											Column: 10,
											Byte:   203,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start: hcl.Pos{
											Line:   10,
											Column: 18,
											Byte:   170,
										},
										End: hcl.Pos{
											Line:   10,
											Column: 19,
											Byte:   171,
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
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   11,
													Column: 11,
													Byte:   182,
												},
												End: hcl.Pos{
													Line:   11,
													Column: 22,
													Byte:   193,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   11,
													Column: 11,
													Byte:   182,
												},
												End: hcl.Pos{
													Line:   11,
													Column: 17,
													Byte:   188,
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
												Filename: "test.tf.json",
												Start: hcl.Pos{
													Line:   12,
													Column: 9,
													Byte:   202,
												},
												End: hcl.Pos{
													Line:   12,
													Column: 9,
													Byte:   202,
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
						Constraint: schema.Reference{
							Address: &schema.ReferenceAddrSchema{
								ScopeId: lang.ScopeId("specialthing"),
							},
						},
					},
				},
			},
			`{"testattr": "${special.test}"}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "special"},
						lang.AttrStep{Name: "test"},
					},
					ScopeId: lang.ScopeId("specialthing"),
					RangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 17, Byte: 16},
						End:      hcl.Pos{Line: 1, Column: 29, Byte: 28},
					},
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
									Constraint: schema.LiteralType{Type: cty.String},
								},
							},
						},
					},
				},
			},
			`{
  "provider": {
    "aws": {
      "alias": "euwest"
    }
  }
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
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   29,
						},
						End: hcl.Pos{
							Line:   5,
							Column: 6,
							Byte:   60,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   29,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   30,
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
			`{
  "variable": {
    "test": {}
  }
}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Type: cty.DynamicPseudoType,
					RangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   30,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 15,
							Byte:   32,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   30,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 14,
							Byte:   31,
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
									Constraint: schema.TypeDeclaration{},
								},
							},
						},
					},
				},
			},
			`{
  "variable": {
    "test": {
      "type": "map(string)"
    }
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Type: cty.Map(cty.String),
					RangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   30,
						},
						End: hcl.Pos{
							Line:   5,
							Column: 6,
							Byte:   65,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   30,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 14,
							Byte:   31,
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
								AttributeExpr: "type",
							},
						},
						Type: schema.BlockTypeObject,
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"type": {
									IsOptional: true,
									Constraint: schema.TypeDeclaration{},
								},
								"default": {
									IsOptional: true,
									Constraint: schema.LiteralType{Type: cty.DynamicPseudoType},
								},
							},
						},
					},
				},
			},
			`{
  "variable": {
    "test": {
      "default": ["something"]
    }
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Type: cty.DynamicPseudoType,
					RangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   30,
						},
						End: hcl.Pos{
							Line:   5,
							Column: 6,
							Byte:   68,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   30,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 14,
							Byte:   31,
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
								AttributeExpr: "type",
							},
						},
						Type: schema.BlockTypeObject,
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"type": {
									IsOptional: true,
									Constraint: schema.TypeDeclaration{},
								},
								"default": {
									IsOptional: true,
									Constraint: schema.LiteralType{Type: cty.DynamicPseudoType},
								},
							},
						},
					},
				},
			},
			`{
  "variable": {
    "test": {
      "type": "list(any)",
      "default": [ "one" ]
    }
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Type: cty.List(cty.DynamicPseudoType),
					RangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   30,
						},
						End: hcl.Pos{
							Line:   6,
							Column: 6,
							Byte:   91,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   30,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 14,
							Byte:   31,
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
			`{
  "module": {
    "test": {}
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "module"},
						lang.AttrStep{Name: "test"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   28,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 15,
							Byte:   30,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   28,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 14,
							Byte:   29,
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
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   28,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 15,
							Byte:   30,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   28,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 14,
							Byte:   29,
						},
					},
				},
			},
		},
		{
			"additional targetables",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"store": {
						Labels: []*schema.LabelSchema{
							{Name: "name"},
						},
						Address: &schema.BlockAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "store"},
								schema.LabelStep{Index: 0},
							},
							SupportUnknownNestedRefs: true,
						},
					},
				},
			},
			`{
  "store": {
    "test": {}
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "store"},
						lang.AttrStep{Name: "test"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   27,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 15,
							Byte:   29,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   27,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 14,
							Byte:   28,
						},
					},
					Type: cty.DynamicPseudoType,
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
										Constraint: schema.LiteralType{Type: cty.String},
										IsOptional: true,
									},
								},
							},
						},
					},
				},
			},
			`{
  "module": {
    "test": {
      "attr": "foo"
    },
    "different": {
      "attr": "foo"
    }
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "module"},
						lang.AttrStep{Name: "test"},
					},
					LocalAddr: lang.Address{},
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.String,
					}),
					RangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   28,
						},
						End: hcl.Pos{
							Line:   5,
							Column: 6,
							Byte:   55,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 13,
							Byte:   28,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 14,
							Byte:   29,
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
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   4,
									Column: 7,
									Byte:   36,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 20,
									Byte:   49,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start: hcl.Pos{
									Line:   4,
									Column: 7,
									Byte:   36,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 13,
									Byte:   42,
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
								Constraint: schema.OneOf{
									schema.Reference{OfType: cty.DynamicPseudoType},
									schema.LiteralType{Type: cty.DynamicPseudoType},
								},
							},
						},
					},
				},
			},
			`{
  "locals": {
    "top_obj": {
      "first": {
        "attr": "val"
      },
      "second": {
        "attr": "val"
      },
      "third": {
        "attr": "val"
      },
      "fourth": {
        "attr": "val"
      }
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
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 5,
							Byte:   20,
						},
						End: hcl.Pos{
							Line:   16,
							Column: 6,
							Byte:   231,
						},
					},
					DefRangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start: hcl.Pos{
							Line:   3,
							Column: 5,
							Byte:   20,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 14,
							Byte:   29,
						},
					},
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "local"},
								lang.AttrStep{Name: "top_obj"},
								lang.AttrStep{Name: "first"},
							},
							ScopeId: "local",
							RangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start:    hcl.Pos{Line: 4, Column: 7, Byte: 39},
								End:      hcl.Pos{Line: 6, Column: 8, Byte: 79},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start:    hcl.Pos{Line: 4, Column: 7, Byte: 39},
								End:      hcl.Pos{Line: 4, Column: 14, Byte: 46},
							},
							Type: cty.Object(map[string]cty.Type{
								"attr": cty.String,
							}),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "local"},
										lang.AttrStep{Name: "top_obj"},
										lang.AttrStep{Name: "first"},
										lang.AttrStep{Name: "attr"},
									},
									ScopeId: "local",
									Type:    cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 5, Column: 9, Byte: 58},
										End:      hcl.Pos{Line: 5, Column: 22, Byte: 71},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 5, Column: 9, Byte: 58},
										End:      hcl.Pos{Line: 5, Column: 15, Byte: 64},
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
							ScopeId: "local",
							RangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start:    hcl.Pos{Line: 13, Column: 7, Byte: 184},
								End:      hcl.Pos{Line: 15, Column: 8, Byte: 225},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start:    hcl.Pos{Line: 13, Column: 7, Byte: 184},
								End:      hcl.Pos{Line: 13, Column: 15, Byte: 192},
							},
							Type: cty.Object(map[string]cty.Type{
								"attr": cty.String,
							}),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "local"},
										lang.AttrStep{Name: "top_obj"},
										lang.AttrStep{Name: "fourth"},
										lang.AttrStep{Name: "attr"},
									},
									ScopeId: "local",
									Type:    cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 14, Column: 9, Byte: 204},
										End:      hcl.Pos{Line: 14, Column: 22, Byte: 217},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 14, Column: 9, Byte: 204},
										End:      hcl.Pos{Line: 14, Column: 15, Byte: 210},
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
							ScopeId: "local",
							RangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start:    hcl.Pos{Line: 7, Column: 7, Byte: 87},
								End:      hcl.Pos{Line: 9, Column: 8, Byte: 128},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start:    hcl.Pos{Line: 7, Column: 7, Byte: 87},
								End:      hcl.Pos{Line: 7, Column: 15, Byte: 95},
							},
							Type: cty.Object(map[string]cty.Type{
								"attr": cty.String,
							}),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "local"},
										lang.AttrStep{Name: "top_obj"},
										lang.AttrStep{Name: "second"},
										lang.AttrStep{Name: "attr"},
									},
									ScopeId: "local",
									Type:    cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 8, Column: 9, Byte: 107},
										End:      hcl.Pos{Line: 8, Column: 22, Byte: 120},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 8, Column: 9, Byte: 107},
										End:      hcl.Pos{Line: 8, Column: 15, Byte: 113},
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
							ScopeId: "local",
							RangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start:    hcl.Pos{Line: 10, Column: 7, Byte: 136},
								End:      hcl.Pos{Line: 12, Column: 8, Byte: 176},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf.json",
								Start:    hcl.Pos{Line: 10, Column: 7, Byte: 136},
								End:      hcl.Pos{Line: 10, Column: 14, Byte: 143},
							},
							Type: cty.Object(map[string]cty.Type{
								"attr": cty.String,
							}),
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "local"},
										lang.AttrStep{Name: "top_obj"},
										lang.AttrStep{Name: "third"},
										lang.AttrStep{Name: "attr"},
									},
									ScopeId: "local",
									Type:    cty.String,
									RangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 11, Column: 9, Byte: 155},
										End:      hcl.Pos{Line: 11, Column: 22, Byte: 168},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf.json",
										Start:    hcl.Pos{Line: 11, Column: 9, Byte: 155},
										End:      hcl.Pos{Line: 11, Column: 15, Byte: 161},
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
			`{
  "output": {}
}
`,
			reference.Targets{},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			f, diags := json.Parse([]byte(tc.cfg), "test.tf.json")
			if len(diags) > 0 {
				t.Fatal(diags)
			}

			d := testPathDecoder(t, &PathContext{
				Schema: tc.schema,
				Files: map[string]*hcl.File{
					"test.tf.json": f,
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
