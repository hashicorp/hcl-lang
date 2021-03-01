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

func TestDecoder_HoverAtPos_expressions(t *testing.T) {
	testCases := []struct {
		name         string
		attrSchema   map[string]*schema.AttributeSchema
		cfg          string
		pos          hcl.Pos
		expectedData *lang.HoverData
		expectedErr  error
	}{
		{
			"string as type",
			map[string]*schema.AttributeSchema{
				"str_attr": {Expr: schema.LiteralTypeOnly(cty.String)},
			},
			`str_attr = "test"`,
			hcl.Pos{Line: 1, Column: 15, Byte: 14},
			&lang.HoverData{
				Content: lang.Markdown("`\"test\"` _string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 12,
						Byte:   11,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 18,
						Byte:   17,
					},
				},
			},
			nil,
		},
		{
			"single-line heredoc string as type",
			map[string]*schema.AttributeSchema{
				"str_attr": {Expr: schema.LiteralTypeOnly(cty.String)},
			},
			`str_attr = <<EOT
hello world
EOT
`,
			hcl.Pos{Line: 2, Column: 3, Byte: 19},
			&lang.HoverData{
				Content: lang.Markdown("```\n" +
					"hello world\n```\n" +
					`_string_`),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 12,
						Byte:   11,
					},
					End: hcl.Pos{
						Line:   3,
						Column: 4,
						Byte:   32,
					},
				},
			},
			nil,
		},
		{
			"multi-line heredoc string as type",
			map[string]*schema.AttributeSchema{
				"str_attr": {Expr: schema.LiteralTypeOnly(cty.String)},
			},
			`str_attr = <<EOT
hello
world
EOT
`,
			hcl.Pos{Line: 2, Column: 3, Byte: 19},
			&lang.HoverData{
				Content: lang.Markdown("```\n" +
					"hello\nworld\n```\n" +
					`_string_`),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 12,
						Byte:   11,
					},
					End: hcl.Pos{
						Line:   4,
						Column: 4,
						Byte:   32,
					},
				},
			},
			nil,
		},
		{
			"integer as type",
			map[string]*schema.AttributeSchema{
				"int_attr": {Expr: schema.LiteralTypeOnly(cty.Number)},
			},
			`int_attr = 4222524`,
			hcl.Pos{Line: 1, Column: 15, Byte: 14},
			&lang.HoverData{
				Content: lang.Markdown("`4222524` _number_"),
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
			nil,
		},
		{
			"float as type",
			map[string]*schema.AttributeSchema{
				"float_attr": {Expr: schema.LiteralTypeOnly(cty.Number)},
			},
			`float_attr = 42.3212`,
			hcl.Pos{Line: 1, Column: 16, Byte: 15},
			&lang.HoverData{
				Content: lang.Markdown("`42.3212` _number_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 14,
						Byte:   13,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 21,
						Byte:   20,
					},
				},
			},
			nil,
		},
		{
			"incompatible type",
			map[string]*schema.AttributeSchema{
				"not_str": {Expr: schema.LiteralTypeOnly(cty.Bool)},
			},
			`not_str = "blah"`,
			hcl.Pos{Line: 1, Column: 14, Byte: 13},
			nil,
			&ConstraintMismatch{},
		},
		{
			"string as value",
			map[string]*schema.AttributeSchema{
				"lit1": {Expr: schema.ExprConstraints{
					schema.LiteralValue{Val: cty.StringVal("foo")},
				}},
			},
			`lit1 = "foo"`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			&lang.HoverData{
				Content: lang.Markdown("`\"foo\"` _string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 8,
						Byte:   7,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 13,
						Byte:   12,
					},
				},
			},
			nil,
		},
		{
			"mismatching literal value",
			map[string]*schema.AttributeSchema{
				"lit2": {Expr: schema.ExprConstraints{
					schema.LiteralValue{Val: cty.StringVal("bar")},
				}},
			},
			`lit2 = "baz"`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			nil,
			&ConstraintMismatch{},
		},
		{
			"object as type",
			map[string]*schema.AttributeSchema{
				"litobj": {Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{
					"source":     cty.String,
					"bool":       cty.Bool,
					"notbool":    cty.String,
					"nested_map": cty.Map(cty.String),
					"nested_obj": cty.Object(map[string]cty.Type{
						"one": cty.String,
						"two": cty.Number,
					}),
				}))},
			},
			`litobj = {
    "source" = "blah"
    "different" = 42
    "bool" = true
    "notbool" = "test"
  }`,
			hcl.Pos{Line: 4, Column: 12, Byte: 65},
			&lang.HoverData{
				Content: lang.Markdown("```" + `
{
  bool = bool
  nested_map = map of string
  nested_obj = {
    one = string
    two = number
  }
  notbool = string
  source = string
}
` + "```" + `
_object_`),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 10,
						Byte:   9,
					},
					End: hcl.Pos{
						Line:   6,
						Column: 4,
						Byte:   98,
					},
				},
			},
			nil,
		},
		{
			"object as type with unknown key",
			map[string]*schema.AttributeSchema{
				"litobj": {Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{
					"source":     cty.String,
					"bool":       cty.Bool,
					"notbool":    cty.String,
					"nested_map": cty.Map(cty.String),
					"nested_obj": cty.Object(map[string]cty.Type{}),
				}))},
			},
			`litobj = {
    "${var.src}" = "blah"
    "${var.env}.${another}" = "prod"
    "different" = 42
    "bool" = true
    "notbool" = "test"
  }`,
			hcl.Pos{Line: 4, Column: 12, Byte: 65},
			&lang.HoverData{
				Content: lang.Markdown("```" + `
{
  bool = bool
  nested_map = map of string
  nested_obj = object
  notbool = string
  source = string
}
` + "```" + `
_object_`),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 10,
						Byte:   9,
					},
					End: hcl.Pos{
						Line:   7,
						Column: 4,
						Byte:   139,
					},
				},
			},
			nil,
		},
		{
			"list of object expressions",
			map[string]*schema.AttributeSchema{
				"objects": {Expr: schema.ExprConstraints{
					schema.ListExpr{
						Elem: schema.ExprConstraints{
							schema.ObjectExpr{
								Attributes: schema.ObjectExprAttributes{
									"source": &schema.AttributeSchema{
										Expr: schema.LiteralTypeOnly(cty.String),
									},
									"bool": &schema.AttributeSchema{
										Expr: schema.LiteralTypeOnly(cty.Bool),
									},
									"notbool": &schema.AttributeSchema{
										Expr: schema.LiteralTypeOnly(cty.String),
									},
									"nested_map": &schema.AttributeSchema{
										Expr: schema.LiteralTypeOnly(cty.Map(cty.String)),
									},
									"nested_obj": &schema.AttributeSchema{
										Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{})),
									},
								},
							},
						},
					},
				}},
			},
			`objects = [
    {
        source = "blah"
        different = 42
        bool = true
        notbool = "test"
    }
]`,
			hcl.Pos{Line: 1, Column: 3, Byte: 2},
			&lang.HoverData{
				Content: lang.Markdown(`**objects** _list_`),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 1,
						Byte:   0,
					},
					End: hcl.Pos{
						Line:   8,
						Column: 2,
						Byte:   117,
					},
				},
			},
			nil,
		},
		{
			"list of object expressions element",
			map[string]*schema.AttributeSchema{
				"objects": {Expr: schema.ExprConstraints{
					schema.ListExpr{
						Elem: schema.ExprConstraints{
							schema.ObjectExpr{
								Attributes: schema.ObjectExprAttributes{
									"source": &schema.AttributeSchema{
										Expr: schema.LiteralTypeOnly(cty.String),
									},
									"bool": &schema.AttributeSchema{
										Expr: schema.LiteralTypeOnly(cty.Bool),
									},
									"notbool": &schema.AttributeSchema{
										Expr: schema.LiteralTypeOnly(cty.String),
									},
									"nested_map": &schema.AttributeSchema{
										Expr: schema.LiteralTypeOnly(cty.Map(cty.String)),
									},
									"nested_obj": &schema.AttributeSchema{
										Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{})),
									},
								},
							},
						},
					},
				}},
			},
			`objects = [
    {
        source = "blah"
        different = 42
        bool = true
        notbool = "test"
    }
]`,
			hcl.Pos{Line: 3, Column: 12, Byte: 29},
			&lang.HoverData{
				Content: lang.Markdown(`**source** _string_`),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   3,
						Column: 9,
						Byte:   26,
					},
					End: hcl.Pos{
						Line:   3,
						Column: 24,
						Byte:   41,
					},
				},
			},
			nil,
		},
		{
			"nested object expression",
			map[string]*schema.AttributeSchema{
				"object": {Expr: schema.ExprConstraints{
					schema.ObjectExpr{
						Attributes: schema.ObjectExprAttributes{
							"nested": &schema.AttributeSchema{
								Expr: schema.ExprConstraints{
									schema.ObjectExpr{
										Attributes: schema.ObjectExprAttributes{
											"source": &schema.AttributeSchema{
												Expr: schema.LiteralTypeOnly(cty.String),
											},
											"bool": &schema.AttributeSchema{
												Expr: schema.LiteralTypeOnly(cty.Bool),
											},
											"notbool": &schema.AttributeSchema{
												Expr: schema.LiteralTypeOnly(cty.String),
											},
											"nested_map": &schema.AttributeSchema{
												Expr: schema.LiteralTypeOnly(cty.Map(cty.String)),
											},
											"nested_obj": &schema.AttributeSchema{
												Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{})),
											},
										},
									},
								},
							},
						},
					},
				}},
			},
			`object = {
    nested = {
        source = "blah"
        different = 42
        bool = true
        notbool = "test"
    }
}`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			&lang.HoverData{
				Content: lang.Markdown("```" + `
{
  nested = {
    bool = bool
    nested_map = map of string
    nested_obj = object
    notbool = string
    source = string
  }
}
` + "```\n_object_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 10,
						Byte:   9,
					},
					End: hcl.Pos{
						Line:   8,
						Column: 2,
						Byte:   125,
					},
				},
			},
			nil,
		},
		{
			"nested object expression inside",
			map[string]*schema.AttributeSchema{
				"object": {Expr: schema.ExprConstraints{
					schema.ObjectExpr{
						Attributes: schema.ObjectExprAttributes{
							"nested": &schema.AttributeSchema{
								Expr: schema.ExprConstraints{
									schema.ObjectExpr{
										Attributes: schema.ObjectExprAttributes{
											"source": &schema.AttributeSchema{
												Expr: schema.LiteralTypeOnly(cty.String),
											},
											"bool": &schema.AttributeSchema{
												Expr: schema.LiteralTypeOnly(cty.Bool),
											},
											"notbool": &schema.AttributeSchema{
												Expr: schema.LiteralTypeOnly(cty.String),
											},
											"nested_map": &schema.AttributeSchema{
												Expr: schema.LiteralTypeOnly(cty.Map(cty.String)),
											},
											"nested_obj": &schema.AttributeSchema{
												Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{})),
											},
										},
									},
								},
							},
						},
					},
				}},
			},
			`object = {
    nested = {
        source = "blah"
        different = 42
        bool = true
        notbool = "test"
    }
}`,
			hcl.Pos{Line: 3, Column: 12, Byte: 37},
			&lang.HoverData{
				Content: lang.Markdown(`**source** _string_`),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   3,
						Column: 9,
						Byte:   34,
					},
					End: hcl.Pos{
						Line:   3,
						Column: 24,
						Byte:   49,
					},
				},
			},
			nil,
		},
		{
			"object as expression",
			map[string]*schema.AttributeSchema{
				"obj": {Expr: schema.ExprConstraints{
					schema.ObjectExpr{
						Attributes: schema.ObjectExprAttributes{
							"source": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.String),
							},
							"bool": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.Bool),
							},
							"notbool": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.String),
							},
							"nested_map": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.Map(cty.String)),
							},
							"nested_obj": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{})),
							},
						},
					},
				}},
			},
			`obj = {
    source = "blah"
    different = 42
    bool = true
    notbool = "test"
}`,
			hcl.Pos{Line: 1, Column: 3, Byte: 2},
			&lang.HoverData{
				Content: lang.Markdown(`**obj** _object_`),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 1,
						Byte:   0,
					},
					End: hcl.Pos{
						Line:   6,
						Column: 2,
						Byte:   85,
					},
				},
			},
			nil,
		},
		{
			"object as expression with unknown key",
			map[string]*schema.AttributeSchema{
				"obj": {Expr: schema.ExprConstraints{
					schema.ObjectExpr{
						Attributes: schema.ObjectExprAttributes{
							"source": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.String),
							},
							"bool": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.Bool),
							},
							"notbool": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.String),
							},
							"nested_map": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.Map(cty.String)),
							},
							"nested_obj": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{})),
							},
						},
					},
				}},
			},
			`obj = {
    var.src = "blah"
    "${var.env}.${another}" = "prod"
    different = 42
    bool = true
    notbool = "test"
}`,
			hcl.Pos{Line: 1, Column: 3, Byte: 2},
			&lang.HoverData{
				Content: lang.Markdown(`**obj** _object_`),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 1,
						Byte:   0,
					},
					End: hcl.Pos{
						Line:   7,
						Column: 2,
						Byte:   123,
					},
				},
			},
			nil,
		},
		{
			"object as expression - expression",
			map[string]*schema.AttributeSchema{
				"obj": {Expr: schema.ExprConstraints{
					schema.ObjectExpr{
						Attributes: schema.ObjectExprAttributes{
							"source": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.String),
							},
							"bool": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.Bool),
							},
							"notbool": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.String),
							},
							"nested_map": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.Map(cty.String)),
							},
							"nested_obj": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{})),
							},
						},
					},
				}},
			},
			`obj = {
    source = "blah"
    different = 42
    bool = true
    notbool = "test"
}`,
			hcl.Pos{Line: 2, Column: 2, Byte: 9},
			&lang.HoverData{
				Content: lang.Markdown("```" + `
{
  bool = bool
  nested_map = map of string
  nested_obj = object
  notbool = string
  source = string
}
` + "```" + `
_object_`),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 7,
						Byte:   6,
					},
					End: hcl.Pos{
						Line:   6,
						Column: 2,
						Byte:   85,
					},
				},
			},
			nil,
		},
		{
			"object as expression - attribute",
			map[string]*schema.AttributeSchema{
				"obj": {Expr: schema.ExprConstraints{
					schema.ObjectExpr{
						Attributes: schema.ObjectExprAttributes{
							"source": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.String),
							},
							"bool": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.Bool),
							},
							"notbool": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.String),
							},
							"nested_map": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.Map(cty.String)),
							},
							"nested_obj": &schema.AttributeSchema{
								Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{})),
							},
						},
					},
				}},
			},
			`obj = {
    source = "blah"
    different = 42
    bool = true
    notbool = "test"
}`,
			hcl.Pos{Line: 2, Column: 8, Byte: 15},
			&lang.HoverData{
				Content: lang.Markdown(`**source** _string_`),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   2,
						Column: 5,
						Byte:   12,
					},
					End: hcl.Pos{
						Line:   2,
						Column: 20,
						Byte:   27,
					},
				},
			},
			nil,
		},
		{
			"map as type",
			map[string]*schema.AttributeSchema{
				"nummap": {Expr: schema.LiteralTypeOnly(cty.Map(cty.Number))},
			},
			`nummap = {
  first = 12
  second = 24
}`,
			hcl.Pos{Line: 2, Column: 9, Byte: 19},
			&lang.HoverData{
				Content: lang.Markdown("_map of number_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 10,
						Byte:   9,
					},
					End: hcl.Pos{
						Line:   4,
						Column: 2,
						Byte:   39,
					},
				},
			},
			nil,
		},
		{
			"map as value",
			map[string]*schema.AttributeSchema{
				"nummap": {Expr: schema.ExprConstraints{
					schema.LiteralValue{
						Val: cty.MapVal(map[string]cty.Value{
							"first":  cty.NumberIntVal(12),
							"second": cty.NumberIntVal(24),
						}),
					},
				}},
			},
			`nummap = {
  "first" = 12
  "second" = 24
}`,
			hcl.Pos{Line: 2, Column: 9, Byte: 19},
			&lang.HoverData{
				Content: lang.Markdown("```\n" +
					"{\n  \"first\" = 12\n  \"second\" = 24\n}\n```\n_map of number_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 10,
						Byte:   9,
					},
					End: hcl.Pos{
						Line:   4,
						Column: 2,
						Byte:   43,
					},
				},
			},
			nil,
		},
		{
			"list as type",
			map[string]*schema.AttributeSchema{
				"mylist": {Expr: schema.LiteralTypeOnly(cty.List(cty.String))},
			},
			`mylist = [ "one", "two" ]`,
			hcl.Pos{Line: 1, Column: 16, Byte: 15},
			&lang.HoverData{
				Content: lang.Markdown("_list of string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 10,
						Byte:   9,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 26,
						Byte:   25,
					},
				},
			},
			nil,
		},
		{
			"set as type",
			map[string]*schema.AttributeSchema{
				"myset": {Expr: schema.LiteralTypeOnly(cty.Set(cty.String))},
			},
			`myset = [ "aa", "bb", "cc" ]`,
			hcl.Pos{Line: 1, Column: 16, Byte: 15},
			&lang.HoverData{
				Content: lang.Markdown("_set of string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 9,
						Byte:   8,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 29,
						Byte:   28,
					},
				},
			},
			nil,
		},
		{
			"matching keyword",
			map[string]*schema.AttributeSchema{
				"keyword": {Expr: schema.ExprConstraints{
					schema.KeywordExpr{
						Keyword: "foobar",
					},
				}},
			},
			`keyword = foobar`,
			hcl.Pos{Line: 1, Column: 14, Byte: 13},
			&lang.HoverData{
				Content: lang.Markdown("`foobar` _keyword_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 11,
						Byte:   10,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 17,
						Byte:   16,
					},
				},
			},
			nil,
		},
		{
			"map expression",
			map[string]*schema.AttributeSchema{
				"mapexpr": {Expr: schema.ExprConstraints{
					schema.MapExpr{
						Name: "special map",
						Elem: schema.LiteralTypeOnly(cty.String),
					},
				}},
			},
			`mapexpr = {
  key1 = "val1"
  key2 = "val2"
}`,
			hcl.Pos{Line: 2, Column: 8, Byte: 19},
			&lang.HoverData{
				Content: lang.Markdown("_special map_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 11,
						Byte:   10,
					},
					End: hcl.Pos{
						Line:   4,
						Column: 2,
						Byte:   45,
					},
				},
			},
			nil,
		},
		{
			"tuple constant expression",
			map[string]*schema.AttributeSchema{
				"tuplecons": {Expr: schema.ExprConstraints{
					schema.TupleConsExpr{
						Name:    "special tuple",
						AnyElem: schema.LiteralTypeOnly(cty.String),
					},
				}},
			},
			`tuplecons = [ "one", "two" ]`,
			hcl.Pos{Line: 1, Column: 18, Byte: 17},
			&lang.HoverData{
				Content: lang.Markdown("_special tuple_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 13,
						Byte:   12,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 29,
						Byte:   28,
					},
				},
			},
			nil,
		},
		{
			"list expression",
			map[string]*schema.AttributeSchema{
				"list": {Expr: schema.ExprConstraints{
					schema.ListExpr{
						Description: lang.Markdown("Special list"),
						Elem:        schema.LiteralTypeOnly(cty.String),
					},
				}},
			},
			`list = [ "one", "two" ]`,
			hcl.Pos{Line: 1, Column: 8, Byte: 7},
			&lang.HoverData{
				Content: lang.Markdown("_list_\n\nSpecial list"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 8,
						Byte:   7,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 24,
						Byte:   23,
					},
				},
			},
			nil,
		},
		{
			"list expression element",
			map[string]*schema.AttributeSchema{
				"list": {Expr: schema.ExprConstraints{
					schema.ListExpr{
						Elem: schema.LiteralTypeOnly(cty.String),
					},
				}},
			},
			`list = [ "one", "two" ]`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			&lang.HoverData{
				Content: lang.Markdown("`\"one\"` _string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 10,
						Byte:   9,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 15,
						Byte:   14,
					},
				},
			},
			nil,
		},
		{
			"set expression",
			map[string]*schema.AttributeSchema{
				"set": {Expr: schema.ExprConstraints{
					schema.SetExpr{
						Description: lang.Markdown("Special set"),
						Elem:        schema.LiteralTypeOnly(cty.String),
					},
				}},
			},
			`set = [ "one", "two" ]`,
			hcl.Pos{Line: 1, Column: 7, Byte: 6},
			&lang.HoverData{
				Content: lang.Markdown("_set_\n\nSpecial set"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 7,
						Byte:   6,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 23,
						Byte:   22,
					},
				},
			},
			nil,
		},
		{
			"set expression element",
			map[string]*schema.AttributeSchema{
				"set": {Expr: schema.ExprConstraints{
					schema.SetExpr{
						Elem: schema.LiteralTypeOnly(cty.String),
					},
				}},
			},
			`set = [ "one", "two" ]`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			&lang.HoverData{
				Content: lang.Markdown("`\"one\"` _string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 9,
						Byte:   8,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 14,
						Byte:   13,
					},
				},
			},
			nil,
		},
		{
			"tuple expression",
			map[string]*schema.AttributeSchema{
				"tup": {Expr: schema.ExprConstraints{
					schema.TupleExpr{
						Description: lang.Markdown("Special tuple"),
						Elems: []schema.ExprConstraints{
							schema.LiteralTypeOnly(cty.String),
						},
					},
				}},
			},
			`tup = [ "one", "two" ]`,
			hcl.Pos{Line: 1, Column: 7, Byte: 6},
			&lang.HoverData{
				Content: lang.Markdown("_tuple_\n\nSpecial tuple"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 7,
						Byte:   6,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 23,
						Byte:   22,
					},
				},
			},
			nil,
		},
		{
			"tuple expression element",
			map[string]*schema.AttributeSchema{
				"tup": {Expr: schema.ExprConstraints{
					schema.TupleExpr{
						Elems: []schema.ExprConstraints{
							schema.LiteralTypeOnly(cty.String),
						},
					},
				}},
			},
			`tup = [ "one", "two" ]`,
			hcl.Pos{Line: 1, Column: 12, Byte: 11},
			&lang.HoverData{
				Content: lang.Markdown("`\"one\"` _string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 9,
						Byte:   8,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 14,
						Byte:   13,
					},
				},
			},
			nil,
		},
		{
			"object as value",
			map[string]*schema.AttributeSchema{
				"objval": {Expr: schema.ExprConstraints{
					schema.LiteralValue{
						Val: cty.ObjectVal(map[string]cty.Value{
							"source": cty.StringVal("blah"),
							"bool":   cty.True,
							"nested_obj": cty.ObjectVal(map[string]cty.Value{
								"greetings": cty.StringVal("hello"),
							}),
							"nested_tuple": cty.TupleVal([]cty.Value{
								cty.NumberIntVal(42),
							}),
						}),
					},
				}},
			},
			`objval = {
  source = "blah"
  bool = true
  nested_obj = {
    greetings = "hello"
  }
  nested_tuple = [ 42 ]
}`,
			hcl.Pos{Line: 3, Column: 8, Byte: 36},
			&lang.HoverData{
				Content: lang.Markdown("```\n" +
					`{
  bool = true
  nested_obj = {
    greetings = "hello"
  }
  nested_tuple = [
    42,
  ]
  source = "blah"
}` +
					"\n```\n_object_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 10,
						Byte:   9,
					},
					End: hcl.Pos{
						Line:   8,
						Column: 2,
						Byte:   113,
					},
				},
			},
			nil,
		},
		{
			"list as value",
			map[string]*schema.AttributeSchema{
				"listval": {Expr: schema.ExprConstraints{
					schema.LiteralValue{
						Val: cty.ListVal([]cty.Value{
							cty.StringVal("first"),
							cty.StringVal("second"),
						}),
					},
				}},
			},
			`listval = [ "first", "second" ]`,
			hcl.Pos{Line: 1, Column: 16, Byte: 15},
			&lang.HoverData{
				Content: lang.Markdown("```\n[\n  \"first\",\n  \"second\",\n]\n```\n_list of string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 11,
						Byte:   10,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 32,
						Byte:   31,
					},
				},
			},
			nil,
		},
		{
			"set as value",
			map[string]*schema.AttributeSchema{
				"setval": {Expr: schema.ExprConstraints{
					schema.LiteralValue{
						Val: cty.SetVal([]cty.Value{
							cty.StringVal("west"),
							cty.StringVal("east"),
						}),
					},
				}},
			},
			`setval = [ "west", "east" ]`,
			hcl.Pos{Line: 1, Column: 16, Byte: 15},
			&lang.HoverData{
				Content: lang.Markdown("```\n[\n  \"east\",\n  \"west\",\n]\n```\n_set of string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 10,
						Byte:   9,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 28,
						Byte:   27,
					},
				},
			},
			nil,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			d := NewDecoder()
			d.SetSchema(&schema.BodySchema{
				Attributes: tc.attrSchema,
			})

			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			err := d.LoadFile("test.tf", f)
			if err != nil {
				t.Fatal(err)
			}

			data, err := d.HoverAtPos("test.tf", tc.pos)

			if err != nil {
				if tc.expectedErr != nil && !errors.As(err, &tc.expectedErr) {
					t.Fatalf("unexpected error: %s\nexpected: %s\n",
						err, tc.expectedErr)
				} else if tc.expectedErr == nil {
					t.Fatal(err)
				}
			} else if tc.expectedErr != nil {
				t.Fatalf("expected error: %s", tc.expectedErr)
			}

			if diff := cmp.Diff(tc.expectedData, data, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("hover data mismatch: %s", diff)
			}
		})
	}
}

