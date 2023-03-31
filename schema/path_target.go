// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
)

type PathTarget struct {
	Address     Address
	Path        lang.Path
	Constraints Constraints
}

func (pt *PathTarget) Copy() *PathTarget {
	if pt == nil {
		return nil
	}

	return &PathTarget{
		Address:     pt.Address.Copy(),
		Path:        pt.Path,
		Constraints: pt.Constraints,
	}
}

type Constraints struct {
	ScopeId lang.ScopeId
	Type    cty.Type
}
