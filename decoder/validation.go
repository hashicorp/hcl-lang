// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (d *PathDecoder) ValidateFilePerSchema(ctx context.Context, filename string) (hcl.Diagnostics, error) {
	f, err := d.fileByName(filename)
	if err != nil {
		return hcl.Diagnostics{}, err
	}
	if d.pathCtx.Schema == nil {
		return hcl.Diagnostics{}, &NoSchemaError{}
	}

	// Check if body is hcl, else return early
	body, ok := f.Body.(*hclsyntax.Body)
	if !ok {
		return hcl.Diagnostics{}, nil // TODO! error
	}

	// compare targets and origins for diags

	// "walk" the configuration, compare it against the schema
	return d.validateFilePerSchema(ctx, body, d.pathCtx.Schema)
}

func (d *PathDecoder) validateFilePerSchema(ctx context.Context, body *hclsyntax.Body, bodySchema *schema.BodySchema) (hcl.Diagnostics, error) {
	diags := hcl.Diagnostics{}

	for name, attr := range body.Attributes {
		attrSchema, ok := bodySchema.Attributes[name]
		if !ok {
			// diag ERR unknown attribute, range from attr
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unexpected attribute",
				Detail:   fmt.Sprintf("An argument named %q is not expected here.", name),
				Subject:  &attr.SrcRange,
			})
			continue
		}

		if attrSchema.IsDeprecated {
			// diag WARN deprecated
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  fmt.Sprintf("%q is deprecated", name),
				Subject:  &attr.SrcRange,
			})
		}
		// track attributes
	}
	// compare required attributes with tracked ones

	for _, block := range body.Blocks {
		blockSchema, ok := bodySchema.Blocks[block.Type]
		if !ok {
			// dig ERR unknown block
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unexpected block",
				Detail:   fmt.Sprintf("Blocks of type %q are not expected here.", block.Type),
				Subject:  &block.TypeRange,
			})
			continue
		}

		validLabelNum := len(blockSchema.Labels)
		// check extraneous for block labels
		for i := range block.Labels {
			if i >= validLabelNum {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Too many labels",
					Detail:   fmt.Sprintf("Only %d labels are expected for %s blocks.", validLabelNum, block.Type),
					Subject:  &block.LabelRanges[i],
				})
			}
		}
		if validLabelNum > len(block.Labels) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Missing label",
				Detail:   fmt.Sprintf("All %s blocks must have %d label(s).", block.Type, validLabelNum),
				Subject:  &block.TypeRange,
			})
		}

		// track block-type count

		// Dig deeper
		if block.Body != nil {
			mergedSchema, err := mergeBlockBodySchemas(block.AsHCLBlock(), blockSchema)
			if err != nil {
				return diags, err
			}

			d, err := d.validateFilePerSchema(ctx, block.Body, mergedSchema)
			if err != nil {
				return diags, err
			}
			diags = append(diags, d...)
		}
	}
	// compare block-type counts against schema min max

	return diags, nil
}
