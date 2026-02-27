// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl-lang/decoder/internal/schemahelper"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (d *PathDecoder) HoverAtPos(ctx context.Context, filename string, pos hcl.Pos) (*lang.HoverData, error) {
	f, err := d.fileByName(filename)
	if err != nil {
		return nil, err
	}

	rootBody, err := d.bodyForFileAndPos(filename, f, pos)
	if err != nil {
		return nil, err
	}

	if d.pathCtx.Schema == nil {
		return nil, &NoSchemaError{}
	}

	data, err := d.hoverAtPos(ctx, rootBody, d.pathCtx.Schema, pos)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *PathDecoder) hoverAtPos(ctx context.Context, body *hclsyntax.Body, bodySchema *schema.BodySchema, pos hcl.Pos) (*lang.HoverData, error) {
	if bodySchema == nil {
		return nil, nil
	}

	filename := body.Range().Filename

	for name, attr := range body.Attributes {
		if attr.Range().ContainsPos(pos) {
			var aSchema *schema.AttributeSchema
			if bodySchema.Extensions != nil && bodySchema.Extensions.SelfRefs {
				ctx = schema.WithActiveSelfRefs(ctx)
			}

			if bodySchema.Extensions != nil && bodySchema.Extensions.Count && name == "count" {
				aSchema = schemahelper.CountAttributeSchema()
			} else if bodySchema.Extensions != nil && bodySchema.Extensions.ForEach && name == "for_each" {
				aSchema = schemahelper.ForEachAttributeSchema()
			} else {
				var ok bool
				aSchema, ok = bodySchema.Attributes[attr.Name]
				if !ok {
					if bodySchema.AnyAttribute == nil {
						return nil, &PositionalError{
							Filename: filename,
							Pos:      pos,
							Msg:      fmt.Sprintf("unknown attribute %q", attr.Name),
						}
					}
					aSchema = bodySchema.AnyAttribute
				}
			}

			if attr.NameRange.ContainsPos(pos) {
				return &lang.HoverData{
					Content: hoverContentForAttribute(name, aSchema),
					Range:   attr.Range(),
				}, nil
			}

			if attr.Expr.Range().ContainsPos(pos) {
				return d.newExpression(attr.Expr, aSchema.Constraint).HoverAtPos(ctx, pos), nil
			}
		}
	}

	for _, block := range body.Blocks {
		if block.Range().ContainsPos(pos) {
			blockSchema, ok := bodySchema.Blocks[block.Type]
			if !ok {
				return nil, &PositionalError{
					Filename: filename,
					Pos:      pos,
					Msg:      fmt.Sprintf("unknown block type %q", block.Type),
				}
			}

			if block.TypeRange.ContainsPos(pos) {
				return &lang.HoverData{
					Content: d.hoverContentForBlock(block.Type, blockSchema),
					Range:   block.TypeRange,
				}, nil
			}

			for i, labelRange := range block.LabelRanges {
				if labelRange.ContainsPos(pos) {
					if i+1 > len(blockSchema.Labels) {
						return nil, &PositionalError{
							Filename: filename,
							Pos:      pos,
							Msg:      fmt.Sprintf("unexpected label (%d) %q", i, block.Labels[i]),
						}
					}

					return &lang.HoverData{
						Content: d.hoverContentForLabel(i, block, blockSchema),
						Range:   labelRange,
					}, nil
				}
			}

			if isPosOutsideBody(block, pos) {
				return nil, &PositionalError{
					Filename: filename,
					Pos:      pos,
					Msg:      fmt.Sprintf("position outside of %q body", block.Type),
				}
			}

			if block.Body != nil && block.Body.Range().ContainsPos(pos) {
				mergedSchema, _ := schemahelper.MergeBlockBodySchemas(block.AsHCLBlock(), blockSchema)
				return d.hoverAtPos(ctx, block.Body, mergedSchema, pos)
			}
		}
	}

	// Position outside of any attribute or block
	return nil, &PositionalError{
		Filename: filename,
		Pos:      pos,
		Msg:      "position outside of any attribute name, value or block",
	}
}

