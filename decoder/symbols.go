package decoder

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// Symbols returns a hierarchy of symbols within the config file
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

func symbolsForBody(body *hclsyntax.Body) []Symbol {
	symbols := make([]Symbol, 0)
	if body == nil {
		return symbols
	}

	for name, attr := range body.Attributes {
		var typ cty.Type

		switch expr := attr.Expr.(type) {
		case *hclsyntax.LiteralValueExpr:
			typ = expr.Val.Type()
		case *hclsyntax.TemplateExpr:
			if expr.IsStringLiteral() {
				typ = cty.String
			}
		}

		symbols = append(symbols, &AttributeSymbol{
			AttrName: name,
			Type:     typ,
			rng:      attr.Range(),
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

	return symbols
}
