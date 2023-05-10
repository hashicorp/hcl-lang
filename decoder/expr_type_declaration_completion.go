// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (td TypeDeclaration) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	if isEmptyExpression(td.expr) {
		editRange := hcl.Range{
			Filename: td.expr.Range().Filename,
			Start:    pos,
			End:      pos,
		}
		return allTypeDeclarationsAsCandidates("", editRange)
	}

	switch eType := td.expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		if len(eType.Traversal) != 1 {
			return []lang.Candidate{}
		}

		prefixLen := pos.Byte - eType.Range().Start.Byte
		if prefixLen > len(eType.Traversal.RootName()) {
			// The user has probably typed an extra character, such as a
			// period, that is not (yet) part of the expression. This prefix
			// won't match anything, so we'll return early.
			return []lang.Candidate{}
		}
		prefix := eType.Traversal.RootName()[0:prefixLen]

		editRange := hcl.Range{
			Filename: eType.Range().Filename,
			Start:    eType.Range().Start,
			End:      eType.Range().End,
		}

		return allTypeDeclarationsAsCandidates(prefix, editRange)
	case *hclsyntax.FunctionCallExpr:
		// position in complex type name
		if eType.NameRange.ContainsPos(pos) {
			prefixLen := pos.Byte - eType.NameRange.Start.Byte
			prefix := eType.Name[0:prefixLen]

			editRange := eType.Range()
			return allTypeDeclarationsAsCandidates(prefix, editRange)
		}

		// position inside paranthesis
		if hcl.RangeBetween(eType.OpenParenRange, eType.CloseParenRange).ContainsPos(pos) {
			if isTypeNameWithElementOnly(eType.Name) {
				if len(eType.Args) == 0 {
					editRange := hcl.Range{
						Filename: eType.Range().Filename,
						Start:    eType.OpenParenRange.End,
						End:      eType.CloseParenRange.Start,
					}

					return allTypeDeclarationsAsCandidates("", editRange)
				}

				if len(eType.Args) == 1 && eType.Args[0].Range().ContainsPos(pos) {
					cons := TypeDeclaration{
						expr:    eType.Args[0],
						pathCtx: td.pathCtx,
					}
					return cons.CompletionAtPos(ctx, pos)
				}

				return []lang.Candidate{}
			}

			if eType.Name == "object" {
				return td.objectCompletionAtPos(ctx, eType, pos)
			}

			if eType.Name == "tuple" {
				return td.tupleCompletionAtPos(ctx, eType, pos)
			}
		}
	}

	return []lang.Candidate{}
}

