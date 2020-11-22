package lang

import (
	"github.com/hashicorp/hcl/v2"
)

type SemanticToken struct {
	Type      SemanticTokenType
	Modifiers []SemanticTokenModifier
	Range     hcl.Range
}

type SemanticTokenType uint

const (
	TokenNil SemanticTokenType = iota
	TokenAttrName
	TokenBlockType
	TokenBlockLabel
)

type SemanticTokenModifier uint

const (
	TokenModifierNil SemanticTokenModifier = iota
	TokenModifierDependent
	TokenModifierDeprecated
)
