// Copyright IBM Corp. 2020, 2026
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"github.com/hashicorp/hcl-lang/decoder/internal/ast"
	"github.com/hashicorp/hcl-lang/decoder/internal/schemahelper"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

type WriteOnlyAttribute struct {
	Name     string
	Resource string
}

type WriteOnlyAttributes []WriteOnlyAttribute

func (d *PathDecoder) CollectWriteOnlyAttributes() (WriteOnlyAttributes, error) {
	if d.pathCtx.Schema == nil {
		// unable to collect write-only attributes without schema
		return nil, &NoSchemaError{}
	}

	attrs := make(WriteOnlyAttributes, 0)
	files := d.filenames()
	for _, filename := range files {
		f, err := d.fileByName(filename)
		if err != nil {
			// skip unparseable file
			continue
		}
		attrs = append(attrs, d.decodeWriteOnlyAttributesForBody(f.Body, d.pathCtx.Schema)...)
	}

	return attrs, nil
}

func (d *PathDecoder) decodeWriteOnlyAttributesForBody(body hcl.Body, bodySchema *schema.BodySchema) WriteOnlyAttributes {
	woAttrs := make(WriteOnlyAttributes, 0)

	if bodySchema == nil {
		return WriteOnlyAttributes{}
	}

	content := ast.DecodeBody(body, bodySchema)

	for _, block := range content.Blocks {
		if block.Type == "resource" {
			blockSchema, ok := bodySchema.Blocks[block.Type]
			if !ok {
				// unknown block (no schema)
				continue
			}

			mergedSchema, _ := schemahelper.MergeBlockBodySchemas(block.Block, blockSchema)

			blockContent := ast.DecodeBody(block.Body, blockSchema.Body)

			for _, attr := range blockContent.Attributes {
				attrSchema, ok := mergedSchema.Attributes[attr.Name]
				if ok && attrSchema.IsWriteOnly {

					woAttrs = append(woAttrs, WriteOnlyAttribute{
						Name:     attr.Name,
						Resource: block.Labels[0],
					})

				}
			}
		}
	}

	return woAttrs
}
