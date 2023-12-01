package decodercontext

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
)

type codeActionCtxKey struct{}

func CodeAction(ctx context.Context) CodeActionContext {
	return ctx.Value(codeActionCtxKey{}).(CodeActionContext)
}

func WithCodeAction(ctx context.Context, caCtx CodeActionContext) context.Context {
	return context.WithValue(ctx, codeActionCtxKey{}, caCtx)
}

type CodeActionContext struct {
	Diagnostics hcl.Diagnostics
	Only        []lang.CodeActionKind
	TriggerKind lang.CodeActionTriggerKind
}
