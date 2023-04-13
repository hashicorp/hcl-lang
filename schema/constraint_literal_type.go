// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
)

// LiteralType represents literal type constraint
// e.g. any literal string ("foo"), number (42), etc.
//
// Non-literal expressions (even if these evaluate to the given
// type) are excluded.
//
// Complex types are supported, but dedicated List,
// Set, Map and other types are preferred, as these can
// convey more details, such as description, unlike
// e.g. LiteralType{Type: cty.List(...)}.
type LiteralType struct {
	Type cty.Type
	// TODO: object defaults

	// SkipComplexTypes avoids descending into complex literal types, such as {} and [].
	// It might be required when LiteralType is used in OneOf to avoid duplicates.
	SkipComplexTypes bool
}

func (LiteralType) isConstraintImpl() constraintSigil {
	return constraintSigil{}
}

func (lt LiteralType) FriendlyName() string {
	return lt.Type.FriendlyNameForConstraint()
}

func (lt LiteralType) Copy() Constraint {
	return LiteralType{
		Type:             lt.Type,
		SkipComplexTypes: lt.SkipComplexTypes,
	}
}

func (lt LiteralType) Validate() error {
	if lt.Type == cty.NilType {
		return errors.New("expected Type not to be nil")
	}
	return nil
}

func (lt LiteralType) EmptyCompletionData(ctx context.Context, nextPlaceholder int, nestingLevel int) CompletionData {
	if lt.Type.IsPrimitiveType() {
		var newText, snippet string

		switch lt.Type {
		case cty.Bool:
			newText = fmt.Sprintf("%t", false)
			// TODO: consider using snippet "choice"
			// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#snippet_syntax
			snippet = fmt.Sprintf("${%d:false}", nextPlaceholder)
		case cty.String:
			newText = fmt.Sprintf("%q", "value")
			snippet = fmt.Sprintf("\"${%d:%s}\"", nextPlaceholder, "value")
		case cty.Number:
			newText = "0"
			snippet = fmt.Sprintf("${%d:0}", nextPlaceholder)
		}

		nextPlaceholder++

		return CompletionData{
			NewText:         newText,
			Snippet:         snippet,
			NextPlaceholder: nextPlaceholder,
		}
	}

	if lt.Type.IsListType() {
		listCons := List{
			Elem: LiteralType{
				Type: lt.Type.ElementType(),
			},
		}
		return listCons.EmptyCompletionData(ctx, nextPlaceholder, nestingLevel)
	}
	if lt.Type.IsSetType() {
		setCons := Set{
			Elem: LiteralType{
				Type: lt.Type.ElementType(),
			},
		}
		return setCons.EmptyCompletionData(ctx, nextPlaceholder, nestingLevel)
	}
	if lt.Type.IsMapType() {
		mapCons := Map{
			Elem: LiteralType{
				Type: lt.Type.ElementType(),
			},
		}
		return mapCons.EmptyCompletionData(ctx, nextPlaceholder, nestingLevel)
	}
	if lt.Type.IsTupleType() {
		types := lt.Type.TupleElementTypes()
		tupleCons := Tuple{
			Elems: make([]Constraint, len(types)),
		}
		for i, typ := range types {
			tupleCons.Elems[i] = LiteralType{
				Type: typ,
			}
		}
		return tupleCons.EmptyCompletionData(ctx, nextPlaceholder, nestingLevel)
	}
	if lt.Type.IsObjectType() {
		attrTypes := lt.Type.AttributeTypes()
		attrs := make(ObjectAttributes, 0)
		for name, attrType := range attrTypes {
			aSchema := &AttributeSchema{
				Constraint: LiteralType{
					Type: attrType,
				},
			}
			if lt.Type.AttributeOptional(name) {
				aSchema.IsOptional = true
			} else {
				aSchema.IsRequired = true
			}

			attrs[name] = aSchema
		}
		cons := Object{
			Attributes: attrs,
		}
		return cons.EmptyCompletionData(ctx, nextPlaceholder, nestingLevel)
	}

	return CompletionData{
		NextPlaceholder: nextPlaceholder,
	}
}

func (lt LiteralType) EmptyHoverData(nestingLevel int) *HoverData {
	if lt.Type.IsPrimitiveType() {
		return &HoverData{
			Content: lang.Markdown(lt.Type.FriendlyNameForConstraint()),
		}
	}
	if lt.Type.IsListType() {
		cons := List{
			Elem: LiteralType{
				Type: lt.Type.ElementType(),
			},
		}
		return cons.EmptyHoverData(nestingLevel)
	}
	if lt.Type.IsSetType() {
		cons := Set{
			Elem: LiteralType{
				Type: lt.Type.ElementType(),
			},
		}
		return cons.EmptyHoverData(nestingLevel)
	}
	if lt.Type.IsMapType() {
		cons := Map{
			Elem: LiteralType{
				Type: lt.Type.ElementType(),
			},
		}
		return cons.EmptyHoverData(nestingLevel)
	}
	if lt.Type.IsTupleType() {
		types := lt.Type.TupleElementTypes()
		elemCons := make([]Constraint, len(types))
		for i, typ := range types {
			elemCons[i] = LiteralType{
				Type: typ,
			}
		}
		cons := Tuple{
			Elems: elemCons,
		}
		return cons.EmptyHoverData(nestingLevel)
	}
	if lt.Type.IsObjectType() {
		attrTypes := lt.Type.AttributeTypes()
		attrs := make(ObjectAttributes, 0)
		for name, attrType := range attrTypes {
			aSchema := &AttributeSchema{
				Constraint: LiteralType{
					Type: attrType,
				},
			}
			if lt.Type.AttributeOptional(name) {
				aSchema.IsOptional = true
			} else {
				aSchema.IsRequired = true
			}

			attrs[name] = aSchema
		}
		cons := Object{
			Attributes: attrs,
		}
		return cons.EmptyHoverData(nestingLevel)
	}

	return nil
}

func (lt LiteralType) ConstraintType() (cty.Type, bool) {
	return lt.Type, true
}
