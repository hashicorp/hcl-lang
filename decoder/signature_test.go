// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

func TestSignatureAtPos(t *testing.T) {
	testCases := []struct {
		testName          string
		functions         map[string]schema.FunctionSignature
		cfg               string
		pos               hcl.Pos
		expectedSignature *lang.FunctionSignature
	}{
		{
			"no function call expr",
			map[string]schema.FunctionSignature{},
			`x = "hello"`,
			hcl.Pos{Line: 1, Column: 3, Byte: 4},
			nil,
		},
		{
			"unknown function",
			map[string]schema.FunctionSignature{},
			`x = foo()`,
			hcl.Pos{Line: 1, Column: 7, Byte: 8},
			nil,
		},
		{
			"no description",
			map[string]schema.FunctionSignature{
				"foo": {
					ReturnType: cty.String,
				},
			},
			`x = foo()`,
			hcl.Pos{Line: 1, Column: 7, Byte: 8},
			&lang.FunctionSignature{
				Name:        "foo() string",
				Description: lang.Markdown(""),
			},
		},
		{
			"no parameter, no var param",
			map[string]schema.FunctionSignature{
				"foo": {
					ReturnType:  cty.String,
					Description: "`foo` description",
				},
			},
			`x = foo()`,
			hcl.Pos{Line: 1, Column: 7, Byte: 8},
			&lang.FunctionSignature{
				Name:        "foo() string",
				Description: lang.Markdown("`foo` description"),
			},
		},
		{
			"one parameter, pos not inside parenthesis",
			map[string]schema.FunctionSignature{
				"foo": {
					Params: []function.Parameter{
						{
							Name:        "input",
							Type:        cty.String,
							Description: "`input` description",
						},
					},
					ReturnType:  cty.String,
					Description: "`foo` description.",
				},
			},
			`x = foo()`,
			hcl.Pos{Line: 1, Column: 5, Byte: 6},
			nil,
		},
		{
			"one parameter, empty first parameter",
			map[string]schema.FunctionSignature{
				"foo": {
					Params: []function.Parameter{
						{
							Name:        "input",
							Type:        cty.String,
							Description: "`input` description",
						},
					},
					ReturnType:  cty.String,
					Description: "`foo` description",
				},
			},
			`x = foo()`,
			hcl.Pos{Line: 1, Column: 7, Byte: 8},
			&lang.FunctionSignature{
				Name:        "foo(input string) string",
				Description: lang.Markdown("`foo` description"),
				Parameters: []lang.FunctionParameter{
					{
						Name:        "input",
						Description: lang.Markdown("`input` description"),
					},
				},
				ActiveParameter: 0,
			},
		},
		{
			"two parameters, empty second parameter",
			map[string]schema.FunctionSignature{
				"foo": {
					Params: []function.Parameter{
						{
							Name:        "input",
							Type:        cty.String,
							Description: "`input` description",
						},
						{
							Name:        "input2",
							Type:        cty.Number,
							Description: "`input2` description",
						},
					},
					ReturnType:  cty.String,
					Description: "`foo` description",
				},
			},
			`x = foo("a", )`,
			hcl.Pos{Line: 1, Column: 12, Byte: 13},
			&lang.FunctionSignature{
				Name:        "foo(input string, input2 number) string",
				Description: lang.Markdown("`foo` description"),
				Parameters: []lang.FunctionParameter{
					{
						Name:        "input",
						Description: lang.Markdown("`input` description"),
					},
					{
						Name:        "input2",
						Description: lang.Markdown("`input2` description"),
					},
				},
				ActiveParameter: 1,
			},
		},
		{
			"one parameter, one variadic, empty variadic",
			map[string]schema.FunctionSignature{
				"foo": {
					Params: []function.Parameter{
						{
							Name:        "input",
							Type:        cty.String,
							Description: "`input` description",
						},
					},
					VarParam: &function.Parameter{
						Name:        "vinput",
						Type:        cty.Number,
						Description: "`vinput` description",
					},
					ReturnType:  cty.String,
					Description: "`foo` description",
				},
			},
			`x = foo("a", )`,
			hcl.Pos{Line: 1, Column: 12, Byte: 13},
			&lang.FunctionSignature{
				Name:        "foo(input string, …vinput number) string",
				Description: lang.Markdown("`foo` description"),
				Parameters: []lang.FunctionParameter{
					{
						Name:        "input",
						Description: lang.Markdown("`input` description"),
					},
					{
						Name:        "vinput",
						Description: lang.Markdown("`vinput` description"),
					},
				},
				ActiveParameter: 1,
			},
		},
		{
			"one parameter, one variadic, some variadic",
			map[string]schema.FunctionSignature{
				"foo": {
					Params: []function.Parameter{
						{
							Name:        "input",
							Type:        cty.String,
							Description: "`input` description",
						},
					},
					VarParam: &function.Parameter{
						Name:        "vinput",
						Type:        cty.Number,
						Description: "`vinput` description",
					},
					ReturnType:  cty.String,
					Description: "`foo` description",
				},
			},
			`x = foo("a", 1, 2, )`,
			hcl.Pos{Line: 1, Column: 18, Byte: 19},
			&lang.FunctionSignature{
				Name:        "foo(input string, …vinput number) string",
				Description: lang.Markdown("`foo` description"),
				Parameters: []lang.FunctionParameter{
					{
						Name:        "input",
						Description: lang.Markdown("`input` description"),
					},
					{
						Name:        "vinput",
						Description: lang.Markdown("`vinput` description"),
					},
				},
				ActiveParameter: 1,
			},
		},
		{
			"two complex parameters, empty second parameter",
			map[string]schema.FunctionSignature{
				"foo": {
					Params: []function.Parameter{
						{
							Name:        "input",
							Type:        cty.List(cty.String),
							Description: "`input` description",
						},
						{
							Name:        "input2",
							Type:        cty.List(cty.Number),
							Description: "`input2` description",
						},
					},
					ReturnType:  cty.Map(cty.Number),
					Description: "`foo` description",
				},
			},
			`x = foo(["a", "b"], )`,
			hcl.Pos{Line: 1, Column: 19, Byte: 20},
			&lang.FunctionSignature{
				Name:        "foo(input list of string, input2 list of number) map of number",
				Description: lang.Markdown("`foo` description"),
				Parameters: []lang.FunctionParameter{
					{
						Name:        "input",
						Description: lang.Markdown("`input` description"),
					},
					{
						Name:        "input2",
						Description: lang.Markdown("`input2` description"),
					},
				},
				ActiveParameter: 1,
			},
		},
		{
			"three parameters, filling middle parameter",
			map[string]schema.FunctionSignature{
				"foo": {
					Params: []function.Parameter{
						{
							Name:        "input",
							Type:        cty.String,
							Description: "`input` description",
						},
						{
							Name:        "input2",
							Type:        cty.Number,
							Description: "`input2` description",
						},
						{
							Name:        "input3",
							Type:        cty.Bool,
							Description: "`input3` description",
						},
					},
					ReturnType:  cty.String,
					Description: "`foo` description",
				},
			},
			`x = foo("a",  , false)`,
			hcl.Pos{Line: 1, Column: 12, Byte: 13},
			&lang.FunctionSignature{
				Name:        "foo(input string, input2 number, input3 bool) string",
				Description: lang.Markdown("`foo` description"),
				Parameters: []lang.FunctionParameter{
					{
						Name:        "input",
						Description: lang.Markdown("`input` description"),
					},
					{
						Name:        "input2",
						Description: lang.Markdown("`input2` description"),
					},
					{
						Name:        "input3",
						Description: lang.Markdown("`input3` description"),
					},
				},
				ActiveParameter: 1,
			},
		},
		{
			"two parameters, adding a third",
			map[string]schema.FunctionSignature{
				"foo": {
					Params: []function.Parameter{
						{
							Name:        "foo",
							Type:        cty.String,
							Description: "`foo` description",
						},
						{
							Name:        "foo2",
							Type:        cty.Number,
							Description: "`foo2` description",
						},
					},
					ReturnType:  cty.String,
					Description: "`foo` description",
				},
			},
			`x = foo("a", "b", )`,
			hcl.Pos{Line: 1, Column: 17, Byte: 18},
			nil,
		},
		{
			"no parameter, one variadic, some variadic",
			map[string]schema.FunctionSignature{
				"foo": {
					VarParam: &function.Parameter{
						Name:        "vinput",
						Type:        cty.Number,
						Description: "`vinput` description",
					},
					ReturnType:  cty.String,
					Description: "`foo` description",
				},
			},
			`x = foo(1, 2, 3, )`,
			hcl.Pos{Line: 1, Column: 16, Byte: 17},
			&lang.FunctionSignature{
				Name:        "foo(…vinput number) string",
				Description: lang.Markdown("`foo` description"),
				Parameters: []lang.FunctionParameter{
					{
						Name:        "vinput",
						Description: lang.Markdown("`vinput` description"),
					},
				},
				ActiveParameter: 0,
			},
		},
		{
			"no parameter, one variadic, second arg",
			map[string]schema.FunctionSignature{
				"concat": {
					VarParam: &function.Parameter{
						Name: "seqs",
						Type: cty.DynamicPseudoType,
					},
					ReturnType:  cty.DynamicPseudoType,
					Description: "`concat` description",
				},
			},
			`x = concat([],)`,
			hcl.Pos{Line: 1, Column: 13, Byte: 14},
			&lang.FunctionSignature{
				Name:        "concat(…seqs dynamic) dynamic",
				Description: lang.Markdown("`concat` description"),
				Parameters: []lang.FunctionParameter{
					{
						Name:        "seqs",
						Description: lang.Markdown(""),
					},
				},
				ActiveParameter: 0,
			},
		},
		{
			"multi-line two complex parameters, empty first parameter",
			map[string]schema.FunctionSignature{
				"foo": {
					Params: []function.Parameter{
						{
							Name:        "input",
							Type:        cty.List(cty.String),
							Description: "`input` description",
						},
						{
							Name:        "input2",
							Type:        cty.List(cty.Number),
							Description: "`input2` description",
						},
					},
					ReturnType:  cty.Map(cty.Number),
					Description: "`foo` description",
				},
			},
			`x = foo(
  
)`,
			hcl.Pos{Line: 2, Column: 1, Byte: 11},
			&lang.FunctionSignature{
				Name:        "foo(input list of string, input2 list of number) map of number",
				Description: lang.Markdown("`foo` description"),
				Parameters: []lang.FunctionParameter{
					{
						Name:        "input",
						Description: lang.Markdown("`input` description"),
					},
					{
						Name:        "input2",
						Description: lang.Markdown("`input2` description"),
					},
				},
				ActiveParameter: 0,
			},
		},
		{
			"multi-line two complex parameters, empty second parameter",
			map[string]schema.FunctionSignature{
				"foo": {
					Params: []function.Parameter{
						{
							Name:        "input",
							Type:        cty.List(cty.String),
							Description: "`input` description",
						},
						{
							Name:        "input2",
							Type:        cty.List(cty.Number),
							Description: "`input2` description",
						},
					},
					ReturnType:  cty.Map(cty.Number),
					Description: "`foo` description",
				},
			},
			`x = foo(
  ["a", "b"],
  
)`,
			hcl.Pos{Line: 3, Column: 1, Byte: 25},
			&lang.FunctionSignature{
				Name:        "foo(input list of string, input2 list of number) map of number",
				Description: lang.Markdown("`foo` description"),
				Parameters: []lang.FunctionParameter{
					{
						Name:        "input",
						Description: lang.Markdown("`input` description"),
					},
					{
						Name:        "input2",
						Description: lang.Markdown("`input2` description"),
					},
				},
				ActiveParameter: 1,
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%2d-%s", i, tc.testName), func(t *testing.T) {
			f, _ := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			d := testPathDecoder(t, &PathContext{
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
				Functions: tc.functions,
			})

			signature, err := d.SignatureAtPos("test.tf", tc.pos)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedSignature, signature); diff != "" {
				t.Fatalf("unexpected signature: %s", diff)
			}
		})
	}

}

func TestSignatureAtPos_json(t *testing.T) {
	f, pDiags := json.Parse([]byte(`{
		"attribute": "${abs(-1)}"
	}`), "test.tf.json")
	if len(pDiags) > 0 {
		t.Fatal(pDiags)
	}

	d := testPathDecoder(t, &PathContext{
		Files: map[string]*hcl.File{
			"test.tf.json": f,
		},
	})

	_, err := d.SignatureAtPos("test.tf.json", hcl.InitialPos)
	unknownFormatErr := &UnknownFileFormatError{}
	if !errors.As(err, &unknownFormatErr) {
		t.Fatal("expected UnknownFileFormatError for JSON body")
	}
}
