package decoder

import (
	"context"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type Object struct {
	expr hcl.Expression
	cons schema.Object
}

func (obj Object) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	// TODO
	return nil
}

func (obj Object) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	// TODO
	return nil
}

func (obj Object) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	// TODO
	return nil
}

func (obj Object) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	// TODO
	return nil
}

func (obj Object) ReferenceTargets(ctx context.Context, attrAddr *schema.AttributeAddrSchema) reference.Targets {
	// TODO
	return nil
}

type ObjectAttributes struct {
	expr hcl.Expression
	cons schema.ObjectAttributes
}

func (oa ObjectAttributes) CompletionAtPos(ctx context.Context, pos hcl.Pos) []lang.Candidate {
	// TODO
	return nil
}

func (oa ObjectAttributes) HoverAtPos(ctx context.Context, pos hcl.Pos) *lang.HoverData {
	// TODO
	return nil
}

func (oa ObjectAttributes) SemanticTokens(ctx context.Context) []lang.SemanticToken {
	// TODO
	return nil
}

func (oa ObjectAttributes) ReferenceOrigins(ctx context.Context, allowSelfRefs bool) reference.Origins {
	// TODO
	return nil
}

func (oa ObjectAttributes) ReferenceTargets(ctx context.Context, attrAddr *schema.AttributeAddrSchema) reference.Targets {
	// TODO
	return nil
}
