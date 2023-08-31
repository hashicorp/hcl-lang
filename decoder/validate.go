// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/decoder/internal/validator"
	"github.com/hashicorp/hcl-lang/decoder/internal/walker"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

var builtinValidators = []validator.Validator{
	validator.BlockLabelsLength{},
	validator.DeprecatedAttribute{},
	validator.DeprecatedBlock{},
	validator.MaxBlocks{},
	validator.MinBlocks{},
	validator.MissingRequiredAttribute{},
	validator.UnexpectedAttribute{},
	validator.UnexpectedBlock{},
}

// Validate returns a set of Diagnostics for all known files
func (d *PathDecoder) Validate(ctx context.Context) (lang.DiagnosticsMap, error) {
	diags := make(lang.DiagnosticsMap)
	if d.pathCtx.Schema == nil {
		return diags, &NoSchemaError{}
	}

	// Validate module files per schema
	for filename, f := range d.pathCtx.Files {
		body, ok := f.Body.(*hclsyntax.Body)
		if !ok {
			// TODO! error
			continue
		}

		diags[filename] = walker.Walk(ctx, body, d.pathCtx.Schema, validationWalker{
			validators: builtinValidators,
		})
	}

	return diags, nil
}

// ValidateFile validates given file and returns a list of Diagnostics for that file
func (d *PathDecoder) ValidateFile(ctx context.Context, filename string) (hcl.Diagnostics, error) {
	if d.pathCtx.Schema == nil {
		return hcl.Diagnostics{}, &NoSchemaError{}
	}

	f, err := d.fileByName(filename)
	if err != nil {
		return hcl.Diagnostics{}, err
	}

	body, ok := f.Body.(*hclsyntax.Body)
	if !ok {
		return hcl.Diagnostics{}, &UnknownFileFormatError{Filename: filename}
	}

	return walker.Walk(ctx, body, d.pathCtx.Schema, validationWalker{
		validators: builtinValidators,
	}), nil
}

type validationWalker struct {
	validators []validator.Validator
}

func (vw validationWalker) Visit(ctx context.Context, node hclsyntax.Node, nodeSchema schema.Schema) (diags hcl.Diagnostics) {
	for _, v := range vw.validators {
		diags = append(diags, v.Visit(ctx, node, nodeSchema)...)
	}

	return
}
