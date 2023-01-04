package schema

var (
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
)
