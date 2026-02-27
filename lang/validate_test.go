// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package lang

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty-debug/ctydebug"
)

func TestDiagnosticsMap_Extend(t *testing.T) {
	testCases := []struct {
		name          string
		baseMap       DiagnosticsMap
		additionalMap DiagnosticsMap
		expectedMap   DiagnosticsMap
	}{
		{
			"empty base",
			DiagnosticsMap{},
			DiagnosticsMap{
				"main.tf": hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  "An error",
						Subject: &hcl.Range{
							Filename: "main.tf",
						},
					},
				},
			},
			DiagnosticsMap{
				"main.tf": hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  "An error",
						Subject: &hcl.Range{
							Filename: "main.tf",
						},
					},
				},
			},
		},
		{
			"same file diag",
			DiagnosticsMap{
				"main.tf": hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  "An error",
						Subject: &hcl.Range{
							Filename: "main.tf",
						},
					},
				},
			},
			DiagnosticsMap{
				"main.tf": hcl.Diagnostics{
					{
						Severity: hcl.DiagWarning,
						Summary:  "A warning",
						Subject: &hcl.Range{
							Filename: "main.tf",
						},
					},
				},
			},
			DiagnosticsMap{
				"main.tf": hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  "An error",
						Subject: &hcl.Range{
							Filename: "main.tf",
						},
					},
					{
						Severity: hcl.DiagWarning,
						Summary:  "A warning",
						Subject: &hcl.Range{
							Filename: "main.tf",
						},
					},
				},
			},
		},
		{
			"new file diag",
			DiagnosticsMap{
				"main.tf": hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  "An error",
						Subject: &hcl.Range{
							Filename: "main.tf",
						},
					},
				},
			},
			DiagnosticsMap{
				"test.tf": hcl.Diagnostics{
					{
						Severity: hcl.DiagWarning,
						Summary:  "A warning",
						Subject: &hcl.Range{
							Filename: "test.tf",
						},
					},
				},
			},
			DiagnosticsMap{
				"main.tf": hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  "An error",
						Subject: &hcl.Range{
							Filename: "main.tf",
						},
					},
				},
				"test.tf": hcl.Diagnostics{
					{
						Severity: hcl.DiagWarning,
						Summary:  "A warning",
						Subject: &hcl.Range{
							Filename: "test.tf",
						},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%2d-%s", i, tc.name), func(t *testing.T) {
			result := tc.baseMap.Extend(tc.additionalMap)

			if diff := cmp.Diff(tc.expectedMap, result, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch: %s", diff)
			}
		})
	}
}
