package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type OneOf struct {
	expr    hcl.Expression
	cons    schema.OneOf
	pathCtx *PathContext
}
