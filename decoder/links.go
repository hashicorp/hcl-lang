package decoder

import (
	"net/url"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// LinksInFile returns links relevant to parts of config in the given file
//
// A link (URI) typically points to the documentation.
func (d *Decoder) LinksInFile(filename string) ([]lang.Link, error) {
	f, err := d.fileByName(filename)
	if err != nil {
		return nil, err
	}

	body, err := d.bodyForFileAndPos(filename, f, hcl.InitialPos)
	if err != nil {
		return nil, err
	}

	d.rootSchemaMu.RLock()
	defer d.rootSchemaMu.RUnlock()

	if d.rootSchema == nil {
		return []lang.Link{}, &NoSchemaError{}
	}

	return d.linksInBody(body, d.rootSchema)
}

func (d *Decoder) linksInBody(body *hclsyntax.Body, bodySchema *schema.BodySchema) ([]lang.Link, error) {
	links := make([]lang.Link, 0)

	for _, block := range body.Blocks {
		blockSchema, ok := bodySchema.Blocks[block.Type]
		if !ok {
			// Ignore unknown block
			continue
		}

		// Currently only block bodies have links associated
		if block.Body != nil {
			depSchema, dk, ok := NewBlockSchema(blockSchema).DependentBodySchema(block)
			if ok && depSchema.DocsLink != nil {
				for _, labelDep := range dk.Labels {
					link := depSchema.DocsLink
					u, err := d.docsURL(link.URL, "documentLink")
					if err == nil {
						links = append(links, lang.Link{
							URI:     u.String(),
							Tooltip: link.Tooltip,
							Range:   block.LabelRanges[labelDep.Index],
						})
					}
				}
			}
		}

	}

	return links, nil
}

func (d *Decoder) docsURL(uri, utmContent string) (*url.URL, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	if d.utmSource != "" {
		q.Set("utm_source", d.utmSource)
	}
	if d.utmMedium != "" {
		q.Set("utm_medium", d.utmMedium)
	}
	if d.useUtmContent {
		q.Set("utm_content", utmContent)
	}
	u.RawQuery = q.Encode()

	return u, nil
}
