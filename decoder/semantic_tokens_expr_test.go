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
			"undefined tuple expression",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.TupleExpr{
							Elems: []schema.ExprConstraints{},
						},
					},
				},
			},
			`attr = [ "one" ]
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
			"undefined tuple type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.LiteralTypeOnly(cty.Tuple([]cty.Type{})),
				},
			},
			`attr = [ "one" ]
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
			bodySchema := &schema.BodySchema{
				Attributes: tc.attrSchema,
			}

			f, pDiags := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			if len(pDiags) > 0 {
				t.Fatal(pDiags)
			}

			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			})

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
		refs           reference.Targets
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
			reference.Targets{},
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
			reference.Targets{
				reference.Target{
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
			reference.Targets{
				reference.Target{
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
		{
			"known scope matching traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.TraversalExpr{OfScopeId: lang.ScopeId("foo")},
					},
				},
			},
			reference.Targets{
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "blah"},
					},
					ScopeId: lang.ScopeId("foo"),
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
		{
			"matching traversal with indexes",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.TraversalExpr{},
					},
				},
			},
			reference.Targets{
				reference.Target{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
						lang.IndexStep{Key: cty.StringVal("test")},
						lang.AttrStep{Name: "bar"},
						lang.IndexStep{Key: cty.NumberIntVal(4)},
					},
				},
			},
			`attr = var.foo["test"].bar[4]
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
				{ // foo
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
							Column: 15,
							Byte:   14,
						},
					},
				},
				{ // "test"
					Type:      lang.TokenMapKey,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 16,
							Byte:   15,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 22,
							Byte:   21,
						},
					},
				},
				{ // bar
					Type:      lang.TokenTraversalStep,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 24,
							Byte:   23,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 27,
							Byte:   26,
						},
					},
				},
				{ // 4
					Type:      lang.TokenNumber,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 28,
							Byte:   27,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 29,
							Byte:   28,
						},
					},
				},
			},
		},
		{
			"loosely matching traversal of unknown type",
			map[string]*schema.AttributeSchema{
				"attr": {
					Expr: schema.ExprConstraints{
						schema.TraversalExpr{OfType: cty.String},
					},
				},
			},
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "var"},
						lang.AttrStep{Name: "foo"},
					},
					Type: cty.DynamicPseudoType,
				},
			},
			`attr = var.foo.bar
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
				{ // foo
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
							Column: 15,
							Byte:   14,
						},
					},
				},
				{ // bar
					Type:      lang.TokenTraversalStep,
					Modifiers: []lang.SemanticTokenModifier{},
					Range: hcl.Range{
						Filename: "test.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 16,
							Byte:   15,
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
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			bodySchema := &schema.BodySchema{
				Attributes: tc.attrSchema,
			}

			f, pDiags := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			if len(pDiags) > 0 {
				t.Fatal(pDiags)
			}

			d := testPathDecoder(t, &PathContext{
				Schema:           bodySchema,
				ReferenceTargets: tc.refs,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			})

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
