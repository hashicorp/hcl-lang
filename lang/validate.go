// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lang

import (
	"context"

	"github.com/hashicorp/hcl/v2"
)

type ValidationFunc func(ctx context.Context) hcl.Diagnostics
