// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validator

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl-lang/schemacontext"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type MinBlocks struct{}

func (v MinBlocks) Visit(ctx context.Context, node hclsyntax.Node, nodeSchema schema.Schema) (diags hcl.Diagnostics) {
	_, ok := node.(*hclsyntax.Body)
	if !ok {
		return
	}

	if nodeSchema == nil {
		return
	}

	foundBlocks := schemacontext.FoundBlocks(ctx)
	dynamicBlocks := schemacontext.DynamicBlocks(ctx)

	bodySchema := nodeSchema.(*schema.BodySchema)
	for name, blockSchema := range bodySchema.Blocks {
		if blockSchema.MinItems != 0 {
			foundBlocks, ok := foundBlocks[name]
			if (!ok || foundBlocks < blockSchema.MinItems) && !hasDynamicBlockInBody(bodySchema, dynamicBlocks, name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Too few blocks specified for %q", name),
					Detail:   fmt.Sprintf("At least %d block(s) are expected for %q", blockSchema.MinItems, name),
					Subject:  node.Range().Ptr(),
				})
			}
		}
	}

	return
}

func hasDynamicBlockInBody(bodySchema *schema.BodySchema, dynamicBlocks map[string]uint64, blockName string) bool {
	if bodySchema.Extensions == nil || !bodySchema.Extensions.DynamicBlocks {
		return false
	}

	if count, ok := dynamicBlocks[blockName]; ok && count > 0 {
		return true
	}

	return false
}
