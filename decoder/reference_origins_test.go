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
		expectedOrigin *lang.ReferenceOrigin
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
			&lang.ReferenceOrigin{
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
				Constraints: lang.ReferenceOriginConstraints{{}},
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
			&lang.ReferenceOrigin{
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
				Constraints: lang.ReferenceOriginConstraints{{}},
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
			&lang.ReferenceOrigin{
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
				Constraints: lang.ReferenceOriginConstraints{{}},
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
			&lang.ReferenceOrigin{
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
				Constraints: lang.ReferenceOriginConstraints{{}},
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
			&lang.ReferenceOrigin{
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
				Constraints: lang.ReferenceOriginConstraints{{}},
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
			&lang.ReferenceOrigin{
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
				Constraints: lang.ReferenceOriginConstraints{
					{OfScopeId: lang.ScopeId("test")},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			d := NewDecoder()
			d.SetSchema(tc.bodySchema)

			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			err := d.LoadFile("test.tf", f)
			if err != nil {
				t.Fatal(err)
			}

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
	d := NewDecoder()

	f, diags := json.Parse([]byte(`{}`), "test.tf.json")
	if len(diags) > 0 {
		t.Fatal(diags)
	}
	err := d.LoadFile("test.tf.json", f)
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.ReferenceOriginAtPos("test.tf.json", hcl.InitialPos)
	unknownFormatErr := &UnknownFileFormatError{}
	if !errors.As(err, &unknownFormatErr) {
		t.Fatal("expected UnknownFileFormatError for JSON body")
	}
}

func TestReferenceOriginsTargeting(t *testing.T) {
	testCases := []struct {
		name            string
		allOrigins      lang.ReferenceOrigins
		refTarget       lang.ReferenceTarget
		expectedOrigins lang.ReferenceOrigins
	}{
		{
			"no origins",
			lang.ReferenceOrigins{},
			lang.ReferenceTarget{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
				},
				Type: cty.String,
			},
			lang.ReferenceOrigins{},
		},
		{
			"exact address match",
			lang.ReferenceOrigins{
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
					Constraints: lang.ReferenceOriginConstraints{
						{OfType: cty.String},
					},
				},
			},
			lang.ReferenceTarget{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
					lang.AttrStep{Name: "secondstep"},
				},
				Type: cty.String,
			},
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "secondstep"},
					},
					Constraints: lang.ReferenceOriginConstraints{
						{OfType: cty.String},
					},
				},
			},
		},
		{
			"no match",
			lang.ReferenceOrigins{
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
			lang.ReferenceTarget{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
					lang.AttrStep{Name: "different"},
				},
				Type: cty.String,
			},
			lang.ReferenceOrigins{},
		},
		{
			"match of nested target - two matches",
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Constraints: lang.ReferenceOriginConstraints{
						{OfType: cty.DynamicPseudoType},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "second"},
					},
					Constraints: lang.ReferenceOriginConstraints{
						{OfType: cty.String},
					},
				},
			},
			lang.ReferenceTarget{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
				},
				Type: cty.Object(map[string]cty.Type{
					"second": cty.String,
				}),
				NestedTargets: lang.ReferenceTargets{
					{
						Addr: lang.Address{
							lang.RootStep{Name: "test"},
							lang.AttrStep{Name: "second"},
						},
						Type: cty.String,
					},
				},
			},
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Constraints: lang.ReferenceOriginConstraints{
						{OfType: cty.DynamicPseudoType},
					},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "second"},
					},
					Constraints: lang.ReferenceOriginConstraints{
						{OfType: cty.String},
					},
				},
			},
		},
		{
			"loose match of target of unknown type",
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					Constraints: lang.ReferenceOriginConstraints{{}},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Constraints: lang.ReferenceOriginConstraints{{}},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "second"},
					},
					Constraints: lang.ReferenceOriginConstraints{{}},
				},
			},
			lang.ReferenceTarget{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
				},
				Type: cty.DynamicPseudoType,
			},
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Constraints: lang.ReferenceOriginConstraints{{}},
				},
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
						lang.AttrStep{Name: "second"},
					},
					Constraints: lang.ReferenceOriginConstraints{{}},
				},
			},
		},
		{
			"mismatch of target nil type",
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "test"},
					},
					Constraints: lang.ReferenceOriginConstraints{
						{OfScopeId: lang.ScopeId("test")},
					},
				},
			},
			lang.ReferenceTarget{
				Addr: lang.Address{
					lang.RootStep{Name: "test"},
				},
				ScopeId: lang.ScopeId("test"),
				Type:    cty.String,
			},
			lang.ReferenceOrigins{},
		},
		// JSON edge cases
		{
			"constraint-less origin mismatching scope-only target",
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "alpha"},
					},
					Constraints: nil,
				},
			},
			lang.ReferenceTarget{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "alpha"},
				},
				ScopeId: "variable",
				Type:    cty.NilType,
			},
			lang.ReferenceOrigins{},
		},
		{
			"constraint-less origin matching type-aware target",
			lang.ReferenceOrigins{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "beta"},
					},
					Constraints: nil,
				},
			},
			lang.ReferenceTarget{
				Addr: lang.Address{
					lang.RootStep{Name: "var"},
					lang.AttrStep{Name: "beta"},
				},
				ScopeId: "variable",
				Type:    cty.DynamicPseudoType,
			},
			lang.ReferenceOrigins{
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
			d := NewDecoder()
			d.SetReferenceOriginReader(func() lang.ReferenceOrigins {
				return tc.allOrigins
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
