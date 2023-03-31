// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2"
)

func (obj Object) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	items, diags := hcl.ExprMap(obj.expr)
	if diags.HasErrors() {
		return reference.Origins{}
	}

	if len(items) == 0 || len(obj.cons.Attributes) == 0 {
		return reference.Origins{}
	}

	origins := make(reference.Origins, 0)

	for _, item := range items {
		attrName, _, ok := rawObjectKey(item.Key)
		if !ok {
			continue
		}

		aSchema, ok := obj.cons.Attributes[attrName]
		if !ok {
			// skip unknown attribute
			continue
		}

		expr := newExpression(obj.pathCtx, item.Value, aSchema.Constraint)

		if elemExpr, ok := expr.(ReferenceOriginsExpression); ok {
			origins = append(origins, elemExpr.ReferenceOrigins(ctx, allowSelfRefs)...)
		}
	}

	return origins
}
