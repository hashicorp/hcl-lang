// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (td TypeDeclaration) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	switch eType := td.expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		if len(eType.Traversal) != 1 {
			return nil
		}

		if eType.Range().ContainsPos(pos) {
			typ, _ := typeexpr.TypeConstraint(eType)
			content, err := hoverContentForType(typ, 0)
			if err != nil {
				return nil
			}
			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   eType.Range(),
			}
		}
	case *hclsyntax.FunctionCallExpr:
		// position in complex type name
		if eType.NameRange.ContainsPos(pos) {
			typ, diags := typeexpr.TypeConstraint(eType)
			if len(diags) > 0 {
				return nil
			}

			content, err := hoverContentForType(typ, 0)
			if err != nil {
				return nil
			}
			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   eType.Range(),
			}
		}

		// position inside paranthesis
		if hcl.RangeBetween(eType.OpenParenRange, eType.CloseParenRange).ContainsPos(pos) {
			if isTypeNameWithElementOnly(eType.Name) {
				if len(eType.Args) == 0 {
					return nil
				}

				if len(eType.Args) == 1 && eType.Args[0].Range().ContainsPos(pos) {
					cons := TypeDeclaration{
						expr:    eType.Args[0],
						pathCtx: td.pathCtx,
					}
					return cons.HoverAtPos(ctx, pos)
				}

				return nil
			}

			if eType.Name == "object" {
				return td.objectHoverAtPos(ctx, eType, pos)
			}

			if eType.Name == "tuple" {
				return td.tupleHoverAtPos(ctx, eType, pos)
			}
		}
	}
	return nil
}

func (td TypeDeclaration) objectHoverAtPos(ctx context.Context, funcExpr *hclsyntax.FunctionCallExpr, pos hcl.Pos) *lang.HoverData {
	if len(funcExpr.Args) != 1 {
		return nil
	}
	objExpr, ok := funcExpr.Args[0].(*hclsyntax.ObjectConsExpr)
	if !ok {
		return nil
	}
	if !objExpr.Range().ContainsPos(pos) {
		return nil
	}

	// account for position on {} braces
	closeRange := hcl.Range{
		Filename: objExpr.Range().Filename,
		Start: hcl.Pos{
			Line:   objExpr.Range().End.Line,
			Column: objExpr.Range().End.Column - 1,
			Byte:   objExpr.Range().End.Byte - 1,
		},
		End: objExpr.Range().End,
	}
	if objExpr.OpenRange.ContainsPos(pos) || closeRange.ContainsPos(pos) {
		typ, diags := typeexpr.TypeConstraint(funcExpr)
		if len(diags) > 0 {
			return nil
		}
		content, err := hoverContentForType(typ, 0)
		if err != nil {
			return nil
		}
		return &lang.HoverData{
			Content: lang.Markdown(content),
			Range:   objExpr.Range(),
		}
	}

	for _, item := range objExpr.Items {
		if item.KeyExpr.Range().ContainsPos(pos) {
			rawKey, _, ok := rawObjectKey(item.KeyExpr)
			if !ok {
				// un-decodable key expression
				return nil
			}

			typ, _ := typeexpr.TypeConstraint(item.ValueExpr)
			return &lang.HoverData{
				Content: lang.Markdown(fmt.Sprintf("`%s` = _%s_", rawKey, typ.FriendlyNameForConstraint())),
				Range:   hcl.RangeBetween(item.KeyExpr.Range(), item.ValueExpr.Range()),
			}
		}
		if item.ValueExpr.Range().ContainsPos(pos) {
			cons := TypeDeclaration{
				expr:    item.ValueExpr,
				pathCtx: td.pathCtx,
			}
			return cons.HoverAtPos(ctx, pos)
		}
	}
	return nil
}

func (td TypeDeclaration) tupleHoverAtPos(ctx context.Context, funcExpr *hclsyntax.FunctionCallExpr, pos hcl.Pos) *lang.HoverData {
	if len(funcExpr.Args) != 1 {
		return nil
	}
	tupleExpr, ok := funcExpr.Args[0].(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil
	}
	if !tupleExpr.Range().ContainsPos(pos) {
		return nil
	}

	// account for position on [] brackets
	closeRange := hcl.Range{
		Filename: tupleExpr.Range().Filename,
		Start: hcl.Pos{
			Line:   tupleExpr.Range().End.Line,
			Column: tupleExpr.Range().End.Column - 1,
			Byte:   tupleExpr.Range().End.Byte - 1,
		},
		End: tupleExpr.Range().End,
	}
	if tupleExpr.OpenRange.ContainsPos(pos) || closeRange.ContainsPos(pos) {
		typ, diags := typeexpr.TypeConstraint(funcExpr)
		if len(diags) > 0 {
			return nil
		}
		content, err := hoverContentForType(typ, 0)
		if err != nil {
			return nil
		}
		return &lang.HoverData{
			Content: lang.Markdown(content),
			Range:   funcExpr.Range(),
		}
	}

	for _, expr := range tupleExpr.Exprs {
		if expr.Range().ContainsPos(pos) {
			cons := TypeDeclaration{
				expr:    expr,
				pathCtx: td.pathCtx,
			}
			return cons.HoverAtPos(ctx, pos)
		}
	}

	return nil
}
