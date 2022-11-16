package reference

import (
	"errors"
	"strings"

	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Targets []Target

func (refs Targets) Copy() Targets {
	if refs == nil {
		return nil
	}

	newRefs := make(Targets, len(refs))
	for i, ref := range refs {
		newRefs[i] = ref.Copy()
	}

	return newRefs
}

func (r Targets) Len() int {
	return len(r)
}

func (r Targets) Less(i, j int) bool {
	return r[i].LocalAddr.String() < r[j].LocalAddr.String() ||
		r[i].Addr.String() < r[j].Addr.String()
}

func (r Targets) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

type TargetWalkFunc func(Target) error

var stopWalking error = errors.New("stop walking")

const InfiniteDepth = -1

func (refs Targets) deepWalk(f TargetWalkFunc, depth int) {
	w := refTargetDeepWalker{
		WalkFunc: f,
		Depth:    depth,
	}
	w.walk(refs)
}

type refTargetDeepWalker struct {
	WalkFunc TargetWalkFunc
	Depth    int

	currentDepth int
}

func (w refTargetDeepWalker) walk(refTargets Targets) {
	for _, ref := range refTargets {
		err := w.WalkFunc(ref)
		if err == stopWalking {
			return
		}

		if len(ref.NestedTargets) > 0 && (w.Depth == InfiniteDepth || w.Depth > w.currentDepth) {
			w.currentDepth++
			w.walk(ref.NestedTargets)
			w.currentDepth--
		}
	}
}

func (refs Targets) MatchWalk(te schema.TraversalExpr, prefix string, outermostBodyRng, originRng hcl.Range, f TargetWalkFunc) {
	for _, ref := range refs {
		matched := false
		if len(ref.LocalAddr) > 0 && strings.HasPrefix(ref.LocalAddr.String(), prefix) {
			// Check if origin is inside the targetable range
			if ref.TargetableFromRangePtr == nil || rangeOverlaps(*ref.TargetableFromRangePtr, originRng) {
				nestedMatches := ref.NestedTargets.containsMatch(te, prefix)
				if ref.MatchesConstraint(te) || nestedMatches {
					matched = true
				}
			}
		}
		if len(ref.Addr) > 0 && strings.HasPrefix(ref.Addr.String(), prefix) {
			nestedMatches := ref.NestedTargets.containsMatch(te, prefix)

			// avoid suggesting references to block's own fields from within
			if !referenceTargetIsInRange(ref, outermostBodyRng) &&
				(ref.MatchesConstraint(te) || nestedMatches) {
				matched = true
			}
		}
		if matched {
			f(ref)
			continue
		}

		ref.NestedTargets.MatchWalk(te, prefix, outermostBodyRng, originRng, f)
	}
}

func referenceTargetIsInRange(target Target, bodyRange hcl.Range) bool {
	return target.RangePtr != nil &&
		bodyRange.Filename == target.RangePtr.Filename &&
		(bodyRange.ContainsPos(target.RangePtr.Start) ||
			posEqual(bodyRange.End, target.RangePtr.End))
}

func posEqual(pos, other hcl.Pos) bool {
	return pos.Line == other.Line &&
		pos.Column == other.Column &&
		pos.Byte == other.Byte
}

func (refs Targets) containsMatch(te schema.TraversalExpr, prefix string) bool {
	for _, ref := range refs {
		if strings.HasPrefix(ref.LocalAddr.String(), prefix) &&
			ref.MatchesConstraint(te) {
			return true
		}
		if strings.HasPrefix(ref.Addr.String(), prefix) &&
			ref.MatchesConstraint(te) {
			return true
		}
		if len(ref.NestedTargets) > 0 {
			if match := ref.NestedTargets.containsMatch(te, prefix); match {
				return true
			}
		}
	}
	return false
}

func (refs Targets) Match(origin MatchableOrigin) (Targets, bool) {
	matchingReferences := make(Targets, 0)

	refs.deepWalk(func(ref Target) error {
		if ref.Matches(origin) {
			matchingReferences = append(matchingReferences, ref)
		}

		return nil
	}, InfiniteDepth)

	return matchingReferences, len(matchingReferences) > 0
}

func (refs Targets) OutermostInFile(file string) Targets {
	targets := make(Targets, 0)

	for _, target := range refs {
		if target.RangePtr == nil {
			continue
		}
		if target.RangePtr.Filename == file {
			targets = append(targets, target)
		}
	}

	return targets
}

func (refs Targets) InnermostAtPos(file string, pos hcl.Pos) (Targets, bool) {
	matchingTargets := make(Targets, 0)

	for _, target := range refs {
		if target.RangePtr == nil {
			continue
		}
		if target.RangePtr.Filename == file && target.RangePtr.ContainsPos(pos) {
			matchingTargets = append(matchingTargets, target)
		}
	}

	innermostTargets := make(Targets, 0)

	for _, target := range matchingTargets {
		if target.DefRangePtr != nil {
			if target.DefRangePtr.Filename == file &&
				target.DefRangePtr.ContainsPos(pos) {
				innermostTargets = append(innermostTargets, target)
				continue
			}
		}

		nestedTargets, ok := target.NestedTargets.InnermostAtPos(file, pos)
		if ok {
			innermostTargets = nestedTargets
			continue
		}

		innermostTargets = append(innermostTargets, target)
	}

	return innermostTargets, len(innermostTargets) > 0
}
