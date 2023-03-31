// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reference

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
)

type OriginConstraint struct {
	OfScopeId lang.ScopeId
	OfType    cty.Type
}

type OriginConstraints []OriginConstraint

func (roc OriginConstraints) Copy() OriginConstraints {
	if roc == nil {
		return nil
	}

	cons := make(OriginConstraints, 0)
	for _, oc := range roc {
		cons = append(cons, oc)
	}

	return cons
}
