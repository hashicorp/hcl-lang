package lang

import (
	"github.com/hashicorp/hcl/v2"
)

type HoverData struct {
	Content MarkupContent
	Range   hcl.Range
}
