// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lang

type CompletionHook struct {
	Name string
}

type ResolveHook struct {
	Name string `json:"resolve_hook,omitempty"`
	Path string `json:"path,omitempty"`
}

type CompletionHooks []CompletionHook

func (chs CompletionHooks) Copy() CompletionHooks {
	if chs == nil {
		return nil
	}

	hooksCopy := make(CompletionHooks, len(chs))
	copy(hooksCopy, chs)
	return hooksCopy
}
