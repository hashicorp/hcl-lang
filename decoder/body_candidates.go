// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"sort"
	"strings"

	"github.com/hashicorp/hcl-lang/decoder/internal/schemahelper"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// bodySchemaCandidates returns candidates for completion of fields inside a body or block.
func (d *PathDecoder) bodySchemaCandidates(ctx context.Context, body *hclsyntax.Body, schema *schema.BodySchema, prefixRng, editRng hcl.Range) lang.Candidates {
	prefix, _ := d.bytesFromRange(prefixRng)

	candidates := lang.NewCandidates()
	count := 0

	if schema.Extensions != nil {
		// check if count attribute "extension" is enabled here
		if schema.Extensions.Count {
			// check if count attribute is already declared, so we don't
			// suggest a duplicate
			if _, ok := body.Attributes["count"]; !ok {
				candidates.List = append(candidates.List, attributeSchemaToCandidate(ctx, "count", schemahelper.CountAttributeSchema(), editRng))
			}
		}

		if schema.Extensions.ForEach {
			// check if for_each attribute is already declared, so we don't
			// suggest a duplicate
			if _, present := body.Attributes["for_each"]; !present {
				candidates.List = append(candidates.List, attributeSchemaToCandidate(ctx, "for_each", schemahelper.ForEachAttributeSchema(), editRng))
			}
		}
	}

	if len(schema.Attributes) > 0 {
		attrNames := sortedAttributeNames(schema.Attributes)
		for _, name := range attrNames {
			attr := schema.Attributes[name]

			if !isAttributeDeclarable(body, name, attr) {
				continue
			}
			if len(prefix) > 0 && !strings.HasPrefix(name, string(prefix)) {
				continue
			}
			if uint(count) >= d.maxCandidates {
				return candidates
			}

			candidates.List = append(candidates.List, attributeSchemaToCandidate(ctx, name, attr, editRng))
			count++
		}
	} else if attr := schema.AnyAttribute; attr != nil && len(prefix) == 0 {
		if uint(count) >= d.maxCandidates {
			return candidates
		}

		candidates.List = append(candidates.List, attributeSchemaToCandidate(ctx, "name", attr, editRng))
		count++
	}

	blockTypes := sortedBlockTypes(schema.Blocks)
	for _, bType := range blockTypes {
		block := schema.Blocks[bType]

		// In Terraform duplicates should never occur when providers
		// use the official plugin SDK, except when a field uses
		// SchemaConfigMode to turn blocks into attributes
		// for backwards compatibility reasons.
		//
		// Decoder allows duplicates but they should be avoided if possible.
		//
		// Here we prefer attribute completion in case of a duplicate
		// to mimic how Terraform Core treats duplicate list(object)
		// and set(object) attributes.
		if _, ok := schema.Attributes[bType]; ok {
			continue
		}

		if !isBlockDeclarable(body, bType, block) {
			continue
		}
		if len(prefix) > 0 && !strings.HasPrefix(bType, string(prefix)) {
			continue
		}
		if uint(count) >= d.maxCandidates {
			return candidates
		}

		candidates.List = append(candidates.List, d.blockSchemaToCandidate(bType, block, editRng))
		count++
	}

	candidates.IsComplete = true

	sort.Sort(candidates)

	return candidates
}

func sortedAttributeNames(attrs map[string]*schema.AttributeSchema) []string {
	names := make([]string, len(attrs))
	i := 0
	for name := range attrs {
		names[i] = name
		i++
	}
	sort.Strings(names)
	return names
}

func sortedBlockTypes(blocks map[string]*schema.BlockSchema) []string {
	bTypes := make([]string, len(blocks))
	i := 0
	for bType := range blocks {
		bTypes[i] = bType
		i++
	}
	sort.Strings(bTypes)
	return bTypes
}

func isAttributeDeclarable(body *hclsyntax.Body, name string, attr *schema.AttributeSchema) bool {
	if attr.IsComputed && !attr.IsOptional {
		return false
	}

	for attrName := range body.Attributes {
		if attrName == name {
			return false
		}
	}

	return true
}

func isBlockDeclarable(body *hclsyntax.Body, blockType string, bSchema *schema.BlockSchema) bool {
	if bSchema.MaxItems == 0 {
		return true
	}

	itemCount := uint64(0)
	for _, block := range body.Blocks {
		if block.Type == blockType {
			itemCount++
			if itemCount >= bSchema.MaxItems {
				return false
			}
		}
	}
	return true
}
