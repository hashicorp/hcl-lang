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

func (d *PathDecoder) Validate(ctx context.Context) (hcl.Diagnostics, error) {
	if d.pathCtx.Schema == nil {
		return hcl.Diagnostics{}, &NoSchemaError{}
	}

	diags := hcl.Diagnostics{}
	// Validate module files per schema
	for _, f := range d.pathCtx.Files {
		body, ok := f.Body.(*hclsyntax.Body)
		if !ok {
			// TODO! error
			continue
		}

		diags = diags.Extend(d.validateBody(ctx, body, d.pathCtx.Schema))
	}

	// Run validation functions
	for _, vFunc := range d.decoderCtx.Validations {
		diags = diags.Extend(vFunc(ctx))
	}

	return diags, nil
}

func (d *PathDecoder) validateBody(ctx context.Context, body *hclsyntax.Body, bodySchema *schema.BodySchema) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	for name, attribute := range body.Attributes {
		attributeSchema, ok := bodySchema.Attributes[name]
		if !ok {
			// ---------- diag ERR unknown attribute
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unexpected attribute",
				Detail:   fmt.Sprintf("An attribute named %q is not expected here", name),
				Subject:  &attribute.SrcRange,
			})
			// don't check futher because this isn't a valid attribute
			continue
		}

		// ---------- diag WARN deprecated attribute
		if attributeSchema.IsDeprecated {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  fmt.Sprintf("%q is deprecated", name),
				Detail:  fmt.Sprintf("Reason: %q", attributeSchema.Description.Value),
				Subject:  &attribute.SrcRange,
			})
		}
	}

	for _, block := range body.Blocks {
		blockSchema, ok := bodySchema.Blocks[block.Type]
		if !ok {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unexpected block",
				Detail:   fmt.Sprintf("Blocks of type %q are not expected here", block.Type),
				Subject:  &block.TypeRange,
			})
			// don't check futher because this isn't a valid block
			continue
		}

		// ---------- diag WARN deprecated block
		if blockSchema.IsDeprecated {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  fmt.Sprintf("%q is deprecated", block.Type),
				Detail:  fmt.Sprintf("Reason: %q", blockSchema.Description.Value),
				Subject:  &block.TypeRange,
			})
		}

		// ---------- daig ERR extraneous block labels
		validLabelNum := len(blockSchema.Labels)
		for i := range block.Labels {
			if i >= validLabelNum {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Too many labels specified for %q", block.Type),
					Detail:   fmt.Sprintf("Only %d label(s) are expected for %q blocks", validLabelNum, block.Type),
					Subject:  &block.LabelRanges[i],
				})
			}
		}

		// ---------- diag ERR missing labels
		if validLabelNum > len(block.Labels) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Not enough labels specified for %q", block.Type),
				Detail:   fmt.Sprintf("All %q blocks must have %d label(s)", block.Type, validLabelNum),
				Subject:  &block.TypeRange,
			})
		}

		if block.Body != nil {
			mergedSchema, err := mergeBlockBodySchemas(block.AsHCLBlock(), blockSchema)
			if err != nil {
				// TODO! err
			}

			diags = diags.Extend(d.validateBody(ctx, block.Body, mergedSchema))
		}
	}

	return diags
}
