package decoder

import (
	"errors"
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

func TestReferenceOriginAtPos(t *testing.T) {
	testCases := []struct {
		name           string
		cfg            string
		bodySchema     *schema.BodySchema
		pos            hcl.Pos
		expectedOrigin *reference.Origin
	}{
		{
			"empty config",
			``,
			&schema.BodySchema{},
			hcl.InitialPos,
			nil,
		},
		{
			"single-step traversal in root attribute",
			`attr = blah`,
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						IsOptional: true,
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{},
						},
					},
				},
			},
			hcl.Pos{
				Line:   1,
				Column: 9,
				Byte:   8,
			},
			&reference.Origin{
				Addr: lang.Address{
					lang.RootStep{Name: "blah"},
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 8,
						Byte:   7,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 12,
						Byte:   11,
					},
				},
				Constraints: reference.OriginConstraints{{}},
			},
		},
		{
			"string literal in root attribute",
			`attr = "blah"`,
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						IsOptional: true,
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{},
						},
					},
				},
			},
			hcl.Pos{
				Line:   1,
				Column: 9,
				Byte:   8,
			},
			nil,
		},
		{
			"multi-step traversal in root attribute",
			`attr = var.myobj.attr.foo.bar`,
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						IsOptional: true,
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{},
						},
					},
				},
			},
			hcl.Pos{
				Line:   1,
				Column: 9,
				Byte:   8,
			},
			&reference.Origin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "myobj"},
					lang.AttrStep{Name: "attr"},
					lang.AttrStep{Name: "foo"},
					lang.AttrStep{Name: "bar"},
				},
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
				Constraints: reference.OriginConstraints{{}},
			},
		},
		{
			"multi-step traversal with map index step in root attribute",
			`attr = var.myobj.mapattr["key"]`,
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						IsOptional: true,
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{},
						},
					},
				},
			},
			hcl.Pos{
				Line:   1,
				Column: 9,
				Byte:   8,
			},
			&reference.Origin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "myobj"},
					lang.AttrStep{Name: "mapattr"},
					lang.IndexStep{Key: cty.StringVal("key")},
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 8,
						Byte:   7,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 32,
						Byte:   31,
					},
				},
				Constraints: reference.OriginConstraints{{}},
			},
		},
		{
			"multi-step traversal with list index step in root attribute",
			`attr = var.myobj.listattr[4]`,
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						IsOptional: true,
						Expr: schema.ExprConstraints{
							schema.TraversalExpr{},
						},
					},
				},
			},
			hcl.Pos{
				Line:   1,
				Column: 9,
				Byte:   8,
			},
			&reference.Origin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "myobj"},
					lang.AttrStep{Name: "listattr"},
					lang.IndexStep{Key: cty.NumberIntVal(4)},
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 8,
						Byte:   7,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 29,
						Byte:   28,
					},
				},
				Constraints: reference.OriginConstraints{{}},
			},
		},
		{
			"multi-step traversal in block body",
			`customblock "foo" {
  attr = var.myobj.listattr[4]
}
`,
			&schema.BodySchema{
				Blocks: map[string]*schema.BlockSchema{
					"customblock": {
						Labels: []*schema.LabelSchema{
							{Name: "type"},
						},
						Body: &schema.BodySchema{
							Attributes: map[string]*schema.AttributeSchema{
								"attr": {
									IsOptional: true,
									Expr: schema.ExprConstraints{
										schema.TraversalExpr{},
									},
								},
							},
						},
					},
				},
			},
			hcl.Pos{
				Line:   2,
				Column: 11,
				Byte:   30,
			},
			&reference.Origin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "myobj"},
					lang.AttrStep{Name: "listattr"},
					lang.IndexStep{Key: cty.NumberIntVal(4)},
				},
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   2,
						Column: 10,
						Byte:   29,
					},
					End: hcl.Pos{
						Line:   2,
						Column: 31,
						Byte:   50,
					},
				},
				Constraints: reference.OriginConstraints{{}},
			},
		},
		{
			"traversal inside collection type",
			`attr = [ var.test ]`,
			&schema.BodySchema{
				Attributes: map[string]*schema.AttributeSchema{
					"attr": {
						IsOptional: true,
						Expr: schema.ExprConstraints{
							schema.SetExpr{
								Elem: schema.ExprConstraints{
									schema.TraversalExpr{OfScopeId: lang.ScopeId("test")},
								},
							},
						},
					},
				},
			},
			hcl.Pos{
				Line:   1,
				Column: 11,
				Byte:   12,
			},
			&reference.Origin{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "test"},
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
						Column: 18,
						Byte:   17,
					},
				},
				Constraints: reference.OriginConstraints{
					{OfScopeId: lang.ScopeId("test")},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)

			d := testPathDecoder(t, &PathContext{
				Schema: tc.bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			})

			refOrigin, err := d.ReferenceOriginAtPos("test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedOrigin, refOrigin, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatched reference origin: %s", diff)
			}
		})
	}
}

