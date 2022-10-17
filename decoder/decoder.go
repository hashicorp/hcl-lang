package decoder

import (
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type Decoder struct {
	ctx        DecoderContext
	pathReader PathReader
}

// NewDecoder creates a new Decoder
//
// Decoder is safe for use without any schema, but configuration files are loaded
// via LoadFile and (optionally) schema is set via SetSchema.
func NewDecoder(pathReader PathReader) *Decoder {
	return &Decoder{
		pathReader: pathReader,
	}
}

func posEqual(pos, other hcl.Pos) bool {
	return pos.Line == other.Line &&
		pos.Column == other.Column &&
		pos.Byte == other.Byte
}

func mergeBlockBodySchemas(block *hcl.Block, blockSchema *schema.BlockSchema) (*schema.BodySchema, error) {
	if len(blockSchema.DependentBody) == 0 {
		return blockSchema.Body, nil
	}

	mergedSchema := &schema.BodySchema{}
	if blockSchema.Body != nil {
		mergedSchema = blockSchema.Body.Copy()
	}
	if mergedSchema.Attributes == nil {
		mergedSchema.Attributes = make(map[string]*schema.AttributeSchema, 0)
	}
	if mergedSchema.Blocks == nil {
		mergedSchema.Blocks = make(map[string]*schema.BlockSchema, 0)
	}
	if mergedSchema.TargetableAs == nil {
		mergedSchema.TargetableAs = make([]*schema.Targetable, 0)
	}
	if mergedSchema.ImpliedOrigins == nil {
		mergedSchema.ImpliedOrigins = make([]schema.ImpliedOrigin, 0)
	}

	depSchema, _, ok := NewBlockSchema(blockSchema).DependentBodySchema(block)
	if ok {
		for name, attr := range depSchema.Attributes {
			if _, exists := mergedSchema.Attributes[name]; !exists {
				mergedSchema.Attributes[name] = attr
			} else {
				// Skip duplicate attribute
				continue
			}
		}
		for bType, block := range depSchema.Blocks {
			if _, exists := mergedSchema.Blocks[bType]; !exists {
				mergedSchema.Blocks[bType] = block
			} else {
				// Skip duplicate block type
				continue
			}
		}

		mergedSchema.TargetableAs = append(mergedSchema.TargetableAs, depSchema.TargetableAs...)
		mergedSchema.ImpliedOrigins = append(mergedSchema.ImpliedOrigins, depSchema.ImpliedOrigins...)

		// TODO: avoid resetting?
		mergedSchema.Targets = depSchema.Targets.Copy()

		// TODO: avoid resetting?
		mergedSchema.DocsLink = depSchema.DocsLink.Copy()

		// use extensions of DependentBody if not nil
		// (to avoid resetting to nil)
		if depSchema.Extensions != nil {
			mergedSchema.Extensions = depSchema.Extensions.Copy()
		}
	}

	return mergedSchema, nil
}

// blockContent represents HCL or JSON block content
type blockContent struct {
	*hcl.Block

	// Range represents range of the block in HCL syntax
	// or closest available representative range in JSON
	Range hcl.Range
}

// bodyContent represents an HCL or JSON body content
type bodyContent struct {
	Attributes hcl.Attributes
	Blocks     []*blockContent
}

// decodeBody produces content of either HCL or JSON body
//
// JSON body requires schema for decoding, empty bodyContent
// is returned if nil schema is provided
func decodeBody(body hcl.Body, bodySchema *schema.BodySchema) bodyContent {
	content := bodyContent{
		Attributes: make(hcl.Attributes, 0),
		Blocks:     make([]*blockContent, 0),
	}

	// More common HCL syntax is processed directly (without schema)
	// which also better represents the reality in symbol lookups
	// i.e. expressions written as opposed to schema requirements
	if hclBody, ok := body.(*hclsyntax.Body); ok {
		for name, attr := range hclBody.Attributes {
			content.Attributes[name] = attr.AsHCLAttribute()
		}

		for _, block := range hclBody.Blocks {
			content.Blocks = append(content.Blocks, &blockContent{
				Block: block.AsHCLBlock(),
				Range: block.Range(),
			})
		}

		return content
	}

	// JSON syntax cannot be decoded without schema as attributes
	// and blocks are otherwise ambiguous
	if bodySchema != nil {
		hclSchema := bodySchema.ToHCLSchema()
		bContent, remainingBody, _ := body.PartialContent(hclSchema)

		content.Attributes = bContent.Attributes
		if bodySchema.AnyAttribute != nil {
			// Remaining unknown fields may also be blocks in JSON,
			// but we blindly treat them as attributes here
			// as we cannot do any better without upstream HCL changes.
			remainingAttrs, _ := remainingBody.JustAttributes()
			for name, attr := range remainingAttrs {
				content.Attributes[name] = attr
			}
		}

		for _, block := range bContent.Blocks {
			// hcl.Block interface (as the only way of accessing block in JSON)
			// does not come with Range for the block, so we calculate it here
			rng := hcl.RangeBetween(block.DefRange, block.Body.MissingItemRange())

			content.Blocks = append(content.Blocks, &blockContent{
				Block: block,
				Range: rng,
			})
		}
	}

	return content
}

func stringPos(pos hcl.Pos) string {
	return fmt.Sprintf("%d,%d", pos.Line, pos.Column)
}

func countAttributeSchema() *schema.AttributeSchema {
	return &schema.AttributeSchema{
		IsOptional: true,
		Expr: schema.ExprConstraints{
			schema.TraversalExpr{OfType: cty.Number},
			schema.LiteralTypeExpr{Type: cty.Number},
		},
		Description: lang.Markdown("The distinct index number (starting with 0) corresponding to the instance"),
	}
}

func forEachAttributeSchema() *schema.AttributeSchema {
	return &schema.AttributeSchema{
		IsOptional: true,
		Expr: schema.ExprConstraints{
			schema.TraversalExpr{OfType: cty.Map(cty.DynamicPseudoType)},
			schema.TraversalExpr{OfType: cty.Set(cty.String)},
			schema.LiteralTypeExpr{Type: cty.Map(cty.DynamicPseudoType)},
			schema.LiteralTypeExpr{Type: cty.Set(cty.String)},
		},
		Description: lang.Markdown("A meta-argument that accepts a map or a set of strings, and creates an instance for each item in that map or set.\n\n" +
			"**Note**: A given block cannot use both `count` and `for_each`."),
	}
}

func dynamicBlockSchema() *schema.BlockSchema {
	return &schema.BlockSchema{
		Description: lang.Markdown("A dynamic block to produce blocks dynamically by iterating over a given complex value"),
		Type:        schema.BlockTypeMap,
		Labels: []*schema.LabelSchema{
			{Name: "name"},
		},
		Body: &schema.BodySchema{
			Attributes: map[string]*schema.AttributeSchema{
				"for_each": {
					Expr: schema.ExprConstraints{
						schema.TraversalExpr{OfType: cty.Map(cty.DynamicPseudoType)},
						schema.TraversalExpr{OfType: cty.Set(cty.String)},
						schema.LiteralTypeExpr{Type: cty.Map(cty.DynamicPseudoType)},
						schema.LiteralTypeExpr{Type: cty.Set(cty.String)},
					},
					IsRequired: true,
					Description: lang.Markdown("A meta-argument that accepts a map or a set of strings, and creates an instance for each item in that map or set.\n\n" +
						"**Note**: A given block cannot use both `count` and `for_each`."),
				},
				"iterator": {
					Expr:       schema.LiteralTypeOnly(cty.String),
					IsOptional: true,
					Description: lang.Markdown("The name of a temporary variable that represents the current " +
						"element of the complex value. Defaults to the label of the dynamic block."),
				},
				"labels": {
					Expr: schema.ExprConstraints{
						schema.ListExpr{
							Elem: schema.ExprConstraints{
								schema.LiteralTypeExpr{Type: cty.String},
								schema.TraversalExpr{OfType: cty.String},
							},
						},
					},
					IsOptional: true,
					Description: lang.Markdown("A list of strings that specifies the block labels, " +
						"in order, to use for each generated block."),
				},
			},
			Blocks: map[string]*schema.BlockSchema{
				"content": {
					Description: lang.PlainText("The body of each generated block"),
				},
			},
		},
	}
}

func countIndexHoverData(rng hcl.Range) *lang.HoverData {
	return &lang.HoverData{
		Content: lang.Markdown("`count.index` _number_\n\nThe distinct index number (starting with 0) corresponding to the instance"),
		Range:   rng,
	}
}

func countIndexCandidate(editRng hcl.Range) lang.Candidate {
	return lang.Candidate{
		Label:       "count.index",
		Detail:      "number",
		Description: lang.PlainText("The distinct index number (starting with 0) corresponding to the instance"),
		Kind:        lang.TraversalCandidateKind,
		TextEdit: lang.TextEdit{
			NewText: "count.index",
			Snippet: "count.index",
			Range:   editRng,
		},
	}
}

func foreachEachCandidate(editRng hcl.Range) []lang.Candidate {
	return []lang.Candidate{
		{
			Label:  "each.key",
			Detail: "string",
			Description: lang.MarkupContent{
				Value: "The map key (or set member) corresponding to this instance",
				Kind:  lang.MarkdownKind,
			},
			Kind: lang.TraversalCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: "each.key",
				Snippet: "each.key",
				Range:   editRng,
			},
		},
		{
			Label:  "each.value",
			Detail: "any type",
			Description: lang.MarkupContent{
				Value: "The map value corresponding to this instance. (If a set was provided, this is the same as `each.key`.)",
				Kind:  lang.MarkdownKind,
			},
			Kind: lang.TraversalCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: "each.value",
				Snippet: "each.value",
				Range:   editRng,
			},
		},
	}
}

func eachKeyHoverData(rng hcl.Range) *lang.HoverData {
	return &lang.HoverData{
		Content: lang.Markdown("`each.key` _string_\n\nThe map key (or set member) corresponding to this instance"),
		Range:   rng,
	}
}

func eachValueHoverData(rng hcl.Range) *lang.HoverData {
	return &lang.HoverData{
		Content: lang.Markdown("`each.value` _any type_\n\nThe map value corresponding to this instance. (If a set was provided, this is the same as `each.key`.)"),
		Range:   rng,
	}
}
