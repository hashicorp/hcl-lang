package decoder

var (
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
)
