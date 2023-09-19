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

type UnexpectedAttribute struct{}

func (v UnexpectedAttribute) Visit(ctx context.Context, node hclsyntax.Node, nodeSchema schema.Schema) (diags hcl.Diagnostics) {
	if schemacontext.HasUnknownSchema(ctx) {
		// Avoid checking for unexpected attributes
		// if we cannot tell which ones are expected.
		return
	}

	attr, ok := node.(*hclsyntax.Attribute)
	if !ok {
		return
	}

	if nodeSchema == nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unexpected attribute",
			Detail:   fmt.Sprintf("An attribute named %q is not expected here", attr.Name),
			Subject:  attr.SrcRange.Ptr(),
		})
	}

	return
}
