package reference

import (
	"github.com/hashicorp/hcl-lang/lang"
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

func (ro Origins) AtPos(file string, pos hcl.Pos) (Origins, bool) {
	matchingOrigins := make(Origins, 0)
	for _, origin := range ro {
		if origin.OriginRange().Filename == file && origin.OriginRange().ContainsPos(pos) {
			matchingOrigins = append(matchingOrigins, origin)
		}
	}

	return matchingOrigins, len(matchingOrigins) > 0
}

func (ro Origins) Match(refTarget Target, atPath lang.Path) Origins {
	origins := make(Origins, 0)

	for _, refOrigin := range ro {
		switch origin := refOrigin.(type) {
		case LocalOrigin:
			if refTarget.Matches(origin.Address(), origin.OriginConstraints()) {
				origins = append(origins, refOrigin)
			}
		case PathOrigin:
			if origin.TargetPath.Equals(atPath) && refTarget.Matches(origin.Address(), origin.OriginConstraints()) {
				origins = append(origins, refOrigin)
			}
		}
	}

	for _, iTarget := range refTarget.NestedTargets {
		origins = append(origins, ro.Match(iTarget, atPath)...)
	}

	return origins
}
