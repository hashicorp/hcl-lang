package decoder

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (a Any) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	switch e := a.expr.(type) {
	// TODO! Support LiteralType
	// TODO! Support References
	case *hclsyntax.FunctionCallExpr:
		f, ok := a.pathCtx.Functions[e.Name]
		if ok {
			// TODO! check for name range
			// TODO! loop over arguments, do hoverAtPos for that argument using the constraint
			return &lang.HoverData{
				// TODO? add link to docs
				Content: lang.Markdown(fmt.Sprintf("```terraform\n%s(%s) %s\n```\n\n%s", e.Name, parameterNamesAsString(f), f.ReturnType.FriendlyName(), f.Description)),
				Range:   a.expr.Range(),
			}
		}
	}

	return nil
}
