package decoder

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

// PathContext represents any context relevant to the lang.Path
// i.e. anything that is tied either to path or language ID
type PathContext struct {
	Schema           *schema.BodySchema
	ReferenceOrigins lang.ReferenceOrigins
	ReferenceTargets lang.ReferenceTargets
	Files            map[string]*hcl.File
}
