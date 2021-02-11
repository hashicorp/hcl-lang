package decoder

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/zclconf/go-cty/cty"
)

func TestSnippetForAttribute(t *testing.T) {
	testCases := []struct {
		testName        string
		attrName        string
		attrSchema      *schema.AttributeSchema
		expectedSnippet string
	}{
		{
			"primitive type",
			"primitive",
			&schema.AttributeSchema{
				Expr: schema.LiteralTypeOnly(cty.String),
			},
			`primitive = "${1:value}"`,
		},
		{
			"map of strings",
			"mymap",
			&schema.AttributeSchema{
				Expr: schema.LiteralTypeOnly(cty.Map(cty.String)),
			},
			`mymap = {
  "${1:key}" = "${2:value}"
}`,
		},
		{
			"map of numbers",
			"mymap",
			&schema.AttributeSchema{
				Expr: schema.LiteralTypeOnly(cty.Map(cty.Number)),
			},
			`mymap = {
  "${1:key}" = ${2:1}
}`,
		},
		{
			"list of numbers",
			"mylist",
			&schema.AttributeSchema{
				Expr: schema.LiteralTypeOnly(cty.List(cty.Number)),
			},
			`mylist = [ ${1:1} ]`,
		},
		{
			"list of objects",
			"mylistobj",
			&schema.AttributeSchema{
				Expr: schema.LiteralTypeOnly(cty.List(cty.Object(map[string]cty.Type{
					"first":  cty.String,
					"second": cty.Number,
				}))),
			},
			`mylistobj = [ {
  first = "${1:value}"
  second = ${2:1}
} ]`,
		},
		{
			"set of numbers",
			"myset",
			&schema.AttributeSchema{
				Expr: schema.LiteralTypeOnly(cty.Set(cty.Number)),
			},
			`myset = [ ${1:1} ]`,
		},
		{
			"object",
			"myobj",
			&schema.AttributeSchema{
				Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{
					"keystr":  cty.String,
					"keynum":  cty.Number,
					"keybool": cty.Bool,
				})),
			},
			`myobj = {
  keybool = ${1:false}
  keynum = ${2:1}
  keystr = "${3:value}"
}`,
		},
		{
			"unknown type",
			"mynil",
			&schema.AttributeSchema{
				Expr: schema.LiteralTypeOnly(cty.DynamicPseudoType),
			},
			`mynil = ${1}`,
		},
		{
			"nested object",
			"myobj",
			&schema.AttributeSchema{
				Expr: schema.LiteralTypeOnly(cty.Object(map[string]cty.Type{
					"keystr": cty.String,
					"another": cty.Object(map[string]cty.Type{
						"nestedstr":     cty.String,
						"nested_number": cty.Number,
					}),
				})),
			},
			`myobj = {
  another = {
    nested_number = ${1:1}
    nestedstr = "${2:value}"
  }
  keystr = "${3:value}"
}`,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			snippet := snippetForAttribute(tc.attrName, tc.attrSchema)
			if diff := cmp.Diff(tc.expectedSnippet, snippet); diff != "" {
				t.Fatalf("unexpected snippet: %s", diff)
			}
		})
	}
}
