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
	Visit(ctx context.Context, node hclsyntax.Node, nodeSchema schema.Schema) hcl.Diagnostics
}

// Walk walks the given node while providing schema relevant to the node.
//
// This is similar to upstream hclsyntax.Walk() which does not make it possible
// to keep track of schema.
func Walk(ctx context.Context, node hclsyntax.Node, nodeSchema schema.Schema, w Walker) hcl.Diagnostics {
	var diags hcl.Diagnostics

	switch nodeType := node.(type) {
	case *hclsyntax.Body:
		foundBlocks := make(map[string]uint64)
		dynamicBlocks := make(map[string]uint64)
		bodySchema, ok := nodeSchema.(*schema.BodySchema)
		if ok {
			for _, attr := range nodeType.Attributes {
				var attrSchema schema.Schema = nil
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

				diags = diags.Extend(Walk(ctx, attr, attrSchema, w))
			}

			for _, block := range nodeType.Blocks {
				var blockSchema schema.Schema = nil
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

				diags = diags.Extend(Walk(ctx, block, blockSchema, w))
			}
		}
		ctx = schemacontext.WithFoundBlocks(ctx, foundBlocks)
		ctx = schemacontext.WithDynamicBlocks(ctx, dynamicBlocks)

		diags = diags.Extend(w.Visit(ctx, node, nodeSchema))

	case *hclsyntax.Attribute:
		diags = diags.Extend(w.Visit(ctx, node, nodeSchema))
	case *hclsyntax.Block:
		diags = diags.Extend(w.Visit(ctx, node, nodeSchema))

		var blockBodySchema schema.Schema = nil
		bSchema, ok := nodeSchema.(*schema.BlockSchema)
		if ok && bSchema.Body != nil {
			mergedSchema, ok := schemahelper.MergeBlockBodySchemas(nodeType.AsHCLBlock(), bSchema)
			if !ok {
				ctx = schemacontext.WithUnknownSchema(ctx)
			}
			blockBodySchema = mergedSchema
		}

		diags = diags.Extend(Walk(ctx, nodeType.Body, blockBodySchema, w))

		// TODO: case hclsyntax.Expression
	}

	return diags
}
