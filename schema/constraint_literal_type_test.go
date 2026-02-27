// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty/cty"
)

func TestLiteralType_EmptyCompletionData(t *testing.T) {
	testCases := []struct {
		typ                   cty.Type
		prefillRequiredFields bool
		expectedData          CompletionData
	}{
		{
			cty.Bool,
			false,
			CompletionData{
				NewText:         "false",
				Snippet:         "${1:false}",
				NextPlaceholder: 2,
			},
		},
		{
			cty.List(cty.String),
			false,
			CompletionData{
				NewText:         `[ "value" ]`,
				Snippet:         `[ "${1:value}" ]`,
				NextPlaceholder: 2,
			},
		},
		{
			cty.Set(cty.String),
			false,
			CompletionData{
				NewText:         `[ "value" ]`,
				Snippet:         `[ "${1:value}" ]`,
				NextPlaceholder: 2,
			},
		},
		{
			cty.Tuple([]cty.Type{}),
			false,
			CompletionData{
				NewText:         "[ ]",
				Snippet:         "[ ${1} ]",
				NextPlaceholder: 2,
			},
		},
		{
			cty.Tuple([]cty.Type{cty.String, cty.Number}),
			false,
			CompletionData{
				NewText:         `[ "value", 0 ]`,
				Snippet:         `[ "${1:value}", ${2:0} ]`,
				NextPlaceholder: 3,
			},
		},
		{
			cty.Map(cty.String),
			false,
			CompletionData{
				NewText:         "{\n  \"name\" = \"value\"\n}",
				Snippet:         "{\n  \"${1:name}\" = \"${2:value}\"\n}",
				NextPlaceholder: 3,
			},
		},
		{
			cty.Object(map[string]cty.Type{}),
			false,
			CompletionData{
				NewText:         "{\n  \n}",
				Snippet:         "{\n  ${1}\n}",
				NextPlaceholder: 2,
			},
		},
		{
			cty.Object(map[string]cty.Type{
				"foo": cty.String,
				"bar": cty.Number,
			}),
			false,
			CompletionData{
				NewText:         "{\n  \n}",
				Snippet:         "{\n  ${1}\n}",
				TriggerSuggest:  true,
				NextPlaceholder: 2,
			},
		},
		{
			cty.Object(map[string]cty.Type{
				"foo": cty.String,
				"bar": cty.Number,
			}),
			true,
			CompletionData{
				NewText:         "{\n  bar = 0\n  foo = \"value\"\n}",
				Snippet:         "{\n  bar = ${1:0}\n  foo = \"${2:value}\"\n}",
				NextPlaceholder: 3,
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%2d-%s", i, tc.typ.FriendlyNameForConstraint()), func(t *testing.T) {
			lt := LiteralType{
				Type: tc.typ,
			}
			ctx := WithPrefillRequiredFields(context.Background(), tc.prefillRequiredFields)
			cData := lt.EmptyCompletionData(ctx, 1, 0)

			if diff := cmp.Diff(tc.expectedData, cData); diff != "" {
				t.Fatalf("unexpected data: %s", diff)
			}
		})
	}
}
