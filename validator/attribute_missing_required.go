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

type MissingRequiredAttribute struct{}

func (v MissingRequiredAttribute) Visit(ctx context.Context, node hclsyntax.Node, nodeSchema schema.Schema) (diags hcl.Diagnostics) {
	body, ok := node.(*hclsyntax.Body)
	if !ok {
		return
	}

	if nodeSchema == nil {
		return
	}

	bodySchema := nodeSchema.(*schema.BodySchema)
	if bodySchema.Attributes == nil {
		return
	}

	for name, attr := range bodySchema.Attributes {
		if attr.IsRequired {
			_, ok := body.Attributes[name]
			if !ok {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Required attribute %q not specified", name),
					Detail:   fmt.Sprintf("An attribute named %q is required here", name),
					Subject:  body.SrcRange.Ptr(),
				})
			}
		}
	}

	return
}
