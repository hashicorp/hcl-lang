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
				Expr: LiteralTypeOnly(cty.String),
			},
			errors.New("one of IsRequired, IsOptional, or IsComputed must be set"),
		},
		{
			&AttributeSchema{
				Expr: ExprConstraints{
					LiteralTypeExpr{Type: cty.String},
					LiteralTypeExpr{Type: cty.Number},
				},
				IsComputed: true,
			},
			nil,
		},
		{
			&AttributeSchema{
				Expr:       LiteralTypeOnly(cty.String),
				IsRequired: true,
				IsOptional: true,
			},
			errors.New("IsOptional or IsRequired must be set, not both"),
		},
		{
			&AttributeSchema{
				Expr:       LiteralTypeOnly(cty.String),
				IsRequired: true,
				IsComputed: true,
			},
			errors.New("cannot be both IsRequired and IsComputed"),
		},
		{
			&AttributeSchema{
				Expr:       LiteralTypeOnly(cty.String),
				IsOptional: true,
				IsComputed: true,
			},
			nil,
		},
		{
			&AttributeSchema{
				Expr: ExprConstraints{
					TraversalExpr{OfType: cty.String},
				},
				IsOptional: true,
			},
			nil,
		},
		{
			&AttributeSchema{
				Expr: ExprConstraints{
					TraversalExpr{OfScopeId: lang.ScopeId("blah")},
				},
				IsOptional: true,
			},
			nil,
		},
		{
			&AttributeSchema{
				Expr: ExprConstraints{
					TraversalExpr{OfType: cty.Number, OfScopeId: lang.ScopeId("blah")},
				},
				IsOptional: true,
			},
			nil,
		},
		{
			&AttributeSchema{
				Expr: ExprConstraints{
					TraversalExpr{OfType: cty.Number, Address: &TraversalAddrSchema{
						ScopeId: lang.ScopeId("test"),
					}},
				},
				IsOptional: true,
			},
			errors.New("(0: schema.TraversalExpr) cannot be have both Address and OfType/OfScopeId set"),
		},
		{
			&AttributeSchema{
				Expr: ExprConstraints{
					TraversalExpr{Address: &TraversalAddrSchema{}},
				},
				IsOptional: true,
			},
			errors.New("(0: schema.TraversalExpr) Address requires non-emmpty ScopeId"),
		},
		{
			&AttributeSchema{
				Expr: ExprConstraints{
					TraversalExpr{Address: &TraversalAddrSchema{
						ScopeId: lang.ScopeId("blah"),
					}},
				},
				IsOptional: true,
			},
			nil,
		},
		{
			&AttributeSchema{
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
				Address: &AttributeAddrSchema{
					Steps: []AddrStep{
						AttrNameStep{},
					},
				},
				IsOptional: true,
			},
			errors.New("Address: at least one of AsData or AsReference must be set"),
		},
		{
			&AttributeSchema{
				Address: &AttributeAddrSchema{
					Steps: []AddrStep{
						AttrNameStep{},
					},
					AsReference: true,
					AsData:      true,
				},
				IsOptional: true,
			},
			nil,
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
