// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"
)

type bodyActiveSelfRefsCtxKey struct{}

func WithActiveSelfRefs(ctx context.Context) context.Context {
	return context.WithValue(ctx, bodyActiveSelfRefsCtxKey{}, true)
}

func ActiveSelfRefsFromContext(ctx context.Context) bool {
	return ctx.Value(bodyActiveSelfRefsCtxKey{}) != nil
}
