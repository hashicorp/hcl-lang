package decoder

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (tuple Tuple) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	eType, ok := tuple.expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil
	}

	for i, elemExpr := range eType.Exprs {
		if i+1 > len(tuple.cons.Elems) {
			return nil
		}

		if elemExpr.Range().ContainsPos(pos) {
			expr := newExpression(tuple.pathCtx, elemExpr, tuple.cons.Elems[i])
			return expr.HoverAtPos(ctx, pos)
		}
	}

	content := fmt.Sprintf("_%s_", tuple.cons.FriendlyName())
	if tuple.cons.Description.Value != "" {
		content += "\n\n" + tuple.cons.Description.Value
	}

	return &lang.HoverData{
		Content: lang.Markdown(content),
		Range:   eType.Range(),
	}
}
