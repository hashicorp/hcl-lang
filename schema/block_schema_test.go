package schema

import (
	"errors"
	"fmt"
	"testing"
)

func TestBlockSchema_Validate(t *testing.T) {
	testCases := []struct {
		schema      *BlockSchema
		expectedErr error
	}{
		{ // empty block (with just type) is valid
			&BlockSchema{},
			nil,
		},
		{
			&BlockSchema{
				Address: &BlockAddrSchema{
					Steps: []AddrStep{
						AttrNameStep{},
					},
				},
			},
			errors.New("Address: Steps[0]: AttrNameStep is not valid for attribute"),
		},
		{
			&BlockSchema{
				Labels: []*LabelSchema{
					{Name: "name"},
				},
				Address: &BlockAddrSchema{
					Steps: []AddrStep{
						LabelStep{Index: 0},
					},
				},
			},
			nil,
		},
		{
			&BlockSchema{
				Address: &BlockAddrSchema{
					Steps: []AddrStep{
						StaticStep{Name: "bleh"},
					},
					InferBody: true,
				},
			},
			errors.New("Address: InferBody requires BodyAsData"),
		},
		{
			&BlockSchema{
				Address: &BlockAddrSchema{
					Steps: []AddrStep{
						StaticStep{Name: "bleh"},
					},
					InferDependentBody: true,
				},
			},
			errors.New("Address: InferDependentBody requires DependentBodyAsData"),
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
