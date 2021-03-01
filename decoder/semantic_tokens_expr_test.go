package decoder

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestDecoder_SemanticTokensInFile_expressions(t *testing.T) {
	testCases := []struct {
		name           string
		attrSchema     map[string]*schema.AttributeSchema
		cfg            string
		expectedTokens []lang.SemanticToken
	}{
		{
			"string as known value",
			map[string]*schema.AttributeSchema{
				"str": {
					Expr: schema.ExprConstraints{
						schema.LiteralValue{
							Val: cty.StringVal("blablah"),
						},
					},
				},
			},
			`str = "blablah"`,
			[]lang.SemanticToken{
				{ // str
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 4,
							Byte:   3,
						},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 7,
							Byte:   6,
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
			"heredoc string as known type",
			map[string]*schema.AttributeSchema{
				"str": {
					Expr: schema.LiteralTypeOnly(cty.String),
				},
			},
			`str = <<EOT
blablah
EOT
`,
			[]lang.SemanticToken{
				{ // str
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 4,
							Byte:   3,
						},
					},
				},
				{
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 7,
							Byte:   6,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 4,
							Byte:   23,
						},
					},
				},
			},
		},
		{
			"string as unknown value",
			map[string]*schema.AttributeSchema{
				"str": {
					Expr: schema.ExprConstraints{
						schema.LiteralValue{
							Val: cty.StringVal("blablah"),
						},
					},
				},
			},
			`str = "blablax"`,
			[]lang.SemanticToken{
				{ // str
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 4,
							Byte:   3,
						},
					},
				},
			},
		},
		{
			"object as value",
			map[string]*schema.AttributeSchema{
				"obj": {
					Expr: schema.ExprConstraints{
						schema.LiteralValue{
							Val: cty.ObjectVal(map[string]cty.Value{
								"first":  cty.NumberIntVal(42),
								"second": cty.StringVal("boo"),
							}),
						},
					},
				},
			},
			`obj = {
  first = 42
  second = "boo"
}`,
			[]lang.SemanticToken{
				{ // obj
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 4,
							Byte:   3,
						},
					},
				},
				{ // first
					Type:      lang.TokenObjectKey,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   10,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 8,
							Byte:   15,
						},
					},
				},
				{ // 42
					Type:      lang.TokenNumber,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 11,
							Byte:   18,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 13,
							Byte:   20,
						},
					},
				},
				{ // second
					Type:      lang.TokenObjectKey,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   3,
							Column: 3,
							Byte:   23,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 9,
							Byte:   29,
						},
					},
				},
				{ // "boo"
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   3,
							Column: 12,
							Byte:   32,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 17,
							Byte:   37,
						},
					},
				},
			},
		},
		{
			"object as mismatching value",
			map[string]*schema.AttributeSchema{
				"obj": {
					Expr: schema.ExprConstraints{
						schema.LiteralValue{
							Val: cty.ObjectVal(map[string]cty.Value{
								"knownkey": cty.NumberIntVal(43),
							}),
						},
					},
				},
			},
			`obj = {
  knownkey = 43
  unknownkey = "boo"
}`,
			[]lang.SemanticToken{
				{ // obj
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 4,
							Byte:   3,
						},
					},
				},
			},
		},

		{
			"object as type",
			map[string]*schema.AttributeSchema{
				"obj": {
					Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{
						"knownkey": cty.Number,
					})),
				},
			},
			`obj = {
  knownkey = 42
  unknownkey = "boo"
}`,
			[]lang.SemanticToken{
				{ // obj
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 4,
							Byte:   3,
						},
					},
				},
				{ // knownkey
					Type:      lang.TokenObjectKey,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   10,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 11,
							Byte:   18,
						},
					},
				},
				{ // 42
					Type:      lang.TokenNumber,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 14,
							Byte:   21,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 16,
							Byte:   23,
						},
					},
				},
			},
		},
		{
			"object as type with unknown key",
			map[string]*schema.AttributeSchema{
				"obj": {
					Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{
						"knownkey": cty.Number,
					})),
				},
			},
			`obj = {
  knownkey = 42
  "${var.env}.${another}" = "prod"
  var.test = "boo"
}`,
			[]lang.SemanticToken{
				{ // obj
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 4,
							Byte:   3,
						},
					},
				},
				{ // knownkey
					Type:      lang.TokenObjectKey,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   10,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 11,
							Byte:   18,
						},
					},
				},
				{ // 42
					Type:      lang.TokenNumber,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 14,
							Byte:   21,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 16,
							Byte:   23,
						},
					},
				},
			},
		},
		{
			"object as expression",
			map[string]*schema.AttributeSchema{
				"obj": {
					Expr: schema.ExprConstraints{
						schema.ObjectExpr{
							Attributes: schema.ObjectExprAttributes{
								"knownkey": {
									Expr: schema.LiteralTypeOnly(cty.Number),
								},
							},
						},
					},
				},
			},
			`obj = {
  knownkey = 42
  unknownkey = "boo"
}`,
			[]lang.SemanticToken{
				{ // obj
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 4,
							Byte:   3,
						},
					},
				},
				{ // knownkey
					Type:      lang.TokenObjectKey,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   10,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 11,
							Byte:   18,
						},
					},
				},
				{ // 42
					Type:      lang.TokenNumber,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 14,
							Byte:   21,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 16,
							Byte:   23,
						},
					},
				},
			},
		},
		{
			"object as expression with unknown key",
			map[string]*schema.AttributeSchema{
				"obj": {
					Expr: schema.ExprConstraints{
						schema.ObjectExpr{
							Attributes: schema.ObjectExprAttributes{
								"knownkey": {
									Expr: schema.LiteralTypeOnly(cty.Number),
								},
							},
						},
					},
				},
			},
			`obj = {
  knownkey = 42
  var.test = 32
  "${var.env}.${another}" = "prod"
}`,
			[]lang.SemanticToken{
				{ // obj
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 4,
							Byte:   3,
						},
					},
				},
				{ // knownkey
					Type:      lang.TokenObjectKey,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   10,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 11,
							Byte:   18,
						},
					},
				},
				{ // 42
					Type:      lang.TokenNumber,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 14,
							Byte:   21,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 16,
							Byte:   23,
						},
					},
				},
			},
		},
		{
			"map literal keys",
			map[string]*schema.AttributeSchema{
				"mapkey": {
					Expr: schema.LiteralTypeOnly(cty.Map(cty.String)),
				},
			},
			`mapkey = {
  bla = "blablah"
  nada = "yada"
  wrong = 42
}
`,
			[]lang.SemanticToken{
				{ // mapkey
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 7,
							Byte:   6,
						},
					},
				},
				{ // bla
					Type:      lang.TokenMapKey,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   13,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 6,
							Byte:   16,
						},
					},
				},
				{ // "blablah"
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 9,
							Byte:   19,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 18,
							Byte:   28,
						},
					},
				},
				{ // nada
					Type:      lang.TokenMapKey,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
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
				{ // "yada"
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   3,
							Column: 10,
							Byte:   38,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 16,
							Byte:   44,
						},
					},
				},
				{ // wrong
					Type:      lang.TokenMapKey,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   4,
							Column: 3,
							Byte:   47,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 8,
							Byte:   52,
						},
					},
				},
			},
		},
		{
			"map expression",
			map[string]*schema.AttributeSchema{
				"mapkey": {
					Expr: schema.ExprConstraints{
						schema.MapExpr{
							Name: "special map",
							Elem: schema.LiteralTypeOnly(cty.String),
						},
					},
				},
			},
			`mapkey = {
  bla = "blablah"
  nada = "yada"
  wrong = 42
}
`,
			[]lang.SemanticToken{
				{ // mapkey
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 7,
							Byte:   6,
						},
					},
				},
				{ // bla
					Type:      lang.TokenMapKey,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 3,
							Byte:   13,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 6,
							Byte:   16,
						},
					},
				},
				{ // "blablah"
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 9,
							Byte:   19,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 18,
							Byte:   28,
						},
					},
				},
				{ // nada
					Type:      lang.TokenMapKey,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
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
				{ // "yada"
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   3,
							Column: 10,
							Byte:   38,
						},
						End: hcl.Pos{
							Line:   3,
							Column: 16,
							Byte:   44,
						},
					},
				},
				{ // wrong
					Type:      lang.TokenMapKey,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   4,
							Column: 3,
							Byte:   47,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 8,
							Byte:   52,
						},
					},
				},
			},
		},
		{
			"known keyword",
			map[string]*schema.AttributeSchema{
				"keyword": {
					Expr: schema.ExprConstraints{
						schema.KeywordExpr{
							Keyword: "foobar",
							Name:    "special type",
						},
					},
				},
			},
			`keyword = foobar
`,
			[]lang.SemanticToken{
				{ // keyword
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 8,
							Byte:   7,
						},
					},
				},
				{ // foobar
					Type:      lang.TokenKeyword,
					Modifiers: []lang.SemanticTokenModifier{},
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
			},
		},
		{
			"unknown keyword",
			map[string]*schema.AttributeSchema{
				"keyword": {
					Expr: schema.ExprConstraints{
						schema.KeywordExpr{
							Keyword: "foobar",
							Name:    "special type",
						},
					},
				},
			},
			`keyword = abcxyz
`,
			[]lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
							Byte:   0,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 8,
							Byte:   7,
						},
					},
				},
			},
		},
		{
			"tuple constant expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.TupleConsExpr{
							AnyElem: schema.LiteralTypeOnly(cty.String),
						},
					},
				},
			},
			`attr = [ "one", 42, "two" ]
`,
			[]lang.SemanticToken{
				{ // attr
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
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
				{ // "one"
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
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
				{ // "two"
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 21,
							Byte:   20,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 26,
							Byte:   25,
						},
					},
				},
			},
		},
		{
			"list expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.ListExpr{
							Elem: schema.LiteralTypeOnly(cty.String),
						},
					},
				},
			},
			`attr = [ "one", 42, "two" ]
`,
			[]lang.SemanticToken{
				{ // attr
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
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
				{ // "one"
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
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
				{ // "two"
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 21,
							Byte:   20,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 26,
							Byte:   25,
						},
					},
				},
			},
		},
		{
			"set expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.SetExpr{
							Elem: schema.LiteralTypeOnly(cty.String),
						},
					},
				},
			},
			`attr = [ "one", 42, "two" ]
`,
			[]lang.SemanticToken{
				{ // attr
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
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
				{ // "one"
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
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
				{ // "two"
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 21,
							Byte:   20,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 26,
							Byte:   25,
						},
					},
				},
			},
		},
		{
			"tuple expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.TupleExpr{
							Elems: []schema.ExprConstraints{
								schema.LiteralTypeOnly(cty.String),
								schema.LiteralTypeOnly(cty.Number),
								schema.LiteralTypeOnly(cty.Bool),
							},
						},
					},
				},
			},
			`attr = [ "one", 42, "two" ]
`,
			[]lang.SemanticToken{
				{ // attr
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
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
				{ // "one"
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
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
				{ // 42
					Type:      lang.TokenNumber,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 17,
							Byte:   16,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 19,
							Byte:   18,
						},
					},
				},
			},
		},
		{
			"tuple as list",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.LiteralTypeOnly(cty.List(cty.String)),
				},
			},
			`attr = [ "one", 42, "two" ]
`,
			[]lang.SemanticToken{
				{ // attr
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
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
				{ // "one"
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
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
				{ // "two"
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 21,
							Byte:   20,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 26,
							Byte:   25,
						},
					},
				},
			},
		},
		{
			"list as value",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.LiteralValue{
							Val: cty.ListVal([]cty.Value{
								cty.StringVal("one"),
								cty.StringVal("two"),
							}),
						},
					},
				},
			},
			`attr = [ "one", "two" ]
`,
			[]lang.SemanticToken{
				{ // attr
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
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
				{ // "one"
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
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
				{ // "two"
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 17,
							Byte:   16,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 22,
							Byte:   21,
						},
					},
				},
			},
		},
		{
			"tuple as set",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.LiteralTypeOnly(cty.Set(cty.String)),
				},
			},
			`attr = [ "one", 42, "two" ]
`,
			[]lang.SemanticToken{
				{ // attr
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
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
				{ // "one"
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
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
				{ // "two"
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 21,
							Byte:   20,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 26,
							Byte:   25,
						},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			d := NewDecoder()
			d.SetSchema(&schema.BodySchema{
				Attributes: tc.attrSchema,
			})

			f, pDiags := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			if len(pDiags) > 0 {
				t.Fatal(pDiags)
			}
			err := d.LoadFile("test.tf", f)
			if err != nil {
				t.Fatal(err)
			}

			tokens, err := d.SemanticTokensInFile("test.tf")
			if err != nil {
				t.Fatal(err)
			}

			diff := cmp.Diff(tc.expectedTokens, tokens)
			if diff != "" {
				t.Fatalf("unexpected tokens: %s", diff)
			}
		})
	}
}

