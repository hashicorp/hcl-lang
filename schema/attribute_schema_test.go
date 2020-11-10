package schema

import (
	"errors"
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestAttributeSchema_Validate(t *testing.T) {
	testCases := []struct {
		schema      *AttributeSchema
		expectedErr error
	}{
		{
			&AttributeSchema{},
			errors.New("one of ValueType or ValueTypes must be specified"),
		},
		{
			&AttributeSchema{
				ValueType: cty.String,
			},
			errors.New("one of IsRequired, IsOptional, or IsComputed must be set"),
		},
		{
			&AttributeSchema{
				ValueTypes: []cty.Type{cty.String, cty.Number},
				IsComputed: true,
			},
			nil,
		},
		{
			&AttributeSchema{
				ValueType:  cty.String,
				ValueTypes: []cty.Type{cty.String, cty.Number},
				IsComputed: true,
			},
			errors.New("ValueType or ValueTypes must be specified, not both"),
		},
		{
			&AttributeSchema{
				ValueType:  cty.String,
				IsRequired: true,
				IsOptional: true,
			},
			errors.New("IsOptional or IsRequired must be set, not both"),
		},
		{
			&AttributeSchema{
				ValueType:  cty.String,
				IsRequired: true,
				IsComputed: true,
			},
			errors.New("cannot be both IsRequired and IsComputed"),
		},
		{
			&AttributeSchema{
				ValueType:  cty.String,
				IsOptional: true,
				IsComputed: true,
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
