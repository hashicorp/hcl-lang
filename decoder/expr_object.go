package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Object struct {
	expr    hcl.Expression
	cons    schema.Object
	pathCtx *PathContext
}
