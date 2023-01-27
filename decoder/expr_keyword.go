package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Keyword struct {
	expr hcl.Expression
	cons schema.Keyword
}
