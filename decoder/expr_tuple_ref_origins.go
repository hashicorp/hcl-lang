// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2"
)

func (tuple Tuple) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	elems, diags := hcl.ExprList(tuple.expr)
	if diags.HasErrors() {
		return reference.Origins{}
	}

	if len(elems) == 0 || len(tuple.cons.Elems) == 0 {
		return reference.Origins{}
	}

	origins := make(reference.Origins, 0)

	for i, elemExpr := range elems {
		if i+1 > len(tuple.cons.Elems) {
			break
		}

		expr := newExpression(tuple.pathCtx, elemExpr, tuple.cons.Elems[i])
		if e, ok := expr.(ReferenceOriginsExpression); ok {
			origins = append(origins, e.ReferenceOrigins(ctx, allowSelfRefs)...)
		}
	}

	return origins
}
