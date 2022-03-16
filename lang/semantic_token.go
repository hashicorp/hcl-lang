package lang

import (
	"github.com/hashicorp/hcl/v2"
)

type SemanticToken struct {
	Type      SemanticTokenType
	Modifiers []SemanticTokenModifier
	Range     hcl.Range
}

type SemanticTokenType string

type SemanticTokenTypes []SemanticTokenType

const (
	// structural tokens
	TokenAttrName   SemanticTokenType = "hcl-attrName"
	TokenBlockType  SemanticTokenType = "hcl-blockType"
	TokenBlockLabel SemanticTokenType = "hcl-blockLabel"

	// expressions
	TokenBool          SemanticTokenType = "hcl-bool"
	TokenString        SemanticTokenType = "hcl-string"
	TokenNumber        SemanticTokenType = "hcl-number"
	TokenObjectKey     SemanticTokenType = "hcl-objectKey"
	TokenMapKey        SemanticTokenType = "hcl-mapKey"
	TokenKeyword       SemanticTokenType = "hcl-keyword"
	TokenTraversalStep SemanticTokenType = "hcl-traversalStep"
	TokenTypeCapsule   SemanticTokenType = "hcl-typeCapsule"
	TokenTypePrimitive SemanticTokenType = "hcl-typePrimitive"
)

var SupportedSemanticTokenTypes = SemanticTokenTypes{
	TokenAttrName,
	TokenBlockType,
	TokenBlockLabel,
	TokenBool,
	TokenString,
	TokenNumber,
	TokenObjectKey,
	TokenMapKey,
	TokenKeyword,
	TokenTraversalStep,
	TokenTypeCapsule,
	TokenTypePrimitive,
}

type SemanticTokenModifier string

const (
	TokenModifierDependent = SemanticTokenModifier("hcl-dependent")
)
