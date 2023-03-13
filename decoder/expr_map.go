package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Map struct {
	expr    hcl.Expression
	cons    schema.Map
	pathCtx *PathContext
}
