// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (ref Reference) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	if ref.cons.Address != nil {
		// no candidates if traversal itself is addressable
		return []lang.Candidate{}
	}

	if ref.pathCtx.ReferenceTargets == nil {
		return []lang.Candidate{}
	}

	file := ref.pathCtx.Files[ref.expr.Range().Filename]
	rootBody, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return []lang.Candidate{}
	}

	outerBodyRng := rootBody.Range()
	// Find outer block body range to allow filtering
	// of references pointing back to the same block
	outerBlock := rootBody.OutermostBlockAtPos(pos)
	if outerBlock != nil {
		ob := outerBlock.Body.(*hclsyntax.Body)
		outerBodyRng = ob.Range()
	}

	if isEmptyExpression(ref.expr) {
		editRng := hcl.Range{
			Filename: ref.expr.Range().Filename,
			Start:    pos,
			End:      pos,
		}
		candidates := make([]lang.Candidate, 0)
		ref.pathCtx.ReferenceTargets.MatchWalk(ctx, ref.cons, "", outerBodyRng, editRng, func(target reference.Target) error {
			address := target.Address(ctx, editRng.Start).String()

			candidates = append(candidates, lang.Candidate{
				Label:       address,
				Detail:      target.FriendlyName(),
				Description: target.Description,
				Kind:        lang.TraversalCandidateKind,
				TextEdit: lang.TextEdit{
					NewText: address,
					Snippet: address,
					Range:   editRng,
				},
			})
			return nil
		})
		return candidates
	}

	eType, ok := ref.expr.(*hclsyntax.ScopeTraversalExpr)
	if !ok {
		return []lang.Candidate{}
	}

	editRng := eType.Range()
	if !editRng.ContainsPos(pos) {
		// account for trailing character(s) which doesn't appear in AST
		// such as dot, opening bracket etc.
		editRng.End = pos
	}
	prefixRng := hcl.Range{
		Filename: eType.Range().Filename,
		Start:    eType.Range().Start,
		End:      pos,
	}
	prefix := string(prefixRng.SliceBytes(file.Bytes))

	candidates := make([]lang.Candidate, 0)
	ref.pathCtx.ReferenceTargets.MatchWalk(ctx, ref.cons, prefix, outerBodyRng, editRng, func(target reference.Target) error {
		address := target.Address(ctx, editRng.Start).String()

		candidates = append(candidates, lang.Candidate{
			Label:       address,
			Detail:      target.FriendlyName(),
			Description: target.Description,
			Kind:        lang.TraversalCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: address,
				Snippet: address,
				Range:   editRng,
			},
		})
		return nil
	})
	return candidates
}
