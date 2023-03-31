// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type TypeDeclaration struct {
	expr    hcl.Expression
	cons    schema.TypeDeclaration
	pathCtx *PathContext
}

func isTypeNameWithElementOnly(name string) bool {
	return name == "list" || name == "set" || name == "map"
}
