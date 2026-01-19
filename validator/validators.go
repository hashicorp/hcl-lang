// Copyright IBM Corp. 2020, 2025
// SPDX-License-Identifier: MPL-2.0

package validator

import (
	"context"

	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Validator interface {
	Visit(ctx context.Context, node hclsyntax.Node, nodeSchema schema.Schema) (context.Context, hcl.Diagnostics)
}
