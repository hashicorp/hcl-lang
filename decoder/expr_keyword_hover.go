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

func (kw Keyword) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	eType, ok := kw.expr.(*hclsyntax.ScopeTraversalExpr)
	if !ok {
		return nil
	}

	if len(eType.Traversal) != 1 {
		return nil
	}

	if eType.Traversal.RootName() == kw.cons.Keyword {
		content := fmt.Sprintf("`%s` _%s_", kw.cons.Keyword, kw.cons.FriendlyName())
		if kw.cons.Description.Value != "" {
			content += "\n\n" + kw.cons.Description.Value
		}

		return &lang.HoverData{
			Content: lang.Markdown(content),
			Range:   eType.SrcRange,
		}
	}

	return nil
}
