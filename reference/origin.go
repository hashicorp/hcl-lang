package reference

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
)

type Origin struct {
	Addr  lang.Address
	Range hcl.Range

	// Constraints represents any traversal expression constraints
	// for the attribute where the origin was found.
	//
	// Further matching against decoded reference targets is needed
	// for >1 constraints, which is done later at runtime as
	// targets and origins can be decoded at different times.
	Constraints OriginConstraints
}

func (ro Origin) Copy() Origin {
	return Origin{
		Addr:        ro.Addr.Copy(),
		Range:       ro.Range,
		Constraints: ro.Constraints.Copy(),
	}
}
