package lang

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
)

type Address []AddressStep

type addrStepSigil struct{}

type AddressStep interface {
	isRefStepImpl() addrStepSigil
}

func (a Address) Marshal() ([]byte, error) {
	return []byte(a.String()), nil
}

func (r Address) String() string {
	addr := ""
	for _, s := range r {
		switch step := s.(type) {
		case RootStep:
			addr += step.Name
		case AttrStep:
			addr += fmt.Sprintf(".%s", step.Name)
		case IndexStep:
			key := step.Key
			switch key.Type() {
			case cty.Number:
				f := key.AsBigFloat()
				idx, _ := f.Int64()
				addr += fmt.Sprintf("[%d]", idx)
			case cty.String:
				addr += fmt.Sprintf("[%q]", key.AsString())
			default:
				addr += fmt.Sprintf("<INVALIDKEY-%T>", step)
			}
		default:
			addr += fmt.Sprintf("<INVALIDSTEP-%T>", step)
		}
	}
	return addr
}

type RootStep struct {
	Name string `json:"name"`
}

func (RootStep) isRefStepImpl() addrStepSigil {
	return addrStepSigil{}
}

type AttrStep struct {
	Name string `json:"name"`
}

func (AttrStep) isRefStepImpl() addrStepSigil {
	return addrStepSigil{}
}

type IndexStep struct {
	Key cty.Value `json:"key"`
}

func (IndexStep) isRefStepImpl() addrStepSigil {
	return addrStepSigil{}
}
