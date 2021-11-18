package reference

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
)

type originSigil struct{}

type Origin interface {
	isOriginImpl() originSigil
	Copy() Origin
	OriginRange() hcl.Range
	OriginConstraints() OriginConstraints
	Address() lang.Address
}
