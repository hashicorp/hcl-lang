// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package schema

type addrStepImplSigil struct{}

type AddrStep interface {
	isAddrStepImpl() addrStepImplSigil
}

type StaticStep struct {
	Name string
}

func (StaticStep) isAddrStepImpl() addrStepImplSigil {
	return addrStepImplSigil{}
}

type LabelStep struct {
	Index uint
}

func (LabelStep) isAddrStepImpl() addrStepImplSigil {
	return addrStepImplSigil{}
}

type AttrNameStep struct{}

func (AttrNameStep) isAddrStepImpl() addrStepImplSigil {
	return addrStepImplSigil{}
}

type AttrValueStep struct {
	Name       string
	IsOptional bool
}

func (AttrValueStep) isAddrStepImpl() addrStepImplSigil {
	return addrStepImplSigil{}
}

type Skip struct{}

func (Skip) isAddrStepImpl() addrStepImplSigil {
	return addrStepImplSigil{}
}
