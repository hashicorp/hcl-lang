package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type List struct {
	expr    hcl.Expression
	cons    schema.List
	pathCtx *PathContext
}
