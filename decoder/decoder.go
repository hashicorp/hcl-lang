package decoder

import (
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
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
				if mergedSchema.Extensions != nil && mergedSchema.Extensions.DynamicBlocks {
					if block.Body.Extensions == nil {
						block.Body.Extensions = &schema.BodyExtensions{}
					}
					block.Body.Extensions.DynamicBlocks = true
				}
				mergedSchema.Blocks[bType] = block
			} else {
				// Skip duplicate block type
				continue
			}
		}

		if mergedSchema.Extensions != nil && mergedSchema.Extensions.DynamicBlocks && len(depSchema.Blocks) > 0 {
			mergedSchema.Blocks["dynamic"] = buildDynamicBlockSchema(depSchema)
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
	} else if !ok && mergedSchema.Extensions != nil && mergedSchema.Extensions.DynamicBlocks && len(mergedSchema.Blocks) > 0 {
		mergedSchema.Blocks["dynamic"] = buildDynamicBlockSchema(mergedSchema)
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
	RangePtr   *hcl.Range
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

		content.RangePtr = hclBody.Range().Ptr()

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
		Description: lang.Markdown("Total number of instances of this block.\n\n" +
			"**Note**: A given block cannot use both `count` and `for_each`."),
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

func buildDynamicBlockSchema(inputSchema *schema.BodySchema) *schema.BlockSchema {
	dependentBody := make(map[schema.SchemaKey]*schema.BodySchema)
	for blockName, block := range inputSchema.Blocks {
		dependentBody[schema.NewSchemaKey(schema.DependencyKeys{
			Labels: []schema.LabelDependent{
				{Index: 0, Value: blockName},
			},
		})] = &schema.BodySchema{
			Blocks: map[string]*schema.BlockSchema{
				"content": {
					Description: lang.PlainText("The body of each generated block"),
					MaxItems:    1,
					Body:        block.Body.Copy(),
				},
			},
		}
	}

	return &schema.BlockSchema{
		Description: lang.Markdown("A dynamic block to produce blocks dynamically by iterating over a given complex value"),
		Labels: []*schema.LabelSchema{
			{
				Name:        "name",
				Completable: true,
				IsDepKey:    true,
			},
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
					IsRequired:  true,
					Description: lang.Markdown("A meta-argument that accepts a map or a set of strings, and creates an instance for each item in that map or set."),
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
		},
		DependentBody: dependentBody,
	}
}

func countIndexReferenceTarget(attr *hcl.Attribute, bodyRange hcl.Range) reference.Target {
	return reference.Target{
		LocalAddr: lang.Address{
			lang.RootStep{Name: "count"},
			lang.AttrStep{Name: "index"},
		},
		TargetableFromRangePtr: bodyRange.Ptr(),
		Type:                   cty.Number,
		Description:            lang.Markdown("The distinct index number (starting with 0) corresponding to the instance"),
		RangePtr:               attr.Range.Ptr(),
		DefRangePtr:            attr.NameRange.Ptr(),
	}
}

func forEachReferenceTargets(attr *hcl.Attribute, bodyRange hcl.Range) reference.Targets {
	return reference.Targets{
		{
			LocalAddr: lang.Address{
				lang.RootStep{Name: "each"},
				lang.AttrStep{Name: "key"},
			},
			TargetableFromRangePtr: bodyRange.Ptr(),
			Type:                   cty.String,
			Description:            lang.Markdown("The map key (or set member) corresponding to this instance"),
			RangePtr:               attr.Range.Ptr(),
			DefRangePtr:            attr.NameRange.Ptr(),
		},
		{
			LocalAddr: lang.Address{
				lang.RootStep{Name: "each"},
				lang.AttrStep{Name: "value"},
			},
			TargetableFromRangePtr: bodyRange.Ptr(),
			Type:                   cty.DynamicPseudoType,
			Description:            lang.Markdown("The map value corresponding to this instance. (If a set was provided, this is the same as `each.key`.)"),
			RangePtr:               attr.Range.Ptr(),
			DefRangePtr:            attr.NameRange.Ptr(),
		},
	}
}
