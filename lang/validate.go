package lang

import (
	"context"

	"github.com/hashicorp/hcl/v2"
)

type ValidationFunc func(ctx context.Context) hcl.Diagnostics
