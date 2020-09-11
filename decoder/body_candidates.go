package decoder

import (
	"sort"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (d *Decoder) bodySchemaCandidates(body *hclsyntax.Body, schema *schema.BodySchema, prefixRng, editRng hcl.Range) lang.Candidates {
	prefix, _ := d.bytesFromRange(prefixRng)

	candidates := lang.NewCandidates()
	count := 0
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

		candidates.List = append(candidates.List, attributeSchemaToCandidate(name, attr, editRng))
		count++
	}

	blockTypes := sortedBlockTypes(schema.Blocks)
	for _, bType := range blockTypes {
		block := schema.Blocks[bType]

		if !isBlockDeclarable(body, bType, block) {
			continue
		}
		if len(prefix) > 0 && !strings.HasPrefix(bType, string(prefix)) {
			continue
		}
		if uint(count) >= d.maxCandidates {
			return candidates
		}

		candidates.List = append(candidates.List, blockSchemaToCandidate(bType, block, editRng))
		count++
	}

	candidates.IsComplete = true

	// TODO: sort by more metadata, such as IsRequired or IsDeprecated
	sort.Slice(candidates.List, func(i, j int) bool {
		return candidates.List[i].Label < candidates.List[j].Label
	})

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
	if attr.IsReadOnly {
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
