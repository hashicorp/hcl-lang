package decoder

import (
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

func TestCollectReferenceOrigins_hcl(t *testing.T) {
	testCases := []struct {
		name            string
		schema          *schema.BodySchema
		cfg             string
		expectedOrigins lang.ReferenceOrigins
	}{
		{
			"no origins",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attribute": {
						Expr: schema.LiteralTypeOnly(cty.String),
					},
				},
			},
			`attribute = "foo-bar"`,
			lang.ReferenceOrigins{},
		},
		{
			"root attribute single step",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{},
						},
					},
				},
			},
			`attr = onestep`,
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Constraints: lang.ReferenceOriginConstraints{{}},
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
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{},
						},
					},
					"attr2": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{},
						},
					},
					"attr3": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{},
						},
					},
				},
			},
			`attr1 = onestep
attr2 = anotherstep
attr3 = onestep`,
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Constraints: lang.ReferenceOriginConstraints{{}},
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
				{
					Addr: lang.Address{
						lang.RootStep{Name: "anotherstep"},
					},
					Constraints: lang.ReferenceOriginConstraints{{}},
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
				{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Constraints: lang.ReferenceOriginConstraints{{}},
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
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{},
						},
					},
				},
			},
			`attr1 = "${onestep}-${onestep}-${another.foo.bar}"`,
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Constraints: lang.ReferenceOriginConstraints{{}},
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
				{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Constraints: lang.ReferenceOriginConstraints{{}},
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
				{
					Addr: lang.Address{
						lang.RootStep{Name: "another"},
						lang.AttrStep{Name: "foo"},
						lang.AttrStep{Name: "bar"},
					},
					Constraints: lang.ReferenceOriginConstraints{{}},
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
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{},
						},
					},
				},
			},
			`attr = one.two["key"].attr[0]`,
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "one"},
						lang.AttrStep{Name: "two"},
						lang.IndexStep{Key: cty.StringVal("key")},
						lang.AttrStep{Name: "attr"},
						lang.IndexStep{Key: cty.NumberIntVal(0)},
					},
					Constraints: lang.ReferenceOriginConstraints{{}},
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
									Expr: schema.ExprConstraints{
										schema.TraversalExpr{},
									},
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
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Constraints: lang.ReferenceOriginConstraints{{}},
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
								Expr: schema.ExprConstraints{
									schema.TraversalExpr{},
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
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "onestep"},
					},
					Constraints: lang.ReferenceOriginConstraints{{}},
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
									Expr: schema.ExprConstraints{
										schema.TraversalExpr{},
									},
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
										Expr: schema.ExprConstraints{
											schema.TraversalExpr{},
										},
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
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Constraints: lang.ReferenceOriginConstraints{{}},
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
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "second"},
					},
					Constraints: lang.ReferenceOriginConstraints{{}},
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
									Expr: schema.ExprConstraints{
										schema.TraversalExpr{},
									},
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
										Expr: schema.ExprConstraints{
											schema.TraversalExpr{},
										},
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
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "first"},
					},
					Constraints: lang.ReferenceOriginConstraints{{}},
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
						Expr: schema.ExprConstraints{
							schema.ListExpr{
								Elem: schema.ExprConstraints{
									schema.TraversalExpr{
										OfScopeId: lang.ScopeId("test"),
									},
								},
							},
						},
					},
					"set": {
						Expr: schema.ExprConstraints{
							schema.SetExpr{
								Elem: schema.ExprConstraints{
									schema.TraversalExpr{
										OfScopeId: lang.ScopeId("test"),
									},
								},
							},
						},
					},
					"tuple": {
						Expr: schema.ExprConstraints{
							schema.TupleExpr{
								Elems: []schema.ExprConstraints{
									{
										schema.TraversalExpr{
											OfScopeId: lang.ScopeId("test"),
										},
									},
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
			lang.ReferenceOrigins{
				{
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
					Constraints: lang.ReferenceOriginConstraints{
						{OfScopeId: lang.ScopeId("test")},
					},
				},
				{
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
					Constraints: lang.ReferenceOriginConstraints{
						{OfScopeId: lang.ScopeId("test")},
					},
				},
				{
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
					Constraints: lang.ReferenceOriginConstraints{
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
						Expr: schema.ExprConstraints{
							schema.ObjectExpr{
								Attributes: schema.ObjectExprAttributes{
									"attr": &schema.AttributeSchema{
										Expr: schema.ExprConstraints{
											schema.TraversalExpr{
												OfScopeId: lang.ScopeId("test"),
											},
										},
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
			lang.ReferenceOrigins{
				{
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
					Constraints: lang.ReferenceOriginConstraints{
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
						Expr: schema.ExprConstraints{
							schema.MapExpr{
								Elem: schema.ExprConstraints{
									schema.TraversalExpr{
										OfScopeId: lang.ScopeId("test"),
									},
								},
							},
						},
					},
				},
			},
			`map = {
  key = var.first
}`,
			lang.ReferenceOrigins{
				{
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
					Constraints: lang.ReferenceOriginConstraints{
						{OfScopeId: lang.ScopeId("test")},
					},
				},
			},
		},
		{
			"origin inside tuple cons expression",
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"tuple_cons": {
						Expr: schema.ExprConstraints{
							schema.TupleConsExpr{
								AnyElem: schema.ExprConstraints{
									schema.TraversalExpr{
										OfScopeId: lang.ScopeId("test"),
									},
								},
							},
						},
					},
				},
			},
			`tuple_cons = [ var.one ]`,
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "one"},
					},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 16,
							Byte:   15,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 23,
							Byte:   22,
						},
					},
					Constraints: lang.ReferenceOriginConstraints{
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
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{OfType: cty.Map(cty.String)},
							schema.MapExpr{
								Elem: schema.ExprConstraints{
									schema.TraversalExpr{OfType: cty.String},
								},
							},
						},
					},
					"obj": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{OfType: cty.Object(map[string]cty.Type{
								"foo": cty.String,
							})},
							schema.ObjectExpr{
								Attributes: schema.ObjectExprAttributes{
									"foo": &schema.AttributeSchema{
										Expr: schema.ExprConstraints{
											schema.TraversalExpr{OfType: cty.String},
										},
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
			lang.ReferenceOrigins{
				{
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
					Constraints: lang.ReferenceOriginConstraints{
						{OfType: cty.String},
					},
				},
				{
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
					Constraints: lang.ReferenceOriginConstraints{
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
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{OfType: cty.List(cty.String)},
							schema.ListExpr{
								Elem: schema.ExprConstraints{
									schema.TraversalExpr{OfType: cty.String},
								},
							},
						},
					},
					"set": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{OfType: cty.Set(cty.String)},
							schema.SetExpr{
								Elem: schema.ExprConstraints{
									schema.TraversalExpr{OfType: cty.String},
								},
							},
						},
					},
					"tup": {
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{OfType: cty.Tuple([]cty.Type{cty.String})},
							schema.TupleExpr{
								Elems: []schema.ExprConstraints{
									{
										schema.TraversalExpr{OfType: cty.String},
									},
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
			lang.ReferenceOrigins{
				{
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
					Constraints: lang.ReferenceOriginConstraints{
						{OfType: cty.String},
					},
				},
				{
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
					Constraints: lang.ReferenceOriginConstraints{
						{OfType: cty.String},
					},
				},
				{
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
					Constraints: lang.ReferenceOriginConstraints{
						{OfType: cty.String},
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
