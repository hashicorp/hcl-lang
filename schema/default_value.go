// Copyright IBM Corp. 2020, 2025
// SPDX-License-Identifier: MPL-2.0

package schema

import "github.com/zclconf/go-cty/cty"

type defaultSigil struct{}

type Default interface {
	isDefaultImpl() defaultSigil
}

type DefaultValue struct {
	Value cty.Value
}

func (dv DefaultValue) isDefaultImpl() defaultSigil {
	return defaultSigil{}
}

// TODO: DefaultKeyword
// TODO: DefaultTypeDeclaration
// TODO: defaults dependent on other attributes
// TODO: defaults dependent on env variables
