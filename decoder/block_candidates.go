// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

// blockSchemaToCandidate generates a lang.Candidate used for auto-complete inside an editor from a BlockSchema.
// If `prefillRequiredFields` is `false`, it returns a snippet that does not expect any prefilled fields.
// If `prefillRequiredFields` is `true`, it returns a snippet that is compatiable with a list of prefilled fields from `generateRequiredFieldsSnippet`
func (d *PathDecoder) blockSchemaToCandidate(blockType string, block *schema.BlockSchema, rng hcl.Range) lang.Candidate {
	triggerSuggest := false
	if len(block.Labels) > 0 {
		// We make some naive assumptions here for simplicity
		// and because this works just well enough for Terraform.
		// - if there is "completable" label it's the first one and the only one
		// - the label has some candidates based on DependentSchema
		//
		// The implementation can certainly be more sophisticated
		// but it would likely involve changes in snippet placeholder
		// numbering and full understanding of UX implications.
		triggerSuggest = block.Labels[0].IsDepKey
	}

	return lang.Candidate{
		Label:        blockType,
		Detail:       detailForBlock(block),
		Description:  block.Description,
		IsDeprecated: block.IsDeprecated,
		Kind:         lang.BlockCandidateKind,
		TextEdit: lang.TextEdit{
			NewText: blockType,
			Snippet: snippetForBlock(blockType, block, d.PrefillRequiredFields),
			Range:   rng,
		},
		TriggerSuggest: triggerSuggest,
	}
}

// detailForBlock returns a `Detail` info string to display in an editor in a hover event
func detailForBlock(block *schema.BlockSchema) string {
	detail := "Block"
	if block.Type != schema.BlockTypeNil {
		detail += fmt.Sprintf(", %s", block.Type)
	}

	if block.MinItems > 0 {
		detail += fmt.Sprintf(", min: %d", block.MinItems)
	}
	if block.MaxItems > 0 {
		detail += fmt.Sprintf(", max: %d", block.MaxItems)
	}

	return strings.TrimSpace(detail)
}

// snippetForBlock takes a block and returns a formatted snippet for a user to complete inside an editor.
// If `prefillRequiredFields` is `false`, it returns a snippet that does not expect any prefilled fields.
// If `prefillRequiredFields` is `true`, it returns a snippet that is compatiable with a list of prefilled fields from `generateRequiredFieldsSnippet`
func snippetForBlock(blockType string, block *schema.BlockSchema, prefillRequiredFields bool) string {
	if prefillRequiredFields {
		labels := ""

		depKey := false
		for _, l := range block.Labels {
			if l.IsDepKey {
				depKey = true
			}
		}

		if depKey {
			for _, l := range block.Labels {
				if l.IsDepKey {
					labels += ` "${0}"`
				} else {
					labels += fmt.Sprintf(` "%s"`, l.Name)
				}
			}
			return fmt.Sprintf("%s%s {\n}", blockType, labels)
		}

		placeholder := 1
		for _, l := range block.Labels {
			labels += fmt.Sprintf(` "${%d:%s}"`, placeholder, l.Name)
			placeholder++
		}

		return fmt.Sprintf("%s%s {\n  ${%d}\n}", blockType, labels, placeholder)
	}

	labels := ""
	placeholder := 1

	for _, l := range block.Labels {
		if l.IsDepKey {
			labels += fmt.Sprintf(` "${%d}"`, placeholder)
		} else {
			labels += fmt.Sprintf(` "${%d:%s}"`, placeholder, l.Name)
		}
		placeholder++
	}

	return fmt.Sprintf("%s%s {\n  ${%d}\n}", blockType, labels, placeholder)
}
