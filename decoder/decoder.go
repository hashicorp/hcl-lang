package decoder

import (
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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
		mergedSchema.Targets = depSchema.Targets.Copy()
		mergedSchema.DocsLink = depSchema.DocsLink.Copy()
		
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

func countAttributeCandidate(editRng hcl.Range) lang.Candidate {
	return lang.Candidate{
		Label:       "count",
		Detail:      "optional, number",
		Description: lang.PlainText("The distinct index number (starting with 0) corresponding to the instance"),
		Kind:        lang.AttributeCandidateKind,
		TextEdit: lang.TextEdit{
			NewText: "count",
			Snippet: "count = ${1:1}",
			Range:   editRng,
		},
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

func countAttributeHoverData(editRng hcl.Range) *lang.HoverData {
	return &lang.HoverData{
		Content: lang.MarkupContent{
			Kind:  lang.MarkdownKind,
			Value: "**count** _optional, number_\n\nThe distinct index number (starting with 0) corresponding to the instance",
		},
		Range: editRng,
	}
}

func forEachAttributeCandidate(editRng hcl.Range) lang.Candidate {
	return lang.Candidate{
		Label:  "for_each",
		Detail: "optional, set or map of any type",
		Description: lang.MarkupContent{
			Kind: lang.MarkdownKind,
			Value: "A meta-argument that accepts a map or a set of strings, and creates an instance for each item in that map or set." +
				" Each instance has a distinct infrastructure object associated with it, and each is separately created, updated, or" +
				" destroyed when the configuration is applied.\n\n" +
				"**Note**: A given resource or module block cannot use both count and for_each.",
		},
		Kind: lang.AttributeCandidateKind,
		TextEdit: lang.TextEdit{
			NewText: "for_each",
			Snippet: "for_each {\n ${1}\n}",
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
			Detail: "string",
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
