// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/zclconf/go-cty/cty"
)

func (ref Reference) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	// deal with native HCL syntax first
	te, ok := ref.expr.(*hclsyntax.ScopeTraversalExpr)
	if ok {
		origin, ok := reference.TraversalToLocalOrigin(te.Traversal, originConstraintsFromCons(ref.cons), allowSelfRefs)
		if ok {
			return reference.Origins{origin}
		}
	}

	if json.IsJSONExpression(ref.expr) {
		// Given the limited AST/API access to JSON we can only
		// guess whether the expression has exactly a single traversal
		vars := ref.expr.Variables()
		if len(vars) == 1 {
			tRange := vars[0].SourceRange()
			expectedExprRange := hcl.Range{
				Filename: tRange.Filename,
				Start: hcl.Pos{
					Line: tRange.Start.Line,
					// account for "${
					Column: tRange.Start.Column - 3,
					Byte:   tRange.Start.Byte - 3,
				},
				End: hcl.Pos{
					Line: tRange.End.Line,
					// account for }"
					Column: tRange.End.Column + 2,
					Byte:   tRange.End.Byte + 2,
				},
			}

			if rangesEqual(expectedExprRange, ref.expr.Range()) {
				origin, ok := reference.TraversalToLocalOrigin(vars[0], originConstraintsFromCons(ref.cons), allowSelfRefs)
				if ok {
					return reference.Origins{origin}
				}
			}
		}

		// Account for "legacy" string syntax which is still
		// in use by Terraform to date in this context.
		val, diags := ref.expr.Value(nil)
		if diags.HasErrors() {
			return reference.Origins{}
		}
		if val.Type() != cty.String {
			return reference.Origins{}
		}
		startPos := hcl.Pos{
			Line: ref.expr.Range().Start.Line,
			// Account for the leading double quote
			Column: ref.expr.Range().Start.Column + 1,
			Byte:   ref.expr.Range().Start.Byte + 1,
		}

		traversal, diags := hclsyntax.ParseTraversalAbs([]byte(val.AsString()), ref.expr.Range().Filename, startPos)
		if diags.HasErrors() {
			return reference.Origins{}
		}
		origin, ok := reference.TraversalToLocalOrigin(traversal, originConstraintsFromCons(ref.cons), allowSelfRefs)
		if ok {
			return reference.Origins{origin}
		}
	}

	return reference.Origins{}
}

func rangesEqual(first, second hcl.Range) bool {
	return posEqual(first.Start, second.Start) && posEqual(first.End, second.End)
}

func originConstraintsFromCons(cons schema.Reference) reference.OriginConstraints {
	if cons.Address != nil {
		// skip traversals which are targets by themselves (not origins)
		return reference.OriginConstraints{}
	}

	// TODO: Remove condition once legacy tests are gone
	// This is being flagged up as invalid schema
	// but we tolerate it for legacy tests
	if cons.OfType == cty.NilType && cons.OfScopeId == "" {
		return reference.OriginConstraints{}
	}

	return reference.OriginConstraints{
		{
			OfType:    cons.OfType,
			OfScopeId: cons.OfScopeId,
		},
	}
}
