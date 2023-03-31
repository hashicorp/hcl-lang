// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/zclconf/go-cty/cty"
)

func (ref Reference) ReferenceTargets(ctx context.Context, _ *TargetContext) reference.Targets {
	if ref.cons.Address == nil {
		return reference.Targets{}
	}

	// deal with native HCL syntax first
	eType, ok := ref.expr.(*hclsyntax.ScopeTraversalExpr)
	if ok {
		addr, err := lang.TraversalToAddress(eType.Traversal)
		if err != nil {
			return reference.Targets{}
		}

		return reference.Targets{
			reference.Target{
				Addr:     addr,
				ScopeId:  ref.cons.Address.ScopeId,
				RangePtr: eType.SrcRange.Ptr(),
				Name:     ref.cons.Name,
			},
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
				addr, err := lang.TraversalToAddress(vars[0])
				if err != nil {
					return reference.Targets{}
				}

				return reference.Targets{
					reference.Target{
						Addr:     addr,
						ScopeId:  ref.cons.Address.ScopeId,
						RangePtr: vars[0].SourceRange().Ptr(),
						Name:     ref.cons.Name,
					},
				}
			}
		}

		// Account for "legacy" string syntax which is still
		// in use by Terraform to date in this context.
		val, diags := ref.expr.Value(&hcl.EvalContext{})
		if diags.HasErrors() {
			return reference.Targets{}
		}
		if val.Type() != cty.String {
			return reference.Targets{}
		}
		startPos := hcl.Pos{
			Line: ref.expr.Range().Start.Line,
			// Account for the leading double quote
			Column: ref.expr.Range().Start.Column + 1,
			Byte:   ref.expr.Range().Start.Byte + 1,
		}

		traversal, diags := hclsyntax.ParseTraversalAbs([]byte(val.AsString()), ref.expr.Range().Filename, startPos)
		if diags.HasErrors() {
			return reference.Targets{}
		}
		addr, err := lang.TraversalToAddress(traversal)
		if err != nil {
			return reference.Targets{}
		}

		return reference.Targets{
			reference.Target{
				Addr:     addr,
				ScopeId:  ref.cons.Address.ScopeId,
				RangePtr: traversal.SourceRange().Ptr(),
				Name:     ref.cons.Name,
			},
		}
	}

	return reference.Targets{}
}