func TestDecoder_SemanticTokensInFile_traversalExpression(t *testing.T) {
	testCases := []struct {
		name           string
		attrSchema     map[string]*schema.AttributeSchema
		refs           lang.References
		cfg            string
		expectedTokens []lang.SemanticToken
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
			`attr = var.blah
`,
			[]lang.SemanticToken{
				{ // attr
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
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
			"known mismatching traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.TraversalExpr{OfType: cty.String},
					},
				},
			},
			lang.References{
				lang.Reference{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "blah"},
					},
					Type: cty.Bool,
				},
			},
			`attr = var.blah
`,
			[]lang.SemanticToken{
				{ // attr
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
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
			"known matching traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.TraversalExpr{OfType: cty.String},
					},
				},
			},
			lang.References{
				lang.Reference{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "blah"},
					},
					Type: cty.String,
				},
			},
			`attr = var.blah
`,
			[]lang.SemanticToken{
				{ // attr
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
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
				{ // var
					Type:      lang.TokenTraversalStep,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 8,
							Byte:   7,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 11,
							Byte:   10,
						},
					},
				},
				{ // blah
					Type:      lang.TokenTraversalStep,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 12,
							Byte:   11,
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

			f, pDiags := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			if len(pDiags) > 0 {
				t.Fatal(pDiags)
			}
			err := d.LoadFile("test.tf", f)
			if err != nil {
				t.Fatal(err)
			}

			tokens, err := d.SemanticTokensInFile("test.tf")
			if err != nil {
				t.Fatal(err)
			}

			diff := cmp.Diff(tc.expectedTokens, tokens)
			if diff != "" {
				t.Fatalf("unexpected tokens: %s", diff)
			}
		})
	}
}
