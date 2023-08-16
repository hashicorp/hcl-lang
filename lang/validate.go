// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lang

import (
	"context"

	"github.com/hashicorp/hcl/v2"
)

type ValidationFunc func(ctx context.Context) DiagnosticsMap

type DiagnosticsMap map[string]hcl.Diagnostics

func (dm DiagnosticsMap) Extend(diagMap DiagnosticsMap) DiagnosticsMap {
	for fileName, diags := range diagMap {
		_, ok := dm[fileName]
		if !ok {
			dm[fileName] = make(hcl.Diagnostics, 0)
		}

		dm[fileName].Extend(diags)
	}

	return dm
}
