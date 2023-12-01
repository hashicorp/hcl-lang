// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeaction

import (
	"context"

	"github.com/hashicorp/hcl-lang/decodercontext"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
)

type AttributeMissingNewline struct{}

type DataWithCode interface {
	Code() string
}

func (v AttributeMissingNewline) CodeActions(ctx context.Context, path lang.Path, rng hcl.Range) []lang.CodeAction {
	actions := make([]lang.CodeAction, 0)

	caCtx := decodercontext.CodeAction(ctx)
	for _, diag := range caCtx.Diagnostics {
		extra, ok := diag.Extra.(DataWithCode)
		if !ok {
			continue
		}

		if extra.Code() == "AttrMissingNewline" {
			actions = append(actions, lang.CodeAction{
				Title:       "Add missing newline",
				Kind:        "quickfix",
				Diagnostics: hcl.Diagnostics{diag},
				Edit: lang.FileEdits{
					lang.TextEdit{
						Range: hcl.Range{
							Filename: diag.Subject.Filename,
							Start:    diag.Subject.Start,
							End:      diag.Subject.Start,
						},
						NewText: "\n",
						Snippet: "\n",
					},
				},
			})
		}
	}

	return actions
}
