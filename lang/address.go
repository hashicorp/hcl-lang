// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lang

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

type Address []AddressStep

func (a Address) Equals(addr Address) bool {
	// Empty address may come up in context where there are
	// two addresses for the same target and only is declared
	// (LocalAddr / Addr) in which case we don't want the empty
	// one to be treated as a match.
	if len(a) == 0 && len(addr) == 0 {
		return false
	}
	if len(a) != len(addr) {
		return false
	}
	for i, step := range a {
		if step.String() != addr[i].String() {
			return false
		}
	}

	return true
}

func (a Address) FirstSteps(steps uint) Address {
	return a[0:steps]
}

func TraversalToAddress(traversal hcl.Traversal) (Address, error) {
	addr := Address{}
	for _, tr := range traversal {
		switch t := tr.(type) {
		case hcl.TraverseRoot:
			addr = append(addr, RootStep{
				Name: t.Name,
			})
		case hcl.TraverseAttr:
			addr = append(addr, AttrStep{
				Name: t.Name,
			})
		case hcl.TraverseIndex:
			addr = append(addr, IndexStep{
				Key: t.Key,
			})
		default:
			return addr, fmt.Errorf("invalid traversal: %#v", tr)
		}
	}
	return addr, nil
}
