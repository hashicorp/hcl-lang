package schema

import (
	"context"
)

type bodyExtCtxKey struct{}

type bodyActiveCountCtxKey struct{}

func WithActiveCount(ctx context.Context) context.Context {
	return context.WithValue(ctx, bodyActiveCountCtxKey{}, true)
}

func ActiveCountFromContext(ctx context.Context) bool {
	return ctx.Value(bodyActiveCountCtxKey{}) != nil
}

type bodyActiveForEachCtxKey struct{}

func WithActiveForEach(ctx context.Context) context.Context {
	return context.WithValue(ctx, bodyActiveForEachCtxKey{}, true)
}

func ActiveForEachFromContext(ctx context.Context) bool {
	return ctx.Value(bodyActiveForEachCtxKey{}) != nil
}

type bodyActiveSelfRefsCtxKey struct{}

func WithActiveSelfRefs(ctx context.Context) context.Context {
	return context.WithValue(ctx, bodyActiveSelfRefsCtxKey{}, true)
}

func ActiveSelfRefsFromContext(ctx context.Context) bool {
	return ctx.Value(bodyActiveSelfRefsCtxKey{}) != nil
}
