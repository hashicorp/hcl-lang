package decoder

import (
	"context"

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

	for name, _ := range body.Attributes {
		_, ok := bodySchema.Attributes[name]
		if !ok {
			// TODO! unknown attribute validation
		}
		// TODO! validate against schema
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
