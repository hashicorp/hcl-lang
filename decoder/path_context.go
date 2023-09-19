// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl-lang/validator"
	"github.com/hashicorp/hcl/v2"
)

// PathContext represents any context relevant to the lang.Path
// i.e. anything that is tied either to path or language ID
type PathContext struct {
	Schema           *schema.BodySchema
	ReferenceOrigins reference.Origins
	ReferenceTargets reference.Targets
	Files            map[string]*hcl.File
	Functions        map[string]schema.FunctionSignature
	Validators       []validator.Validator
}

type pathCtxKey struct{}

func withPathContext(ctx context.Context, pathCtx *PathContext) context.Context {
	return context.WithValue(ctx, pathCtxKey{}, pathCtx)
}

func PathCtx(ctx context.Context) (*PathContext, error) {
	pathCtx, ok := ctx.Value(pathCtxKey{}).(*PathContext)
	if !ok {
		return nil, fmt.Errorf("path context not found")
	}
	return pathCtx, nil
}
