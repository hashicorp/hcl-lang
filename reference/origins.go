// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

func (ro Origins) Match(localPath lang.Path, target Target, targetPath lang.Path) Origins {
	origins := make(Origins, 0)

	for _, refOrigin := range ro {
		switch origin := refOrigin.(type) {
		case LocalOrigin:
			if localPath.Equals(targetPath) && target.Matches(origin) {
				origins = append(origins, refOrigin)
			}
		case PathOrigin:
			if origin.TargetPath.Equals(targetPath) && target.Matches(origin) {
				origins = append(origins, refOrigin)
			}
		}
	}

	for _, iTarget := range target.NestedTargets {
		origins = append(origins, ro.Match(localPath, iTarget, targetPath)...)
	}

	return origins
}
