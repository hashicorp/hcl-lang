package decoder

import (
	"sort"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// ReferenceOriginAtPos returns the ReferenceOrigin
// enclosing the position in a file, if one exists, else nil
func (d *Decoder) ReferenceOriginAtPos(filename string, pos hcl.Pos) (*lang.ReferenceOrigin, error) {
	f, err := d.fileByName(filename)
	if err != nil {
		return nil, err
	}

	rootBody, err := d.bodyForFileAndPos(filename, f, pos)
	if err != nil {
		return nil, err
	}

	return d.referenceOriginAtPos(rootBody, pos)
}

func (d *Decoder) ReferenceOriginsTargeting(refTarget lang.ReferenceTarget) (lang.ReferenceOrigins, error) {
	if d.refOriginReader == nil {
		return nil, nil
	}

	allOrigins := ReferenceOrigins(d.refOriginReader())

	return allOrigins.Targeting(refTarget), nil
}

func (d *Decoder) CollectReferenceOrigins() (lang.ReferenceOrigins, error) {
	refOrigins := make(lang.ReferenceOrigins, 0)
	files := d.Filenames()
	for _, filename := range files {
		f, err := d.fileByName(filename)
		if err != nil {
			// skip unparseable file
			continue
		}

		body, ok := f.Body.(*hclsyntax.Body)
		if !ok {
			// skip JSON or other body format
			continue
		}

		refOrigins = append(refOrigins, d.referenceOriginsInBody(body)...)
	}

	sort.SliceStable(refOrigins, func(i, j int) bool {
		return refOrigins[i].Range.Filename <= refOrigins[i].Range.Filename &&
			refOrigins[i].Range.Start.Byte < refOrigins[j].Range.Start.Byte
	})

	return refOrigins, nil
}

func (d *Decoder) referenceOriginsInBody(body *hclsyntax.Body) lang.ReferenceOrigins {
	origins := make(lang.ReferenceOrigins, 0)
	for _, attr := range body.Attributes {
		for _, traversal := range attr.Expr.Variables() {
			addr, err := lang.TraversalToAddress(traversal)
			if err != nil {
				continue
			}
			origins = append(origins, lang.ReferenceOrigin{
				Addr:  addr,
				Range: traversal.SourceRange(),
			})
		}
	}

	for _, block := range body.Blocks {
		if block.Body != nil {
			origins = append(origins, d.referenceOriginsInBody(block.Body)...)
		}
	}

	return origins
}

func (d *Decoder) referenceOriginAtPos(body *hclsyntax.Body, pos hcl.Pos) (*lang.ReferenceOrigin, error) {
	for _, attr := range body.Attributes {
		if d.isPosInsideAttrExpr(attr, pos) {
			traversal, ok := d.traversalAtPos(attr.Expr, pos)
			if ok {
				addr, err := lang.TraversalToAddress(traversal)
				if err == nil {
					return &lang.ReferenceOrigin{
						Addr:  addr,
						Range: traversal.SourceRange(),
					}, nil
				}
			}

			return nil, nil
		}
	}

	for _, block := range body.Blocks {
		if block.Range().ContainsPos(pos) {
			if block.Body != nil && block.Body.Range().ContainsPos(pos) {
				return d.referenceOriginAtPos(block.Body, pos)
			}
		}
	}

	return nil, nil
}

func (d *Decoder) traversalAtPos(expr hclsyntax.Expression, pos hcl.Pos) (hcl.Traversal, bool) {
	for _, traversal := range expr.Variables() {
		if traversal.SourceRange().ContainsPos(pos) {
			return traversal, true
		}
	}

	return nil, false
}

type ReferenceOrigins lang.ReferenceOrigins

func (ro ReferenceOrigins) Targeting(refTarget lang.ReferenceTarget) lang.ReferenceOrigins {
	origins := make(lang.ReferenceOrigins, 0)

	// The O(n^2) here is not ideal but it should
	// be fine given expected data size

	for _, refOrigin := range ro {
		// TODO: reflect refTarget.Type in comparing
		if Address(refOrigin.Addr).Equals(Address(refTarget.Address())) {
			origins = append(origins, refOrigin)
		}

		for _, iTarget := range refTarget.NestedTargets {
			origins = append(origins, ro.Targeting(iTarget)...)
		}
	}

	return origins
}
