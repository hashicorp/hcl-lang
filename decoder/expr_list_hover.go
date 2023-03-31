// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (list List) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	eType, ok := list.expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil
	}

	for _, elemExpr := range eType.Exprs {
		if elemExpr.Range().ContainsPos(pos) {
			expr := newExpression(list.pathCtx, elemExpr, list.cons.Elem)
			return expr.HoverAtPos(ctx, pos)
		}
	}

	content := fmt.Sprintf("_%s_", list.cons.FriendlyName())
	if list.cons.Description.Value != "" {
		content += "\n\n" + list.cons.Description.Value
	}

	return &lang.HoverData{
		Content: lang.Markdown(content),
		Range:   eType.Range(),
	}
}
