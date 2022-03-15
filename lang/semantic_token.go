package lang

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

type SemanticToken struct {
	Type      SemanticTokenType
	Modifiers []SemanticTokenModifier
	Range     hcl.Range
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=SemanticTokenType -output=semantic_token_type_string.go
type SemanticTokenType uint

const (
	TokenNil SemanticTokenType = iota

	// structural tokens
	TokenAttrName
	TokenBlockType
	TokenBlockLabel

	// expressions
	TokenBool
	TokenString
	TokenNumber
	TokenObjectKey
	TokenMapKey
	TokenKeyword
	TokenTraversalStep
	TokenTypeCapsule
	TokenTypePrimitive
)

func (t SemanticTokenType) GoString() string {
	return fmt.Sprintf("lang.%s", t.String())
}

type SemanticTokenModifier string

const (
	TokenModifierDependent = SemanticTokenModifier("hcl-dependent")
)
