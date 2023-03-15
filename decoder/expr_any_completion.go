package decoder

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty/function"
)

func (a Any) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	if isEmptyExpression(a.expr) {
		editRange := hcl.Range{
			Filename: a.expr.Range().Filename,
			Start:    pos,
			End:      pos,
		}

		candidates := make([]lang.Candidate, 0)
		candidates = append(candidates, a.matchingFunctions("", editRange)...)
		candidates = append(candidates, newExpression(a.pathCtx, a.expr, schema.Reference{OfType: a.cons.OfType}).CompletionAtPos(ctx, pos)...)
		candidates = append(candidates, newExpression(a.pathCtx, a.expr, schema.LiteralType{Type: a.cons.OfType}).CompletionAtPos(ctx, pos)...)

		return candidates
	}

	switch eType := a.expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		if len(eType.Traversal) > 1 {
			// TODO! Reference completion
			return []lang.Candidate{}
		}
		prefixLen := pos.Byte - eType.Traversal.SourceRange().Start.Byte
		prefix := eType.Traversal.RootName()[0:prefixLen]

		// TODO! include references
		return a.matchingFunctions(prefix, eType.Range())
	case *hclsyntax.FunctionCallExpr:
		if eType.NameRange.ContainsPos(pos) {
			prefixLen := pos.Byte - eType.NameRange.Start.Byte
			prefix := eType.Name[0:prefixLen]
			editRange := eType.Range()
			return a.matchingFunctions(prefix, editRange)
		}

		f, ok := a.pathCtx.Functions[eType.Name]
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

				cons := newExpression(a.pathCtx, arg, schema.AnyExpression{OfType: param.Type})
				return cons.CompletionAtPos(ctx, pos)
			}
			lastArgEndPos = arg.Range().End
			lastArgIdx = i
		}

		fileBytes := a.pathCtx.Files[eType.Range().Filename].Bytes
		recoveredBytes := recoverLeftBytes(fileBytes, pos, func(byteOffset int, r rune) bool {
			return (r == ',' || r == '(') && byteOffset > lastArgEndPos.Byte
		})
		trimmedBytes := bytes.TrimRight(recoveredBytes, " \t\n")

		activePar := 0 // default to first parameter
		if string(trimmedBytes) == "," {
			activePar = lastArgIdx + 1
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

		cons := newExpression(a.pathCtx, newEmptyExpressionAtPos(eType.Range().Filename, pos), schema.AnyExpression{OfType: param.Type})
		return cons.CompletionAtPos(ctx, pos)
	}

	return []lang.Candidate{}
}

func (a Any) matchingFunctions(prefix string, editRange hcl.Range) []lang.Candidate {
	candidates := make([]lang.Candidate, 0)

	for name, f := range a.pathCtx.Functions {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		// Only suggest functions that have a matching return type
		if a.cons.OfType.TestConformance(f.ReturnType) != nil {
			continue
		}

		// TODO? see why accepting a completion isn't triggering signatureHelp (it does for gopls)
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
