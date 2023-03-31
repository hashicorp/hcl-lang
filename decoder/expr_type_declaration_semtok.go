// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (td TypeDeclaration) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	switch eType := td.expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		if len(eType.Traversal) != 1 {
			return []lang.SemanticToken{}
		}

		if isPrimitiveTypeDeclaration(eType.Traversal.RootName()) {
			return []lang.SemanticToken{
				{
					Type:      lang.TokenTypePrimitive,
					Modifiers: []lang.SemanticTokenModifier{},
					Range:     eType.Range(),
				},
			}
		}
	case *hclsyntax.FunctionCallExpr:
		if isTypeNameWithElementOnly(eType.Name) {
			tokens := make([]lang.SemanticToken, 0)

			tokens = append(tokens, lang.SemanticToken{
				Type:      lang.TokenTypeCapsule,
				Modifiers: []lang.SemanticTokenModifier{},
				Range:     eType.NameRange,
			})

			if len(eType.Args) == 0 {
				return tokens
			}

			if len(eType.Args) == 1 {
				cons := TypeDeclaration{
					expr:    eType.Args[0],
					pathCtx: td.pathCtx,
				}
				tokens = append(tokens, cons.SemanticTokens(ctx)...)

				return tokens
			}

			return []lang.SemanticToken{}
		}

		if eType.Name == "object" {
			return td.objectSemanticTokens(ctx, eType)
		}

		if eType.Name == "tuple" {
			return td.tupleSemanticTokens(ctx, eType)
		}
	}
	return nil
}

func (td TypeDeclaration) objectSemanticTokens(ctx context.Context, funcExpr *hclsyntax.FunctionCallExpr) []lang.SemanticToken {
	tokens := make([]lang.SemanticToken, 0)
	tokens = append(tokens, lang.SemanticToken{
		Type:      lang.TokenTypeCapsule,
		Modifiers: []lang.SemanticTokenModifier{},
		Range:     funcExpr.NameRange,
	})

	if len(funcExpr.Args) != 1 {
		return tokens
	}

	objExpr, ok := funcExpr.Args[0].(*hclsyntax.ObjectConsExpr)
	if !ok {
		return []lang.SemanticToken{}
	}

	for _, item := range objExpr.Items {
		_, _, ok := rawObjectKey(item.KeyExpr)
		if !ok {
			// avoid reporting un-decodable key
			return tokens
		}

		tokens = append(tokens, lang.SemanticToken{
			Type:      lang.TokenAttrName,
			Modifiers: []lang.SemanticTokenModifier{},
			Range:     item.KeyExpr.Range(),
		})

		cons := TypeDeclaration{
			expr:    item.ValueExpr,
			pathCtx: td.pathCtx,
		}
		tokens = append(tokens, cons.SemanticTokens(ctx)...)
	}

	return tokens
}

func (td TypeDeclaration) tupleSemanticTokens(ctx context.Context, funcExpr *hclsyntax.FunctionCallExpr) []lang.SemanticToken {
	tokens := make([]lang.SemanticToken, 0)
	tokens = append(tokens, lang.SemanticToken{
		Type:      lang.TokenTypeCapsule,
		Modifiers: []lang.SemanticTokenModifier{},
		Range:     funcExpr.NameRange,
	})

	if len(funcExpr.Args) != 1 {
		return tokens
	}

	tupleExpr, ok := funcExpr.Args[0].(*hclsyntax.TupleConsExpr)
	if !ok {
		return []lang.SemanticToken{}
	}

	for _, expr := range tupleExpr.Exprs {
		cons := TypeDeclaration{
			expr:    expr,
			pathCtx: td.pathCtx,
		}
		tokens = append(tokens, cons.SemanticTokens(ctx)...)
	}

	return tokens
}
