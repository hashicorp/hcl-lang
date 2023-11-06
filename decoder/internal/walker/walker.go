// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package walker

import (
	"context"

	"github.com/hashicorp/hcl-lang/decoder/internal/schemahelper"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl-lang/schemacontext"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Walker interface {
	Visit(ctx context.Context, node hclsyntax.Node, nodeSchema schema.Schema) (context.Context, hcl.Diagnostics)
}

// Walk walks the given node while providing schema relevant to the node.
//
// This is similar to upstream hclsyntax.Walk() which does not make it possible
// to keep track of schema.
func Walk(ctx context.Context, node hclsyntax.Node, nodeSchema schema.Schema, w Walker) hcl.Diagnostics {
	var diags hcl.Diagnostics

	blkNestingLvl, ok := schemacontext.BlockNestingLevel(ctx)
	if !ok {
		ctx = schemacontext.WithBlockNestingLevel(ctx, 0)
		blkNestingLvl = 0
	}

	switch nodeType := node.(type) {
	case *hclsyntax.Body:
		bodyCtx := ctx
		foundBlocks := make(map[string]uint64)
		dynamicBlocks := make(map[string]uint64)

		bodySchema, bodySchemaOk := nodeSchema.(*schema.BodySchema)

		for _, attr := range nodeType.Attributes {
			var attrSchema schema.Schema = nil
			if bodySchemaOk {
				aSchema, ok := bodySchema.Attributes[attr.Name]
				if ok {
					attrSchema = aSchema
				} else if bodySchema.AnyAttribute != nil {
					attrSchema = bodySchema.AnyAttribute
				}

				if bodySchema.Extensions != nil && bodySchema.Extensions.Count && attr.Name == "count" {
					attrSchema = schemahelper.CountAttributeSchema()
				}

				if bodySchema.Extensions != nil && bodySchema.Extensions.ForEach && attr.Name == "for_each" {
					attrSchema = schemahelper.ForEachAttributeSchema()
				}
			}

			diags = diags.Extend(Walk(bodyCtx, attr, attrSchema, w))
		}

		for _, block := range nodeType.Blocks {
			var blockSchema schema.Schema = nil
			if bodySchemaOk {
				bs, ok := bodySchema.Blocks[block.Type]
				if ok {
					blockSchema = bs
				}

				if _, ok := foundBlocks[block.Type]; !ok {
					foundBlocks[block.Type] = 0
				}
				foundBlocks[block.Type]++

				if block.Type == "dynamic" {
					if len(block.Labels) > 0 {
						label := block.Labels[0]
						if _, ok := dynamicBlocks[label]; !ok {
							dynamicBlocks[label] = 0
						}
						dynamicBlocks[label]++
					}
				}
			}

			diags = diags.Extend(Walk(bodyCtx, block, blockSchema, w))
		}

		bodyCtx = schemacontext.WithFoundBlocks(bodyCtx, foundBlocks)
		bodyCtx = schemacontext.WithDynamicBlocks(bodyCtx, dynamicBlocks)

		var bodyDiags hcl.Diagnostics
		_, bodyDiags = w.Visit(bodyCtx, node, nodeSchema)
		diags = diags.Extend(bodyDiags)

	case *hclsyntax.Attribute:
		var attrDiags hcl.Diagnostics
		_, attrDiags = w.Visit(ctx, node, nodeSchema)
		diags = diags.Extend(attrDiags)
	case *hclsyntax.Block:
		var blockCtx context.Context
		var blockDiags hcl.Diagnostics

		blockCtx, blockDiags = w.Visit(ctx, node, nodeSchema)
		diags = diags.Extend(blockDiags)

		var blockBodySchema schema.Schema = nil
		bSchema, ok := nodeSchema.(*schema.BlockSchema)
		if ok && bSchema.Body != nil {
			mergedSchema, result := schemahelper.MergeBlockBodySchemas(nodeType.AsHCLBlock(), bSchema)
			if result == schemahelper.LookupFailed || result == schemahelper.LookupPartiallySuccessful {
				blockCtx = schemacontext.WithUnknownSchema(blockCtx)
			}
			blockBodySchema = mergedSchema
		}

		blockCtx = schemacontext.WithBlockNestingLevel(blockCtx, blkNestingLvl+1)
		diags = diags.Extend(Walk(blockCtx, nodeType.Body, blockBodySchema, w))

		// TODO: case hclsyntax.Expression
	}

	return diags
}
