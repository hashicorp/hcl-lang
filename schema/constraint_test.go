package schema

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
)

var (
	_ Constraint = AnyExpression{}
	_ Constraint = Keyword{}
	_ Constraint = List{}
	_ Constraint = LiteralType{}
	_ Constraint = LiteralValue{}
	_ Constraint = Map{}
	_ Constraint = ObjectAttributes{}
	_ Constraint = Object{}
	_ Constraint = Set{}
	_ Constraint = Reference{}
	_ Constraint = Tuple{}
	_ Constraint = TypeDeclaration{}

	_ ConstraintWithHoverData = List{}
	_ ConstraintWithHoverData = LiteralType{}
	_ ConstraintWithHoverData = LiteralValue{}
	_ ConstraintWithHoverData = Map{}
	_ ConstraintWithHoverData = ObjectAttributes{}
	_ ConstraintWithHoverData = Object{}
	_ ConstraintWithHoverData = Set{}
	_ ConstraintWithHoverData = Tuple{}

	_ TypeAwareConstraint = AnyExpression{}
	_ TypeAwareConstraint = List{}
	_ TypeAwareConstraint = LiteralType{}
	_ TypeAwareConstraint = LiteralValue{}
	_ TypeAwareConstraint = Map{}
	_ TypeAwareConstraint = Object{}
	_ TypeAwareConstraint = OneOf{}
	_ TypeAwareConstraint = Set{}
	_ TypeAwareConstraint = Tuple{}
)

func TestConstraint_EmptyHoverData(t *testing.T) {
	testCases := []struct {
		cons              ConstraintWithHoverData
		expectedHoverData *HoverData
	}{
		{
			LiteralType{
				Type: cty.String,
			},
			&HoverData{
				Content: lang.Markdown(`string`),
			},
		},
		{
			List{
				Elem: LiteralType{
					Type: cty.String,
				},
			},
			&HoverData{
				Content: lang.Markdown("list(string)"),
			},
		},
		{
			LiteralType{
				Type: cty.List(cty.String),
			},
			&HoverData{
				Content: lang.Markdown("list(string)"),
			},
		},
		{
			Set{
				Elem: LiteralType{
					Type: cty.String,
				},
			},
			&HoverData{
				Content: lang.Markdown("set(string)"),
			},
		},
		{
			LiteralType{
				Type: cty.Set(cty.String),
			},
			&HoverData{
				Content: lang.Markdown("set(string)"),
			},
		},
		{
			LiteralType{
				Type: cty.Object(map[string]cty.Type{
					"foo": cty.String,
					"bar": cty.Number,
					"baz": cty.List(cty.String),
				}),
			},
			&HoverData{
				Content: lang.Markdown("```" + `
{
  bar = number
  baz = list(string)
  foo = string
}
` + "```\n"),
			},
		},
		{
			LiteralType{
				Type: cty.Object(map[string]cty.Type{
					"foo": cty.String,
					"bar": cty.Number,
					"baz": cty.Object(map[string]cty.Type{
						"foo": cty.String,
						"bar": cty.Number,
					}),
				}),
			},
			&HoverData{
				Content: lang.Markdown("```" + `
{
  bar = number
  baz = {
    bar = number
    foo = string
  }
  foo = string
}
` + "```\n"),
			},
		},
		// literal value
		{
			LiteralValue{
				Value: cty.StringVal("foobar"),
			},
			&HoverData{
				Content: lang.Markdown(`"foobar"`),
			},
		},
		{
			LiteralValue{
				Value: cty.StringVal("foo\nbar"),
			},
			&HoverData{
				Content: lang.Markdown("```\nfoo\nbar\n```\n"),
			},
		},
		{
			LiteralValue{
				Value: cty.NumberIntVal(42),
			},
			&HoverData{
				Content: lang.Markdown(`42`),
			},
		},
		{
			LiteralValue{
				Value: cty.ObjectVal(map[string]cty.Value{
					"foo": cty.StringVal("too"),
					"bar": cty.NumberIntVal(42),
					"baz": cty.ListVal([]cty.Value{cty.StringVal("toot")}),
				}),
			},
			&HoverData{
				Content: lang.Markdown("```" + `
{
  bar = 42
  baz = tolist(["toot"])
  foo = "too"
}
` + "```\n"),
			},
		},
		{
			LiteralValue{
				Value: cty.MapVal(map[string]cty.Value{
					"foo": cty.StringVal("too"),
					"bar": cty.StringVal("boo"),
				}),
			},
			&HoverData{
				Content: lang.Markdown("```" + `
tomap({
  "bar" = "boo"
  "foo" = "too"
})
` + "```\n"),
			},
		},
		{
			LiteralValue{
				Value: cty.MapVal(map[string]cty.Value{
					"foo": cty.MapVal(map[string]cty.Value{
						"noot": cty.StringVal("noot"),
					}),
					"bar": cty.MapVal(map[string]cty.Value{
						"baz": cty.StringVal("toot"),
					}),
				}),
			},
			&HoverData{
				Content: lang.Markdown("```" + `
tomap({
  "bar" = tomap({
    "baz" = "toot"
  })
  "foo" = tomap({
    "noot" = "noot"
  })
})
` + "```\n"),
			},
		},
		{
			LiteralValue{
				Value: cty.ObjectVal(map[string]cty.Value{
					"foo": cty.StringVal("too"),
					"bar": cty.NumberIntVal(43),
					"baz": cty.ObjectVal(map[string]cty.Value{
						"foo": cty.StringVal("boo"),
						"bar": cty.NumberIntVal(32),
					}),
				}),
			},
			&HoverData{
				Content: lang.Markdown("```" + `
{
  bar = 43
  baz = {
    bar = 32
    foo = "boo"
  }
  foo = "too"
}
` + "```\n"),
			},
		},
		{
			LiteralValue{
				Value: cty.ObjectVal(map[string]cty.Value{
					"foo": cty.StringVal("too"),
					"bar": cty.NumberIntVal(43),
					"baz": cty.MapVal(map[string]cty.Value{
						"foo": cty.NumberIntVal(42),
						"bar": cty.NumberIntVal(32),
					}),
				}),
			},
			&HoverData{
				Content: lang.Markdown("```" + `
{
  bar = 43
  baz = tomap({
    "bar" = 32
    "foo" = 42
  })
  foo = "too"
}
` + "```\n"),
			},
		},

		// negative tests
		{
			List{
				Elem: Keyword{
					Keyword: "kw",
				},
			},
			nil,
		},
		{
			Set{
				Elem: Keyword{
					Keyword: "kw",
				},
			},
			nil,
		},
		{
			Tuple{
				Elems: []Constraint{
					Keyword{
						Keyword: "kw",
					},
				},
			},
			nil,
		},
		{
			Map{
				Elem: Keyword{
					Keyword: "kw",
				},
			},
			nil,
		},
		{
			Object{
				Attributes: map[string]*AttributeSchema{
					"foo": {
						Constraint: Keyword{
							Keyword: "kw",
						},
					},
				},
			},
			nil,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			hoverData := tc.cons.EmptyHoverData(0)
			if diff := cmp.Diff(tc.expectedHoverData, hoverData); diff != "" {
				t.Fatalf("unexpected hover data: %s", diff)
			}
		})
	}
}
