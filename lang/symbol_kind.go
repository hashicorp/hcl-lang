package lang

// SymbolKind represents kind of a symbol in configuration
type SymbolKind uint

const (
	NilSymbolKind SymbolKind = iota
	AttributeSymbolKind
	BlockSymbolKind
)
