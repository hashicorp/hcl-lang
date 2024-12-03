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

func (a Any) hoverForExprAtPos(ctx context.Context, pos hcl.Pos) (*lang.HoverData, bool) {
	switch eType := a.expr.(type) {
	case *hclsyntax.ForExpr:
		if !isTypeIterable(a.cons.OfType) {
			return nil, false
		}

		// TODO: eType.KeyVarExpr.Range() to display key type

		// TODO: eType.ValVarExpr.Range() to display value type

		if eType.CollExpr.Range().ContainsPos(pos) {
			return newExpression(a.pathCtx, eType.CollExpr, a.cons).HoverAtPos(ctx, pos), true
		}

		if eType.KeyExpr != nil && eType.KeyExpr.Range().ContainsPos(pos) {
			typ, ok := iterableKeyType(a.cons.OfType)
			if !ok {
				return nil, false
			}
			cons := schema.AnyExpression{
				OfType: typ,
			}
			return newExpression(a.pathCtx, eType.KeyExpr, cons).HoverAtPos(ctx, pos), true
		}

		if eType.ValExpr.Range().ContainsPos(pos) {
			typ, ok := iterableValueType(a.cons.OfType)
			if !ok {
				return nil, false
			}
			cons := schema.AnyExpression{
				OfType: typ,
			}
			return newExpression(a.pathCtx, eType.ValExpr, cons).HoverAtPos(ctx, pos), true
		}

		if eType.CondExpr != nil && eType.CondExpr.Range().ContainsPos(pos) {
			cons := schema.AnyExpression{
				OfType: cty.Bool,
			}
			return newExpression(a.pathCtx, eType.CondExpr, cons).HoverAtPos(ctx, pos), true
		}
	}

	return nil, false
}

func (a Any) semanticTokensForForExpr(ctx context.Context) ([]lang.SemanticToken, bool) {
	tokens := make([]lang.SemanticToken, 0)

	switch eType := a.expr.(type) {
	case *hclsyntax.ForExpr:
		if !isTypeIterable(a.cons.OfType) {
			return nil, false
		}

		// TODO: eType.KeyVarExpr.Range() to report key as keyword
		// TODO: eType.ValVarExpr.Range() to display value as keyword

		tokens = append(tokens, newExpression(a.pathCtx, eType.CollExpr, a.cons).SemanticTokens(ctx)...)

		if eType.KeyExpr != nil {
			typ, ok := iterableKeyType(a.cons.OfType)
			if !ok {
				return nil, false
			}
			cons := schema.AnyExpression{
				OfType: typ,
			}
			tokens = append(tokens, newExpression(a.pathCtx, eType.KeyExpr, cons).SemanticTokens(ctx)...)
		}

		typ, ok := iterableValueType(a.cons.OfType)
		if !ok {
			return nil, false
		}
		cons := schema.AnyExpression{
			OfType: typ,
		}
		tokens = append(tokens, newExpression(a.pathCtx, eType.ValExpr, cons).SemanticTokens(ctx)...)

		if eType.CondExpr != nil {
			cons := schema.AnyExpression{
				OfType: cty.Bool,
			}
			tokens = append(tokens, newExpression(a.pathCtx, eType.CondExpr, cons).SemanticTokens(ctx)...)
		}

		return tokens, true
	}

	return tokens, false
}

func (a Any) refOriginsForForExpr(ctx context.Context) (reference.Origins, bool) {
	origins := make(reference.Origins, 0)

	// There is currently no way of decoding for expressions in JSON
	// so we just collect them using the fallback logic assuming "any"
	// constraint and focus on collecting expressions in HCL with more
	// accurate constraints below.

	switch eType := a.expr.(type) {
	case *hclsyntax.ForExpr:
		if !isTypeIterable(a.cons.OfType) {
			return nil, false
		}

		// TODO: eType.KeyVarExpr.Range() to collect key as origin
		// TODO: eType.ValVarExpr.Range() to collect value as origin

		// A for expression's input can be a list, a set, a tuple, a map, or an object
		collCons := schema.OneOf{
			schema.AnyExpression{OfType: cty.List(cty.DynamicPseudoType)},
			schema.AnyExpression{OfType: cty.Set(cty.DynamicPseudoType)},
			schema.AnyExpression{OfType: cty.EmptyTuple},
			schema.AnyExpression{OfType: cty.Map(cty.DynamicPseudoType)},
			schema.AnyExpression{OfType: cty.EmptyObject},
		}
		if collExpr, ok := newExpression(a.pathCtx, eType.CollExpr, collCons).(ReferenceOriginsExpression); ok {
			origins = append(origins, collExpr.ReferenceOrigins(ctx)...)
		}

		if eType.KeyExpr != nil {
			typ, ok := iterableKeyType(a.cons.OfType)
			if !ok {
				return nil, false
			}
			cons := schema.AnyExpression{
				OfType: typ,
			}
			if keyExpr, ok := newExpression(a.pathCtx, eType.KeyExpr, cons).(ReferenceOriginsExpression); ok {
				origins = append(origins, keyExpr.ReferenceOrigins(ctx)...)
			}
		}

		typ, ok := iterableValueType(a.cons.OfType)
		if !ok {
			return nil, false
		}
		cons := schema.AnyExpression{
			OfType: typ,
		}
		if valExpr, ok := newExpression(a.pathCtx, eType.ValExpr, cons).(ReferenceOriginsExpression); ok {
			origins = append(origins, valExpr.ReferenceOrigins(ctx)...)
		}

		if eType.CondExpr != nil {
			cons := schema.AnyExpression{
				OfType: cty.Bool,
			}

			if condExpr, ok := newExpression(a.pathCtx, eType.CondExpr, cons).(ReferenceOriginsExpression); ok {
				origins = append(origins, condExpr.ReferenceOrigins(ctx)...)
			}
		}

		return origins, true
	}

	return origins, false
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
