// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
)

type functionExpr struct {
	expr       hcl.Expression
	returnType cty.Type
	pathCtx    *PathContext
}

func (fe functionExpr) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	if isEmptyExpression(fe.expr) {
		editRange := hcl.Range{
			Filename: fe.expr.Range().Filename,
			Start:    pos,
			End:      pos,
		}

		return fe.matchingFunctions("", editRange)
	}

	switch eType := fe.expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		if len(eType.Traversal) > 1 {
			// We assume that function names cannot contain dots
			return []lang.Candidate{}
		}

		prefixLen := pos.Byte - eType.Traversal.SourceRange().Start.Byte
		rootName := eType.Traversal.RootName()

		// There can be a single segment with trailing dot which cannot
		// be a function anymore as functions cannot contain dots.
		if prefixLen < 0 || prefixLen > len(rootName) {
			return []lang.Candidate{}
		}

		prefix := rootName[0:prefixLen]
		return fe.matchingFunctions(prefix, eType.Range())

	case *hclsyntax.ExprSyntaxError:
		// Note: this range can range up until the end of the file in case of invalid config
		if eType.SrcRange.ContainsPos(pos) {
			// we are somewhere in the range for this attribute but we don't have an expression range to check
			// so we look back to check whether we are in a partially written provider defined function
			fileBytes := fe.pathCtx.Files[eType.SrcRange.Filename].Bytes

			recoveredPrefixBytes := recoverLeftBytes(fileBytes, pos, func(offset int, r rune) bool {
				return !isNamespacedFunctionNameRune(r)
			})
			// recoveredPrefixBytes also contains the rune before the function name, so we need to trim it
			_, lengthFirstRune := utf8.DecodeRune(recoveredPrefixBytes)
			recoveredPrefixBytes = recoveredPrefixBytes[lengthFirstRune:]

			recoveredSuffixBytes := recoverRightBytes(fileBytes, pos, func(offset int, r rune) bool {
				return !isNamespacedFunctionNameRune(r) && r != '('
			})
			// recoveredSuffixBytes also contains the rune after the function name, so we need to trim it
			_, lengthLastRune := utf8.DecodeLastRune(recoveredSuffixBytes)
			recoveredSuffixBytes = recoveredSuffixBytes[:len(recoveredSuffixBytes)-lengthLastRune]

			recoveredIdentifier := append(recoveredPrefixBytes, recoveredSuffixBytes...)

			// check if our recovered identifier contains "::"
			// Why two colons? For no colons the parser would return a traversal expression
			// and a single colon will apparently be treated as a traversal and a partial object expression
			// (refer to this follow-up issue for more on that case: https://github.com/hashicorp/vscode-terraform/issues/1697)
			if bytes.Contains(recoveredIdentifier, []byte("::")) {
				editRange := hcl.Range{
					Filename: fe.expr.Range().Filename,
					Start: hcl.Pos{
						Line:   pos.Line, // we don't recover newlines, so we can keep the original line number
						Byte:   pos.Byte - len(recoveredPrefixBytes),
						Column: pos.Column - len(recoveredPrefixBytes),
					},
					End: hcl.Pos{
						Line:   pos.Line,
						Byte:   pos.Byte + len(recoveredSuffixBytes),
						Column: pos.Column + len(recoveredSuffixBytes),
					},
				}

				return fe.matchingFunctions(string(recoveredPrefixBytes), editRange)
			}
		}

		return []lang.Candidate{}

	case *hclsyntax.FunctionCallExpr:
		if eType.NameRange.ContainsPos(pos) {
			prefixLen := pos.Byte - eType.NameRange.Start.Byte
			prefix := eType.Name[0:prefixLen]
			editRange := eType.Range()
			return fe.matchingFunctions(prefix, editRange)
		}

		f, ok := fe.pathCtx.Functions[eType.Name]
		if !ok {
			return []lang.Candidate{} // Unknown function
		}

		parensRange := hcl.RangeBetween(eType.OpenParenRange, eType.CloseParenRange)
		if !parensRange.ContainsPos(pos) {
			return []lang.Candidate{} // Not inside parenthesis
		}

		paramsLen := len(f.Params)
		if paramsLen == 0 && f.VarParam == nil {
			return []lang.Candidate{} // Function accepts no parameters
		}

		var lastArgExpr hcl.Expression
		lastArgEndPos := eType.OpenParenRange.Start
		lastArgIdx := 0
		for i, arg := range eType.Args {
			// We overshot the argument and stop
			if arg.Range().Start.Byte > pos.Byte {
				break
			}
			if arg.Range().ContainsPos(pos) || arg.Range().End.Byte == pos.Byte {
				var param function.Parameter
				if i < paramsLen {
					param = f.Params[i]
				} else if f.VarParam != nil {
					param = *f.VarParam
				} else {
					// Too many arguments passed to the function
					return []lang.Candidate{}
				}

				cons := newExpression(fe.pathCtx, arg, schema.AnyExpression{OfType: param.Type})
				return cons.CompletionAtPos(ctx, pos)
			}
			lastArgExpr = arg
			lastArgEndPos = arg.Range().End
			lastArgIdx = i
		}

		fileBytes := fe.pathCtx.Files[eType.Range().Filename].Bytes
		recoveredBytes := recoverLeftBytes(fileBytes, pos, func(byteOffset int, r rune) bool {
			return (r == ',' || r == '(') && byteOffset > lastArgEndPos.Byte
		})
		trimmedBytes := bytes.TrimRight(recoveredBytes, " \t\n")

		activePar := lastArgIdx // default to last seen parameter
		elemExpr := newEmptyExpressionAtPos(fe.expr.Range().Filename, pos)
		if string(trimmedBytes) == "," {
			activePar = lastArgIdx + 1
		} else if len(recoveredBytes) == 0 && lastArgExpr != nil {
			// We were unable to recover any bytes, which suggests
			// we're still (partially) completing the last argument.
			// A common case is trailing dot in func_call(var.foo.)
			elemExpr = lastArgExpr
		}

		var param function.Parameter
		if activePar < paramsLen {
			param = f.Params[activePar]
		} else if f.VarParam != nil {
			param = *f.VarParam
		} else {
			// Too many arguments passed to the function
			return []lang.Candidate{}
		}

		cons := newExpression(fe.pathCtx, elemExpr, schema.AnyExpression{OfType: param.Type})
		return cons.CompletionAtPos(ctx, pos)
	}
	return []lang.Candidate{}
}

