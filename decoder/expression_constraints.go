// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type ExprConstraints schema.ExprConstraints

func (ec ExprConstraints) HasKeywordsOnly() bool {
	hasKeywordExpr := false
	for _, constraint := range ec {
		if _, ok := constraint.(schema.KeywordExpr); ok {
			hasKeywordExpr = true
		} else {
			return false
		}
	}
	return hasKeywordExpr
}

func (ec ExprConstraints) KeywordExpr() (schema.KeywordExpr, bool) {
	for _, c := range ec {
		if kw, ok := c.(schema.KeywordExpr); ok {
			return kw, ok
		}
	}
	return schema.KeywordExpr{}, false
}

func (ec ExprConstraints) TraversalExprs() (schema.TraversalExprs, bool) {
	tes := make([]schema.TraversalExpr, 0)
	for _, c := range ec {
		if te, ok := c.(schema.TraversalExpr); ok {
			tes = append(tes, te)
		}
	}

	return tes, len(tes) > 0
}

func (ec ExprConstraints) MapExpr() (schema.MapExpr, bool) {
	for _, c := range ec {
		if me, ok := c.(schema.MapExpr); ok {
			return me, ok
		}
	}
	return schema.MapExpr{}, false
}

func (ec ExprConstraints) ObjectExpr() (schema.ObjectExpr, bool) {
	for _, c := range ec {
		if me, ok := c.(schema.ObjectExpr); ok {
			return me, ok
		}
	}
	return schema.ObjectExpr{}, false
}

func (ec ExprConstraints) SetExpr() (schema.SetExpr, bool) {
	for _, c := range ec {
		if se, ok := c.(schema.SetExpr); ok {
			return se, ok
		}
	}
	return schema.SetExpr{}, false
}

func (ec ExprConstraints) ListExpr() (schema.ListExpr, bool) {
	for _, c := range ec {
		if le, ok := c.(schema.ListExpr); ok {
			return le, ok
		}
	}
	return schema.ListExpr{}, false
}

func (ec ExprConstraints) TupleExpr() (schema.TupleExpr, bool) {
	for _, c := range ec {
		if te, ok := c.(schema.TupleExpr); ok {
			return te, ok
		}
	}
	return schema.TupleExpr{}, false
}

func (ec ExprConstraints) HasLiteralTypeOf(exprType cty.Type) bool {
	for _, c := range ec {
		if lt, ok := c.(schema.LiteralTypeExpr); ok && lt.Type.Equals(exprType) {
			return true
		}
	}
	return false
}

func (ec ExprConstraints) LiteralType() (cty.Type, bool) {
	for _, c := range ec {
		if lt, ok := c.(schema.LiteralTypeExpr); ok {
			return lt.Type, true
		}
	}
	return cty.NilType, false
}

func (ec ExprConstraints) HasLiteralValueOf(val cty.Value) bool {
	for _, c := range ec {
		if lv, ok := c.(schema.LegacyLiteralValue); ok && lv.Val.RawEquals(val) {
			return true
		}
	}
	return false
}

func (ec ExprConstraints) LiteralValueOf(val cty.Value) (schema.LegacyLiteralValue, bool) {
	for _, c := range ec {
		if lv, ok := c.(schema.LegacyLiteralValue); ok && lv.Val.RawEquals(val) {
			return lv, true
		}
	}
	return schema.LegacyLiteralValue{}, false
}

func (ec ExprConstraints) LiteralTypeOfTupleExpr() (schema.LiteralTypeExpr, bool) {
	for _, c := range ec {
		if lv, ok := c.(schema.LiteralTypeExpr); ok {
			if lv.Type.IsListType() {
				return lv, true
			}
			if lv.Type.IsSetType() {
				return lv, true
			}
			if lv.Type.IsTupleType() {
				return lv, true
			}
		}
	}
	return schema.LiteralTypeExpr{}, false
}

func (ec ExprConstraints) LiteralTypeOfObjectConsExpr() (schema.LiteralTypeExpr, bool) {
	for _, c := range ec {
		if lv, ok := c.(schema.LiteralTypeExpr); ok {
			if lv.Type.IsObjectType() {
				return lv, true
			}
			if lv.Type.IsMapType() {
				return lv, true
			}
		}
	}
	return schema.LiteralTypeExpr{}, false
}

func (ec ExprConstraints) LiteralValueOfTupleExpr(expr *hclsyntax.TupleConsExpr) (schema.LegacyLiteralValue, bool) {
	exprValues := make([]cty.Value, len(expr.Exprs))
	for i, e := range expr.Exprs {
		val, _ := e.Value(nil)
		if !val.IsWhollyKnown() || val.IsNull() {
			return schema.LegacyLiteralValue{}, false
		}
		exprValues[i] = val
	}

	for _, c := range ec {
		if lv, ok := c.(schema.LegacyLiteralValue); ok {
			valType := lv.Val.Type()
			if valType.IsListType() && lv.Val.RawEquals(cty.ListVal(exprValues)) {
				return lv, true
			}
			if valType.IsSetType() && lv.Val.RawEquals(cty.SetVal(exprValues)) {
				return lv, true
			}
			if valType.IsTupleType() && lv.Val.RawEquals(cty.TupleVal(exprValues)) {
				return lv, true
			}
		}
	}

	return schema.LegacyLiteralValue{}, false
}

func (ec ExprConstraints) LiteralValueOfObjectConsExpr(expr *hclsyntax.ObjectConsExpr) (schema.LegacyLiteralValue, bool) {
	exprValues := make(map[string]cty.Value)
	for _, item := range expr.Items {
		key, _ := item.KeyExpr.Value(nil)
		if key.IsNull() || !key.IsWhollyKnown() || key.Type() != cty.String {
			// Avoid building incomplete object with keys
			// that can't be interpolated without further context
			return schema.LegacyLiteralValue{}, false
		}

		val, _ := item.ValueExpr.Value(nil)
		if !val.IsWhollyKnown() || val.IsNull() {
			return schema.LegacyLiteralValue{}, false
		}

		exprValues[key.AsString()] = val
	}

	for _, c := range ec {
		if lv, ok := c.(schema.LegacyLiteralValue); ok {
			valType := lv.Val.Type()
			if valType.IsMapType() && lv.Val.RawEquals(cty.MapVal(exprValues)) {
				return lv, true
			}
			if valType.IsObjectType() && lv.Val.RawEquals(cty.ObjectVal(exprValues)) {
				return lv, true
			}
		}
	}

	return schema.LegacyLiteralValue{}, false
}

func (ec ExprConstraints) TypeDeclarationExpr() (schema.TypeDeclarationExpr, bool) {
	for _, c := range ec {
		if td, ok := c.(schema.TypeDeclarationExpr); ok {
			return td, ok
		}
	}
	return schema.TypeDeclarationExpr{}, false
}
