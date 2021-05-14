package lang

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type ScopeId string

type Reference struct {
	Addr     Address
	ScopeId  ScopeId
	RangePtr *hcl.Range

	Type        cty.Type
	Name        string
	Description MarkupContent

	InsideReferences References
}

type References []Reference

func (r References) Len() int {
	return len(r)
}

func (r References) Less(i, j int) bool {
	return r[i].Addr.String() < r[j].Addr.String()
}

func (r References) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r Reference) Address() Address {
	return r.Addr
}

func (r Reference) FriendlyName() string {
	if r.Name != "" {
		return r.Name
	}

	if r.Type != cty.NilType {
		return r.Type.FriendlyName()
	}

	return "reference"
}

func (r Reference) TargetRange() (hcl.Range, bool) {
	if r.RangePtr == nil {
		return hcl.Range{}, false
	}

	return *r.RangePtr, true
}
