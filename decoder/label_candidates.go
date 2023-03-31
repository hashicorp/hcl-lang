// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (d *PathDecoder) labelCandidatesFromDependentSchema(idx int, db map[schema.SchemaKey]*schema.BodySchema, prefixRng, editRng hcl.Range, block *hclsyntax.Block, labelSchemas []*schema.LabelSchema) (lang.Candidates, error) {
	candidates := lang.NewCandidates()
	candidates.IsComplete = true
	count := 0

	foundCandidateNames := make(map[string]bool, 0)

	prefix, _ := d.bytesFromRange(prefixRng)

	for _, schemaKey := range sortedSchemaKeys(db) {
		depKeys, err := decodeSchemaKey(schemaKey)
		if err != nil {
			// key undecodable
			continue
		}

		if uint(count) >= d.maxCandidates {
			// reached maximum no of candidates
			candidates.IsComplete = false
			break
		}

		bodySchema := db[schemaKey]

		for _, label := range depKeys.Labels {
			if label.Index != idx {
				continue
			}

			if len(prefix) > 0 && !strings.HasPrefix(label.Value, string(prefix)) {
				continue
			}

			// Dependent keys may be duplicated where one
			// key is labels-only and other one contains
			// labels + attributes.
			//
			// Specifically in Terraform this applies to
			// a resource type depending on 'provider' attribute.
			//
			// We do need such dependent keys elsewhere
			// to know how to do completion within a block
			// but this doesn't matter when completing the label itself
			// unless/until we're also completing the dependent attributes.
			if _, ok := foundCandidateNames[label.Value]; ok {
				continue
			}

			te := lang.TextEdit{}
			if d.PrefillRequiredFields {
				snippet := generateRequiredFieldsSnippet(label.Value, bodySchema, labelSchemas, 2, 0)
				te = lang.TextEdit{
					NewText: label.Value,
					Snippet: snippet,
					Range:   hcl.RangeBetween(editRng, block.OpenBraceRange),
				}
			} else {
				te = lang.TextEdit{
					NewText: label.Value,
					Snippet: label.Value,
					Range:   editRng,
				}
			}

			candidates.List = append(candidates.List, lang.Candidate{
				Label:        label.Value,
				Kind:         lang.LabelCandidateKind,
				IsDeprecated: bodySchema.IsDeprecated,
				TextEdit:     te,
				Detail:       bodySchema.Detail,
				Description:  bodySchema.Description,
			})

			foundCandidateNames[label.Value] = true
			count++
		}
	}

	sort.Sort(candidates)

	return candidates, nil
}

// generateRequiredFieldsSnippet returns a properly formatted snippet of all required
// fields (attributes, blocks, etc). It handles the main stanza declaration and calls
// `requiredFieldsSnippet` to handle recursing through the body schema
func generateRequiredFieldsSnippet(label string, bodySchema *schema.BodySchema, labelSchemas []*schema.LabelSchema, placeholder int, indentCount int) string {
	snippetText := ""

	// build a space deliminated string of dependent labels
	// In Terraform, `label` is the resource we're printing, each label after is the dependent labels
	// for example: resource "aws_instance" "foo"
	if len(labelSchemas) > 0 {
		snippetText += fmt.Sprintf("%s\"", label)
		for _, l := range labelSchemas[1:] {
			snippetText += fmt.Sprintf(" \"${%d:%s}\"", placeholder, l.Name)
			placeholder++
		}

		// must end with a newline to have a correctly formated stanza
		snippetText += " {\n"
	}

	// get all required fields and build final snippet
	snippetText += requiredFieldsSnippet(bodySchema, placeholder, indentCount)

	// add a final tabstop so that the user is landed in the correct place when
	// they are finished tabbing through each field
	snippetText += "\t${0}"

	return snippetText
}

// requiredFieldsSnippet returns a properly formatted snippet of all required
// fields (attributes, blocks). It recurses through the Body schema to
// ensure nested fields are accounted for. It takes care to add newlines and
// tabs where necessary to have a snippet be formatted correctly in the target client
func requiredFieldsSnippet(bodySchema *schema.BodySchema, placeholder int, indentCount int) string {
	// there are edge cases where we might not have a body, end early here
	if bodySchema == nil {
		return ""
	}

	snippetText := ""

	// to handle recursion we check the value here. if its 0, this is the
	// first call so we set to 0 to set a reasonable starting indent
	if indentCount == 0 {
		indentCount = 1
	}
	indent := strings.Repeat("\t", indentCount)

	// store how many required attributes there are for the current body
	reqAttr := 0
	attrNames := bodySchema.AttributeNames()
	for _, attrName := range attrNames {
		attr := bodySchema.Attributes[attrName]
		if attr.IsRequired {
			reqAttr++
		}
	}

	// iterate over each attribute, skip if not required, and print snippet
	attrCount := 0
	for _, attrName := range attrNames {
		attr := bodySchema.Attributes[attrName]
		if !attr.IsRequired {
			continue
		}

		var snippet string
		if attr.Constraint != nil {
			// We already know we want to do pre-filling at this point
			// We could plumb through the context here, but it saves us
			// an argument in multiple functions above.
			ctx := schema.WithPrefillRequiredFields(context.Background(), true)
			snippet = attr.Constraint.EmptyCompletionData(ctx, placeholder, indentCount).Snippet
		} else {
			snippet = snippetForExprContraint(uint(placeholder), attr.Expr)
		}
		snippetText += fmt.Sprintf("%s%s = %s", indent, attrName, snippet)

		// attrCount is used to tell if we are at the end of the list of attributes
		// so we don't add a trailing newline. this will affect both attribute
		// and block placement
		attrCount++
		if attrCount <= reqAttr {
			snippetText += "\n"
		}
		placeholder++
	}

	// iterate over each block, skip if not required, and print snippet
	blockTypes := bodySchema.BlockTypes()
	for _, blockType := range blockTypes {
		blockSchema := bodySchema.Blocks[blockType]
		if blockSchema.MinItems <= 0 {
			continue
		}

		// build a space deliminated string of dependent labels, if any
		labels := ""
		if len(blockSchema.Labels) > 0 {
			for _, label := range blockSchema.Labels {
				labels += fmt.Sprintf(` "${%d:%s}"`, placeholder, label.Name)
				placeholder++
			}
		}

		// newlines and indents here affect final snippet, be careful modifying order here
		snippetText += fmt.Sprintf("%s%s%s {\n", indent, blockType, labels)
		// we increment indentCount by 1 to indicate these are nested underneath
		// recurse through the body to find any attributes or blocks and print snippet
		snippetText += requiredFieldsSnippet(blockSchema.Body, placeholder, indentCount+1)
		// final newline is needed here to properly format each block
		snippetText += fmt.Sprintf("%s}\n", indent)
	}

	return snippetText
}

func sortedSchemaKeys(m map[schema.SchemaKey]*schema.BodySchema) []schema.SchemaKey {
	keys := make([]schema.SchemaKey, 0)
	for k := range m {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return string(keys[i]) < string(keys[j])
	})
	return keys
}

func decodeSchemaKey(key schema.SchemaKey) (schema.DependencyKeys, error) {
	var dk schema.DependencyKeys
	err := json.Unmarshal([]byte(key), &dk)
	return dk, err
}