func (td TypeDeclaration) objectCompletionAtPos(ctx context.Context, funcExpr *hclsyntax.FunctionCallExpr, pos hcl.Pos) []lang.Candidate {
	if len(funcExpr.Args) == 0 {
		editRange := hcl.Range{
			Filename: funcExpr.Range().Filename,
			Start:    funcExpr.OpenParenRange.End,
			End:      funcExpr.CloseParenRange.Start,
		}

		return innerObjectTypeAsCompletionCandidates(editRange)
	}

	if len(funcExpr.Args) > 1 {
		return []lang.Candidate{}
	}

	objExpr, isObject := funcExpr.Args[0].(*hclsyntax.ObjectConsExpr)
	if !isObject {
		return []lang.Candidate{}
	}
	if !funcExpr.Args[0].Range().ContainsPos(pos) {
		return []lang.Candidate{}
	}

	editRange := hcl.Range{
		Filename: objExpr.Range().Filename,
		Start:    pos,
		End:      pos,
	}

	if len(objExpr.Items) == 0 {
		// check for incomplete configuration between {}
		betweenBraces := hcl.Range{
			Filename: objExpr.Range().Filename,
			Start:    objExpr.OpenRange.End,
			End:      pos,
		}
		fileBytes := td.pathCtx.Files[objExpr.Range().Filename].Bytes
		remainingBytes := bytes.TrimSpace(betweenBraces.SliceBytes(fileBytes))

		if len(remainingBytes) == 0 {
			return []lang.Candidate{
				objectAttributeItemAsCompletionCandidate(editRange),
			}
		}

		// if last byte is =, then it's incomplete attribute
		if remainingBytes[len(remainingBytes)-1] == '=' {
			// TODO: object optional+default
			return allTypeDeclarationsAsCandidates("", editRange)
		}
	}

	recoveryPos := objExpr.OpenRange.End
	var lastItemRange, nextItemRange *hcl.Range
	for _, item := range objExpr.Items {
		emptyRange := hcl.Range{
			Filename: objExpr.Range().Filename,
			Start:    item.KeyExpr.Range().End,
			End:      item.ValueExpr.Range().Start,
		}
		if emptyRange.ContainsPos(pos) {
			// exit early if we're in empty space between key and value
			return []lang.Candidate{}
		}

		// check if we've just missed the position
		if pos.Byte < item.KeyExpr.Range().Start.Byte {
			// enable recovery between last item's end and position

			// record current (next) item so we can avoid
			// completion on the same line
			nextItemRange = hcl.RangeBetween(item.KeyExpr.Range(), item.ValueExpr.Range()).Ptr()
			break
		}

		lastItemRange = hcl.RangeBetween(item.KeyExpr.Range(), item.ValueExpr.Range()).Ptr()
		recoveryPos = item.ValueExpr.Range().End

		if item.KeyExpr.Range().ContainsPos(pos) {
			return []lang.Candidate{}
		}
		if item.ValueExpr.Range().ContainsPos(pos) || item.ValueExpr.Range().End.Byte == pos.Byte {
			cons := TypeDeclaration{
				expr:    item.ValueExpr,
				pathCtx: td.pathCtx,
			}
			return cons.CompletionAtPos(ctx, pos)
		}
	}

	// check any incomplete configuration up to a terminating charactor
	fileBytes := td.pathCtx.Files[objExpr.Range().Filename].Bytes
	recoveredBytes := recoverLeftBytes(fileBytes, pos, func(offset int, r rune) bool {
		return isObjectItemTerminatingRune(r) && offset > recoveryPos.Byte
	})
	trimmedBytes := bytes.TrimRight(recoveredBytes, " \t")

	if len(trimmedBytes) == 0 {
		return []lang.Candidate{}
	}

	if len(trimmedBytes) == 1 && isObjectItemTerminatingRune(rune(trimmedBytes[0])) {
		// avoid completing on the same line as next item
		if nextItemRange != nil && nextItemRange.Start.Line == pos.Line {
			return []lang.Candidate{}
		}

		// avoid completing on the same line as last item
		if lastItemRange != nil && lastItemRange.End.Line == pos.Line {
			// if it is not single-line notation
			if trimmedBytes[0] != ',' {
				return []lang.Candidate{}
			}
		}

		return []lang.Candidate{
			objectAttributeItemAsCompletionCandidate(editRange),
		}
	}

	// if last byte is =, then it's incomplete attribute
	if trimmedBytes[len(trimmedBytes)-1] == '=' {
		// TODO: object optional+default
		return allTypeDeclarationsAsCandidates("", editRange)
	}

	return []lang.Candidate{}
}

func (td TypeDeclaration) tupleCompletionAtPos(ctx context.Context, funcExpr *hclsyntax.FunctionCallExpr, pos hcl.Pos) []lang.Candidate {
	if len(funcExpr.Args) == 0 {
		editRange := hcl.Range{
			Filename: funcExpr.Range().Filename,
			Start:    funcExpr.OpenParenRange.End,
			End:      funcExpr.CloseParenRange.Start,
		}

		return innerTupleTypeAsCompletionCandidates(editRange)
	}

	if len(funcExpr.Args) != 1 {
		// tuple types have to be wrapped in []
		return []lang.Candidate{}
	}

	tupleExpr, ok := funcExpr.Args[0].(*hclsyntax.TupleConsExpr)
	if !ok {
		return []lang.Candidate{}
	}

	for _, expr := range tupleExpr.Exprs {
		if expr.Range().ContainsPos(pos) || expr.Range().End.Byte == pos.Byte {
			cons := TypeDeclaration{
				expr:    expr,
				pathCtx: td.pathCtx,
			}
			return cons.CompletionAtPos(ctx, pos)
		}
	}

	betweenParens := hcl.Range{
		Filename: tupleExpr.Range().Filename,
		Start:    tupleExpr.OpenRange.End,
		End: hcl.Pos{
			Line: tupleExpr.SrcRange.End.Line,
			// shift left in front of the closing brace }
			Column: tupleExpr.SrcRange.End.Column - 1,
			Byte:   tupleExpr.SrcRange.End.Byte - 1,
		},
	}
	if betweenParens.ContainsPos(pos) || betweenParens.End.Byte == pos.Byte {
		editRange := hcl.Range{
			Filename: tupleExpr.Range().Filename,
			Start:    pos,
			End:      pos,
		}
		return allTypeDeclarationsAsCandidates("", editRange)
	}

	return []lang.Candidate{}
}

func allTypeDeclarationsAsCandidates(prefix string, editRange hcl.Range) []lang.Candidate {
	candidates := make([]lang.Candidate, 0)
	// TODO: any
	candidates = append(candidates, primitiveTypeDeclarationsAsCandidates(prefix, editRange)...)
	candidates = append(candidates, complexTypeDeclarationsAsCandidates(prefix, editRange)...)
	return candidates
}

