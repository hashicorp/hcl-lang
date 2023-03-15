package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (lt LiteralType) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	typ := lt.cons.Type

	// Primitive types are collected separately on attribute level
	if typ.IsPrimitiveType() {
		return reference.Targets{}
	}

	if typ.IsListType() {
		expr, ok := lt.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return nil
		}

		list := List{
			cons: schema.List{
				Elem: schema.LiteralType{
					Type: typ.ElementType(),
				},
			},
			expr:    expr,
			pathCtx: lt.pathCtx,
		}
		return list.ReferenceTargets(ctx, targetCtx)
	}

	if typ.IsSetType() {
		expr, ok := lt.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return nil
		}

		set := Set{
			cons: schema.Set{
				Elem: schema.LiteralType{
					Type: typ.ElementType(),
				},
			},
			expr:    expr,
			pathCtx: lt.pathCtx,
		}
		return set.ReferenceTargets(ctx, targetCtx)
	}

	if typ.IsTupleType() {
		expr, ok := lt.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return nil
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
			cons:    cons,
			expr:    expr,
			pathCtx: lt.pathCtx,
		}

		return tuple.ReferenceTargets(ctx, targetCtx)
	}

	if typ.IsMapType() {
		expr, ok := lt.expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return nil
		}

		m := Map{
			cons: schema.Map{
				Elem: schema.LiteralType{
					Type: typ.ElementType(),
				},
			},
			expr:    expr,
			pathCtx: lt.pathCtx,
		}
		return m.ReferenceTargets(ctx, targetCtx)
	}

	if typ.IsObjectType() {
		expr, ok := lt.expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return nil
		}

		obj := Object{
			cons: schema.Object{
				Attributes: ctyObjectToObjectAttributes(typ),
			},
			expr:    expr,
			pathCtx: lt.pathCtx,
		}
		return obj.ReferenceTargets(ctx, targetCtx)
	}

	return reference.Targets{}
}
