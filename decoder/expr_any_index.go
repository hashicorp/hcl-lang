// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (a Any) complexIndex(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	var candidates []lang.Candidate

	cons := schema.AnyExpression{
		OfType: cty.String, // TODO improve type (could be int)
	}

	switch eType := a.expr.(type) {
	// An empty expression, e.g. `tags[]`, is a scope traversal expression
	// with an empty index step.
	case *hclsyntax.ScopeTraversalExpr:
		if len(eType.Traversal) < 2 {
			return candidates
		}
		// If the last part of the traversal is an index step,
		// we start a new completion to enable completion of
		// references and functions.
		lastTraversal := eType.Traversal[len(eType.Traversal)-1]
		if _, ok := lastTraversal.(hcl.TraverseIndex); ok {
			editRange := hcl.Range{
				Filename: eType.SrcRange.Filename,
				Start:    pos,
				End:      pos,
			}
			expr := &hclsyntax.LiteralValueExpr{
				SrcRange: editRange,
				Val:      cty.UnknownVal(cty.DynamicPseudoType),
			}
			return newExpression(a.pathCtx, expr, cons).CompletionAtPos(ctx, pos)
		}
	// If there is a prefix or valid expression within the index step,
	// we're dealing an index expression and can defer completion for the key.
	case *hclsyntax.IndexExpr:
		return newExpression(a.pathCtx, eType.Key, cons).CompletionAtPos(ctx, pos)
	}

	return candidates
}

func (a Any) hoverIndexExprAtPos(ctx context.Context, pos hcl.Pos) (*lang.HoverData, bool) {
	if eType, ok := a.expr.(*hclsyntax.IndexExpr); ok {
		if eType.Key.Range().ContainsPos(pos) {
			cons := schema.AnyExpression{
				OfType: cty.String, // TODO improve type (could be int)
			}
			return newExpression(a.pathCtx, eType.Key, cons).HoverAtPos(ctx, pos), true
		}
	}

	return nil, false
}
