// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (a Any) ReferenceOrigins(ctx context.Context) reference.Origins {
	typ := a.cons.OfType

	if typ.IsListType() {
		_, ok := a.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return a.refOriginsForNonComplexExpr(ctx)
		}

		list := List{
			expr:    a.expr,
			pathCtx: a.pathCtx,
			cons: schema.List{
				Elem: schema.AnyExpression{
					OfType: typ.ElementType(),
				},
			},
		}
		return list.ReferenceOrigins(ctx)
	}

	if typ.IsSetType() {
		_, ok := a.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return a.refOriginsForNonComplexExpr(ctx)
		}

		set := Set{
			expr:    a.expr,
			pathCtx: a.pathCtx,
			cons: schema.Set{
				Elem: schema.AnyExpression{
					OfType: typ.ElementType(),
				},
			},
		}
		return set.ReferenceOrigins(ctx)
	}

	if typ.IsTupleType() {
		_, ok := a.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return a.refOriginsForNonComplexExpr(ctx)
		}

		elemTypes := typ.TupleElementTypes()
		cons := schema.Tuple{
			Elems: make([]schema.Constraint, len(elemTypes)),
		}
		for i, elemType := range elemTypes {
			cons.Elems[i] = schema.LiteralType{
				Type: elemType,
			}
		}

		tuple := Tuple{
			expr:    a.expr,
			pathCtx: a.pathCtx,
			cons:    cons,
		}
		return tuple.ReferenceOrigins(ctx)
	}

	if typ.IsMapType() {
		_, ok := a.expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return a.refOriginsForNonComplexExpr(ctx)
		}

		m := Map{
			expr:    a.expr,
			pathCtx: a.pathCtx,
			cons: schema.Map{
				Elem: schema.AnyExpression{
					OfType: typ.ElementType(),
				},
				AllowInterpolatedKeys: true,
			},
		}
		return m.ReferenceOrigins(ctx)
	}

	if typ.IsObjectType() {
		_, ok := a.expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return a.refOriginsForNonComplexExpr(ctx)
		}

		obj := Object{
			expr:    a.expr,
			pathCtx: a.pathCtx,
			cons: schema.Object{
				Attributes:            ctyObjectToObjectAttributes(typ),
				AllowInterpolatedKeys: true,
			},
		}
		return obj.ReferenceOrigins(ctx)
	}

	return a.refOriginsForNonComplexExpr(ctx)
}

func (a Any) refOriginsForNonComplexExpr(ctx context.Context) reference.Origins {
	// TODO: Support splat expression https://github.com/hashicorp/terraform-ls/issues/526
	// TODO: Support relative traversals https://github.com/hashicorp/terraform-ls/issues/532

	if origins, ok := a.refOriginsForOperatorExpr(ctx); ok {
		return origins
	}

	if origins, ok := a.refOriginsForTemplateExpr(ctx); ok {
		return origins
	}

	if origins, ok := a.refOriginsForConditionalExpr(ctx); ok {
		return origins
	}

	if origins, ok := a.refOriginsForForExpr(ctx); ok {
		return origins
	}

	// attempt to get accurate constraint for the origins
	// if we recognise the given expression
	funcExpr := functionExpr{
		expr:       a.expr,
		returnType: a.cons.OfType,
		pathCtx:    a.pathCtx,
	}
	origins := funcExpr.ReferenceOrigins(ctx)
	if len(origins) > 0 {
		return origins
	}

	// If we're dealing with a valid function call expression that doesn't contain
	// any origins, there is no more work todo here and nothing below would match,
	// so we can return early.
	_, diags := hcl.ExprCall(a.expr)
	if !diags.HasErrors() {
		return origins
	}

	allowSelfRefs := schema.ActiveSelfRefsFromContext(ctx)
	te, ok := a.expr.(*hclsyntax.ScopeTraversalExpr)
	if ok {
		oCons := reference.OriginConstraints{
			{OfType: a.cons.OfType},
		}
		origin, ok := reference.TraversalToLocalOrigin(te.Traversal, oCons, allowSelfRefs)
		if ok {
			return reference.Origins{origin}
		}

		return reference.Origins{}
	}

	// if not we just collect any/all origins with vague constraint
	// as that is safest
	origins = make(reference.Origins, 0)
	vars := a.expr.Variables()
	for _, traversal := range vars {
		oCons := reference.OriginConstraints{
			{OfType: cty.DynamicPseudoType},
		}
		origin, ok := reference.TraversalToLocalOrigin(traversal, oCons, allowSelfRefs)
		if ok {
			origins = append(origins, origin)
		}
	}
	return origins
}
