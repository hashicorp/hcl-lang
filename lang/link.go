package lang

import (
	"github.com/hashicorp/hcl/v2"
)

type Link struct {
	URI     string
	Tooltip string
	Range   hcl.Range
}
