// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schemahelper

import "context"

type unknownSchemaCtxKey struct{}

// WithUnknownSchema attaches a flag indicating that the schema being passed
// is not wholly known.
func WithUnknownSchema(ctx context.Context) context.Context {
	return context.WithValue(ctx, unknownSchemaCtxKey{}, true)
}

// HasUnknownSchema returns true if the schema being passed
// is not wholly known. This allows each validator to decide
// whether and how to validate the AST.
func HasUnknownSchema(ctx context.Context) bool {
	value := ctx.Value(unknownSchemaCtxKey{})
	uSchema, ok := value.(bool)
	if !ok {
		return false
	}
	return uSchema
}