func (d *PathDecoder) hoverContentForLabel(i int, block *hclsyntax.Block, bSchema *schema.BlockSchema) lang.MarkupContent {
	value := block.Labels[i]
	labelSchema := bSchema.Labels[i]

	if labelSchema.IsDepKey {
		bs, _, result := schemahelper.NewBlockSchema(bSchema).DependentBodySchema(block.AsHCLBlock())
		if result == schemahelper.LookupSuccessful || result == schemahelper.LookupPartiallySuccessful {
			content := fmt.Sprintf("`%s`", value)
			if bs.Detail != "" {
				content += " " + bs.Detail
			} else if labelSchema.Name != "" {
				content += " " + labelSchema.Name
			}
			if bs.Description.Value != "" {
				content += "\n\n" + bs.Description.Value
			} else if labelSchema.Description.Value != "" {
				content += "\n\n" + labelSchema.Description.Value
			}

			if bs.HoverURL != "" {
				u, err := d.docsURL(bs.HoverURL, "documentHover")
				if err == nil {
					content += fmt.Sprintf("\n\n[`%s` on %s](%s)",
						value, u.Hostname(), u.String())
				}
			}

			return lang.Markdown(content)
		}
	}

	content := fmt.Sprintf("%q", value)
	if labelSchema.Name != "" {
		content += fmt.Sprintf(" (%s)", labelSchema.Name)
	}
	content = strings.TrimSpace(content)
	if labelSchema.Description.Value != "" {
		content += "\n\n" + labelSchema.Description.Value
	}

	return lang.Markdown(content)
}

func (d *PathDecoder) hoverContentForBlock(bType string, schema *schema.BlockSchema) lang.MarkupContent {
	value := fmt.Sprintf("**%s** _%s_", bType, detailForBlock(schema))
	if schema.Description.Value != "" {
		value += fmt.Sprintf("\n\n%s", schema.Description.Value)
	}

	if schema.Body != nil && schema.Body.HoverURL != "" {
		u, err := d.docsURL(schema.Body.HoverURL, "documentHover")
		if err == nil {
			value += fmt.Sprintf("\n\n[`%s` on %s](%s)",
				bType, u.Hostname(), u.String())
		}
	}

	return lang.MarkupContent{
		Kind:  lang.MarkdownKind,
		Value: value,
	}
}

func hoverContentForReferenceTarget(ctx context.Context, ref reference.Target, pos hcl.Pos) (string, error) {
	content := fmt.Sprintf("`%s`", ref.Address(ctx, pos))

	var friendlyName string
	if ref.Type != cty.NilType {
		typeContent, err := hoverContentForType(ref.Type, 0)
		if err == nil {
			friendlyName = "\n" + typeContent
		}
	}
	if friendlyName == "" {
		friendlyName = " " + ref.FriendlyName()
	}
	content += friendlyName

	if ref.Description.Value != "" {
		content += fmt.Sprintf("\n\n%s", ref.Description.Value)
	}

	return content, nil
}

func hoverContentForType(attrType cty.Type, nestingLvl int) (string, error) {
	if attrType.IsPrimitiveType() || attrType == cty.DynamicPseudoType {
		if nestingLvl > 0 {
			return attrType.FriendlyName(), nil
		}
		return fmt.Sprintf(`_%s_`, attrType.FriendlyName()), nil
	}

	if attrType.IsObjectType() {
		attrNames := sortedObjectAttrNames(attrType)
		if len(attrNames) == 0 {
			return attrType.FriendlyName(), nil
		}
		value := ""
		if nestingLvl == 0 {
			value += "```\n"
		}
		value += "{\n"
		insideNesting := strings.Repeat("  ", nestingLvl+1)
		for _, name := range attrNames {
			valType := attrType.AttributeType(name)
			valData := valType.FriendlyNameForConstraint()

			data, err := hoverContentForType(valType, nestingLvl+1)
			if err == nil {
				valData = data
			}

			if attrType.AttributeOptional(name) {
				valData = fmt.Sprintf("optional, %s", valData)
			}

			value += fmt.Sprintf("%s%s = %s\n", insideNesting, name, valData)
		}
		endBraceNesting := strings.Repeat("  ", nestingLvl)
		value += fmt.Sprintf("%s}", endBraceNesting)
		if nestingLvl == 0 {
			value += "\n```\n_object_"
		}

		return value, nil
	}

	if attrType.IsMapType() || attrType.IsListType() || attrType.IsSetType() || attrType.IsTupleType() {
		if nestingLvl > 0 {
			return attrType.FriendlyName(), nil
		}
		value := fmt.Sprintf(`_%s_`, attrType.FriendlyName())
		return value, nil
	}

	return "", fmt.Errorf("unsupported type: %q", attrType.FriendlyName())
}
