// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

var (
	_ ExprConstraint = KeywordExpr{}
	_ ExprConstraint = ListExpr{}
	_ ExprConstraint = LiteralTypeExpr{}
	_ ExprConstraint = LegacyLiteralValue{}
	_ ExprConstraint = MapExpr{}
	_ ExprConstraint = MapExpr{}
	_ ExprConstraint = ObjectExprAttributes{}
	_ ExprConstraint = ObjectExpr{}
	_ ExprConstraint = SetExpr{}
	_ ExprConstraint = TraversalExpr{}
	_ ExprConstraint = TupleExpr{}
	_ ExprConstraint = TypeDeclarationExpr{}
)
