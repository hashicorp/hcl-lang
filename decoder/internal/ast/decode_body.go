// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ast

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// DecodeBody produces content of either HCL or JSON body
// JSON body requires schema for decoding, empty bodyContent
// is returned if nil schema is provided
func DecodeBody(body hcl.Body, bodySchema *schema.BodySchema) BodyContent {
	content := BodyContent{
		Attributes: make(hcl.Attributes, 0),
		Blocks:     make([]*BlockContent, 0),
	}

	// More common HCL syntax is processed directly (without schema)
	// which also better represents the reality in symbol lookups
	// i.e. expressions written as opposed to schema requirements
	if hclBody, ok := body.(*hclsyntax.Body); ok {
		for name, attr := range hclBody.Attributes {
			content.Attributes[name] = attr.AsHCLAttribute()
		}

		for _, block := range hclBody.Blocks {
			content.Blocks = append(content.Blocks, &BlockContent{
				Block: block.AsHCLBlock(),
				Range: block.Range(),
			})
		}

		content.RangePtr = hclBody.Range().Ptr()

		return content
	}

	// JSON syntax cannot be decoded without schema as attributes
	// and blocks are otherwise ambiguous
	if bodySchema != nil {
		hclSchema := bodySchema.ToHCLSchema()
		bContent, remainingBody, _ := body.PartialContent(hclSchema)

		content.Attributes = bContent.Attributes
		if bodySchema.AnyAttribute != nil {
			// Remaining unknown fields may also be blocks in JSON,
			// but we blindly treat them as attributes here
			// as we cannot do any better without upstream HCL changes.
			remainingAttrs, _ := remainingBody.JustAttributes()
			for name, attr := range remainingAttrs {
				content.Attributes[name] = attr
			}
		}

		for _, block := range bContent.Blocks {
			// hcl.Block interface (as the only way of accessing block in JSON)
			// does not come with Range for the block, so we calculate it here
			rng := hcl.RangeBetween(block.DefRange, block.Body.MissingItemRange())

			content.Blocks = append(content.Blocks, &BlockContent{
				Block: block,
				Range: rng,
			})
		}
	}

	return content
}
