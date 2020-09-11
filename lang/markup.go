package lang

const (
	NilKind MarkupKind = iota
	PlainTextKind
	MarkdownKind
)

type MarkupKind uint

// MarkupContent represents human-readable content
// which can be represented as Markdown or plaintext
// for backwards-compatible reasons.
type MarkupContent struct {
	Value string
	Kind  MarkupKind
}

func PlainText(value string) MarkupContent {
	return MarkupContent{
		Value: value,
		Kind:  PlainTextKind,
	}
}

func Markdown(value string) MarkupContent {
	return MarkupContent{
		Value: value,
		Kind:  MarkdownKind,
	}
}
