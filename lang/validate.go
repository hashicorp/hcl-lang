// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lang

import (
	"github.com/hashicorp/hcl/v2"
)

type DiagnosticsMap map[string]hcl.Diagnostics

func (dm DiagnosticsMap) Extend(diagMap DiagnosticsMap) DiagnosticsMap {
	for fileName, diags := range diagMap {
		_, ok := dm[fileName]
		if !ok {
			dm[fileName] = make(hcl.Diagnostics, 0)
		}

		dm[fileName] = dm[fileName].Extend(diags)
	}

	return dm
}

// Count returns the number of diagnostics for all files
func (dm DiagnosticsMap) Count() int {
	count := 0
	for _, diags := range dm {
		count += len(diags)
	}
	return count
}
