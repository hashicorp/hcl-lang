// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
)

func (oo OneOf) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	origins := make(reference.Origins, 0)

	for _, con := range oo.cons {
		expr := newExpression(oo.pathCtx, oo.expr, con)
		e, ok := expr.(ReferenceOriginsExpression)
		if !ok {
			continue
		}

		origins = appendOrigins(origins, e.ReferenceOrigins(ctx, allowSelfRefs))
	}

	return origins
}

func appendOrigins(origins, newOrigins reference.Origins) reference.Origins {
	// Deduplicating origins like this is probably not ideal
	// from performance perspective (N^2) but improving it would
	// require redesign of the schema.Reference constraint,
	// such that it doesn't necessitate the need of OneOf for multiple ScopeIds
	// and maintains all possible ScopeIds & Types as a *single* slice.
	for _, newOrigin := range newOrigins {
		newMatchableOrigin, ok := newOrigin.(reference.MatchableOrigin)
		if !ok {
			origins = append(origins, newOrigin)
			continue
		}

		foundMatch := false
		for i, origin := range origins {
			existingOrigin, ok := origin.(reference.MatchableOrigin)
			if ok &&
				existingOrigin.Address().Equals(newMatchableOrigin.Address()) &&
				rangesEqual(existingOrigin.OriginRange(), newMatchableOrigin.OriginRange()) {

				origins[i] = existingOrigin.AppendConstraints(newMatchableOrigin.OriginConstraints())
				foundMatch = true
				break
			}
		}
		if !foundMatch {
			origins = append(origins, newOrigin)
		}
	}

	return origins
}
