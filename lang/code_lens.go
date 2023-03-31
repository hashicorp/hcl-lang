// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lang

import (
	"context"

	"github.com/hashicorp/hcl/v2"
)

type CodeLensFunc func(ctx context.Context, path Path, file string) ([]CodeLens, error)

type CodeLens struct {
	Range   hcl.Range
	Command Command
}

type Command struct {
	Title     string
	ID        string
	Arguments []CommandArgument
}

type CommandArgument interface {
	MarshalJSON() ([]byte, error)
}
