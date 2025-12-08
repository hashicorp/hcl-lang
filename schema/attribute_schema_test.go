// Copyright IBM Corp. 2020, 2025
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
)

func TestAttributeSchema_Validate(t *testing.T) {
	testCases := []struct {
		schema      *AttributeSchema
		expectedErr error
	}{
		{
			&AttributeSchema{
				Constraint: LiteralType{Type: cty.String},
			},
			errors.New("one of IsRequired, IsOptional, or IsComputed must be set"),
		},
		{
			&AttributeSchema{
				Constraint: OneOf{
					LiteralType{Type: cty.String},
					LiteralType{Type: cty.Number},
				},
				IsComputed: true,
			},
			nil,
		},
		{
			&AttributeSchema{
				Constraint: LiteralType{Type: cty.String},
				IsRequired: true,
				IsOptional: true,
			},
			errors.New("IsOptional or IsRequired must be set, not both"),
		},
		{
			&AttributeSchema{
				Constraint: LiteralType{Type: cty.String},
				IsRequired: true,
				IsComputed: true,
			},
			errors.New("cannot be both IsRequired and IsComputed"),
		},
		{
			&AttributeSchema{
				Constraint: LiteralType{Type: cty.String},
				IsOptional: true,
				IsComputed: true,
			},
			nil,
		},
		{
			&AttributeSchema{
				Constraint: Reference{OfType: cty.String},
				IsOptional: true,
			},
			nil,
		},
		{
			&AttributeSchema{
				Constraint: Reference{OfScopeId: lang.ScopeId("blah")},
				IsOptional: true,
			},
			nil,
		},
		{
			&AttributeSchema{
				Constraint: Reference{OfType: cty.Number, OfScopeId: lang.ScopeId("blah")},
				IsOptional: true,
			},
			nil,
		},
		{
			&AttributeSchema{
				Constraint: Reference{OfType: cty.Number, Address: &ReferenceAddrSchema{ScopeId: lang.ScopeId("test")}},
				IsOptional: true,
			},
			errors.New("Constraint: schema.Reference: cannot have both Address and OfType/OfScopeId set"),
		},
		{
			&AttributeSchema{
				Constraint: Reference{Address: &ReferenceAddrSchema{}},
				IsOptional: true,
			},
			errors.New("Constraint: schema.Reference: Address requires non-empty ScopeId"),
		},
		{
			&AttributeSchema{
				Constraint: Reference{Address: &ReferenceAddrSchema{ScopeId: lang.ScopeId("blah")}},
				IsOptional: true,
			},
			nil,
		},
		{
			&AttributeSchema{
				Constraint: LiteralType{Type: cty.String},
				Address: &AttributeAddrSchema{
					Steps: []AddrStep{
						LabelStep{Index: 0},
					},
					AsReference: true,
				},
				IsOptional: true,
			},
			errors.New("Address[0]: LabelStep is not valid for attribute"),
		},
		{
			&AttributeSchema{
				Constraint: LiteralType{Type: cty.String},
				Address: &AttributeAddrSchema{
					Steps: []AddrStep{
						AttrValueStep{Name: "unknown"},
					},
					AsReference: true,
				},
				IsOptional: true,
			},
			errors.New("Address[0]: AttrValueStep is not implemented for attribute"),
		},
		{
			&AttributeSchema{
				Constraint: LiteralType{Type: cty.String},
				Address: &AttributeAddrSchema{
					Steps: []AddrStep{
						AttrNameStep{},
					},
				},
				IsOptional: true,
			},
			errors.New("Address: at least one of AsExprType or AsReference must be set"),
		},
		{
			&AttributeSchema{
				Constraint: LiteralType{Type: cty.String},
				Address: &AttributeAddrSchema{
					Steps: []AddrStep{
						AttrNameStep{},
					},
					AsReference: true,
					AsExprType:  true,
				},
				IsOptional: true,
			},
			nil,
		},
		{
			&AttributeSchema{
				Constraint:  Reference{OfType: cty.Number, OfScopeId: lang.ScopeId("blah")},
				IsRequired:  true,
				IsSensitive: true,
			},
			nil,
		},
		{
			&AttributeSchema{
				Constraint:  LiteralType{},
				IsRequired:  true,
				IsSensitive: true,
			},
			errors.New("Constraint: schema.LiteralType: expected Type not to be nil"),
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			err := tc.schema.Validate()
			if tc.expectedErr == nil && err != nil {
				t.Fatal(err)
			}
			if tc.expectedErr != nil && err == nil {
				t.Fatalf("expected error: %q, none given", tc.expectedErr.Error())
			}
			if tc.expectedErr != nil && tc.expectedErr.Error() != err.Error() {
				t.Fatalf("error mismatch,\nexpected: %q\ngiven: %q", tc.expectedErr.Error(), err.Error())
			}
		})
	}
}
