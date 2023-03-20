package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type LiteralType struct {
	expr hcl.Expression
	cons schema.LiteralType

	pathCtx *PathContext
}
