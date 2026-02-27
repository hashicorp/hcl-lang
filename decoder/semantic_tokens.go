// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"sort"

	"github.com/hashicorp/hcl-lang/decoder/internal/schemahelper"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// SemanticTokensInFile returns a sequence of semantic tokens
// within the config file.
func (d *PathDecoder) SemanticTokensInFile(ctx context.Context, filename string) ([]lang.SemanticToken, error) {
	f, err := d.fileByName(filename)
	if err != nil {
		return nil, err
	}

	body, err := d.bodyForFileAndPos(filename, f, hcl.InitialPos)
	if err != nil {
		return nil, err
	}

	if d.pathCtx.Schema == nil {
		return []lang.SemanticToken{}, nil
	}

	tokens := d.tokensForBody(ctx, body, d.pathCtx.Schema, []lang.SemanticTokenModifier{})

	// TODO decouple semantic tokens for valid references from AST walking
	//   instead of matching targets and origins when encountering a traversal expression,
	//   we can do this way earlier by comparing pathCtx.ReferenceTargets and
	//   d.pathCtx.ReferenceOrigins, to build a list of tokens.
	//   Be sure to sort them afterward!

	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].Range.Start.Byte < tokens[j].Range.Start.Byte
	})

	return tokens, nil
}

func (d *PathDecoder) tokensForBody(ctx context.Context, body *hclsyntax.Body, bodySchema *schema.BodySchema, parentModifiers []lang.SemanticTokenModifier) []lang.SemanticToken {
	tokens := make([]lang.SemanticToken, 0)

	if bodySchema == nil {
		return tokens
	}

	for name, attr := range body.Attributes {
		attrSchema, ok := bodySchema.Attributes[name]
		if !ok {
			if bodySchema.Extensions != nil && name == "count" && bodySchema.Extensions.Count {
				attrSchema = schemahelper.CountAttributeSchema()
			} else if bodySchema.Extensions != nil && name == "for_each" && bodySchema.Extensions.ForEach {
				attrSchema = schemahelper.ForEachAttributeSchema()
			} else {
				if bodySchema.AnyAttribute == nil {
					// unknown attribute
					continue
				}
				attrSchema = bodySchema.AnyAttribute
			}
		}

		attrModifiers := make([]lang.SemanticTokenModifier, 0)
		attrModifiers = append(attrModifiers, parentModifiers...)
		attrModifiers = append(attrModifiers, attrSchema.SemanticTokenModifiers...)

		tokens = append(tokens, lang.SemanticToken{
			Type:      lang.TokenAttrName,
			Modifiers: attrModifiers,
			Range:     attr.NameRange,
		})

		tokens = append(tokens, d.newExpression(attr.Expr, attrSchema.Constraint).SemanticTokens(ctx)...)
	}

	for _, block := range body.Blocks {
		blockSchema, hasDepSchema := bodySchema.Blocks[block.Type]
		if !hasDepSchema {
			// unknown block
			continue
		}

		blockModifiers := make([]lang.SemanticTokenModifier, 0)
		blockModifiers = append(blockModifiers, parentModifiers...)
		blockModifiers = append(blockModifiers, blockSchema.SemanticTokenModifiers...)

		tokens = append(tokens, lang.SemanticToken{
			Type:      lang.TokenBlockType,
			Modifiers: blockModifiers,
			Range:     block.TypeRange,
		})

		for i, labelRange := range block.LabelRanges {
			if i+1 > len(blockSchema.Labels) {
				// unknown label
				continue
			}

			labelSchema := blockSchema.Labels[i]

			labelModifiers := make([]lang.SemanticTokenModifier, 0)
			labelModifiers = append(labelModifiers, parentModifiers...)
			labelModifiers = append(labelModifiers, blockSchema.SemanticTokenModifiers...)
			labelModifiers = append(labelModifiers, labelSchema.SemanticTokenModifiers...)

			tokens = append(tokens, lang.SemanticToken{
				Type:      lang.TokenBlockLabel,
				Modifiers: labelModifiers,
				Range:     labelRange,
			})
		}

		if block.Body != nil {
			mergedSchema, _ := schemahelper.MergeBlockBodySchemas(block.AsHCLBlock(), blockSchema)

			tokens = append(tokens, d.tokensForBody(ctx, block.Body, mergedSchema, blockModifiers)...)
		}
	}

	return tokens
}

func isPrimitiveTypeDeclaration(kw string) bool {
	switch kw {
	case "bool":
		return true
	case "number":
		return true
	case "string":
		return true
	case "null":
		return true
	case "any":
		return true
	}
	return false
}
