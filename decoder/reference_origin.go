package decoder

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
)

type ReferenceOrigin struct {
	Path  lang.Path
	Range hcl.Range
}

type ReferenceOrigins []ReferenceOrigin
