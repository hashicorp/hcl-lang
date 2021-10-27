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
func (d *PathDecoder) CodeLensesForFile(ctx context.Context, file string) ([]lang.CodeLens, error) {
	lenses := make([]lang.CodeLens, 0)

	var result *multierror.Error

	for _, clFunc := range d.decoderCtx.CodeLenses {
		ctx = withPathContext(ctx, d.pathCtx)

		cls, err := clFunc(ctx, d.path, file)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}
		lenses = append(lenses, cls...)
	}

	return lenses, result.ErrorOrNil()
}
