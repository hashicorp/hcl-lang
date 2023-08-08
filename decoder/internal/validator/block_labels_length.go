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

type BlockLabelsLength struct{}

func (v BlockLabelsLength) Visit(ctx context.Context, node hclsyntax.Node, nodeSchema schema.Schema) (diags hcl.Diagnostics) {
	block, ok := node.(*hclsyntax.Block)
	if !ok {
		return
	}

	if nodeSchema == nil {
		return
	}

	blockSchema := nodeSchema.(*schema.BlockSchema)

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

	if validLabelNum > len(block.Labels) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Not enough labels specified for %q", block.Type),
			Detail:   fmt.Sprintf("All %q blocks must have %d label(s)", block.Type, validLabelNum),
			Subject:  block.TypeRange.Ptr(),
		})
	}

	return
}
