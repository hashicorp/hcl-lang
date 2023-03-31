// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lang

type FunctionSignature struct {
	Name string

	// Description is an optional human-readable description
	// of the function.
	Description MarkupContent

	// Parameters is an ordered list of the function's parameters.
	Parameters []FunctionParameter

	// ActiveParameter is an index marking the parameter a user is currently
	// editing. It should lie inside the range of Parameters.
	ActiveParameter uint32
}

type FunctionParameter struct {
	Name string

	// Description is an optional human-readable description
	// of the parameter.
	Description MarkupContent
}
