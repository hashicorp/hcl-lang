package decoder

import (
	"sort"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// SemanticTokensInFile returns a sequence of semantic tokens
// within the config file.
func (d *Decoder) SemanticTokensInFile(filename string) ([]lang.SemanticToken, error) {
	f, err := d.fileByName(filename)
	if err != nil {
		return nil, err
	}

	body, err := d.bodyForFileAndPos(filename, f, hcl.InitialPos)
	if err != nil {
		return nil, err
	}

	if d.rootSchema == nil {
		return []lang.SemanticToken{}, nil
	}

	tokens := tokensForBody(body, d.rootSchema, false)

	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].Range.Start.Byte < tokens[j].Range.Start.Byte
	})

	return tokens, nil
}

func tokensForBody(body *hclsyntax.Body, bodySchema *schema.BodySchema, isDependent bool) []lang.SemanticToken {
	tokens := make([]lang.SemanticToken, 0)

	if bodySchema == nil {
		return tokens
	}

	for name, attr := range body.Attributes {
		attrSchema, ok := bodySchema.Attributes[name]
		if !ok {
			if bodySchema.AnyAttribute == nil {
				// unknown attribute
				continue
			}
			attrSchema = bodySchema.AnyAttribute
		}

		modifiers := make([]lang.SemanticTokenModifier, 0)
		if isDependent {
			modifiers = append(modifiers, lang.TokenModifierDependent)
		}
		if attrSchema.IsDeprecated {
			modifiers = append(modifiers, lang.TokenModifierDeprecated)
		}

		tokens = append(tokens, lang.SemanticToken{
			Type:      lang.TokenAttrName,
			Modifiers: modifiers,
			Range:     attr.NameRange,
		})
	}

	for _, block := range body.Blocks {
		blockSchema, ok := bodySchema.Blocks[block.Type]
		if !ok {
			// unknown block
			continue
		}

		modifiers := make([]lang.SemanticTokenModifier, 0)
		if isDependent {
			modifiers = append(modifiers, lang.TokenModifierDependent)
		}
		if blockSchema.IsDeprecated {
			modifiers = append(modifiers, lang.TokenModifierDeprecated)
		}

		tokens = append(tokens, lang.SemanticToken{
			Type:      lang.TokenBlockType,
			Modifiers: modifiers,
			Range:     block.TypeRange,
		})

		for i, labelRange := range block.LabelRanges {
			if i+1 > len(blockSchema.Labels) {
				// unknown label
				continue
			}

			labelSchema := blockSchema.Labels[i]

			modifiers := make([]lang.SemanticTokenModifier, 0)
			if labelSchema.IsDepKey {
				modifiers = append(modifiers, lang.TokenModifierDependent)
			}

			tokens = append(tokens, lang.SemanticToken{
				Type:      lang.TokenBlockLabel,
				Modifiers: modifiers,
				Range:     labelRange,
			})
		}

		if block.Body != nil {
			tokens = append(tokens, tokensForBody(block.Body, blockSchema.Body, false)...)
		}

		dk := dependencyKeysFromBlock(block, blockSchema)
		depSchema, ok := blockSchema.DependentBodySchema(dk)
		if ok {
			tokens = append(tokens, tokensForBody(block.Body, depSchema, true)...)
		}
	}

	return tokens
}
