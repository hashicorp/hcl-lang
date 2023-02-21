package schema

var (
	_ Constraint = AnyExpression{}
	_ Constraint = Keyword{}
	_ Constraint = List{}
	_ Constraint = LiteralType{}
	_ Constraint = LiteralValue{}
	_ Constraint = Map{}
	_ Constraint = ObjectAttributes{}
	_ Constraint = Object{}
	_ Constraint = Set{}
	_ Constraint = Reference{}
	_ Constraint = Tuple{}
	_ Constraint = TypeDeclaration{}

	_ ConstraintWithHoverData = List{}
	_ ConstraintWithHoverData = LiteralType{}
	_ ConstraintWithHoverData = LiteralValue{}
	_ ConstraintWithHoverData = Map{}
	_ ConstraintWithHoverData = ObjectAttributes{}
	_ ConstraintWithHoverData = Object{}
	_ ConstraintWithHoverData = Set{}
	_ ConstraintWithHoverData = Tuple{}

	_ TypeAwareConstraint = AnyExpression{}
	_ TypeAwareConstraint = List{}
	_ TypeAwareConstraint = LiteralType{}
	_ TypeAwareConstraint = LiteralValue{}
	_ TypeAwareConstraint = Map{}
	_ TypeAwareConstraint = Object{}
	_ TypeAwareConstraint = Set{}
	_ TypeAwareConstraint = Tuple{}
)
