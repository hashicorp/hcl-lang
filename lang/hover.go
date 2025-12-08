// Copyright IBM Corp. 2020, 2025
// SPDX-License-Identifier: MPL-2.0

package lang

import (
	"github.com/hashicorp/hcl/v2"
)

type HoverData struct {
	Content MarkupContent
	Range   hcl.Range
}
