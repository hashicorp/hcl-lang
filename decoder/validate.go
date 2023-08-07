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

// Validate returns a set of Diagnostics for all known files
func (d *PathDecoder) Validate(ctx context.Context) (map[string]hcl.Diagnostics, error) {
	diags := make(map[string]hcl.Diagnostics, 0)
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

		diags[filename] = d.validateBody(ctx, body, d.pathCtx.Schema)
	}

	// Run validation functions
	// for _, vFunc := range d.decoderCtx.Validations {
	// 	diags = diags.Extend(vFunc(ctx))
	// }

	return diags, nil
}

// validateBody returns a set of Diagnostics for a given HCL body
//
// Validations available:
//
//   - unexpected attribute
//
//   - missing required attribute
//
//   - deprecated attribute
//
//   - unexpected block
//
//   - deprecated block
//
//   - min blocks
//
//   - max blocks
func (d *PathDecoder) validateBody(ctx context.Context, body *hclsyntax.Body, bodySchema *schema.BodySchema) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	// Iterate over all Attributes in the body
	for name, attribute := range body.Attributes {
		attributeSchema, ok := bodySchema.Attributes[name]
		if !ok {
			// ---------- diag ERR unknown attribute
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unexpected attribute",
				Detail:   fmt.Sprintf("An attribute named %q is not expected here", name),
				Subject:  attribute.SrcRange.Ptr(),
			})
			// don't check futher because this isn't a valid attribute
			continue
		}

		// ---------- diag WARN deprecated attribute
		if attributeSchema.IsDeprecated {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  fmt.Sprintf("%q is deprecated", name),
				Detail:   fmt.Sprintf("Reason: %q", attributeSchema.Description.Value),
				Subject:  attribute.SrcRange.Ptr(),
			})
		}
	}

	// Iterate over all schema Attributes and check if specified in the configuration
	for name, attribute := range bodySchema.Attributes {
		if attribute.IsRequired {
			_, ok := body.Attributes[name]
			if !ok {
				// ---------- diag ERR unknown attribute
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Required attribute %q not specified", name),
					Detail:   fmt.Sprintf("An attribute named %q is required here", name),
					// TODO This is the closest I could think of
					// maybe block instead ?
					Subject: body.SrcRange.Ptr(),
				})
			}
		}
	}

	// keep track of blocks actually used so we can compare to schema later
	specifiedBlocks := make(map[string]int)

	// Iterate over all Blocks in the body
	for _, block := range body.Blocks {
		blockSchema, ok := bodySchema.Blocks[block.Type]
		if !ok {
			// ---------- diag ERR unknown block
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unexpected block",
				Detail:   fmt.Sprintf("Blocks of type %q are not expected here", block.Type),
				Subject:  block.TypeRange.Ptr(),
			})
			// don't check futher because this isn't a valid block
			continue
		}

		// ---------- diag WARN deprecated block
		if blockSchema.IsDeprecated {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  fmt.Sprintf("%q is deprecated", block.Type),
				// todo check if description is there
				Detail:  fmt.Sprintf("Reason: %q", blockSchema.Description.Value),
				Subject: &block.TypeRange,
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
					Subject:  block.LabelRanges[i].Ptr(),
				})
			}
		}

		// ---------- diag ERR missing labels
		if validLabelNum > len(block.Labels) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Not enough labels specified for %q", block.Type),
				Detail:   fmt.Sprintf("All %q blocks must have %d label(s)", block.Type, validLabelNum),
				Subject:  block.TypeRange.Ptr(),
			})
		}

		if block.Body != nil {
			mergedSchema, err := mergeBlockBodySchemas(block.AsHCLBlock(), blockSchema)
			if err != nil {
				// TODO! err
			}

			// Recurse for nested blocks
			diags = diags.Extend(d.validateBody(ctx, block.Body, mergedSchema))
		}

		// build list of blocks specified
		specifiedBlocks[block.Type]++
	}

	// Iterate over bodySchema Blocks and check if they are specified in configuration
	for name, block := range bodySchema.Blocks {
		// check if the bodySchema Block is specified in the configuration
		numBlocks, ok := specifiedBlocks[name]
		if ok {
			// block is in schema and user specified it in configuration
			// check if schema says there should be maximum number of items for this block
			if block.MaxItems > 0 {
				// ---------- diag ERR too many blocks
				if numBlocks > int(block.MaxItems) {
					subjectRange := &body.Blocks[block.Type].TypeRange
					maxItems := block.MaxItems
					diags = tooManyBlocksDiag(diags, name, maxItems, subjectRange)
				}
			}

			// check if schema says there should be minimum number of items for this block
			if block.MinItems > 0 {
				// ---------- diag ERR too little blocks
				if numBlocks < int(block.MinItems) {
					subjectRange := &body.Blocks[block.Type].TypeRange
					minItems := block.MinItems
					diags = tooFewItemsDiag(diags, name, minItems, subjectRange)
				}
			}
		} else {
			// block is in schema, but user did not specify it in configuration
			// check if schema says there should be maximum number of items for this block
			numBlocks = 0
			if block.MaxItems > 0 {
				// ---------- diag ERR too many blocks
				if numBlocks > int(block.MaxItems) {
					// use current body range as there isn't a block to reference because
					// the user didn't write anything here
					subjectRange := &body.SrcRange
					maxItems := block.MaxItems
					diags = tooManyBlocksDiag(diags, name, maxItems, subjectRange)
				}
			}

			// check if schema says there should be minimum number of items for this block
			if block.MinItems > 0 {
				// ---------- diag ERR too little blocks
				if numBlocks < int(block.MinItems) {
					// use current body range as there isn't a block to reference because
					// the user didn't write anything here
					subjectRange := &body.SrcRange
					minItems := block.MinItems
					diags = tooFewItemsDiag(diags, name, minItems, subjectRange)
				}
			}
		}
	}

	return diags
}

func tooFewItemsDiag(diags hcl.Diagnostics, name string, minItems uint64, subjectRange *hcl.Range) hcl.Diagnostics {
	diags = append(diags, &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("Too few blocks specified for %q", name),
		Detail:   fmt.Sprintf("At least %d block(s) are expected for %q", minItems, name),
		Subject:  subjectRange,
	})
	return diags
}

func tooManyBlocksDiag(diags hcl.Diagnostics, name string, maxItems uint64, subjectRange *hcl.Range) hcl.Diagnostics {
	diags = append(diags, &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("Too many blocks specified for %q", name),
		Detail:   fmt.Sprintf("Only %d block(s) are expected for %q", maxItems, name),
		Subject:  subjectRange,
	})
	return diags
}