func TestReferenceOriginAtPos_json(t *testing.T) {
	f, diags := json.Parse([]byte(`{}`), "test.tf.json")
	if len(diags) > 0 {
		t.Fatal(diags)
	}

	d := testPathDecoder(t, &PathContext{
		Files: map[string]*hcl.File{
			"test.tf.json": f,
		},
	})

	_, err := d.ReferenceOriginAtPos("test.tf.json", hcl.InitialPos)
	unknownFormatErr := &UnknownFileFormatError{}
	if !errors.As(err, &unknownFormatErr) {
		t.Fatal("expected UnknownFileFormatError for JSON body")
	}
}

func TestReferenceOriginsTargeting(t *testing.T) {
	testCases := []struct {
		name            string
		allOrigins      reference.Origins
		refTarget       reference.Target
		expectedOrigins reference.Origins
	}{
		{
			"no origins",
			reference.Origins{},
			reference.Target{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
				},
				Type: cty.String,
			},
			reference.Origins{},
		},
		{
			"exact address match",
			reference.Origins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "secondstep"},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
			},
			reference.Target{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
					lang.AttrStep{Name: "secondstep"},
				},
				Type: cty.String,
			},
			reference.Origins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "secondstep"},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
			},
		},
		{
			"no match",
			reference.Origins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "secondstep"},
					},
				},
			},
			reference.Target{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
					lang.AttrStep{Name: "different"},
				},
				Type: cty.String,
			},
			reference.Origins{},
		},
		{
			"match of nested target - two matches",
			reference.Origins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.DynamicPseudoType},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "second"},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
			},
			reference.Target{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
				},
				Type: cty.Object(map[string]cty.Type{
					"second": cty.String,
				}),
				NestedTargets: reference.Targets{
					{
						Addr: lang.Address{
							lang.RootStep{Name: "test"},
							lang.AttrStep{Name: "second"},
						},
						Type: cty.String,
					},
				},
			},
			reference.Origins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.DynamicPseudoType},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "second"},
					},
					Constraints: reference.OriginConstraints{
						{OfType: cty.String},
					},
				},
			},
		},
		{
			"loose match of target of unknown type",
			reference.Origins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					Constraints: reference.OriginConstraints{{}},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Constraints: reference.OriginConstraints{{}},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "second"},
					},
					Constraints: reference.OriginConstraints{{}},
				},
			},
			reference.Target{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
				},
				Type: cty.DynamicPseudoType,
			},
			reference.Origins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Constraints: reference.OriginConstraints{{}},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "second"},
					},
					Constraints: reference.OriginConstraints{{}},
				},
			},
		},
		{
			"mismatch of target nil type",
			reference.Origins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Constraints: reference.OriginConstraints{
						{OfScopeId: lang.ScopeId("test")},
					},
				},
			},
			reference.Target{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
				},
				ScopeId: lang.ScopeId("test"),
				Type:    cty.String,
			},
			reference.Origins{},
		},
		// JSON edge cases
		{
			"constraint-less origin mismatching scope-only target",
			reference.Origins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "alpha"},
					},
					Constraints: nil,
				},
			},
			reference.Target{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "alpha"},
				},
				ScopeId: "variable",
				Type:    cty.NilType,
			},
			reference.Origins{},
		},
		{
			"constraint-less origin matching type-aware target",
			reference.Origins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "beta"},
					},
					Constraints: nil,
				},
			},
			reference.Target{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "beta"},
				},
				ScopeId: "variable",
				Type:    cty.DynamicPseudoType,
			},
			reference.Origins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "beta"},
					},
					Constraints: nil,
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			d := testPathDecoder(t, &PathContext{
				ReferenceOrigins: tc.allOrigins,
			})

			origins, err := d.ReferenceOriginsTargeting(tc.refTarget)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedOrigins, origins, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatched reference origins: %s", diff)
			}
		})
	}
}
