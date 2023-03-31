// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lang

const (
	NilKind MarkupKind = iota
	PlainTextKind
	MarkdownKind
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=MarkupKind -output=markup_kind_string.go
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
