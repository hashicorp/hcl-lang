package decoder

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

func blockSchemaToCandidate(blockType string, block *schema.BlockSchema, rng hcl.Range) lang.Candidate {
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
			Snippet: snippetForBlock(blockType, block),
			Range:   rng,
		},
		TriggerSuggest: triggerSuggest,
	}
}

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

func snippetForBlock(blockType string, block *schema.BlockSchema) string {
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
