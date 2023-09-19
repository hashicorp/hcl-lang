// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schemacontext

import "context"

type unknownSchemaCtxKey struct{}
type foundBlocksCtxKey struct{}
type dynamicBlocksCtxKey struct{}

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

func WithFoundBlocks(ctx context.Context, foundBlocks map[string]uint64) context.Context {
	return context.WithValue(ctx, foundBlocksCtxKey{}, foundBlocks)
}

func FoundBlocks(ctx context.Context) map[string]uint64 {
	return ctx.Value(foundBlocksCtxKey{}).(map[string]uint64)
}

func WithDynamicBlocks(ctx context.Context, dynamicBlocks map[string]uint64) context.Context {
	return context.WithValue(ctx, dynamicBlocksCtxKey{}, dynamicBlocks)
}

func DynamicBlocks(ctx context.Context) map[string]uint64 {
	return ctx.Value(dynamicBlocksCtxKey{}).(map[string]uint64)
}
