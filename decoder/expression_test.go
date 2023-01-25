package decoder

var (
	_ Expression = Any{}
	_ Expression = Keyword{}
	_ Expression = List{}
	_ Expression = LiteralType{}
	_ Expression = LiteralValue{}
	_ Expression = Map{}
	_ Expression = ObjectAttributes{}
	_ Expression = Object{}
	_ Expression = Set{}
	_ Expression = Reference{}
	_ Expression = Tuple{}
	_ Expression = TypeDeclaration{}

	_ ReferenceOriginsExpression = Any{}
	_ ReferenceOriginsExpression = List{}
	_ ReferenceOriginsExpression = LiteralType{}
	_ ReferenceOriginsExpression = Map{}
	_ ReferenceOriginsExpression = ObjectAttributes{}
	_ ReferenceOriginsExpression = Object{}
	_ ReferenceOriginsExpression = Set{}
	_ ReferenceOriginsExpression = Reference{}
	_ ReferenceOriginsExpression = Tuple{}

	_ ReferenceTargetsExpression = Any{}
	_ ReferenceTargetsExpression = List{}
	_ ReferenceTargetsExpression = LiteralType{}
	_ ReferenceTargetsExpression = Map{}
	_ ReferenceTargetsExpression = ObjectAttributes{}
	_ ReferenceTargetsExpression = Object{}
	_ ReferenceTargetsExpression = Reference{}
	_ ReferenceTargetsExpression = Tuple{}
)
