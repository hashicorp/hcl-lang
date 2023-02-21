package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (obj Object) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	eType, ok := obj.expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return reference.Origins{}
	}

	if len(eType.Items) == 0 || len(obj.cons.Attributes) == 0 {
		return reference.Origins{}
	}

	origins := make(reference.Origins, 0)

	for _, item := range eType.Items {
		attrName, _, ok := rawObjectKey(item.KeyExpr)
		if !ok {
			continue
		}

		aSchema, ok := obj.cons.Attributes[attrName]
		if !ok {
			// skip unknown attribute
			continue
		}

		expr := newExpression(obj.pathCtx, item.ValueExpr, aSchema.Constraint)

		if elemExpr, ok := expr.(ReferenceOriginsExpression); ok {
			origins = append(origins, elemExpr.ReferenceOrigins(ctx, allowSelfRefs)...)
		}
	}

	return origins
}
