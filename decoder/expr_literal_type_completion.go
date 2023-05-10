// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (lt LiteralType) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	typ := lt.cons.Type

	if isEmptyExpression(lt.expr) {
		editRange := hcl.Range{
			Filename: lt.expr.Range().Filename,
			Start:    pos,
			End:      pos,
		}

		if typ.IsPrimitiveType() {
			if typ == cty.Bool {
				return boolLiteralTypeCandidates("", editRange)
			}
			return []lang.Candidate{}
		}

		if typ == cty.DynamicPseudoType {
			return []lang.Candidate{}
		}

		if lt.cons.SkipComplexTypes {
			return []lang.Candidate{}
		}

		return []lang.Candidate{
			{
				Label:  labelForLiteralType(typ),
				Detail: typ.FriendlyName(),
				Kind:   candidateKindForType(typ),
				TextEdit: lang.TextEdit{
					Range:   editRange,
					NewText: newTextForLiteralType(typ),
					Snippet: snippetForLiteralType(1, typ),
				},
			},
		}
	}

	if typ == cty.Bool {
		return lt.completeBoolAtPos(ctx, pos)
	}

	if !lt.cons.SkipComplexTypes && typ.IsListType() {
		expr, ok := lt.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return []lang.Candidate{}
		}

		cons := schema.List{
			Elem: schema.LiteralType{
				Type: typ.ElementType(),
			},
		}

		return newExpression(lt.pathCtx, expr, cons).CompletionAtPos(ctx, pos)
	}

	if !lt.cons.SkipComplexTypes && typ.IsSetType() {
		expr, ok := lt.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return []lang.Candidate{}
		}

		cons := schema.Set{
			Elem: schema.LiteralType{
				Type: typ.ElementType(),
			},
		}

		return newExpression(lt.pathCtx, expr, cons).CompletionAtPos(ctx, pos)
	}

	if !lt.cons.SkipComplexTypes && typ.IsTupleType() {
		expr, ok := lt.expr.(*hclsyntax.TupleConsExpr)
		if !ok {
			return []lang.Candidate{}
		}

		elemTypes := typ.TupleElementTypes()
		cons := schema.Tuple{
			Elems: make([]schema.Constraint, len(elemTypes)),
		}
		for i, elemType := range elemTypes {
			cons.Elems[i] = schema.LiteralType{
				Type: elemType,
			}
		}

		return newExpression(lt.pathCtx, expr, cons).CompletionAtPos(ctx, pos)
	}

	if !lt.cons.SkipComplexTypes && typ.IsMapType() {
		expr, ok := lt.expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return []lang.Candidate{}
		}

		cons := schema.Map{
			Elem: schema.LiteralType{
				Type: typ.ElementType(),
			},
		}
		return newExpression(lt.pathCtx, expr, cons).CompletionAtPos(ctx, pos)
	}

	if !lt.cons.SkipComplexTypes && typ.IsObjectType() {
		expr, ok := lt.expr.(*hclsyntax.ObjectConsExpr)
		if !ok {
			return []lang.Candidate{}
		}

		cons := schema.Object{
			Attributes: ctyObjectToObjectAttributes(typ),
		}
		return newExpression(lt.pathCtx, expr, cons).CompletionAtPos(ctx, pos)
	}

	return []lang.Candidate{}
}

func (lt LiteralType) completeBoolAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	switch eType := lt.expr.(type) {

	case *hclsyntax.ScopeTraversalExpr:
		prefixLen := pos.Byte - eType.Range().Start.Byte
		if prefixLen > len(eType.Traversal.RootName()) {
			// The user has probably typed an extra character, such as a
			// period, that is not (yet) part of the expression. This prefix
			// won't match anything, so we'll return early.
			return []lang.Candidate{}
		}
		prefix := eType.Traversal.RootName()[0:prefixLen]
		return boolLiteralTypeCandidates(prefix, eType.Range())

	case *hclsyntax.LiteralValueExpr:
		if eType.Val.Type() == cty.Bool {
			value := "false"
			if eType.Val.True() {
				value = "true"
			}
			prefixLen := pos.Byte - eType.Range().Start.Byte
			prefix := value[0:prefixLen]
			return boolLiteralTypeCandidates(prefix, eType.Range())
		}
	}

	return []lang.Candidate{}
}

func boolLiteralTypeCandidates(prefix string, editRange hcl.Range) []lang.Candidate {
	candidates := make([]lang.Candidate, 0)

	if strings.HasPrefix("false", prefix) {
		candidates = append(candidates, lang.Candidate{
			Label:  "false",
			Detail: cty.Bool.FriendlyNameForConstraint(),
			Kind:   lang.BoolCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: "false",
				Snippet: "false",
				Range:   editRange,
			},
		})
	}
	if strings.HasPrefix("true", prefix) {
		candidates = append(candidates, lang.Candidate{
			Label:  "true",
			Detail: cty.Bool.FriendlyNameForConstraint(),
			Kind:   lang.BoolCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: "true",
				Snippet: "true",
				Range:   editRange,
			},
		})
	}

	return candidates
}

func ctyObjectToObjectAttributes(objType cty.Type) schema.ObjectAttributes {
	attrTypes := objType.AttributeTypes()
	objAttributes := make(schema.ObjectAttributes, len(attrTypes))

	for name, attrType := range attrTypes {
		aSchema := &schema.AttributeSchema{
			Constraint: schema.LiteralType{
				Type: attrType,
			},
		}
		if objType.AttributeOptional(name) {
			aSchema.IsOptional = true
		} else {
			aSchema.IsRequired = true
		}
		objAttributes[name] = aSchema
	}

	return objAttributes
}
