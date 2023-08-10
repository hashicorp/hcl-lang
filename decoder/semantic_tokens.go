// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"sort"

	"github.com/hashicorp/hcl-lang/decoder/internal/schemahelper"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// SemanticTokensInFile returns a sequence of semantic tokens
// within the config file.
func (d *PathDecoder) SemanticTokensInFile(ctx context.Context, filename string) ([]lang.SemanticToken, error) {
	f, err := d.fileByName(filename)
	if err != nil {
		return nil, err
	}

	body, err := d.bodyForFileAndPos(filename, f, hcl.InitialPos)
	if err != nil {
		return nil, err
	}

	if d.pathCtx.Schema == nil {
		return []lang.SemanticToken{}, nil
	}

	tokens := d.tokensForBody(ctx, body, d.pathCtx.Schema, []lang.SemanticTokenModifier{})

	// TODO decouple semantic tokens for valid references from AST walking
	//   instead of matching targets and origins when encountering a traversal expression,
	//   we can do this way earlier by comparing pathCtx.ReferenceTargets and
	//   d.pathCtx.ReferenceOrigins, to build a list of tokens.
	//   Be sure to sort them afterward!

	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].Range.Start.Byte < tokens[j].Range.Start.Byte
	})

	return tokens, nil
}

func (d *PathDecoder) tokensForBody(ctx context.Context, body *hclsyntax.Body, bodySchema *schema.BodySchema, parentModifiers []lang.SemanticTokenModifier) []lang.SemanticToken {
	tokens := make([]lang.SemanticToken, 0)

	if bodySchema == nil {
		return tokens
	}

	for name, attr := range body.Attributes {
		attrSchema, ok := bodySchema.Attributes[name]
		if !ok {
			if bodySchema.Extensions != nil && name == "count" && bodySchema.Extensions.Count {
				attrSchema = schemahelper.CountAttributeSchema()
			} else if bodySchema.Extensions != nil && name == "for_each" && bodySchema.Extensions.ForEach {
				attrSchema = schemahelper.ForEachAttributeSchema()
			} else {
				if bodySchema.AnyAttribute == nil {
					// unknown attribute
					continue
				}
				attrSchema = bodySchema.AnyAttribute
			}
		}

		attrModifiers := make([]lang.SemanticTokenModifier, 0)
		attrModifiers = append(attrModifiers, parentModifiers...)
		attrModifiers = append(attrModifiers, attrSchema.SemanticTokenModifiers...)

		tokens = append(tokens, lang.SemanticToken{
			Type:      lang.TokenAttrName,
			Modifiers: attrModifiers,
			Range:     attr.NameRange,
		})

		if attrSchema.Constraint != nil {
			tokens = append(tokens, d.newExpression(attr.Expr, attrSchema.Constraint).SemanticTokens(ctx)...)
		} else {
			ec := ExprConstraints(attrSchema.Expr)
			tokens = append(tokens, d.tokensForExpression(ctx, attr.Expr, ec)...)
		}
	}

	for _, block := range body.Blocks {
		blockSchema, hasDepSchema := bodySchema.Blocks[block.Type]
		if !hasDepSchema {
			// unknown block
			continue
		}

		blockModifiers := make([]lang.SemanticTokenModifier, 0)
		blockModifiers = append(blockModifiers, parentModifiers...)
		blockModifiers = append(blockModifiers, blockSchema.SemanticTokenModifiers...)

		tokens = append(tokens, lang.SemanticToken{
			Type:      lang.TokenBlockType,
			Modifiers: blockModifiers,
			Range:     block.TypeRange,
		})

		for i, labelRange := range block.LabelRanges {
			if i+1 > len(blockSchema.Labels) {
				// unknown label
				continue
			}

			labelSchema := blockSchema.Labels[i]

			labelModifiers := make([]lang.SemanticTokenModifier, 0)
			labelModifiers = append(labelModifiers, parentModifiers...)
			labelModifiers = append(labelModifiers, blockSchema.SemanticTokenModifiers...)
			labelModifiers = append(labelModifiers, labelSchema.SemanticTokenModifiers...)

			tokens = append(tokens, lang.SemanticToken{
				Type:      lang.TokenBlockLabel,
				Modifiers: labelModifiers,
				Range:     labelRange,
			})
		}

		if block.Body != nil {
			mergedSchema, _ := schemahelper.MergeBlockBodySchemas(block.AsHCLBlock(), blockSchema)

			tokens = append(tokens, d.tokensForBody(ctx, block.Body, mergedSchema, blockModifiers)...)
		}
	}

	return tokens
}

