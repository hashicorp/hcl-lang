package lang

import "github.com/hashicorp/hcl/v2"

type ReferenceOrigin struct {
	Addr  Address
	Range hcl.Range
}
