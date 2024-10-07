// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
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
		attrName, _, isRawKey := rawObjectKey(item.Key)

		var aSchema *schema.AttributeSchema
		var isKnownAttr bool
		if isRawKey {
			aSchema, isKnownAttr = obj.cons.Attributes[attrName]
		}

		keyExpr, ok := item.Key.(*hclsyntax.ObjectConsKeyExpr)
		if ok {
			parensExpr, ok := keyExpr.Wrapped.(*hclsyntax.ParenthesesExpr)
			if ok {
				keyCons := schema.AnyExpression{
					OfType: cty.String,
				}
				kExpr := newExpression(obj.pathCtx, parensExpr, keyCons)
				if expr, ok := kExpr.(ReferenceOriginsExpression); ok {
					origins = append(origins, expr.ReferenceOrigins(ctx, allowSelfRefs)...)
				}
			}
		}

		if isKnownAttr {
			expr := newExpression(obj.pathCtx, item.Value, aSchema.Constraint)
			if elemExpr, ok := expr.(ReferenceOriginsExpression); ok {
				origins = append(origins, elemExpr.ReferenceOrigins(ctx, allowSelfRefs)...)
			}
		}

		if aSchema != nil && aSchema.OriginForTarget != nil {
			address, ok := resolveObjectAddress(attrName, aSchema.OriginForTarget.Address)
			if !ok {
				continue
			}

			origins = append(origins, reference.PathOrigin{
				Range:      item.Key.Range(),
				TargetAddr: address,
				TargetPath: aSchema.OriginForTarget.Path,
				Constraints: reference.OriginConstraints{
					{
						OfScopeId: aSchema.OriginForTarget.Constraints.ScopeId,
						OfType:    aSchema.OriginForTarget.Constraints.Type,
					},
				},
			})
		}
	}

	return origins
}

func resolveObjectAddress(attrName string, addr schema.Address) (lang.Address, bool) {
	// This function is a simplified version of the original resolveAttributeAddress
	// because we don't have an attribute to pass

	address := make(lang.Address, 0)

	if len(addr) == 0 {
		return lang.Address{}, false
	}

	for i, s := range addr {
		var stepName string

		switch step := s.(type) {
		case schema.StaticStep:
			stepName = step.Name
		case schema.AttrNameStep:
			stepName = attrName
			// TODO: AttrValueStep? Currently no use case for it
		default:
			// unknown step
			return lang.Address{}, false
		}

		if i == 0 {
			address = append(address, lang.RootStep{
				Name: stepName,
			})
			continue
		}
		address = append(address, lang.AttrStep{
			Name: stepName,
		})
	}

	return address, true
}