func (d *PathDecoder) tokensForExpression(ctx context.Context, expr hclsyntax.Expression, constraints ExprConstraints) []lang.SemanticToken {
	tokens := make([]lang.SemanticToken, 0)

	switch eType := expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		exprKeyword := eType.Traversal.RootName()
		kw, ok := constraints.KeywordExpr()
		if ok && len(eType.Traversal) == 1 && exprKeyword == kw.Keyword {
			return []lang.SemanticToken{
				{
					Type:      lang.TokenKeyword,
					Modifiers: []lang.SemanticTokenModifier{},
					Range:     eType.Range(),
				},
			}
		}

		_, ok = constraints.TraversalExprs()
		if ok && d.pathCtx.ReferenceTargets != nil && d.pathCtx.ReferenceOrigins != nil {
			traversal := eType.AsTraversal()

			origins, ok := d.pathCtx.ReferenceOrigins.AtPos(traversal.SourceRange().Filename, traversal.SourceRange().Start)
			if !ok {
				return tokens
			}

			targetFound := false
			for _, origin := range origins {
				matchableOrigin, ok := origin.(reference.MatchableOrigin)
				if !ok {
					continue
				}
				_, ok = d.pathCtx.ReferenceTargets.Match(matchableOrigin)
				if ok {
					targetFound = true
					break
				}
			}

			if !targetFound {
				return tokens
			}

			for _, t := range traversal {
				// TODO: Add meaning to each step/token?
				// This would require declaring the meaning in schema.AddrStep
				// and exposing it via lang.AddressStep
				// See https://github.com/hashicorp/vscode-terraform/issues/574

				switch ts := t.(type) {
				case hcl.TraverseRoot:
					tokens = append(tokens, lang.SemanticToken{
						Type:      lang.TokenTraversalStep,
						Modifiers: []lang.SemanticTokenModifier{},
						Range:     t.SourceRange(),
					})
				case hcl.TraverseAttr:
					rng := t.SourceRange()
					tokens = append(tokens, lang.SemanticToken{
						Type:      lang.TokenTraversalStep,
						Modifiers: []lang.SemanticTokenModifier{},
						Range: hcl.Range{
							Filename: rng.Filename,
							// omit the initial '.'
							Start: hcl.Pos{
								Line:   rng.Start.Line,
								Column: rng.Start.Column + 1,
								Byte:   rng.Start.Byte + 1,
							},
							End: rng.End,
						},
					})
				case hcl.TraverseIndex:
					// for index steps we only report
					// what's inside brackets
					rng := t.SourceRange()
					idxRange := hcl.Range{
						Filename: rng.Filename,
						Start: hcl.Pos{
							Line:   rng.Start.Line,
							Column: rng.Start.Column + 1,
							Byte:   rng.Start.Byte + 1,
						},
						End: hcl.Pos{
							Line:   rng.End.Line,
							Column: rng.End.Column - 1,
							Byte:   rng.End.Byte - 1,
						},
					}

					if ts.Key.Type() == cty.String {
						tokens = append(tokens, lang.SemanticToken{
							Type:      lang.TokenMapKey,
							Modifiers: []lang.SemanticTokenModifier{},
							Range:     idxRange,
						})
					}
					if ts.Key.Type() == cty.Number {
						tokens = append(tokens, lang.SemanticToken{
							Type:      lang.TokenNumber,
							Modifiers: []lang.SemanticTokenModifier{},
							Range:     idxRange,
						})
					}
				}

			}
		}
		_, ok = constraints.TypeDeclarationExpr()
		if ok && isPrimitiveTypeDeclaration(exprKeyword) {
			tokens = append(tokens, lang.SemanticToken{
				Type:      lang.TokenTypePrimitive,
				Modifiers: []lang.SemanticTokenModifier{},
				Range:     expr.Range(),
			})
		}
	case *hclsyntax.FunctionCallExpr:
		_, ok := constraints.TypeDeclarationExpr()
		if ok && isComplexTypeDeclaration(eType.Name) {
			tokens = append(tokens, lang.SemanticToken{
				Type:      lang.TokenTypeCapsule,
				Modifiers: []lang.SemanticTokenModifier{},
				Range:     eType.NameRange,
			})
			for _, arg := range eType.Args {
				tokens = append(tokens, d.tokensForExpression(ctx, arg, constraints)...)
			}
			return tokens
		}

	case *hclsyntax.TemplateExpr:
		// complex templates are not supported yet
		if !eType.IsStringLiteral() && !isMultilineStringLiteral(eType) {
			return tokens
		}
		if constraints.HasLiteralTypeOf(cty.String) {
			return tokenForTypedExpression(eType, cty.String)
		}
		literal := eType.Parts[0].(*hclsyntax.LiteralValueExpr)
		if constraints.HasLiteralValueOf(literal.Val) {
			return tokenForTypedExpression(eType, cty.String)
		}
	case *hclsyntax.TemplateWrapExpr:
		return d.tokensForExpression(ctx, eType.Wrapped, constraints)
	case *hclsyntax.TupleConsExpr:
		se, ok := constraints.SetExpr()
		if ok {
			ec := ExprConstraints(se.Elem)
			for _, expr := range eType.Exprs {
				tokens = append(tokens, d.tokensForExpression(ctx, expr, ec)...)
			}
			return tokens
		}
		le, ok := constraints.ListExpr()
		if ok {
			ec := ExprConstraints(le.Elem)
			for _, expr := range eType.Exprs {
				tokens = append(tokens, d.tokensForExpression(ctx, expr, ec)...)
			}
			return tokens
		}
		te, ok := constraints.TupleExpr()
		if ok {
			for i, expr := range eType.Exprs {
				if i >= len(te.Elems) {
					break
				}
				ec := ExprConstraints(te.Elems[i])
				tokens = append(tokens, d.tokensForExpression(ctx, expr, ec)...)
			}
			return tokens
		}
		lt, ok := constraints.LiteralTypeOfTupleExpr()
		if ok {
			return tokensForTupleConsExpr(eType, lt.Type)
		}
		lv, ok := constraints.LiteralValueOfTupleExpr(eType)
		if ok {
			return tokensForTupleConsExpr(eType, lv.Val.Type())
		}
	case *hclsyntax.ObjectConsExpr:
		oe, ok := constraints.ObjectExpr()
		if ok {
			for _, item := range eType.Items {
				key, _ := item.KeyExpr.Value(nil)
				if key.IsNull() || !key.IsWhollyKnown() || key.Type() != cty.String {
					// skip items keys that can't be interpolated
					// without further context
					continue
				}
				attr, ok := oe.Attributes[key.AsString()]
				if !ok {
					continue
				}

				tokens = append(tokens, lang.SemanticToken{
					Type:      lang.TokenObjectKey,
					Modifiers: []lang.SemanticTokenModifier{},
					Range:     item.KeyExpr.Range(),
				})

				ec := ExprConstraints(attr.Expr)
				tokens = append(tokens, d.tokensForExpression(ctx, item.ValueExpr, ec)...)
			}
			return tokens
		}
		me, ok := constraints.MapExpr()
		if ok {
			for _, item := range eType.Items {
				tokens = append(tokens, lang.SemanticToken{
					Type:      lang.TokenMapKey,
					Modifiers: []lang.SemanticTokenModifier{},
					Range:     item.KeyExpr.Range(),
				})
				ec := ExprConstraints(me.Elem)
				tokens = append(tokens, d.tokensForExpression(ctx, item.ValueExpr, ec)...)
			}
			return tokens
		}
		lt, ok := constraints.LiteralTypeOfObjectConsExpr()
		if ok {
			return tokensForObjectConsExpr(eType, lt.Type)
		}
		litVal, ok := constraints.LiteralValueOfObjectConsExpr(eType)
		if ok {
			return tokensForObjectConsExpr(eType, litVal.Val.Type())
		}
		_, ok = constraints.TypeDeclarationExpr()
		if ok {
			return d.tokensForObjectConsTypeDeclarationExpr(ctx, eType, constraints)
		}
	case *hclsyntax.LiteralValueExpr:
		valType := eType.Val.Type()
		if constraints.HasLiteralTypeOf(valType) {
			return tokenForTypedExpression(eType, valType)
		}
		if constraints.HasLiteralValueOf(eType.Val) {
			return tokenForTypedExpression(eType, valType)
		}
	}
	return tokens
}

