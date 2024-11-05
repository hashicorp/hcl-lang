// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"sort"

	"github.com/hashicorp/hcl-lang/decoder/internal/ast"
	"github.com/hashicorp/hcl-lang/decoder/internal/schemahelper"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

func (d *Decoder) ReferenceOriginsTargetingPos(path lang.Path, file string, pos hcl.Pos) ReferenceOrigins {
	origins := make(ReferenceOrigins, 0)

	ctx := context.Background()

	localCtx, err := d.pathReader.PathContext(path)
	if err != nil {
		return origins
	}

	targets, ok := localCtx.ReferenceTargets.InnermostAtPos(file, pos)
	if !ok {
		return ReferenceOrigins{}
	}

	for _, target := range targets {
		paths := d.pathReader.Paths(ctx)
		for _, p := range paths {
			pathCtx, err := d.pathReader.PathContext(p)
			if err != nil {
				continue
			}

			rawOrigins := pathCtx.ReferenceOrigins.Match(p, target, path)
			for _, origin := range rawOrigins {
				origins = append(origins, ReferenceOrigin{
					Path:  p,
					Range: origin.OriginRange(),
				})
			}
		}
	}

	sort.SliceStable(origins, func(i, j int) bool {
		if origins[i].Path.Path != origins[j].Path.Path {
			return origins[i].Path.Path < origins[j].Path.Path
		}
		if origins[i].Range.Filename != origins[j].Range.Filename {
			return origins[i].Range.Filename < origins[j].Range.Filename
		}
		return origins[i].Range.Start.Byte < origins[j].Range.Start.Byte
	})

	return origins
}

func (d *PathDecoder) CollectReferenceOrigins() (reference.Origins, error) {
	refOrigins := make(reference.Origins, 0)
	impliedOrigins := make([]schema.ImpliedOrigin, 0)

	if d.pathCtx.Schema == nil {
		// unable to collect reference origins without schema
		return refOrigins, &NoSchemaError{}
	}

	files := d.filenames()
	for _, filename := range files {
		f, err := d.fileByName(filename)
		if err != nil {
			// skip unparseable file
			continue
		}

		os, ios := d.referenceOriginsInBody(f.Body, d.pathCtx.Schema)
		refOrigins = append(refOrigins, os...)
		impliedOrigins = append(impliedOrigins, ios...)
	}

	for _, impliedOrigin := range impliedOrigins {
		for _, origin := range refOrigins {
			localOrigin, ok := origin.(reference.LocalOrigin)

			if ok && localOrigin.Addr.Equals(impliedOrigin.OriginAddress) {
				refOrigins = append(refOrigins, reference.PathOrigin{
					Range:      origin.OriginRange(),
					TargetAddr: impliedOrigin.TargetAddress,
					TargetPath: impliedOrigin.Path,
					Constraints: reference.OriginConstraints{
						{
							OfScopeId: impliedOrigin.Constraints.ScopeId,
							OfType:    impliedOrigin.Constraints.Type,
						},
					},
				})
			}
		}
	}

	sort.SliceStable(refOrigins, func(i, j int) bool {
		if refOrigins[i].OriginRange().Filename != refOrigins[j].OriginRange().Filename {
			return refOrigins[i].OriginRange().Filename < refOrigins[j].OriginRange().Filename
		}
		return refOrigins[i].OriginRange().Start.Byte < refOrigins[j].OriginRange().Start.Byte
	})

	return refOrigins, nil
}

func (d *PathDecoder) referenceOriginsInBody(body hcl.Body, bodySchema *schema.BodySchema) (reference.Origins, []schema.ImpliedOrigin) {
	origins := make(reference.Origins, 0)
	impliedOrigins := make([]schema.ImpliedOrigin, 0)

	if bodySchema == nil {
		return origins, impliedOrigins
	}

	ctx := context.Background()

	impliedOrigins = append(impliedOrigins, bodySchema.ImpliedOrigins...)
	content := ast.DecodeBody(body, bodySchema)

	for _, attr := range content.Attributes {
		var aSchema *schema.AttributeSchema
		if bodySchema.Extensions != nil && bodySchema.Extensions.Count && attr.Name == "count" {
			aSchema = schemahelper.CountAttributeSchema()
		} else if bodySchema.Extensions != nil && bodySchema.Extensions.ForEach && attr.Name == "for_each" {
			aSchema = schemahelper.ForEachAttributeSchema()
		} else {
			var ok bool
			aSchema, ok = bodySchema.Attributes[attr.Name]
			if !ok {
				if bodySchema.AnyAttribute == nil {
					// skip unknown attribute
					continue
				}
				aSchema = bodySchema.AnyAttribute
			}
		}

		if aSchema.OriginForTarget != nil {
			targetAddr, ok := resolveAttributeAddress(attr, aSchema.OriginForTarget.Address)
			if ok {
				origins = append(origins, reference.PathOrigin{
					Range:      attr.NameRange,
					TargetAddr: targetAddr,
					TargetPath: aSchema.OriginForTarget.Path,
					Constraints: reference.OriginConstraints{
						{
							OfScopeId: aSchema.OriginForTarget.Constraints.ScopeId,
							OfType:    aSchema.OriginForTarget.Constraints.Type,
						},
					},
				})
			}
		}

		if aSchema.IsDepKey && bodySchema.Targets != nil {
			origins = append(origins, reference.DirectOrigin{
				Range:       attr.Expr.Range(),
				TargetPath:  bodySchema.Targets.Path,
				TargetRange: bodySchema.Targets.Range,
			})
		}

		if bodySchema.Extensions != nil && bodySchema.Extensions.SelfRefs {
			ctx = schema.WithActiveSelfRefs(ctx)
		}
		expr := d.newExpression(attr.Expr, aSchema.Constraint)
		if eType, ok := expr.(ReferenceOriginsExpression); ok {
			origins = append(origins, eType.ReferenceOrigins(ctx)...)
		}
	}

	for _, block := range content.Blocks {
		if block.Body != nil {
			bSchema, ok := bodySchema.Blocks[block.Type]
			if !ok {
				// skip unknown blocks
				continue
			}
			mergedSchema, _ := schemahelper.MergeBlockBodySchemas(block.Block, bSchema)

			os, ios := d.referenceOriginsInBody(block.Body, mergedSchema)
			origins = append(origins, os...)
			impliedOrigins = append(impliedOrigins, ios...)
		}
	}

	return origins, impliedOrigins
}
