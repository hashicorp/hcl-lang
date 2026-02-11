// Copyright IBM Corp. 2020, 2026
// SPDX-License-Identifier: MPL-2.0

package validator

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type DeprecatedAttribute struct{}

func (v DeprecatedAttribute) Visit(ctx context.Context, node hclsyntax.Node, nodeSchema schema.Schema) (context.Context, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	attr, ok := node.(*hclsyntax.Attribute)
	if !ok {
		return ctx, diags
	}

	if nodeSchema == nil {
		return ctx, diags
	}
	attrSchema := nodeSchema.(*schema.AttributeSchema)
	if attrSchema.IsDeprecated {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  fmt.Sprintf("%q is deprecated", attr.Name),
			Detail:   fmt.Sprintf("Reason: %q", attrSchema.Description.Value),
			Subject:  attr.SrcRange.Ptr(),
		})
	}

	return ctx, diags
}