func isPrimitiveTypeDeclaration(kw string) bool {
	switch kw {
	case "bool":
		return true
	case "number":
		return true
	case "string":
		return true
	case "null":
		return true
	case "any":
		return true
	}
	return false
}

func isComplexTypeDeclaration(funcName string) bool {
	switch funcName {
	case "list":
		return true
	case "set":
		return true
	case "tuple":
		return true
	case "map":
		return true
	case "object":
		return true
	}
	return false
}

func (d *PathDecoder) tokensForObjectConsTypeDeclarationExpr(ctx context.Context, expr *hclsyntax.ObjectConsExpr, constraints ExprConstraints) []lang.SemanticToken {
	tokens := make([]lang.SemanticToken, 0)
	for _, item := range expr.Items {
		key, _ := item.KeyExpr.Value(nil)
		if key.IsNull() || !key.IsWhollyKnown() || key.Type() != cty.String {
			// skip items keys that can't be interpolated
			// without further context
			continue
		}

		tokens = append(tokens, lang.SemanticToken{
			Type:      lang.TokenAttrName,
			Modifiers: []lang.SemanticTokenModifier{},
			Range:     item.KeyExpr.Range(),
		})

		tokens = append(tokens, d.tokensForExpression(ctx, item.ValueExpr, constraints)...)
	}
	return tokens
}

