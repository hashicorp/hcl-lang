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

type UnexpectedBlock struct{}

func (v UnexpectedBlock) Visit(ctx context.Context, node hclsyntax.Node, nodeSchema schema.Schema) (diags hcl.Diagnostics) {
	if schemacontext.HasUnknownSchema(ctx) {
		// Avoid checking for unexpected blocks
		// if we cannot tell which ones are expected.
		return
	}

	block, ok := node.(*hclsyntax.Block)
	if !ok {
		return
	}

	if nodeSchema == nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unexpected block",
			Detail:   fmt.Sprintf("Blocks of type %q are not expected here", block.Type),
			Subject:  block.TypeRange.Ptr(),
		})
	}
	return
}
