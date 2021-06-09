package schema

var (
	_ ExprConstraint = KeywordExpr{}
	_ ExprConstraint = ListExpr{}
	_ ExprConstraint = LiteralTypeExpr{}
	_ ExprConstraint = LiteralValue{}
	_ ExprConstraint = MapExpr{}
	_ ExprConstraint = MapExpr{}
	_ ExprConstraint = ObjectExprAttributes{}
	_ ExprConstraint = ObjectExpr{}
	_ ExprConstraint = SetExpr{}
	_ ExprConstraint = TraversalExpr{}
	_ ExprConstraint = TupleConsExpr{}
	_ ExprConstraint = TupleExpr{}
	_ ExprConstraint = TypeDeclarationExpr{}
)
