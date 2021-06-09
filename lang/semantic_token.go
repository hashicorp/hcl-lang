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

//go:generate stringer -type=SemanticTokenType -output=semantic_token_type_string.go
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

//go:generate stringer -type=SemanticTokenModifier -output=semantic_token_modifier_string.go
type SemanticTokenModifier uint

const (
	TokenModifierNil SemanticTokenModifier = iota
	TokenModifierDependent
	TokenModifierDeprecated
)

func (m SemanticTokenModifier) GoString() string {
	return fmt.Sprintf("lang.%s", m.String())
}
