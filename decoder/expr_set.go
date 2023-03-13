package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Set struct {
	expr    hcl.Expression
	cons    schema.Set
	pathCtx *PathContext
}
