package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (lt LiteralType) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	typ := lt.cons.Type

	if targetCtx == nil || len(targetCtx.ParentAddress) == 0 {
		return reference.Targets{}
	}

	if typ.IsPrimitiveType() {
		var rangePtr *hcl.Range
		if targetCtx.ParentRangePtr != nil {
			rangePtr = targetCtx.ParentRangePtr
		} else {
			rangePtr = lt.expr.Range().Ptr()
		}

		var refType cty.Type
		if targetCtx.AsExprType {
			refType = typ
		}

		return reference.Targets{
			{
				Addr:                   targetCtx.ParentAddress,
				LocalAddr:              targetCtx.ParentLocalAddress,
				TargetableFromRangePtr: targetCtx.TargetableFromRangePtr,
				ScopeId:                targetCtx.ScopeId,
				RangePtr:               rangePtr,
				DefRangePtr:            targetCtx.ParentDefRangePtr,
				Type:                   refType,
			},
		}
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
