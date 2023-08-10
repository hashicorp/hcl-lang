// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schemahelper

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

func MergeBlockBodySchemas(block *hcl.Block, blockSchema *schema.BlockSchema) (*schema.BodySchema, bool) {
	mergedSchema := &schema.BodySchema{}
	if blockSchema.Body != nil {
		mergedSchema = blockSchema.Body.Copy()
	}
	if mergedSchema.Attributes == nil {
		mergedSchema.Attributes = make(map[string]*schema.AttributeSchema, 0)
	}
	if mergedSchema.Blocks == nil {
		mergedSchema.Blocks = make(map[string]*schema.BlockSchema, 0)
	}
	if mergedSchema.TargetableAs == nil {
		mergedSchema.TargetableAs = make([]*schema.Targetable, 0)
	}
	if mergedSchema.ImpliedOrigins == nil {
		mergedSchema.ImpliedOrigins = make([]schema.ImpliedOrigin, 0)
	}

	depSchema, depKeys, ok := NewBlockSchema(blockSchema).DependentBodySchema(block)
	if ok {
		for name, attr := range depSchema.Attributes {
			if _, exists := mergedSchema.Attributes[name]; !exists {
				mergedSchema.Attributes[name] = attr
			} else {
				// Skip duplicate attribute
				continue
			}
		}
		for bType, block := range depSchema.Blocks {
			if _, exists := mergedSchema.Blocks[bType]; !exists {
				// propagate DynamicBlocks extension to any nested blocks
				if mergedSchema.Extensions != nil && mergedSchema.Extensions.DynamicBlocks {
					if block.Body.Extensions == nil {
						block.Body.Extensions = &schema.BodyExtensions{}
					}
					block.Body.Extensions.DynamicBlocks = true
				}

				mergedSchema.Blocks[bType] = block
			} else {
				// Skip duplicate block type
				continue
			}
		}

		if mergedSchema.Extensions != nil && mergedSchema.Extensions.DynamicBlocks && len(depSchema.Blocks) > 0 {
			mergedSchema.Blocks["dynamic"] = buildDynamicBlockSchema(depSchema)
		}

		mergedSchema.TargetableAs = append(mergedSchema.TargetableAs, depSchema.TargetableAs...)
		mergedSchema.ImpliedOrigins = append(mergedSchema.ImpliedOrigins, depSchema.ImpliedOrigins...)

		// TODO: avoid resetting?
		mergedSchema.Targets = depSchema.Targets.Copy()

		// TODO: avoid resetting?
		mergedSchema.DocsLink = depSchema.DocsLink.Copy()

		// use extensions of DependentBody if not nil
		// (to avoid resetting to nil)
		if depSchema.Extensions != nil {
			mergedSchema.Extensions = depSchema.Extensions.Copy()
		}
	} else if !ok && mergedSchema.Extensions != nil && mergedSchema.Extensions.DynamicBlocks && len(mergedSchema.Blocks) > 0 {
		// dynamic blocks are only relevant for dependent schemas,
		// but we may end up here because the schema is a result
		// of merged static + dependent schema from previous iteration

		// propagate DynamicBlocks extension to any nested blocks
		if mergedSchema.Extensions != nil && mergedSchema.Extensions.DynamicBlocks {
			for bType, block := range mergedSchema.Blocks {
				if block.Body.Extensions == nil {
					block.Body.Extensions = &schema.BodyExtensions{}
				}
				block.Body.Extensions.DynamicBlocks = true
				mergedSchema.Blocks[bType] = block
			}
		}

		mergedSchema.Blocks["dynamic"] = buildDynamicBlockSchema(mergedSchema)
	}

	expectedDepBody := len(depKeys.Labels) > 0 || len(depKeys.Attributes) > 0

	// report success either if there wasn't any dependent body merging to do
	// or if the merging was successful

	return mergedSchema, !expectedDepBody || ok
}
