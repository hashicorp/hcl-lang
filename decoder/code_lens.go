// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl-lang/lang"
)

// CodeLensesForFile executes any code lenses in the order declared
// and returns any errors from all of them.
// i.e. Error from an earlier lens doesn't prevent future lens
// from being executed
func (d *Decoder) CodeLensesForFile(ctx context.Context, path lang.Path, file string) ([]lang.CodeLens, error) {
	lenses := make([]lang.CodeLens, 0)

	var result *multierror.Error

	pathCtx, err := d.pathReader.PathContext(path)
	if err == nil {
		ctx = withPathContext(ctx, pathCtx)
	}

	ctx = withPathReader(ctx, d.pathReader)

	for _, clFunc := range d.ctx.CodeLenses {
		cls, err := clFunc(ctx, path, file)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		lenses = append(lenses, cls...)
	}

	return lenses, result.ErrorOrNil()
}
