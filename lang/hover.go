// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package lang

import (
	"github.com/hashicorp/hcl/v2"
)

type HoverData struct {
	Content MarkupContent
	Range   hcl.Range
}
