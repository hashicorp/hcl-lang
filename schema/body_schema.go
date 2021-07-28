package schema

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl-lang/lang"
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
	TargetableAs []*Targetable

	// TODO: Functions
}

type DocsLink struct {
	URL     string
	Tooltip string
}

func (*BodySchema) isSchemaImpl() schemaImplSigil {
	return schemaImplSigil{}
}

// NewBodySchema creates a new BodySchema instance
func NewBodySchema() *BodySchema {
	return &BodySchema{
		Blocks:     make(map[string]*BlockSchema, 0),
		Attributes: make(map[string]*AttributeSchema, 0),
	}
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
		newBs.TargetableAs = make([]*Targetable, len(bs.TargetableAs))
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