func (fe functionExpr) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	funcExpr, ok := fe.expr.(*hclsyntax.FunctionCallExpr)
	if !ok {
		return nil
	}

	funcSig, ok := fe.pathCtx.Functions[funcExpr.Name]
	if !ok {
		return nil
	}

	if funcExpr.NameRange.ContainsPos(pos) {
		return &lang.HoverData{
			Content: hoverContentForFunction(funcExpr.Name, funcSig),
			Range:   fe.expr.Range(),
		}
	}

	paramsLen := len(funcSig.Params)
	if paramsLen == 0 && funcSig.VarParam == nil {
		return nil // Function accepts no parameters
	}

	for i, arg := range funcExpr.Args {
		var param function.Parameter
		if i < paramsLen {
			param = funcSig.Params[i]
		} else if funcSig.VarParam != nil {
			param = *funcSig.VarParam
		} else {
			// Too many arguments passed to the function
			return nil
		}

		if arg.Range().ContainsPos(pos) {
			return newExpression(fe.pathCtx, arg, schema.AnyExpression{
				OfType: param.Type,
			}).HoverAtPos(ctx, pos)
		}
	}

	return nil
}

func (fe functionExpr) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	funcExpr, ok := fe.expr.(*hclsyntax.FunctionCallExpr)
	if !ok {
		return []lang.SemanticToken{}
	}
	funcSig, ok := fe.pathCtx.Functions[funcExpr.Name]
	if !ok {
		return []lang.SemanticToken{}
	}

	tokens := make([]lang.SemanticToken, 0)

	tokens = append(tokens, lang.SemanticToken{
		Type:      lang.TokenFunctionName,
		Modifiers: []lang.SemanticTokenModifier{},
		Range:     funcExpr.NameRange,
	})

	paramsLen := len(funcSig.Params)
	if paramsLen == 0 && funcSig.VarParam == nil {
		return tokens // Function accepts no parameters
	}

	for i, arg := range funcExpr.Args {
		var param function.Parameter
		if i < paramsLen {
			param = funcSig.Params[i]
		} else if funcSig.VarParam != nil {
			param = *funcSig.VarParam
		} else {
			// Too many arguments passed to the function
			break
		}

		tokens = append(tokens, newExpression(fe.pathCtx, arg, schema.AnyExpression{
			OfType: param.Type,
		}).SemanticTokens(ctx)...)
	}

	return tokens
}

func (fe functionExpr) ReferenceOrigins(ctx context.Context) reference.Origins {
	funcExpr, diags := hcl.ExprCall(fe.expr)
	if diags.HasErrors() {
		return reference.Origins{}
	}

	funcSig, ok := fe.pathCtx.Functions[funcExpr.Name]
	if !ok {
		return nil
	}

	paramsLen := len(funcSig.Params)
	if paramsLen == 0 && funcSig.VarParam == nil {
		return nil // Function accepts no parameters
	}

	origins := make(reference.Origins, 0)
	for i, arg := range funcExpr.Arguments {
		var param function.Parameter
		if i < paramsLen {
			param = funcSig.Params[i]
		} else if funcSig.VarParam != nil {
			param = *funcSig.VarParam
		} else {
			// Too many arguments passed to the function
			break
		}

		expr := Any{
			pathCtx: fe.pathCtx,
			expr:    arg,
			cons: schema.AnyExpression{
				OfType: param.Type,
			},
		}
		origins = append(origins, expr.ReferenceOrigins(ctx)...)
	}

	return origins
}

func (fe functionExpr) matchingFunctions(prefix string, editRange hcl.Range) []lang.Candidate {
	candidates := make([]lang.Candidate, 0)

	for name, f := range fe.pathCtx.Functions {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		// Reject functions that have a non-convertible return type
		if _, err := convert.Convert(cty.UnknownVal(f.ReturnType), fe.returnType); err != nil {
			continue
		}

		candidates = append(candidates, lang.Candidate{
			Label:       name,
			Detail:      fmt.Sprintf("%s(%s) %s", name, parameterNamesAsString(f), f.ReturnType.FriendlyName()),
			Kind:        lang.FunctionCandidateKind,
			Description: lang.Markdown(f.Description),
			TextEdit: lang.TextEdit{
				NewText: fmt.Sprintf("%s()", name),
				Snippet: fmt.Sprintf("%s(${0})", name),
				Range:   editRange,
			},
		})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Label < candidates[j].Label
	})

	return candidates
}

func hoverContentForFunction(name string, funcSig schema.FunctionSignature) lang.MarkupContent {
	rawMd := fmt.Sprintf("```terraform\n%s(%s) %s\n```\n\n%s",
		name, parameterNamesAsString(funcSig), funcSig.ReturnType.FriendlyName(), funcSig.Description)

	if funcSig.Detail != "" {
		rawMd += fmt.Sprintf("\n\n%s", funcSig.Detail)
	}

	return lang.Markdown(rawMd)
}