func TestDecoder_HoverAtPos_traversalExpressions(t *testing.T) {
	testCases := []struct {
		name         string
		attrSchema   map[string]*schema.AttributeSchema
		refs         lang.References
		cfg          string
		pos          hcl.Pos
		expectedData *lang.HoverData
		expectedErr  error
	}{
		{
			"unknown traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.TraversalExpr{OfType: cty.String},
					},
				},
			},
			lang.References{},
			`attr = var.blah`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			nil,
			&NoReferenceFound{},
		},
		{
			"known mismatching traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.TraversalExpr{OfType: cty.String},
					},
				},
			},
			lang.References{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "blah"},
					},
					Type: cty.Bool,
				},
			},
			`attr = var.blah`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			nil,
			&NoReferenceFound{},
		},
		{
			"known type matching traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.TraversalExpr{OfType: cty.String},
					},
				},
			},
			lang.References{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "blah"},
					},
					Type: cty.String,
				},
			},
			`attr = var.blah`,
			hcl.Pos{Line: 1, Column: 10, Byte: 9},
			&lang.HoverData{
				Content: lang.Markdown("`var.blah`\n_string_"),
				Range: hcl.Range{
					Filename: "test.tf",
					Start: hcl.Pos{
						Line:   1,
						Column: 8,
						Byte:   7,
					},
					End: hcl.Pos{
						Line:   1,
						Column: 16,
						Byte:   15,
					},
				},
			},
			nil,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			d := NewDecoder()
			d.SetSchema(&schema.BodySchema{
				Attributes: tc.attrSchema,
			})
			d.SetReferenceReader(func() lang.References {
				return tc.refs
			})

			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			err := d.LoadFile("test.tf", f)
			if err != nil {
				t.Fatal(err)
			}

			data, err := d.HoverAtPos("test.tf", tc.pos)
			if err != nil {
				if tc.expectedErr != nil && !errors.As(err, &tc.expectedErr) {
					t.Fatalf("unexpected error: %s\nexpected: %s\n",
						err, tc.expectedErr)
				} else if tc.expectedErr == nil {
					t.Fatal(err)
				}
			} else if tc.expectedErr != nil {
				t.Fatalf("expected error: %s", tc.expectedErr)
			}

			if diff := cmp.Diff(tc.expectedData, data, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("hover data mismatch: %s", diff)
			}
		})
	}
}
