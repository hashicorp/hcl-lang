package reference

import (
	"github.com/hashicorp/hcl/v2"
)

type Origins []Origin

func (ro Origins) Copy() Origins {
	if ro == nil {
		return nil
	}

	newOrigins := make(Origins, len(ro))
	for i, origin := range ro {
		newOrigins[i] = origin.Copy()
	}

	return newOrigins
}

func (ro Origins) AtPos(file string, pos hcl.Pos) (*Origin, bool) {
	for _, origin := range ro {
		if origin.Range.Filename == file && origin.Range.ContainsPos(pos) {
			return &origin, true
		}
	}

	return nil, false
}

func (ro Origins) Targeting(refTarget Target) Origins {
	origins := make(Origins, 0)

	for _, refOrigin := range ro {
		if refTarget.IsTargetableBy(refOrigin) {
			origins = append(origins, refOrigin)
		}
	}

	for _, iTarget := range refTarget.NestedTargets {
		origins = append(origins, ro.Targeting(iTarget)...)
	}

	return origins
}
