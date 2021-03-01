package decoder

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// SymbolsInFile returns a hierarchy of symbols within the config file
//
// A symbol is typically represented by a block or an attribute.
func (d *Decoder) SymbolsInFile(filename string) ([]Symbol, error) {
	symbols := make([]Symbol, 0)

	f, err := d.fileByName(filename)
	if err != nil {
		return nil, err
	}

	body, err := d.bodyForFileAndPos(filename, f, hcl.InitialPos)
	if err != nil {
		return nil, err
	}
	symbols = append(symbols, symbolsForBody(body)...)

	return symbols, nil
}

// Symbols returns a hierarchy of symbols matching the query
// in all loaded files (typically whole module).
// Query can be empty, as per LSP's workspace/symbol request,
// in which case all symbols are returned.
//
// A symbol is typically represented by a block or an attribute.
func (d *Decoder) Symbols(query string) ([]Symbol, error) {
	symbols := make([]Symbol, 0)

	files := d.Filenames()

	for _, filename := range files {
		fSymbols, err := d.SymbolsInFile(filename)
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

func symbolsForBody(body *hclsyntax.Body) []Symbol {
	symbols := make([]Symbol, 0)
	if body == nil {
		return symbols
	}

	for name, attr := range body.Attributes {
		symbols = append(symbols, &AttributeSymbol{
			AttrName:      name,
			ExprKind:      symbolExprKind(attr.Expr),
			rng:           attr.Range(),
			nestedSymbols: nestedSymbolsForExpr(attr.Expr),
		})
	}
	for _, block := range body.Blocks {
		symbols = append(symbols, &BlockSymbol{
			Type:          block.Type,
			Labels:        block.Labels,
			rng:           block.Range(),
			nestedSymbols: symbolsForBody(block.Body),
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
	}
	return nil
}

func nestedSymbolsForExpr(expr hcl.Expression) []Symbol {
	symbols := make([]Symbol, 0)

	switch e := expr.(type) {
	case *hclsyntax.TupleConsExpr:
		for i, item := range e.ExprList() {
			symbols = append(symbols, &ExprSymbol{
				ExprName:      fmt.Sprintf("%d", i),
				ExprKind:      symbolExprKind(item),
				rng:           item.Range(),
				nestedSymbols: nestedSymbolsForExpr(item),
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
				rng:           hcl.RangeBetween(item.KeyExpr.Range(), item.ValueExpr.Range()),
				nestedSymbols: nestedSymbolsForExpr(item.ValueExpr),
			})
		}
	}

	return symbols
}
