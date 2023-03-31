// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"bytes"
	"context"
	"sort"
	"strings"
	"unicode"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type declaredAttributes map[string]hcl.Range

func (obj Object) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	if isEmptyExpression(obj.expr) {
		cData := obj.cons.EmptyCompletionData(ctx, 1, 0)
		return []lang.Candidate{
			{ // TODO: Consider rendering first N elements in Label?
				Label:       "{…}",
				Detail:      "object",
				Kind:        lang.ObjectCandidateKind,
				Description: obj.cons.Description,
				TextEdit: lang.TextEdit{
					NewText: cData.NewText,
					Snippet: cData.Snippet,
					Range: hcl.Range{
						Filename: obj.expr.Range().Filename,
						Start:    pos,
						End:      pos,
					},
				},
				TriggerSuggest: cData.TriggerSuggest,
			},
		}
	}

	eType, ok := obj.expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return []lang.Candidate{}
	}

	betweenBraces := hcl.Range{
		Filename: eType.Range().Filename,
		Start:    eType.OpenRange.End,
		End:      eType.Range().End,
	}

	if !betweenBraces.ContainsPos(pos) {
		return []lang.Candidate{}
	}

	if len(obj.cons.Attributes) == 0 {
		return []lang.Candidate{}
	}

	editRange := hcl.Range{
		Filename: eType.Range().Filename,
		Start:    pos,
		End:      pos,
	}

	declared := make(declaredAttributes, 0)
	recoveryPos := eType.OpenRange.Start
	var lastItemRange, nextItemRange *hcl.Range

	for _, item := range eType.Items {
		emptyRange := hcl.Range{
			Filename: eType.Range().Filename,
			Start:    item.KeyExpr.Range().End,
			End:      item.ValueExpr.Range().Start,
		}
		if emptyRange.ContainsPos(pos) {
			// exit early if we're in empty space between key and value
			return []lang.Candidate{}
		}

		attrName, attrRange, ok := rawObjectKey(item.KeyExpr)
		if !ok {
			continue
		}

		// collect all declared attributes
		declared[attrName] = hcl.RangeBetween(item.KeyExpr.Range(), item.ValueExpr.Range())

		if nextItemRange != nil {
			continue
		}
		// check if we've just missed the position
		if pos.Byte < item.KeyExpr.Range().Start.Byte {
			// record current (next) item so we can avoid completion
			// on the same line in multi-line mode (without comma)
			nextItemRange = hcl.RangeBetween(item.KeyExpr.Range(), item.ValueExpr.Range()).Ptr()

			// enable recovery of incomplete configuration
			// between last item's end and position
			continue
		}
		lastItemRange = hcl.RangeBetween(item.KeyExpr.Range(), item.ValueExpr.Range()).Ptr()
		recoveryPos = item.ValueExpr.Range().End

		if item.KeyExpr.Range().ContainsPos(pos) {
			prefix := ""

			// if we're before start of the attribute
			// it means the attribute is likely quoted
			if pos.Byte >= attrRange.Start.Byte {
				prefixLen := pos.Byte - attrRange.Start.Byte
				prefix = attrName[0:prefixLen]
			}

			editRange := hcl.RangeBetween(item.KeyExpr.Range(), item.ValueExpr.Range())

			return objectAttributesToCandidates(ctx, prefix, obj.cons.Attributes, declared, editRange)
		}
		if item.ValueExpr.Range().ContainsPos(pos) || item.ValueExpr.Range().End.Byte == pos.Byte {
			aSchema, ok := obj.cons.Attributes[attrName]
			if !ok {
				// unknown attribute
				return []lang.Candidate{}
			}

			cons := newExpression(obj.pathCtx, item.ValueExpr, aSchema.Constraint)

			return cons.CompletionAtPos(ctx, pos)
		}
	}

	// check any incomplete configuration up to a terminating character
	fileBytes := obj.pathCtx.Files[eType.Range().Filename].Bytes
	leftBytes := recoverLeftBytes(fileBytes, pos, func(offset int, r rune) bool {
		return isObjectItemTerminatingRune(r) && offset > recoveryPos.Byte
	})
	trimmedBytes := bytes.TrimRight(leftBytes, " \t")

	if len(trimmedBytes) == 0 {
		// no terminating character was found which indicates
		// we're on the same line as an existing item
		// and we're missing preceding comma
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

		return objectAttributesToCandidates(ctx, "", obj.cons.Attributes, declared, editRange)
	}

	// trime left side as well now
	// to make prefix/attribute extraction easier below
	trimmedBytes = bytes.TrimLeftFunc(trimmedBytes, func(r rune) bool {
		return isObjectItemTerminatingRune(r) || unicode.IsSpace(r)
	})

	// if last byte is =, then it's incomplete attribute
	if len(trimmedBytes) > 0 && trimmedBytes[len(trimmedBytes)-1] == '=' {
		emptyExpr := newEmptyExpressionAtPos(eType.Range().Filename, pos)

		attrName := string(bytes.TrimFunc(trimmedBytes[:len(trimmedBytes)-1], func(r rune) bool {
			return unicode.IsSpace(r) || r == '"'
		}))
		aSchema, ok := obj.cons.Attributes[attrName]
		if !ok {
			// unknown attribute
			return []lang.Candidate{}
		}

		cons := newExpression(obj.pathCtx, emptyExpr, aSchema.Constraint)

		return cons.CompletionAtPos(ctx, pos)
	}

	prefix := string(bytes.TrimFunc(trimmedBytes, func(r rune) bool {
		return unicode.IsSpace(r) || r == '"'
	}))

	// calculate appropriate edit range in case there
	// are also characters on the right from position
	// which are worth replacing
	remainingRange := hcl.Range{
		Filename: eType.Range().Filename,
		Start:    pos,
		End:      eType.SrcRange.End,
	}
	editRange = objectItemPrefixBasedEditRange(remainingRange, fileBytes, trimmedBytes)

	return objectAttributesToCandidates(ctx, prefix, obj.cons.Attributes, declared, editRange)
}

