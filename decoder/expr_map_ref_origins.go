// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2"
)

func (m Map) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	items, diags := hcl.ExprMap(m.expr)
	if diags.HasErrors() {
		return reference.Origins{}
	}

	if len(items) == 0 || m.cons.Elem == nil {
		return reference.Origins{}
	}

	origins := make(reference.Origins, 0)

	for _, item := range items {
		expr := newExpression(m.pathCtx, item.Value, m.cons.Elem)

		if elemExpr, ok := expr.(ReferenceOriginsExpression); ok {
			origins = append(origins, elemExpr.ReferenceOrigins(ctx, allowSelfRefs)...)
		}
	}

	return origins
}
