// Copyright IBM Corp. 2020, 2026
// SPDX-License-Identifier: MPL-2.0

package reference

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty-debug/ctydebug"
)

func TestTarget_Address(t *testing.T) {
	testCases := []struct {
		name            string
		pos             hcl.Pos
		activeSelfRefs  bool
		target          Target
		expectedAddress lang.Address
	}{
		{
			"absolute address and no local address",
			hcl.InitialPos,
			false,
			Target{
				Addr: lang.Address{
					lang.RootStep{Name: "aws_instance"},
					lang.AttrStep{Name: "instance_size"},
				},
			},
			lang.Address{
				lang.RootStep{Name: "aws_instance"},
				lang.AttrStep{Name: "instance_size"},
			},
		},
		{
			"local address and no absolute address",
			hcl.InitialPos,
			false,
			Target{
				LocalAddr: lang.Address{
					lang.RootStep{Name: "count"},
					lang.AttrStep{Name: "index"},
				},
			},
			lang.Address{
				lang.RootStep{Name: "count"},
				lang.AttrStep{Name: "index"},
			},
		},
		{
			"self address with active self and matching range",
			hcl.Pos{Line: 2, Column: 2, Byte: 2},
			true,
			Target{
				Addr: lang.Address{
					lang.RootStep{Name: "aws_instance"},
					lang.AttrStep{Name: "instance_size"},
				},
				LocalAddr: lang.Address{
					lang.RootStep{Name: "self"},
					lang.AttrStep{Name: "instance_size"},
				},
				TargetableFromRangePtr: &hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:      hcl.Pos{Line: 3, Column: 1, Byte: 10},
				},
			},
			lang.Address{
				lang.RootStep{Name: "self"},
				lang.AttrStep{Name: "instance_size"},
			},
		},
		{
			"self address without active self but matching range",
			hcl.Pos{Line: 2, Column: 2, Byte: 2},
			false,
			Target{
				Addr: lang.Address{
					lang.RootStep{Name: "aws_instance"},
					lang.AttrStep{Name: "instance_size"},
				},
				LocalAddr: lang.Address{
					lang.RootStep{Name: "self"},
					lang.AttrStep{Name: "instance_size"},
				},
				TargetableFromRangePtr: &hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:      hcl.Pos{Line: 3, Column: 1, Byte: 10},
				},
			},
			lang.Address{
				lang.RootStep{Name: "aws_instance"},
				lang.AttrStep{Name: "instance_size"},
			},
		},
		{
			"self address with active self but no matching range",
			hcl.Pos{Line: 5, Column: 2, Byte: 15},
			true,
			Target{
				Addr: lang.Address{
					lang.RootStep{Name: "aws_instance"},
					lang.AttrStep{Name: "instance_size"},
				},
				LocalAddr: lang.Address{
					lang.RootStep{Name: "self"},
					lang.AttrStep{Name: "instance_size"},
				},
				TargetableFromRangePtr: &hcl.Range{
					Filename: "test.tf",
					Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:      hcl.Pos{Line: 3, Column: 1, Byte: 10},
				},
			},
			lang.Address{
				lang.RootStep{Name: "aws_instance"},
				lang.AttrStep{Name: "instance_size"},
			},
		},
		{
			"self address with active self and missing targetable",
			hcl.Pos{Line: 5, Column: 2, Byte: 15},
			true,
			Target{
				Addr: lang.Address{
					lang.RootStep{Name: "aws_instance"},
					lang.AttrStep{Name: "instance_size"},
				},
				LocalAddr: lang.Address{
					lang.RootStep{Name: "self"},
					lang.AttrStep{Name: "instance_size"},
				},
			},
			lang.Address{
				lang.RootStep{Name: "aws_instance"},
				lang.AttrStep{Name: "instance_size"},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			ctx := context.Background()

			if tc.activeSelfRefs {
				ctx = schema.WithActiveSelfRefs(ctx)
			}

			address := tc.target.Address(ctx, tc.pos)
			if diff := cmp.Diff(tc.expectedAddress, address, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of address: %s", diff)
			}
		})
	}
}
