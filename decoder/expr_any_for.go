// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (a Any) completeForExprAtPos(ctx context.Context, pos hcl.Pos) ([]lang.Candidate, bool) {
	candidates := make([]lang.Candidate, 0)

	switch eType := a.expr.(type) {
	case *hclsyntax.ForExpr:
		if !isTypeIterable(a.cons.OfType) {
			return candidates, true
		}

		if eType.CollExpr.Range().ContainsPos(pos) || eType.CollExpr.Range().End.Byte == pos.Byte {
			return newExpression(a.pathCtx, eType.CollExpr, a.cons).CompletionAtPos(ctx, pos), true
		}

		if eType.KeyExpr != nil && (eType.KeyExpr.Range().ContainsPos(pos) || eType.KeyExpr.Range().End.Byte == pos.Byte) {
			typ, ok := iterableKeyType(a.cons.OfType)
			if !ok {
				return candidates, true
			}
			cons := schema.AnyExpression{
				OfType: typ,
			}

			return newExpression(a.pathCtx, eType.KeyExpr, cons).CompletionAtPos(ctx, pos), true
		}

		if eType.ValExpr.Range().ContainsPos(pos) || eType.ValExpr.Range().End.Byte == pos.Byte {
			typ, ok := iterableValueType(a.cons.OfType)
			if !ok {
				return candidates, true
			}
			cons := schema.AnyExpression{
				OfType: typ,
			}

			return newExpression(a.pathCtx, eType.ValExpr, cons).CompletionAtPos(ctx, pos), true
		}

		if eType.CondExpr != nil && (eType.CondExpr.Range().ContainsPos(pos) || eType.CondExpr.Range().End.Byte == pos.Byte) {
			cons := schema.AnyExpression{
				OfType: cty.Bool,
			}
			return newExpression(a.pathCtx, eType.CondExpr, cons).CompletionAtPos(ctx, pos), true
		}
		return candidates, false
	}

	return candidates, true
}

func isTypeIterable(typ cty.Type) bool {
	if typ == cty.DynamicPseudoType {
		return true
	}
	if typ.IsListType() {
		return true
	}
	if typ.IsMapType() {
		return true
	}
	if typ.IsSetType() {
		return true
	}
	if typ.IsTupleType() {
		return true
	}
	if typ.IsObjectType() {
		return true
	}
	return false
}

func iterableKeyType(typ cty.Type) (cty.Type, bool) {
	if typ == cty.DynamicPseudoType {
		return cty.DynamicPseudoType, true
	}
	if typ.IsListType() {
		return cty.Number, true
	}
	if typ.IsSetType() {
		// This looks awkward but we just mimic Terraform's behaviour
		return *typ.SetElementType(), true
	}
	if typ.IsTupleType() {
		return cty.Number, true
	}
	if typ.IsMapType() {
		return cty.String, true
	}
	if typ.IsObjectType() {
		return cty.String, true
	}
	return cty.NilType, false
}

func iterableValueType(typ cty.Type) (cty.Type, bool) {
	if typ == cty.DynamicPseudoType {
		return cty.DynamicPseudoType, true
	}
	if typ.IsListType() {
		return *typ.ListElementType(), true
	}
	if typ.IsSetType() {
		return *typ.SetElementType(), true
	}
	if typ.IsTupleType() {
		// This is not accurate but pragmatic
		return cty.DynamicPseudoType, true
	}
	if typ.IsMapType() {
		return *typ.MapElementType(), true
	}
	if typ.IsObjectType() {
		// This is not accurate but pragmatic
		return cty.DynamicPseudoType, true
	}
	return cty.NilType, false
}
