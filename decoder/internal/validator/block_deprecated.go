// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validator

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type DeprecatedBlock struct{}

func (v DeprecatedBlock) Visit(ctx context.Context, node hclsyntax.Node, nodeSchema schema.Schema) (diags hcl.Diagnostics) {
	block, ok := node.(*hclsyntax.Block)
	if !ok {
		return
	}

	if nodeSchema == nil {
		return
	}
	blockSchema := nodeSchema.(*schema.BlockSchema)
	if blockSchema.IsDeprecated {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  fmt.Sprintf("%q is deprecated", block.Type),
			Detail:   fmt.Sprintf("Reason: %q", blockSchema.Description.Value),
			Subject:  block.TypeRange.Ptr(),
		})
	}

	return
}
