// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lang

import (
	"github.com/hashicorp/hcl/v2"
)

type SemanticToken struct {
	Type      SemanticTokenType
	Modifiers SemanticTokenModifiers
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
	TokenFunctionName  SemanticTokenType = "hcl-functionName"
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
	TokenFunctionName,
}

type SemanticTokenModifier string

type SemanticTokenModifiers []SemanticTokenModifier

func (stm SemanticTokenModifiers) Copy() SemanticTokenModifiers {
	if stm == nil {
		return nil
	}

	modifiersCopy := make(SemanticTokenModifiers, len(stm))
	copy(modifiersCopy, stm)
	return modifiersCopy
}

const (
	TokenModifierDependent = SemanticTokenModifier("hcl-dependent")
)
