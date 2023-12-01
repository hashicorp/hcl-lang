// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lang

import (
	"context"

	"github.com/hashicorp/hcl/v2"
)

type CodeActionImpl interface {
	CodeActionKind() CodeActionKind
	CodeActions(ctx context.Context, path Path, rng hcl.Range) []CodeAction
	// TODO: ResolveCodeAction(CodeAction) CodeAction
}

type CodeAction struct {
	Title string
	Kind  CodeActionKind

	Diagnostics hcl.Diagnostics
	Edit        Edit
	Command     Command
}

type editSigil struct{}

type Edit interface {
	isEditImpl() editSigil
}

type FileEdits []TextEdit

func (fe FileEdits) isEditImpl() editSigil {
	return editSigil{}
}

type CodeActionKind string
type CodeActionTriggerKind string
