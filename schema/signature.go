// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

type FunctionSignature struct {
	// Description is an optional human-readable description
	// of the function.
	Description string

	Detail string

	// ReturnType is the ctyjson representation of the function's
	// return types based on supplying all parameters using
	// dynamic types. Functions can have dynamic return types.
	ReturnType cty.Type

	// Params describes the function's fixed positional parameters.
	Params []function.Parameter

	// VarParam describes the function's variadic
	// parameter if it is supported.
	VarParam *function.Parameter
}

func (fs *FunctionSignature) Copy() *FunctionSignature {
	newFS := &FunctionSignature{
		Description: fs.Description,
		Detail:      fs.Detail,
		ReturnType:  fs.ReturnType,
	}

	newFS.Params = make([]function.Parameter, len(fs.Params))
	copy(newFS.Params, fs.Params)

	if fs.VarParam != nil {
		vpCpy := *fs.VarParam
		newFS.VarParam = &vpCpy
	}

	return newFS
}
