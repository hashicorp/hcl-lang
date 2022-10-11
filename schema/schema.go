package schema

import (
	"context"
)

type schemaImplSigil struct{}

// Schema represents any schema (e.g. attribute, label, or a block)
type Schema interface {
	isSchemaImpl() schemaImplSigil
}

type bodyExtCtxKey struct{}

type bodyActiveCountCtxKey struct{}

func WithActiveCount(ctx context.Context) context.Context {
	return context.WithValue(ctx, bodyActiveCountCtxKey{}, true)
}

func ActiveCountFromContext(ctx context.Context) bool {
	return ctx.Value(bodyActiveCountCtxKey{}) != nil
}
