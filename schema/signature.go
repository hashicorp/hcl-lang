package schema

import (
	"fmt"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

type FunctionSignature struct {
	// Description is an optional human-readable description
	// of the function.
	Description string

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

// ParameterNames returns a list of all parameter names of a function.
func (fs FunctionSignature) ParameterNames() []string {
	paramsLen := len(fs.Params)
	if fs.VarParam != nil {
		paramsLen += 1
	}
	names := make([]string, 0, paramsLen)

	for _, p := range fs.Params {
		names = append(names, p.Name)
	}
	if fs.VarParam != nil {
		names = append(names, "..."+fs.VarParam.Name)
	}

	return names
}

// ParameterSignature returns a string containing all function parameters
// with their respective types.
//
// Useful for displaying as part of a function signature.
func (fs FunctionSignature) ParameterSignature() string {
	paramsLen := len(fs.Params)
	if fs.VarParam != nil {
		paramsLen += 1
	}
	names := make([]string, 0, paramsLen)

	for _, p := range fs.Params {
		names = append(names, fmt.Sprintf("%s %s", p.Name, p.Type.FriendlyName()))
	}
	if fs.VarParam != nil {
		names = append(names, fmt.Sprintf("...%s %s", fs.VarParam.Name, fs.VarParam.Type.FriendlyName()))
	}

	return strings.Join(names, ", ")
}