func objectItemPrefixBasedEditRange(remainingRange hcl.Range, fileBytes []byte, rawPrefixBytes []byte) hcl.Range {
	remainingBytes := remainingRange.SliceBytes(fileBytes)
	roughEndByteOffset := bytes.IndexFunc(remainingBytes, func(r rune) bool {
		return r == '\n' || r == '}'
	})
	// avoid editing over whitespace
	trimmedRightBytes := bytes.TrimRightFunc(remainingBytes[:roughEndByteOffset], func(r rune) bool {
		return unicode.IsSpace(r)
	})
	trimmedOffset := len(trimmedRightBytes)

	return hcl.Range{
		Filename: remainingRange.Filename,
		Start: hcl.Pos{
			// TODO: Calculate Line+Column for multi-line keys?
			Line:   remainingRange.Start.Line,
			Column: remainingRange.Start.Column - len(rawPrefixBytes),
			Byte:   remainingRange.Start.Byte - len(rawPrefixBytes),
		},
		End: hcl.Pos{
			// TODO: Calculate Line+Column for multi-line values?
			Line:   remainingRange.Start.Line,
			Column: remainingRange.Start.Column + trimmedOffset,
			Byte:   remainingRange.Start.Byte + trimmedOffset,
		},
	}
}

func objectAttributesToCandidates(ctx context.Context, prefix string, attrs schema.ObjectAttributes, declared declaredAttributes, editRange hcl.Range) []lang.Candidate {
	if len(attrs) == 0 {
		return []lang.Candidate{}
	}

	candidates := make([]lang.Candidate, 0)

	attrNames := sortedObjectAttributeNames(attrs)

	for _, name := range attrNames {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		// avoid suggesting already declared attribute
		// unless we're overriding it
		if declaredRng, ok := declared[name]; ok && !declaredRng.Overlaps(editRange) {
			continue
		}

		candidates = append(candidates, attributeSchemaToCandidate(ctx, name, attrs[name], editRange))
	}

	return candidates
}

func sortedObjectAttributeNames(objAttributes schema.ObjectAttributes) []string {
	names := make([]string, 0, len(objAttributes))
	for name := range objAttributes {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}
