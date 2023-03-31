// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type LiteralValue struct {
	expr    hcl.Expression
	cons    schema.LiteralValue
	pathCtx *PathContext
}

func formatNumberVal(val cty.Value) string {
	bf := val.AsBigFloat()

	if bf.IsInt() {
		intNum, _ := bf.Int64()
		return fmt.Sprintf("%d", intNum)
	}

	fNum, _ := bf.Float64()
	return strconv.FormatFloat(fNum, 'f', -1, 64)

}
