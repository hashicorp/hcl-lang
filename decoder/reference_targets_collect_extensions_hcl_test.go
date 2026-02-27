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

func TestCollectReferenceTargets_extension_hcl(t *testing.T) {
	testCases := []struct {
		name         string
		schema       *schema.BodySchema
		cfg          string
		expectedRefs reference.Targets
	}{
		{
			"self references collection - attributes",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Address: &schema.BlockAddrSchema{
							Steps: schema.Address{
								schema.LabelStep{Index: 0},
								schema.LabelStep{Index: 1},
							},
							DependentBodyAsData:  true,
							InferDependentBody:   true,
							DependentBodySelfRef: true,
						},
						Labels: []*schema.LabelSchema{
							{
								Name:     "type",
								IsDepKey: true,
							},
							{
								Name: "name",
							},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"static": {IsOptional: true, Constraint: schema.LiteralType{Type: cty.String}},
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{
										Index: 0,
										Value: "aws_instance",
									},
								},
							}): {
								Attributes: map[string]*schema.AttributeSchema{
									"bar": {IsOptional: true, Constraint: schema.LiteralType{Type: cty.Number}},
									"foo": {IsOptional: true, Constraint: schema.LiteralType{Type: cty.String}},
								},
							},
						},
					},
				},
			},
			`resource "aws_instance" "blah" {
  static = "test"
  foo = "test"
  bar = 42
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_instance"},
						lang.AttrStep{Name: "blah"},
					},
					LocalAddr: lang.Address{
						lang.RootStep{Name: "self"},
					},
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   5,
							Column: 2,
							Byte:   78,
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
							Column: 31,
							Byte:   30,
						},
					},
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   5,
							Column: 2,
							Byte:   78,
						},
					},
					Type: cty.Object(map[string]cty.Type{
						"bar": cty.Number,
						"foo": cty.String,
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_instance"},
								lang.AttrStep{Name: "blah"},
								lang.AttrStep{Name: "bar"},
							},
							LocalAddr: lang.Address{
								lang.RootStep{Name: "self"},
								lang.AttrStep{Name: "bar"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   4,
									Column: 3,
									Byte:   68,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 11,
									Byte:   76,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   4,
									Column: 3,
									Byte:   68,
								},
								End: hcl.Pos{
									Line:   4,
									Column: 6,
									Byte:   71,
								},
							},
							Type: cty.Number,
							TargetableFromRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 32,
									Byte:   31,
								},
								End: hcl.Pos{
									Line:   5,
									Column: 2,
									Byte:   78,
								},
							},
						},
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_instance"},
								lang.AttrStep{Name: "blah"},
								lang.AttrStep{Name: "foo"},
							},
							LocalAddr: lang.Address{
								lang.RootStep{Name: "self"},
								lang.AttrStep{Name: "foo"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 3,
									Byte:   53,
								},
								End: hcl.Pos{
									Line:   3,
									Column: 15,
									Byte:   65,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 3,
									Byte:   53,
								},
								End: hcl.Pos{
									Line:   3,
									Column: 6,
									Byte:   56,
								},
							},
							Type: cty.String,
							TargetableFromRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 32,
									Byte:   31,
								},
								End: hcl.Pos{
									Line:   5,
									Column: 2,
									Byte:   78,
								},
							},
						},
					},
				},
			},
		},
		{
			"self references collection - object block",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Address: &schema.BlockAddrSchema{
							Steps: schema.Address{
								schema.LabelStep{Index: 0},
								schema.LabelStep{Index: 1},
							},
							DependentBodyAsData:  true,
							InferDependentBody:   true,
							DependentBodySelfRef: true,
						},
						Labels: []*schema.LabelSchema{
							{
								Name:     "type",
								IsDepKey: true,
							},
							{
								Name: "name",
							},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"static": {IsOptional: true, Constraint: schema.LiteralType{Type: cty.String}},
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{
										Index: 0,
										Value: "aws_instance",
									},
								},
							}): {
								Blocks: map[string]*schema.BlockSchema{
									"foo": {
										Type: schema.BlockTypeObject,
										Body: &schema.BodySchema{
											Attributes: map[string]*schema.AttributeSchema{
												"bar": {IsOptional: true, Constraint: schema.LiteralType{Type: cty.Number}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			`resource "aws_instance" "blah" {
  static = "test"
  foo {
    bar = 42
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_instance"},
						lang.AttrStep{Name: "blah"},
					},
					LocalAddr: lang.Address{
						lang.RootStep{Name: "self"},
					},
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
							Byte:   77,
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
							Column: 31,
							Byte:   30,
						},
					},
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   6,
							Column: 2,
							Byte:   77,
						},
					},
					Type: cty.Object(map[string]cty.Type{
						"foo": cty.Object(map[string]cty.Type{
							"bar": cty.Number,
						}),
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_instance"},
								lang.AttrStep{Name: "blah"},
								lang.AttrStep{Name: "foo"},
							},
							LocalAddr: lang.Address{
								lang.RootStep{Name: "self"},
								lang.AttrStep{Name: "foo"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 3,
									Byte:   53,
								},
								End: hcl.Pos{
									Line:   5,
									Column: 4,
									Byte:   75,
								},
							},
							DefRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 3,
									Byte:   53,
								},
								End: hcl.Pos{
									Line:   3,
									Column: 6,
									Byte:   56,
								},
							},
							Type: cty.Object(map[string]cty.Type{
								"bar": cty.Number,
							}),
							TargetableFromRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 32,
									Byte:   31,
								},
								End: hcl.Pos{
									Line:   6,
									Column: 2,
									Byte:   77,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws_instance"},
										lang.AttrStep{Name: "blah"},
										lang.AttrStep{Name: "foo"},
										lang.AttrStep{Name: "bar"},
									},
									LocalAddr: lang.Address{
										lang.RootStep{Name: "self"},
										lang.AttrStep{Name: "foo"},
										lang.AttrStep{Name: "bar"},
									},
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   4,
											Column: 5,
											Byte:   63,
										},
										End: hcl.Pos{
											Line:   4,
											Column: 13,
											Byte:   71,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   4,
											Column: 5,
											Byte:   63,
										},
										End: hcl.Pos{
											Line:   4,
											Column: 8,
											Byte:   66,
										},
									},
									Type: cty.Number,
									TargetableFromRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   1,
											Column: 32,
											Byte:   31,
										},
										End: hcl.Pos{
											Line:   6,
											Column: 2,
											Byte:   77,
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
			"self references collection - list block",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Address: &schema.BlockAddrSchema{
							Steps: schema.Address{
								schema.LabelStep{Index: 0},
								schema.LabelStep{Index: 1},
							},
							DependentBodyAsData:  true,
							InferDependentBody:   true,
							DependentBodySelfRef: true,
						},
						Labels: []*schema.LabelSchema{
							{
								Name:     "type",
								IsDepKey: true,
							},
							{
								Name: "name",
							},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"static": {IsOptional: true, Constraint: schema.LiteralType{Type: cty.String}},
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{
										Index: 0,
										Value: "aws_instance",
									},
								},
							}): {
								Blocks: map[string]*schema.BlockSchema{
									"foo": {
										Type: schema.BlockTypeList,
										Body: &schema.BodySchema{
											Attributes: map[string]*schema.AttributeSchema{
												"bar": {IsOptional: true, Constraint: schema.LiteralType{Type: cty.Number}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			`resource "aws_instance" "blah" {
  static = "test"
  foo {
    bar = 42
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_instance"},
						lang.AttrStep{Name: "blah"},
					},
					LocalAddr: lang.Address{
						lang.RootStep{Name: "self"},
					},
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
							Byte:   77,
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
							Column: 31,
							Byte:   30,
						},
					},
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   6,
							Column: 2,
							Byte:   77,
						},
					},
					Type: cty.Object(map[string]cty.Type{
						"foo": cty.List(cty.Object(map[string]cty.Type{
							"bar": cty.Number,
						})),
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_instance"},
								lang.AttrStep{Name: "blah"},
								lang.AttrStep{Name: "foo"},
							},
							LocalAddr: lang.Address{
								lang.RootStep{Name: "self"},
								lang.AttrStep{Name: "foo"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 3,
									Byte:   53,
								},
								End: hcl.Pos{
									Line:   5,
									Column: 4,
									Byte:   75,
								},
							},
							Type: cty.List(cty.Object(map[string]cty.Type{
								"bar": cty.Number,
							})),
							TargetableFromRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 32,
									Byte:   31,
								},
								End: hcl.Pos{
									Line:   6,
									Column: 2,
									Byte:   77,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws_instance"},
										lang.AttrStep{Name: "blah"},
										lang.AttrStep{Name: "foo"},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
									},
									LocalAddr: lang.Address{
										lang.RootStep{Name: "self"},
										lang.AttrStep{Name: "foo"},
										lang.IndexStep{Key: cty.NumberIntVal(0)},
									},
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   3,
											Column: 3,
											Byte:   53,
										},
										End: hcl.Pos{
											Line:   5,
											Column: 4,
											Byte:   75,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   3,
											Column: 3,
											Byte:   53,
										},
										End: hcl.Pos{
											Line:   3,
											Column: 6,
											Byte:   56,
										},
									},
									Type: cty.Object(map[string]cty.Type{
										"bar": cty.Number,
									}),
									TargetableFromRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   1,
											Column: 32,
											Byte:   31,
										},
										End: hcl.Pos{
											Line:   6,
											Column: 2,
											Byte:   77,
										},
									},
									NestedTargets: reference.Targets{
										{
											Addr: lang.Address{
												lang.RootStep{Name: "aws_instance"},
												lang.AttrStep{Name: "blah"},
												lang.AttrStep{Name: "foo"},
												lang.IndexStep{Key: cty.NumberIntVal(0)},
												lang.AttrStep{Name: "bar"},
											},
											LocalAddr: lang.Address{
												lang.RootStep{Name: "self"},
												lang.AttrStep{Name: "foo"},
												lang.IndexStep{Key: cty.NumberIntVal(0)},
												lang.AttrStep{Name: "bar"},
											},
											RangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   4,
													Column: 5,
													Byte:   63,
												},
												End: hcl.Pos{
													Line:   4,
													Column: 13,
													Byte:   71,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   4,
													Column: 5,
													Byte:   63,
												},
												End: hcl.Pos{
													Line:   4,
													Column: 8,
													Byte:   66,
												},
											},
											Type: cty.Number,
											TargetableFromRangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   1,
													Column: 32,
													Byte:   31,
												},
												End: hcl.Pos{
													Line:   6,
													Column: 2,
													Byte:   77,
												},
											},
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
			"self references collection - set block",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Address: &schema.BlockAddrSchema{
							Steps: schema.Address{
								schema.LabelStep{Index: 0},
								schema.LabelStep{Index: 1},
							},
							DependentBodyAsData:  true,
							InferDependentBody:   true,
							DependentBodySelfRef: true,
						},
						Labels: []*schema.LabelSchema{
							{
								Name:     "type",
								IsDepKey: true,
							},
							{
								Name: "name",
							},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"static": {IsOptional: true, Constraint: schema.LiteralType{Type: cty.String}},
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{
										Index: 0,
										Value: "aws_instance",
									},
								},
							}): {
								Blocks: map[string]*schema.BlockSchema{
									"foo": {
										Type: schema.BlockTypeSet,
										Body: &schema.BodySchema{
											Attributes: map[string]*schema.AttributeSchema{
												"bar": {IsOptional: true, Constraint: schema.LiteralType{Type: cty.Number}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			`resource "aws_instance" "blah" {
  static = "test"
  foo {
    bar = 42
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_instance"},
						lang.AttrStep{Name: "blah"},
					},
					LocalAddr: lang.Address{
						lang.RootStep{Name: "self"},
					},
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
							Byte:   77,
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
							Column: 31,
							Byte:   30,
						},
					},
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   6,
							Column: 2,
							Byte:   77,
						},
					},
					Type: cty.Object(map[string]cty.Type{
						"foo": cty.Set(cty.Object(map[string]cty.Type{
							"bar": cty.Number,
						})),
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_instance"},
								lang.AttrStep{Name: "blah"},
								lang.AttrStep{Name: "foo"},
							},
							LocalAddr: lang.Address{
								lang.RootStep{Name: "self"},
								lang.AttrStep{Name: "foo"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 3,
									Byte:   53,
								},
								End: hcl.Pos{
									Line:   5,
									Column: 4,
									Byte:   75,
								},
							},
							Type: cty.Set(cty.Object(map[string]cty.Type{
								"bar": cty.Number,
							})),
							TargetableFromRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 32,
									Byte:   31,
								},
								End: hcl.Pos{
									Line:   6,
									Column: 2,
									Byte:   77,
								},
							},
						},
					},
				},
			},
		},
		{
			"self references collection - map block",
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"resource": {
						Address: &schema.BlockAddrSchema{
							Steps: schema.Address{
								schema.LabelStep{Index: 0},
								schema.LabelStep{Index: 1},
							},
							DependentBodyAsData:  true,
							InferDependentBody:   true,
							DependentBodySelfRef: true,
						},
						Labels: []*schema.LabelSchema{
							{
								Name:     "type",
								IsDepKey: true,
							},
							{
								Name: "name",
							},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"static": {IsOptional: true, Constraint: schema.LiteralType{Type: cty.String}},
							},
						},
						DependentBody: map[schema.SchemaKey]*schema.BodySchema{
							schema.NewSchemaKey(schema.DependencyKeys{
								Labels: []schema.LabelDependent{
									{
										Index: 0,
										Value: "aws_instance",
									},
								},
							}): {
								Blocks: map[string]*schema.BlockSchema{
									"foo": {
										Type: schema.BlockTypeMap,
										Body: &schema.BodySchema{
											Attributes: map[string]*schema.AttributeSchema{
												"bar": {IsOptional: true, Constraint: schema.LiteralType{Type: cty.Number}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			`resource "aws_instance" "blah" {
  static = "test"
  foo "dog" {
    bar = 42
  }
}
`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "aws_instance"},
						lang.AttrStep{Name: "blah"},
					},
					LocalAddr: lang.Address{
						lang.RootStep{Name: "self"},
					},
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
							Column: 31,
							Byte:   30,
						},
					},
					TargetableFromRangePtr: &hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   6,
							Column: 2,
							Byte:   83,
						},
					},
					Type: cty.Object(map[string]cty.Type{
						"foo": cty.Map(cty.Object(map[string]cty.Type{
							"bar": cty.Number,
						})),
					}),
					NestedTargets: reference.Targets{
						{
							Addr: lang.Address{
								lang.RootStep{Name: "aws_instance"},
								lang.AttrStep{Name: "blah"},
								lang.AttrStep{Name: "foo"},
							},
							LocalAddr: lang.Address{
								lang.RootStep{Name: "self"},
								lang.AttrStep{Name: "foo"},
							},
							RangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   3,
									Column: 3,
									Byte:   53,
								},
								End: hcl.Pos{
									Line:   5,
									Column: 4,
									Byte:   81,
								},
							},
							Type: cty.Map(cty.Object(map[string]cty.Type{
								"bar": cty.Number,
							})),
							TargetableFromRangePtr: &hcl.Range{
								Filename: "test.tf",
								Start: hcl.Pos{
									Line:   1,
									Column: 32,
									Byte:   31,
								},
								End: hcl.Pos{
									Line:   6,
									Column: 2,
									Byte:   83,
								},
							},
							NestedTargets: reference.Targets{
								{
									Addr: lang.Address{
										lang.RootStep{Name: "aws_instance"},
										lang.AttrStep{Name: "blah"},
										lang.AttrStep{Name: "foo"},
										lang.IndexStep{Key: cty.StringVal("dog")},
									},
									LocalAddr: lang.Address{
										lang.RootStep{Name: "self"},
										lang.AttrStep{Name: "foo"},
										lang.IndexStep{Key: cty.StringVal("dog")},
									},
									RangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   3,
											Column: 3,
											Byte:   53,
										},
										End: hcl.Pos{
											Line:   5,
											Column: 4,
											Byte:   81,
										},
									},
									DefRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   3,
											Column: 3,
											Byte:   53,
										},
										End: hcl.Pos{
											Line:   3,
											Column: 12,
											Byte:   62,
										},
									},
									Type: cty.Object(map[string]cty.Type{
										"bar": cty.Number,
									}),
									TargetableFromRangePtr: &hcl.Range{
										Filename: "test.tf",
										Start: hcl.Pos{
											Line:   1,
											Column: 32,
											Byte:   31,
										},
										End: hcl.Pos{
											Line:   6,
											Column: 2,
											Byte:   83,
										},
									},
									NestedTargets: reference.Targets{
										{
											Addr: lang.Address{
												lang.RootStep{Name: "aws_instance"},
												lang.AttrStep{Name: "blah"},
												lang.AttrStep{Name: "foo"},
												lang.IndexStep{Key: cty.StringVal("dog")},
												lang.AttrStep{Name: "bar"},
											},
											LocalAddr: lang.Address{
												lang.RootStep{Name: "self"},
												lang.AttrStep{Name: "foo"},
												lang.IndexStep{Key: cty.StringVal("dog")},
												lang.AttrStep{Name: "bar"},
											},
											RangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   4,
													Column: 5,
													Byte:   69,
												},
												End: hcl.Pos{
													Line:   4,
													Column: 13,
													Byte:   77,
												},
											},
											DefRangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   4,
													Column: 5,
													Byte:   69,
												},
												End: hcl.Pos{
													Line:   4,
													Column: 8,
													Byte:   72,
												},
											},
											Type: cty.Number,
											TargetableFromRangePtr: &hcl.Range{
												Filename: "test.tf",
												Start: hcl.Pos{
													Line:   1,
													Column: 32,
													Byte:   31,
												},
												End: hcl.Pos{
													Line:   6,
													Column: 2,
													Byte:   83,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
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
