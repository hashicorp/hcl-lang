// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/hcl/v2/json"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestCollectRefTargets_exprReference_hcl(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefTargets reference.Targets
	}{
		{
			"no traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						Address: &schema.ReferenceAddrSchema{
							ScopeId: lang.ScopeId("foo"),
						},
					},
					IsOptional: true,
				},
			},
			`attr = "foo"`,
			reference.Targets{},
		},
		{
			"wrapped traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						Address: &schema.ReferenceAddrSchema{
							ScopeId: lang.ScopeId("foo"),
						},
					},
					IsOptional: true,
				},
			},
			`attr = "${foo}"`,
			reference.Targets{},
		},
		{
			"traversal with string",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						Address: &schema.ReferenceAddrSchema{
							ScopeId: lang.ScopeId("foo"),
						},
					},
					IsOptional: true,
				},
			},
			`attr = "${foo}-bar"`,
			reference.Targets{},
		},
		{
			"non-addressable traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
					},
					IsOptional: true,
				},
			},
			`attr = foo`,
			reference.Targets{},
		},
		{
			"addressable traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
						Address: &schema.ReferenceAddrSchema{
							ScopeId: lang.ScopeId("foobar"),
						},
						Name: "custom name",
					},
					IsOptional: true,
				},
			},
			`attr = foo`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					ScopeId: lang.ScopeId("foobar"),
					RangePtr: &hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:      hcl.Pos{Line: 1, Column: 11, Byte: 10},
					},
					Name: "custom name",
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			bodySchema := &schema.BodySchema{
				Attributes: tc.attrSchema,
			}

			f, diags := hclsyntax.ParseConfig([]byte(tc.cfg), "test.tf", hcl.InitialPos)
			if len(diags) > 0 {
				t.Error(diags)
			}
			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.tf": f,
				},
			})

			targets, err := d.CollectReferenceTargets()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedRefTargets, targets, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("unexpected targets: %s", diff)
			}
		})
	}
}

func TestCollectRefTargets_exprReference_json(t *testing.T) {
	testCases := []struct {
		testName           string
		attrSchema         map[string]*schema.AttributeSchema
		cfg                string
		expectedRefTargets reference.Targets
	}{
		{
			"no traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						Address: &schema.ReferenceAddrSchema{
							ScopeId: lang.ScopeId("foo"),
						},
					},
					IsOptional: true,
				},
			},
			`{"attr": 422}`,
			reference.Targets{},
		},
		{
			"traversal with string",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						Address: &schema.ReferenceAddrSchema{
							ScopeId: lang.ScopeId("foo"),
						},
					},
					IsOptional: true,
				},
			},
			`{"attr": "${foo}-bar"}`,
			reference.Targets{},
		},
		{
			"non-addressable traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
					},
					IsOptional: true,
				},
			},
			`{"attr": "${foo}"}`,
			reference.Targets{},
		},
		{
			"addressable traversal",
			map[string]*schema.AttributeSchema{
				"attr": {
					Constraint: schema.Reference{
						OfType: cty.String,
						Address: &schema.ReferenceAddrSchema{
							ScopeId: lang.ScopeId("foobar"),
						},
						Name: "custom name",
					},
					IsOptional: true,
				},
			},
			`{"attr": "${foo}"}`,
			reference.Targets{
				{
					Addr: lang.Address{
						lang.RootStep{Name: "foo"},
					},
					ScopeId: lang.ScopeId("foobar"),
					RangePtr: &hcl.Range{
						Filename: "test.tf.json",
						Start:    hcl.Pos{Line: 1, Column: 13, Byte: 12},
						End:      hcl.Pos{Line: 1, Column: 16, Byte: 15},
					},
					Name: "custom name",
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.testName), func(t *testing.T) {
			bodySchema := &schema.BodySchema{
				Attributes: tc.attrSchema,
			}

			f, diags := json.ParseWithStartPos([]byte(tc.cfg), "test.tf.json", hcl.InitialPos)
			if len(diags) > 0 {
				t.Error(diags)
			}
			d := testPathDecoder(t, &PathContext{
				Schema: bodySchema,
				Files: map[string]*hcl.File{
					"test.tf.json": f,
				},
			})

			targets, err := d.CollectReferenceTargets()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedRefTargets, targets, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("unexpected targets: %s", diff)
			}
		})
	}
}
