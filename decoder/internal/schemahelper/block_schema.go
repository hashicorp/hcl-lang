// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package schemahelper

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

func MergeBlockBodySchemas(block *hcl.Block, blockSchema *schema.BlockSchema) (*schema.BodySchema, LookupResult) {
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

	depSchema, _, result := NewBlockSchema(blockSchema).DependentBodySchema(block)
	if result == LookupSuccessful || result == LookupPartiallySuccessful {
		for name, attr := range depSchema.Attributes {
			mergedSchema.Attributes[name] = attr
		}
		for bType, block := range depSchema.Blocks {
			copiedBlock := block.Copy()
			// propagate DynamicBlocks extension to any nested blocks
			if mergedSchema.Extensions != nil && mergedSchema.Extensions.DynamicBlocks {
				if copiedBlock.Body.Extensions == nil {
					copiedBlock.Body.Extensions = &schema.BodyExtensions{}
				}
				copiedBlock.Body.Extensions.DynamicBlocks = true
			}
			mergedSchema.Blocks[bType] = copiedBlock
		}

		if mergedSchema.Extensions != nil && mergedSchema.Extensions.DynamicBlocks && len(depSchema.Blocks) > 0 {
			mergedSchema.Blocks["dynamic"] = buildDynamicBlockSchema(depSchema, mergedSchema)
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
	} else if (result == LookupFailed || result == NoDependentKeys) && mergedSchema.Extensions != nil && mergedSchema.Extensions.DynamicBlocks && len(mergedSchema.Blocks) > 0 {
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

		mergedSchema.Blocks["dynamic"] = buildDynamicBlockSchema(mergedSchema, mergedSchema)
	}

	return mergedSchema, result
}
