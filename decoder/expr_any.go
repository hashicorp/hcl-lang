// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Any struct {
	expr    hcl.Expression
	cons    schema.AnyExpression
	pathCtx *PathContext
}
