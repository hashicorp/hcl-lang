package decoder

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestAddress_Equals_basic(t *testing.T) {
	originalAddr := Address(lang.Address{
		lang.RootStep{Name: "provider"},
		lang.AttrStep{Name: "aws"},
	})

	matchingAddr := lang.Address{
		lang.RootStep{Name: "provider"},
		lang.AttrStep{Name: "aws"},
	}
	if !originalAddr.Equals(Address(matchingAddr)) {
		t.Fatalf("expected %q to match %q", originalAddr, matchingAddr)
	}

	mismatchingAddr := lang.Address{
		lang.RootStep{Name: "provider"},
		lang.AttrStep{Name: "aaa"},
	}
	if originalAddr.Equals(Address(mismatchingAddr)) {
		t.Fatalf("expected %q not to match %q", originalAddr, mismatchingAddr)
	}
}

func TestAddress_Equals_numericIndexStep(t *testing.T) {
	originalAddr := Address(lang.Address{
		lang.RootStep{Name: "aws_alb"},
		lang.AttrStep{Name: "example"},
		lang.IndexStep{Key: cty.NumberIntVal(0)},
	})

	matchingAddr := lang.Address{
		lang.RootStep{Name: "aws_alb"},
		lang.AttrStep{Name: "example"},
		lang.IndexStep{Key: cty.NumberIntVal(0)},
	}
	if !originalAddr.Equals(Address(matchingAddr)) {
		t.Fatalf("expected %q to match %q", originalAddr, matchingAddr)
	}

	mismatchingAddr := lang.Address{
		lang.RootStep{Name: "aws_alb"},
		lang.AttrStep{Name: "example"},
		lang.IndexStep{Key: cty.NumberIntVal(4)},
	}
	if originalAddr.Equals(Address(mismatchingAddr)) {
		t.Fatalf("expected %q not to match %q", originalAddr, mismatchingAddr)
	}
}

func TestAddress_Equals_stringIndexStep(t *testing.T) {
	originalAddr := Address(lang.Address{
		lang.RootStep{Name: "aws_alb"},
		lang.AttrStep{Name: "example"},
		lang.IndexStep{Key: cty.StringVal("first")},
	})

	matchingAddr := lang.Address{
		lang.RootStep{Name: "aws_alb"},
		lang.AttrStep{Name: "example"},
		lang.IndexStep{Key: cty.StringVal("first")},
	}
	if !originalAddr.Equals(Address(matchingAddr)) {
		t.Fatalf("expected %q to match %q", originalAddr, matchingAddr)
	}

	mismatchingAddr := lang.Address{
		lang.RootStep{Name: "aws_alb"},
		lang.AttrStep{Name: "example"},
		lang.IndexStep{Key: cty.StringVal("second")},
	}
	if originalAddr.Equals(Address(mismatchingAddr)) {
		t.Fatalf("expected %q not to match %q", originalAddr, mismatchingAddr)
	}
}

func TestDecodeReferences_noSchema(t *testing.T) {
	d := NewDecoder()
	_, err := d.DecodeReferences()
	if err == nil {
		t.Fatal("expected error when no schema is set")
	}

	noSchemaErr := &NoSchemaError{}
	if !errors.As(err, &noSchemaErr) {
		t.Fatalf("unexpected error: %#v, expected %#v", err, noSchemaErr)
	}
}

func TestDecodeReferences_basic(t *testing.T) {
	testCases := []struct {
		name         string
		schema       *schema.BodySchema
		cfg          string
		expectedRefs lang.References
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
			lang.References{
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
				},
			},
		},
		{
			"root attribute as data",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"testattr": {
						Address: &schema.AttributeAddrSchema{
							Steps: []schema.AddrStep{
								schema.StaticStep{Name: "special"},
								schema.AttrNameStep{},
							},
							AsData: true,
						},
						IsOptional: true,
						Expr:       schema.LiteralTypeOnly(cty.String),
					},
				},
			},
			`testattr = "example"
`,
			lang.References{
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
				},
			},
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
			lang.References{},
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
			lang.References{
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
			lang.References{
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
			lang.References{
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
			lang.References{
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
			lang.References{
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
}
`,
			lang.References{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws"},
					},
					Type: cty.Object(map[string]cty.Type{
						"attr":      cty.Number,
						"name":      cty.String,
						"map_attr":  cty.Map(cty.String),
						"list_attr": cty.List(cty.String),
					}),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   9,
							Column: 2,
							Byte:   133,
						},
					},
					InsideReferences: lang.References{
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
									Column: 2,
									Byte:   103,
								},
								End: hcl.Pos{
									Line:   8,
									Column: 30,
									Byte:   131,
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
									Column: 3,
									Byte:   101,
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
							// InferDependentBody:  true,
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
								},
							},
						},
					},
				},
			},
			`provider "aws" {
  attr = 42
  name = "hello world"
}
provider "test" {
  attr = 42
  name = "hello world"
}
`,
			lang.References{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws"},
					},
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.Number,
						"name": cty.String,
					}),
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
							Byte:   53,
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
								},
							},
						},
					},
				},
			},
			`provider "aws" {
  attr = 42
  name = "hello world"
}
provider "test" {
  attr = 42
  name = "hello world"
}
`,
			lang.References{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws"},
					},
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.Number,
						"name": cty.String,
					}),
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
							Byte:   53,
						},
					},
					InsideReferences: lang.References{
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
			lang.References{
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
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.Number,
						"objblock": cty.Object(map[string]cty.Type{
							"port":     cty.Number,
							"protocol": cty.String,
						}),
					}),
					InsideReferences: lang.References{
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
							Type: cty.Object(map[string]cty.Type{
								"port":     cty.Number,
								"protocol": cty.String,
							}),
							InsideReferences: lang.References{
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
			lang.References{
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
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.Number,
						"listblock": cty.List(cty.Object(map[string]cty.Type{
							"port":     cty.Number,
							"protocol": cty.String,
						})),
					}),
					InsideReferences: lang.References{
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
									Line:   11,
									Column: 2,
									Byte:   138,
								},
								End: hcl.Pos{
									Line:   11,
									Column: 2,
									Byte:   138,
								},
							},
							Type: cty.List(cty.Object(map[string]cty.Type{
								"port":     cty.Number,
								"protocol": cty.String,
							})),
							InsideReferences: lang.References{
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
									Type: cty.String,
								},
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
									Type: cty.String,
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
			lang.References{
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
					Type: cty.Object(map[string]cty.Type{
						"attr": cty.Number,
						"listener": cty.Map(cty.Object(map[string]cty.Type{
							"port":     cty.Number,
							"protocol": cty.String,
						})),
					}),
					InsideReferences: lang.References{
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
									Line:   10,
									Column: 2,
									Byte:   134,
								},
								End: hcl.Pos{
									Line:   10,
									Column: 2,
									Byte:   134,
								},
							},
							Type: cty.Map(cty.Object(map[string]cty.Type{
								"port":     cty.Number,
								"protocol": cty.String,
							})),
							InsideReferences: lang.References{
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
									Type: cty.String,
								},
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
											Line:   9,
											Column: 4,
											Byte:   132,
										},
										End: hcl.Pos{
											Line:   9,
											Column: 4,
											Byte:   132,
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
			lang.References{
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
			lang.References{
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
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			d := NewDecoder()
			d.SetSchema(tc.schema)

			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			err := d.LoadFile("test.tf", f)
			if err != nil {
				t.Fatal(err)
			}

			refs, err := d.DecodeReferences()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedRefs, refs, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of references: %s", diff)
			}
		})
	}
}
