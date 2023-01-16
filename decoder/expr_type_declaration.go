package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type TypeDeclaration struct {
	expr    hcl.Expression
	cons    schema.TypeDeclaration
	pathCtx *PathContext
}

func (td TypeDeclaration) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	// TODO
	return nil
}

func (td TypeDeclaration) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	// TODO
	return nil
}

func isTypeNameWithElementOnly(name string) bool {
	return name == "list" || name == "set" || name == "map"
}
