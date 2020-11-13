package decoder

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func (d *Decoder) HoverAtPos(filename string, pos hcl.Pos) (*lang.HoverData, error) {
	f, err := d.fileByName(filename)
	if err != nil {
		return nil, err
	}

	rootBody, err := d.bodyForFileAndPos(filename, f, pos)
	if err != nil {
		return nil, err
	}

	d.rootSchemaMu.RLock()
	defer d.rootSchemaMu.RUnlock()

	if d.rootSchema == nil {
		return nil, &NoSchemaError{}
	}

	data, err := d.hoverAtPos(rootBody, d.rootSchema, pos)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *Decoder) hoverAtPos(body *hclsyntax.Body, bodySchema *schema.BodySchema, pos hcl.Pos) (*lang.HoverData, error) {
	if bodySchema == nil {
		return nil, nil
	}

	filename := body.Range().Filename

	for name, attr := range body.Attributes {
		if attr.Range().ContainsPos(pos) {
			aSchema, ok := bodySchema.Attributes[attr.Name]
			if !ok {
				return nil, &PositionalError{
					Filename: filename,
					Pos:      pos,
					Msg:      fmt.Sprintf("unknown attribute %q", attr.Name),
				}
			}

			return &lang.HoverData{
				Content: hoverContentForAttribute(name, aSchema),
				Range:   attr.Range(),
			}, nil
		}
	}

	for _, block := range body.Blocks {
		if block.Range().ContainsPos(pos) {
			bSchema, ok := bodySchema.Blocks[block.Type]
			if !ok {
				return nil, &PositionalError{
					Filename: filename,
					Pos:      pos,
					Msg:      fmt.Sprintf("unknown block type %q", block.Type),
				}
			}

			if block.TypeRange.ContainsPos(pos) {
				return &lang.HoverData{
					Content: hoverContentForBlock(block.Type, bSchema),
					Range:   block.TypeRange,
				}, nil
			}

			for i, labelRange := range block.LabelRanges {
				if labelRange.ContainsPos(pos) {
					if i+1 > len(bSchema.Labels) {
						return nil, &PositionalError{
							Filename: filename,
							Pos:      pos,
							Msg:      fmt.Sprintf("unexpected label (%d) %q", i, block.Labels[i]),
						}
					}

					return &lang.HoverData{
						Content: hoverContentForLabel(i, block, bSchema),
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
				mergedSchema, err := mergeBlockBodySchemas(block, bSchema)
				if err != nil {
					return nil, err
				}

				return d.hoverAtPos(block.Body, mergedSchema, pos)
			}
		}
	}

	// Position outside of any attribute or block
	return nil, &PositionalError{
		Filename: filename,
		Pos:      pos,
		Msg:      "position outside of any attribute or block",
	}
}

func hoverContentForLabel(i int, block *hclsyntax.Block, bSchema *schema.BlockSchema) lang.MarkupContent {
	value := block.Labels[i]
	labelSchema := bSchema.Labels[i]

	if labelSchema.IsDepKey {
		dk := dependencyKeysFromBlock(block, bSchema)
		bs, ok := bSchema.DependentBodySchema(dk)
		if ok {
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

func hoverContentForAttribute(name string, schema *schema.AttributeSchema) lang.MarkupContent {
	value := fmt.Sprintf("**%s** _%s_", name, detailForAttribute(schema))
	if schema.Description.Value != "" {
		value += fmt.Sprintf("\n\n%s", schema.Description.Value)
	}
	return lang.MarkupContent{
		Kind:  lang.MarkdownKind,
		Value: value,
	}
}

func hoverContentForBlock(bType string, schema *schema.BlockSchema) lang.MarkupContent {
	value := fmt.Sprintf("**%s** _%s_", bType, detailForBlock(schema))
	if schema.Description.Value != "" {
		value += fmt.Sprintf("\n\n%s", schema.Description.Value)
	}
	return lang.MarkupContent{
		Kind:  lang.MarkdownKind,
		Value: value,
	}
}
