package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (a Any) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	switch e := a.expr.(type) {
	// TODO! Support LiteralType
	// TODO! Support References
	case *hclsyntax.FunctionCallExpr:
		_, ok := a.pathCtx.Functions[e.Name]
		if ok {
			// TODO! check for name range
			// TODO! loop over arguments, do SemanticTokens for that argument using the constraint
			// TODO? if we introduce a new token type, we need to add it in the extension's package.json and docs
			return []lang.SemanticToken{
				{
					Type:      lang.TokenAttrName,
					Modifiers: []lang.SemanticTokenModifier{},
					// Range:     eType.Range(),
				},
			}
		}
	}

	return nil
}
