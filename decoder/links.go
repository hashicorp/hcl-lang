// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"net/url"

	"github.com/hashicorp/hcl-lang/decoder/internal/schemahelper"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// LinksInFile returns links relevant to parts of config in the given file
//
// A link (URI) typically points to the documentation.
func (d *PathDecoder) LinksInFile(filename string) ([]lang.Link, error) {
	f, err := d.fileByName(filename)
	if err != nil {
		return nil, err
	}

	body, err := d.bodyForFileAndPos(filename, f, hcl.InitialPos)
	if err != nil {
		return nil, err
	}

	if d.pathCtx.Schema == nil {
		return []lang.Link{}, &NoSchemaError{}
	}

	return d.linksInBody(body, d.pathCtx.Schema)
}

func (d *PathDecoder) linksInBody(body *hclsyntax.Body, bodySchema *schema.BodySchema) ([]lang.Link, error) {
	links := make([]lang.Link, 0)

	for _, block := range body.Blocks {
		blockSchema, ok := bodySchema.Blocks[block.Type]
		if !ok {
			// Ignore unknown block
			continue
		}

		// Currently only block bodies have links associated
		if block.Body != nil {
			depSchema, dk, ok := schemahelper.NewBlockSchema(blockSchema).DependentBodySchema(block.AsHCLBlock())
			if ok && depSchema.DocsLink != nil {
				link := depSchema.DocsLink
				u, err := d.docsURL(link.URL, "documentLink")
				if err != nil {
					continue
				}
				for _, labelDep := range dk.Labels {
					links = append(links, lang.Link{
						URI:     u.String(),
						Tooltip: link.Tooltip,
						Range:   block.LabelRanges[labelDep.Index],
					})
				}
				for _, attrDep := range dk.Attributes {
					links = append(links, lang.Link{
						URI:     u.String(),
						Tooltip: link.Tooltip,
						Range:   block.Body.Attributes[attrDep.Name].Expr.Range(),
					})
				}
			}
		}

	}

	return links, nil
}

func (d *PathDecoder) docsURL(uri, utmContent string) (*url.URL, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	if d.decoderCtx.UtmSource != "" {
		q.Set("utm_source", d.decoderCtx.UtmSource)
	}
	if d.decoderCtx.UtmMedium != "" {
		q.Set("utm_medium", d.decoderCtx.UtmMedium)
	}
	if d.decoderCtx.UseUtmContent {
		q.Set("utm_content", utmContent)
	}
	u.RawQuery = q.Encode()

	return u, nil
}
