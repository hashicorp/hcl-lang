// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
)

type PathReader interface {
	Paths(ctx context.Context) []lang.Path
	PathContext(path lang.Path) (*PathContext, error)
}

type pathReaderKey struct{}

func withPathReader(ctx context.Context, pathReader PathReader) context.Context {
	return context.WithValue(ctx, pathReaderKey{}, pathReader)
}

func PathReaderFromContext(ctx context.Context) (PathReader, error) {
	pathReader, ok := ctx.Value(pathReaderKey{}).(PathReader)
	if !ok {
		return nil, fmt.Errorf("path reader not found")
	}
	return pathReader, nil
}
