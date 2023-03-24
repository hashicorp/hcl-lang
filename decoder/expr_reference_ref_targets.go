package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
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
		if len(vars) != 1 {
			return reference.Targets{}
		}

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

	return reference.Targets{}
}
