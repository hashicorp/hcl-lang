// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl-lang/decoder/internal/ast"
	"github.com/hashicorp/hcl-lang/decoder/internal/schemahelper"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// SymbolsInFile returns a hierarchy of symbols within the config file
//
// A symbol is typically represented by a block or an attribute.
//
// Symbols within JSON files require schema to be present for decoding.
func (d *PathDecoder) SymbolsInFile(filename string) ([]Symbol, error) {
	f, err := d.fileByName(filename)
	if err != nil {
		return nil, err
	}

	_, isHcl := f.Body.(*hclsyntax.Body)
	if !isHcl {
		return nil, &UnknownFileFormatError{Filename: filename}
	}

	return d.symbolsForBody(f.Body, d.pathCtx.Schema), nil
}

func (d *PathDecoder) symbolsInFile(filename string) ([]Symbol, error) {
	f, err := d.fileByName(filename)
	if err != nil {
		return nil, err
	}

	return d.symbolsForBody(f.Body, d.pathCtx.Schema), nil
}

// Symbols returns a hierarchy of symbols matching the query in all paths.
// Query can be empty, as per LSP's workspace/symbol request,
// in which case all symbols are returned.
//
// A symbol is typically represented by a block or an attribute.
//
// Symbols within JSON files require schema to be present for decoding.
func (d *Decoder) Symbols(ctx context.Context, query string) ([]Symbol, error) {
	symbols := make([]Symbol, 0)

	for _, path := range d.pathReader.Paths(ctx) {
		pathDecoder, err := d.Path(path)
		if err != nil {
			continue
		}
		dirSymbols, err := pathDecoder.symbols(query)
		if err != nil {
			continue
		}

		symbols = append(symbols, dirSymbols...)
	}

	return symbols, nil
}

func (d *PathDecoder) symbols(query string) ([]Symbol, error) {
	symbols := make([]Symbol, 0)
	files := d.filenames()

	for _, filename := range files {
		fSymbols, err := d.symbolsInFile(filename)
		if err != nil {
			return nil, err
		}

		for _, symbol := range fSymbols {
			if query == "" || strings.Contains(symbol.Name(), query) {
				symbols = append(symbols, symbol)
			}
		}
	}

	return symbols, nil
}

func (d *PathDecoder) symbolsForBody(body hcl.Body, bodySchema *schema.BodySchema) []Symbol {
	symbols := make([]Symbol, 0)
	if body == nil {
		return symbols
	}

	content := ast.DecodeBody(body, bodySchema)

	for name, attr := range content.Attributes {
		symbols = append(symbols, &AttributeSymbol{
			AttrName:      name,
			ExprKind:      symbolExprKind(attr.Expr),
			path:          d.path,
			rng:           attr.Range,
			nestedSymbols: d.nestedSymbolsForExpr(attr.Expr),
		})
	}

	for _, block := range content.Blocks {
		var bSchema *schema.BodySchema
		if bodySchema != nil {
			bs, ok := bodySchema.Blocks[block.Type]
			if ok {
				bSchema = bs.Body
				mergedSchema, _ := schemahelper.MergeBlockBodySchemas(block.Block, bs)
				bSchema = mergedSchema
			}
		}

		symbols = append(symbols, &BlockSymbol{
			Type:          block.Type,
			Labels:        block.Labels,
			path:          d.path,
			rng:           block.Range,
			nestedSymbols: d.symbolsForBody(block.Body, bSchema),
		})
	}

	sort.SliceStable(symbols, func(i, j int) bool {
		return symbols[i].Range().Start.Byte < symbols[j].Range().Start.Byte
	})

	return symbols
}

func symbolExprKind(expr hcl.Expression) lang.SymbolExprKind {
	switch e := expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		return lang.TraversalExprKind{}
	case *hclsyntax.LiteralValueExpr:
		// String constant may also be a traversal in some cases, but currently not recognized
		// TODO: https://github.com/hashicorp/terraform-ls/issues/674
		return lang.LiteralTypeKind{Type: e.Val.Type()}
	case *hclsyntax.TemplateExpr:
		if e.IsStringLiteral() {
			return lang.LiteralTypeKind{Type: cty.String}
		}
		if isMultilineStringLiteral(e) {
			return lang.LiteralTypeKind{Type: cty.String}
		}
	case *hclsyntax.TupleConsExpr:
		return lang.TupleConsExprKind{}
	case *hclsyntax.ObjectConsExpr:
		return lang.ObjectConsExprKind{}
	default:
		// TODO Determine expression types for JSON
	}
	return nil
}

func (d *PathDecoder) nestedSymbolsForExpr(expr hcl.Expression) []Symbol {
	symbols := make([]Symbol, 0)

	switch e := expr.(type) {
	case *hclsyntax.TupleConsExpr:
		for i, item := range e.ExprList() {
			symbols = append(symbols, &ExprSymbol{
				ExprName:      fmt.Sprintf("%d", i),
				ExprKind:      symbolExprKind(item),
				path:          d.path,
				rng:           item.Range(),
				nestedSymbols: d.nestedSymbolsForExpr(item),
			})
		}
	case *hclsyntax.ObjectConsExpr:
		for _, item := range e.Items {
			key, _ := item.KeyExpr.Value(nil)
			if key.IsNull() || !key.IsWhollyKnown() || key.Type() != cty.String {
				// skip items keys that can't be interpolated
				// without further context
				continue
			}
			symbols = append(symbols, &ExprSymbol{
				ExprName:      key.AsString(),
				ExprKind:      symbolExprKind(item.ValueExpr),
				path:          d.path,
				rng:           hcl.RangeBetween(item.KeyExpr.Range(), item.ValueExpr.Range()),
				nestedSymbols: d.nestedSymbolsForExpr(item.ValueExpr),
			})
		}
	}

	return symbols
}
