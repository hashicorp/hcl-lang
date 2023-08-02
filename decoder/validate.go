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
			// TODO! unknown block validation
		}
		// TODO! validate against schema

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
