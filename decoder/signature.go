// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// SignatureAtPos returns a function signature for the given pos if pos
// is inside a FunctionCallExpr
func (d *PathDecoder) SignatureAtPos(filename string, pos hcl.Pos) (*lang.FunctionSignature, error) {
	file, err := d.fileByName(filename)
	if err != nil {
		return nil, err
	}

	body, err := d.bodyForFileAndPos(filename, file, pos)
	if err != nil {
		return nil, err
	}

	var signature *lang.FunctionSignature
	hclsyntax.VisitAll(body, func(node hclsyntax.Node) hcl.Diagnostics {
		if !node.Range().ContainsPos(pos) {
			return nil // Node not under cursor
		}

		fNode, isFunc := node.(*hclsyntax.FunctionCallExpr)
		if !isFunc {
			return nil // No function call expression
		}

		f, ok := d.pathCtx.Functions[fNode.Name]
		if !ok {
			return nil // Unknown function
		}

		if len(f.Params) == 0 && f.VarParam == nil {
			signature = &lang.FunctionSignature{
				Name:        fmt.Sprintf("%s(%s) %s", fNode.Name, parameterNamesAsString(f), f.ReturnType.FriendlyName()),
				Description: lang.Markdown(f.Description),
			}

			return nil // Function accepts no parameters, return early
		}

		pRange := hcl.RangeBetween(fNode.OpenParenRange, fNode.CloseParenRange)
		if !pRange.ContainsPos(pos) {
			return nil // Not inside parenthesis
		}

		activePar := 0 // default to first parameter
		foundActivePar := false
		lastArgEndPos := fNode.OpenParenRange.Start
		lastArgIdx := 0
		for i, v := range fNode.Args {
			// We overshot the argument and stop
			if v.Range().Start.Byte > pos.Byte {
				break
			}
			if v.Range().ContainsPos(pos) || v.Range().End.Byte == pos.Byte {
				activePar = i
				foundActivePar = true
				break
			}
			lastArgEndPos = v.Range().End
			lastArgIdx = i
		}

		if !foundActivePar {
			recoveredBytes := recoverLeftBytes(file.Bytes, pos, func(byteOffset int, r rune) bool {
				return r == ',' && byteOffset > lastArgEndPos.Byte
			})
			trimmedBytes := bytes.TrimRight(recoveredBytes, " \t\n")
			if string(trimmedBytes) == "," {
				activePar = lastArgIdx + 1
			}
		}

		paramsLen := len(f.Params)
		if f.VarParam != nil {
			paramsLen += 1
		}

		if activePar >= paramsLen && f.VarParam == nil {
			return nil // too many arguments passed to the function
		}

		if activePar >= paramsLen {
			// there are multiple variadic arguments passed, so
			// we want to highlight the variadic parameter in the
			// function signature
			activePar = paramsLen - 1
		}

		parameters := make([]lang.FunctionParameter, 0, paramsLen)
		for _, p := range f.Params {
			parameters = append(parameters, lang.FunctionParameter{
				Name:        p.Name,
				Description: lang.Markdown(p.Description),
			})
		}
		if f.VarParam != nil {
			parameters = append(parameters, lang.FunctionParameter{
				Name:        f.VarParam.Name,
				Description: lang.Markdown(f.VarParam.Description),
			})
		}
		signature = &lang.FunctionSignature{
			Name:            fmt.Sprintf("%s(%s) %s", fNode.Name, parameterNamesAsString(f), f.ReturnType.FriendlyName()),
			Description:     lang.Markdown(f.Description),
			Parameters:      parameters,
			ActiveParameter: uint32(activePar),
		}

		return nil // We don't want to add any diagnostics
	})

	return signature, nil
}

// parameterNamesAsString returns a string containing all function parameters
// with their respective types.
//
// Useful for displaying as part of a function signature.
func parameterNamesAsString(fs schema.FunctionSignature) string {
	paramsLen := len(fs.Params)
	if fs.VarParam != nil {
		paramsLen += 1
	}
	names := make([]string, 0, paramsLen)

	for _, p := range fs.Params {
		names = append(names, fmt.Sprintf("%s %s", p.Name, p.Type.FriendlyName()))
	}
	if fs.VarParam != nil {
		names = append(names, fmt.Sprintf("…%s %s", fs.VarParam.Name, fs.VarParam.Type.FriendlyName()))
	}

	return strings.Join(names, ", ")
}
