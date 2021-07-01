package lang

import "github.com/hashicorp/hcl/v2"

type ReferenceOrigin struct {
	Addr  Address
	Range hcl.Range
}

type ReferenceOrigins []ReferenceOrigin

func (ro ReferenceOrigins) Copy() ReferenceOrigins {
	if ro == nil {
		return nil
	}

	newOrigins := make(ReferenceOrigins, len(ro))
	for i, origin := range ro {
		newOrigins[i] = origin.Copy()
	}

	return newOrigins
}

func (ro ReferenceOrigin) Copy() ReferenceOrigin {
	return ReferenceOrigin{
		Addr:  ro.Addr,
		Range: ro.Range,
	}
}
