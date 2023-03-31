// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2"
)

func (list List) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	elems, diags := hcl.ExprList(list.expr)
	if diags.HasErrors() {
		return reference.Origins{}
	}

	if len(elems) == 0 || list.cons.Elem == nil {
		return reference.Origins{}
	}

	origins := make(reference.Origins, 0)

	for _, elemExpr := range elems {
		expr := newExpression(list.pathCtx, elemExpr, list.cons.Elem)
		if e, ok := expr.(ReferenceOriginsExpression); ok {
			origins = append(origins, e.ReferenceOrigins(ctx, allowSelfRefs)...)
		}
	}

	return origins
}
