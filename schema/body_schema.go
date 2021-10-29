package schema

import (
	"fmt"
	"sort"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
)

// BodySchema describes schema of a body comprised of blocks or attributes
// (if any), where body can be root or body of any block in the hierarchy.
type BodySchema struct {
	Blocks       map[string]*BlockSchema
	Attributes   map[string]*AttributeSchema
	AnyAttribute *AttributeSchema
	IsDeprecated bool
	Detail       string
	Description  lang.MarkupContent

	// DocsLink represents a link to docs that will be exposed
	// as part of LinksInFile()
	DocsLink *DocsLink

	// HoverURL represents a URL that will be appended to the end
	// of hover data in HoverAtPos(). This can differ from DocsLink,
	// but often will match.
	HoverURL string

	// TargetableAs represents how else the body may be targeted
	// if not by its declarable attributes or blocks.
	TargetableAs Targetables
}

type DocsLink struct {
	URL     string
	Tooltip string
}

func (*BodySchema) isSchemaImpl() schemaImplSigil {
	return schemaImplSigil{}
}

func (bs *BodySchema) ToHCLSchema() *hcl.BodySchema {
	attributes := make([]hcl.AttributeSchema, 0)
	for name, attr := range bs.Attributes {
		attributes = append(attributes, hcl.AttributeSchema{
			Name:     name,
			Required: attr.IsRequired,
		})
	}

	blocks := make([]hcl.BlockHeaderSchema, 0)
	for blockType, block := range bs.Blocks {
		labelNames := make([]string, len(block.Labels))
		for i, label := range block.Labels {
			labelNames[i] = label.Name
		}

		blocks = append(blocks, hcl.BlockHeaderSchema{
			Type:       blockType,
			LabelNames: labelNames,
		})
	}

	return &hcl.BodySchema{
		Attributes: attributes,
		Blocks:     blocks,
	}
}

// NewBodySchema creates a new BodySchema instance
func NewBodySchema() *BodySchema {
	return &BodySchema{
		Blocks:     make(map[string]*BlockSchema, 0),
		Attributes: make(map[string]*AttributeSchema, 0),
	}
}

func (as *BodySchema) AttributeNames() []string {
	keys := make([]string, 0, len(as.Attributes))
	for k := range as.Attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (as *BodySchema) BlockTypes() []string {
	keys := make([]string, 0, len(as.Blocks))
	for k := range as.Blocks {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (bs *BodySchema) Validate() error {
	if len(bs.Attributes) > 0 && bs.AnyAttribute != nil {
		return fmt.Errorf("one of Attributes or AnyAttribute must be set, not both")
	}

	var result *multierror.Error
	for name, attr := range bs.Attributes {
		err := attr.Validate()
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("%s: %w", name, err))
		}
	}

	for bType, block := range bs.Blocks {
		err := block.Validate()
		if err != nil {
			if me, ok := err.(*multierror.Error); ok {
				for _, err := range me.Errors {
					result = multierror.Append(result, fmt.Errorf("%s: %w", bType, err))
				}
			} else {
				result = multierror.Append(result, fmt.Errorf("%s: %w", bType, err))
			}
		}
	}

	return result.ErrorOrNil()
}

func (bs *BodySchema) Copy() *BodySchema {
	if bs == nil {
		return nil
	}

	newBs := &BodySchema{
		IsDeprecated: bs.IsDeprecated,
		Detail:       bs.Detail,
		Description:  bs.Description,
		AnyAttribute: bs.AnyAttribute.Copy(),
		HoverURL:     bs.HoverURL,
		DocsLink:     bs.DocsLink.Copy(),
	}

	if bs.TargetableAs != nil {
		newBs.TargetableAs = make(Targetables, len(bs.TargetableAs))
		for id, target := range bs.TargetableAs {
			newBs.TargetableAs[id] = target.Copy()
		}
	}

	if bs.Attributes != nil {
		newBs.Attributes = make(map[string]*AttributeSchema, len(bs.Attributes))
		for name, attr := range bs.Attributes {
			newBs.Attributes[name] = attr.Copy()
		}
	}

	if bs.Blocks != nil {
		newBs.Blocks = make(map[string]*BlockSchema, len(bs.Blocks))
		for name, block := range bs.Blocks {
			newBs.Blocks[name] = block.Copy()
		}
	}

	return newBs
}

func (dl *DocsLink) Copy() *DocsLink {
	if dl == nil {
		return nil
	}

	return &DocsLink{
		URL:     dl.URL,
		Tooltip: dl.Tooltip,
	}
}
