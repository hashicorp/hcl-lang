package lang

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Address []AddressStep

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
