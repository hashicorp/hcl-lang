package schema

var (
	_ ExprConstraint = LiteralTypeExpr{}
	_ ExprConstraint = LiteralValue{}
	_ ExprConstraint = TupleConsExpr{}
	_ ExprConstraint = MapExpr{}
	_ ExprConstraint = KeywordExpr{}
	_ ExprConstraint = TraversalExpr{}
)
