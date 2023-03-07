package decoder

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (set Set) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	eType, ok := set.expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil
	}

	for _, elemExpr := range eType.Exprs {
		if elemExpr.Range().ContainsPos(pos) {
			expr := newExpression(set.pathCtx, elemExpr, set.cons.Elem)
			return expr.HoverAtPos(ctx, pos)
		}
	}

	content := fmt.Sprintf("_%s_", set.cons.FriendlyName())
	if set.cons.Description.Value != "" {
		content += "\n\n" + set.cons.Description.Value
	}

	return &lang.HoverData{
		Content: lang.Markdown(content),
		Range:   eType.Range(),
	}
}
