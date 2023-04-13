// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

func (a Any) ReferenceTargets(ctx context.Context, targetCtx *TargetContext) reference.Targets {
	typ := a.cons.OfType

	if typ == cty.DynamicPseudoType {
		val, diags := a.expr.Value(&hcl.EvalContext{})
		if !diags.HasErrors() {
			typ = val.Type()
		}
	}

	if targetCtx == nil || len(targetCtx.ParentAddress) == 0 {
		return reference.Targets{}
	}

	if typ.IsPrimitiveType() || typ == cty.DynamicPseudoType {
		var rangePtr *hcl.Range
		if targetCtx.ParentRangePtr != nil {
			rangePtr = targetCtx.ParentRangePtr
		} else {
			rangePtr = a.expr.Range().Ptr()
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
		list := List{
			cons: schema.List{
				Elem: schema.LiteralType{
					Type: typ.ElementType(),
				},
			},
			expr:    a.expr,
			pathCtx: a.pathCtx,
		}
		return list.ReferenceTargets(ctx, targetCtx)
	}

	if typ.IsSetType() {
		set := Set{
			cons: schema.Set{
				Elem: schema.LiteralType{
					Type: typ.ElementType(),
				},
			},
			expr:    a.expr,
			pathCtx: a.pathCtx,
		}
		return set.ReferenceTargets(ctx, targetCtx)
	}

	if typ.IsTupleType() {
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
			expr:    a.expr,
			pathCtx: a.pathCtx,
		}

		return tuple.ReferenceTargets(ctx, targetCtx)
	}

	if typ.IsMapType() {
		m := Map{
			cons: schema.Map{
				Elem: schema.LiteralType{
					Type: typ.ElementType(),
				},
			},
			expr:    a.expr,
			pathCtx: a.pathCtx,
		}
		return m.ReferenceTargets(ctx, targetCtx)
	}

	if typ.IsObjectType() {
		obj := Object{
			cons: schema.Object{
				Attributes: ctyObjectToObjectAttributes(typ),
			},
			expr:    a.expr,
			pathCtx: a.pathCtx,
		}
		return obj.ReferenceTargets(ctx, targetCtx)
	}

	return reference.Targets{}
}
