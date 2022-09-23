package context

import (
	"context"

	"github.com/hashicorp/hcl-lang/schema"
)

type bodyExtCtxKey struct{}

type bodyActiveCountCtxKey struct{}

func WithExtensions(ctx context.Context, ext *schema.BodyExtensions) context.Context {
	return context.WithValue(ctx, bodyExtCtxKey{}, ext)
}

func WithActiveCount(ctx context.Context) context.Context {
	return context.WithValue(ctx, bodyActiveCountCtxKey{}, true)
}

func ExtensionsFromContext(ctx context.Context) (*schema.BodyExtensions, bool) {
	ext, ok := ctx.Value(bodyExtCtxKey{}).(*schema.BodyExtensions)
	return ext, ok
}

func ActiveCountFromContext(ctx context.Context) bool {
	return ctx.Value(bodyActiveCountCtxKey{}) != nil
}