func tokenForTypedExpression(expr hclsyntax.Expression, consType cty.Type) []lang.SemanticToken {
	switch eType := expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		if consType.IsPrimitiveType() {
			return tokensForLiteralValueExpr(eType, consType)
		}
	case *hclsyntax.TemplateExpr:
		if eType.IsStringLiteral() {
			literal := eType.Parts[0].(*hclsyntax.LiteralValueExpr)
			if !literal.Val.Type().Equals(consType) {
				return []lang.SemanticToken{}
			}

			return []lang.SemanticToken{
				{
					Type:      lang.TokenString,
					Modifiers: []lang.SemanticTokenModifier{},
					Range:     expr.Range(),
				},
			}
		}
	case *hclsyntax.ObjectConsExpr:
		return tokensForObjectConsExpr(eType, consType)
	case *hclsyntax.TupleConsExpr:
		return tokensForTupleConsExpr(eType, consType)
	}

	return []lang.SemanticToken{}
}

func tokensForLiteralValueExpr(expr *hclsyntax.LiteralValueExpr, consType cty.Type) []lang.SemanticToken {
	tokens := make([]lang.SemanticToken, 0)

	if !expr.Val.Type().Equals(consType) {
		// type mismatch
		return tokens
	}

	switch consType {
	case cty.Bool:
		tokens = append(tokens, lang.SemanticToken{
			Type:      lang.TokenBool,
			Modifiers: []lang.SemanticTokenModifier{},
			Range:     expr.Range(),
		})
	case cty.String:
		tokens = append(tokens, lang.SemanticToken{
			Type:      lang.TokenString,
			Modifiers: []lang.SemanticTokenModifier{},
			Range:     expr.Range(),
		})
	case cty.Number:
		tokens = append(tokens, lang.SemanticToken{
			Type:      lang.TokenNumber,
			Modifiers: []lang.SemanticTokenModifier{},
			Range:     expr.Range(),
		})
	}

	return tokens
}

