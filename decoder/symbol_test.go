package decoder

var (
	_ Symbol = &AttributeSymbol{}
	_ Symbol = &BlockSymbol{}
	_ Symbol = &ExprSymbol{}
)
