// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lang

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
)

type addrStepSigil struct{}

type AddressStep interface {
	isRefStepImpl() addrStepSigil
	String() string
}

func (a Address) Marshal() ([]byte, error) {
	return []byte(a.String()), nil
}

func (a Address) String() string {
	addr := ""
	for _, s := range a {
		addr += s.String()
	}
	return addr
}

func (a Address) Copy() Address {
	addrCopy := make(Address, len(a))
	copy(addrCopy, a)
	return addrCopy
}

type RootStep struct {
	Name string `json:"name"`
}

func (s RootStep) String() string {
	return s.Name
}

func (RootStep) isRefStepImpl() addrStepSigil {
	return addrStepSigil{}
}

type AttrStep struct {
	Name string `json:"name"`
}

func (s AttrStep) String() string {
	return fmt.Sprintf(".%s", s.Name)
}

func (AttrStep) isRefStepImpl() addrStepSigil {
	return addrStepSigil{}
}

type IndexStep struct {
	Key cty.Value `json:"key"`
}

func (s IndexStep) String() string {
	switch s.Key.Type() {
	case cty.Number:
		f := s.Key.AsBigFloat()
		idx, _ := f.Int64()
		return fmt.Sprintf("[%d]", idx)
	case cty.String:
		return fmt.Sprintf("[%q]", s.Key.AsString())
	}

	return fmt.Sprintf("<INVALIDKEY-%T>", s)
}

func (IndexStep) isRefStepImpl() addrStepSigil {
	return addrStepSigil{}
}