func primitiveTypeDeclarationsAsCandidates(prefix string, editRange hcl.Range) []lang.Candidate {
	candidates := make([]lang.Candidate, 0)

	if strings.HasPrefix("bool", prefix) {
		candidates = append(candidates, lang.Candidate{
			Label:  cty.Bool.FriendlyNameForConstraint(),
			Detail: cty.Bool.FriendlyNameForConstraint(),
			Kind:   lang.BoolCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: "bool",
				Snippet: "bool",
				Range:   editRange,
			},
		})
	}
	if strings.HasPrefix("number", prefix) {
		candidates = append(candidates, lang.Candidate{
			Label:  cty.Number.FriendlyNameForConstraint(),
			Detail: cty.Number.FriendlyNameForConstraint(),
			Kind:   lang.NumberCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: "number",
				Snippet: "number",
				Range:   editRange,
			},
		})
	}
	if strings.HasPrefix("string", prefix) {
		candidates = append(candidates, lang.Candidate{
			Label:  cty.String.FriendlyNameForConstraint(),
			Detail: cty.String.FriendlyNameForConstraint(),
			Kind:   lang.StringCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: "string",
				Snippet: "string",
				Range:   editRange,
			},
		})
	}

	return candidates
}

func complexTypeDeclarationsAsCandidates(prefix string, editRange hcl.Range) []lang.Candidate {
	candidates := make([]lang.Candidate, 0)
	// TODO: indentation

	if strings.HasPrefix("list", prefix) {
		candidates = append(candidates, lang.Candidate{
			Label:  "list(…)",
			Detail: "list",
			Kind:   lang.ListCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: "list()",
				Snippet: fmt.Sprintf("list(${%d})", 0),
				Range:   editRange,
			},
			TriggerSuggest: true,
		})
	}
	if strings.HasPrefix("set", prefix) {
		candidates = append(candidates, lang.Candidate{
			Label:  "set(…)",
			Detail: "set",
			Kind:   lang.SetCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: "set()",
				Snippet: fmt.Sprintf("set(${%d})", 0),
				Range:   editRange,
			},
			TriggerSuggest: true,
		})
	}
	if strings.HasPrefix("tuple", prefix) {
		candidates = append(candidates, lang.Candidate{
			Label:  "tuple([…])",
			Detail: "tuple",
			Kind:   lang.TupleCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: "tuple([])",
				Snippet: fmt.Sprintf("tuple([ ${%d} ])", 0),
				Range:   editRange,
			},
			TriggerSuggest: true,
		})
	}
	if strings.HasPrefix("map", prefix) {
		candidates = append(candidates, lang.Candidate{
			Label:  "map(…)",
			Detail: "map",
			Kind:   lang.MapCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: "map()",
				Snippet: fmt.Sprintf("map(${%d})", 0),
				Range:   editRange,
			},
			TriggerSuggest: true,
		})
	}
	if strings.HasPrefix("object", prefix) {
		candidates = append(candidates, lang.Candidate{
			Label:  "object({…})",
			Detail: "object",
			Kind:   lang.ObjectCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: "object({\n\n})",
				Snippet: fmt.Sprintf("object({\n  ${%d:name} = ${%d}\n})", 1, 2),
				Range:   editRange,
			},
		})
	}

	return candidates
}

func objectAttributeItemAsCompletionCandidate(editRange hcl.Range) lang.Candidate {
	return lang.Candidate{
		Label:  "name = type",
		Detail: "type",
		Kind:   lang.AttributeCandidateKind,
		TextEdit: lang.TextEdit{
			NewText: "name = ",
			Snippet: fmt.Sprintf("${%d:name} = ", 1),
			Range:   editRange,
		},
	}
}

func innerObjectTypeAsCompletionCandidates(editRange hcl.Range) []lang.Candidate {
	return []lang.Candidate{
		{
			Label:  "{…}",
			Detail: "object",
			Kind:   lang.ObjectCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: "{\n\n}",
				Snippet: fmt.Sprintf("{\n  ${%d:name} = ${%d}\n}", 1, 2),
				Range:   editRange,
			},
		},
	}
}

func innerTupleTypeAsCompletionCandidates(editRange hcl.Range) []lang.Candidate {
	return []lang.Candidate{
		{
			Label:  "[…]",
			Detail: "tuple",
			Kind:   lang.TupleCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: "[]",
				Snippet: "[ ${0} ]",
				Range:   editRange,
			},
		},
	}
}
