// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import "fmt"

type Address []AddrStep

func (addr Address) Copy() Address {
	if addr == nil {
		return nil
	}

	newAddr := make([]AddrStep, len(addr))
	for i, step := range addr {
		newAddr[i] = step
	}
	return newAddr
}

func (addr Address) AttributeValidate() error {
	if addr == nil {
		return nil
	}

	for i, step := range addr {
		if _, ok := step.(LabelStep); ok {
			return fmt.Errorf("Address[%d]: LabelStep is not valid for attribute", i)
		}
		if _, ok := step.(AttrValueStep); ok {
			return fmt.Errorf("Address[%d]: AttrValueStep is not implemented for attribute", i)
		}
	}

	return nil
}

func (addr Address) BlockValidate() error {
	if addr == nil {
		return nil
	}

	for i, step := range addr {
		if _, ok := step.(AttrNameStep); ok {
			return fmt.Errorf("Steps[%d]: AttrNameStep is not valid for block", i)
		}
	}

	return nil
}
