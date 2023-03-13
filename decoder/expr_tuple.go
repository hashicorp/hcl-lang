package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Tuple struct {
	expr    hcl.Expression
	cons    schema.Tuple
	pathCtx *PathContext
}
