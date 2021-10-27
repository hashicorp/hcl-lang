package decoder

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
)

type ReferenceTarget struct {
	OriginRange hcl.Range

	Path        lang.Path
	Range       hcl.Range
	DefRangePtr *hcl.Range
}