func tokensForObjectConsExpr(expr *hclsyntax.ObjectConsExpr, exprType cty.Type) []lang.SemanticToken {
	tokens := make([]lang.SemanticToken, 0)

	if exprType.IsObjectType() {
		attrTypes := exprType.AttributeTypes()
		for _, item := range expr.Items {
			key, _ := item.KeyExpr.Value(nil)
			if key.IsNull() || !key.IsWhollyKnown() || key.Type() != cty.String {
				// skip items keys that can't be interpolated
				// without further context
				continue
			}

			valType, ok := attrTypes[key.AsString()]
			if !ok {
				// unknown attribute
				continue
			}
			tokens = append(tokens, lang.SemanticToken{
				Type:      lang.TokenObjectKey,
				Modifiers: []lang.SemanticTokenModifier{},
				Range:     item.KeyExpr.Range(),
			})
			tokens = append(tokens, tokenForTypedExpression(item.ValueExpr, valType)...)
		}
	}
	if exprType.IsMapType() {
		elemType := *exprType.MapElementType()
		for _, item := range expr.Items {
			tokens = append(tokens, lang.SemanticToken{
				Type:      lang.TokenMapKey,
				Modifiers: []lang.SemanticTokenModifier{},
				Range:     item.KeyExpr.Range(),
			})
			tokens = append(tokens, tokenForTypedExpression(item.ValueExpr, elemType)...)
		}
	}

	return tokens
}

func tokensForTupleConsExpr(expr *hclsyntax.TupleConsExpr, exprType cty.Type) []lang.SemanticToken {
	tokens := make([]lang.SemanticToken, 0)

	for i, e := range expr.Exprs {
		var elemType cty.Type
		if exprType.IsListType() {
			elemType = *exprType.ListElementType()
		}
		if exprType.IsSetType() {
			elemType = *exprType.SetElementType()
		}
		if exprType.IsTupleType() {
			elTypes := exprType.TupleElementTypes()
			if len(elTypes) < i+1 {
				// skip unknown tuple element
				continue
			}
			elemType = elTypes[i]
		}

		tokens = append(tokens, tokenForTypedExpression(e, elemType)...)
	}

	return tokens
}

func semanticTokensForTraversalExpression(traversal hcl.Traversal) []lang.SemanticToken {
	if len(traversal) == 0 {
		return nil
	}

	tokens := make([]lang.SemanticToken, 0)

	tokens = append(tokens, lang.SemanticToken{
		Type:      lang.TokenTraversalStep,
		Modifiers: []lang.SemanticTokenModifier{},
		Range:     traversal[0].SourceRange(),
	})

	for i := 1; i < len(traversal); i++ {
		tokens = append(tokens, lang.SemanticToken{
			Type:      lang.TokenTraversalStep,
			Modifiers: []lang.SemanticTokenModifier{},
			Range: hcl.Range{
				Filename: traversal[i].SourceRange().Filename,
				Start: hcl.Pos{
					Line:   traversal[i].SourceRange().Start.Line,
					Column: traversal[i].SourceRange().Start.Column + 1,
					Byte:   traversal[i].SourceRange().Start.Byte + 1,
				},
				End: hcl.Pos{
					Line:   traversal[i].SourceRange().End.Line,
					Column: traversal[i].SourceRange().End.Column + 1,
					Byte:   traversal[i].SourceRange().End.Byte + 1,
				},
			},
		})
	}

	return tokens
}
