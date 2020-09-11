package lang

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
)

type Reference []ReferenceStep

type refStepSigil struct{}

type ReferenceStep interface {
	isRefStepImpl() refStepSigil
}

func (r Reference) Marshal() ([]byte, error) {
	ref := ""
	for _, s := range r {
		switch step := s.(type) {
		case RootStep:
			ref += step.Name
		case AttrStep:
			ref += fmt.Sprintf(".%s", step.Name)
		case IndexStep:
			key := step.Key
			switch key.Type() {
			case cty.Number:
				f := key.AsBigFloat()
				idx, _ := f.Int64()
				ref += fmt.Sprintf("[%d]", idx)
			case cty.String:
				ref += fmt.Sprintf("[%q]", key.AsString())
			default:
				ref += fmt.Sprintf("<INVALIDKEY-%T>", step)
			}
		default:
			ref += fmt.Sprintf("<INVALIDSTEP-%T>", step)
		}
	}
	return []byte(ref), nil
}

type RootStep struct {
	Name string `json:"name"`
}

func (RootStep) isRefStepImpl() refStepSigil {
	return refStepSigil{}
}

type AttrStep struct {
	Name string `json:"name"`
}

func (AttrStep) isRefStepImpl() refStepSigil {
	return refStepSigil{}
}

type IndexStep struct {
	Key cty.Value `json:"key"`
}

func (IndexStep) isRefStepImpl() refStepSigil {
	return refStepSigil{}
}
