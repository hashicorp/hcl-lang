package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Reference struct {
	expr    hcl.Expression
	cons    schema.Reference
	pathCtx *PathContext
}
