package decoder

import (
	"fmt"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// CandidatesAtPos returns completion candidates for a given position in a file
//
// Schema is required in order to return any candidates and method will return
// error if there isn't one.
func (d *Decoder) CandidatesAtPos(filename string, pos hcl.Pos) (lang.Candidates, error) {
	f, err := d.fileByName(filename)
	if err != nil {
		return lang.ZeroCandidates(), err
	}

	rootBody, err := d.bodyForFileAndPos(filename, f, pos)
	if err != nil {
		return lang.ZeroCandidates(), err
	}

	d.rootSchemaMu.RLock()
	defer d.rootSchemaMu.RUnlock()

	if d.rootSchema == nil {
		return lang.ZeroCandidates(), &NoSchemaError{}
	}

	return d.candidatesAtPos(rootBody, d.rootSchema, pos)
}

func (d *Decoder) candidatesAtPos(body *hclsyntax.Body, bodySchema *schema.BodySchema, pos hcl.Pos) (lang.Candidates, error) {
	if bodySchema == nil {
		return lang.ZeroCandidates(), nil
	}

	filename := body.Range().Filename

	for _, attr := range body.Attributes {
		if attr.Expr.Range().ContainsPos(pos) || attr.EqualsRange.End.Byte == pos.Byte {
			if aSchema, ok := bodySchema.Attributes[attr.Name]; ok {
				return d.attrValueCandidatesAtPos(attr, aSchema, pos)
			}
			if bodySchema.AnyAttribute != nil {
				return d.attrValueCandidatesAtPos(attr, bodySchema.AnyAttribute, pos)
			}

			return lang.ZeroCandidates(), nil
		}
		if attr.NameRange.ContainsPos(pos) {
			prefixRng := attr.NameRange
			prefixRng.End = pos
			return d.bodySchemaCandidates(body, bodySchema, prefixRng, attr.Range()), nil
		}
		if attr.EqualsRange.ContainsPos(pos) {
			return lang.ZeroCandidates(), nil
		}
	}

	rng := hcl.Range{
		Filename: filename,
		Start:    pos,
		End:      pos,
	}

	for _, block := range body.Blocks {
		if block.Range().ContainsPos(pos) {
			bSchema, ok := bodySchema.Blocks[block.Type]
			if !ok {
				return lang.ZeroCandidates(), &PositionalError{
					Filename: filename,
					Pos:      pos,
					Msg:      fmt.Sprintf("unknown block type %q", block.Type),
				}
			}

			if block.TypeRange.ContainsPos(pos) {
				prefixRng := block.TypeRange
				prefixRng.End = pos
				return d.bodySchemaCandidates(body, bodySchema, prefixRng, block.Range()), nil
			}

			for i, labelRange := range block.LabelRanges {
				if labelRange.ContainsPos(pos) {
					if i+1 > len(bSchema.Labels) {
						return lang.ZeroCandidates(), &PositionalError{
							Filename: filename,
							Pos:      pos,
							Msg:      fmt.Sprintf("unexpected label (%d) %q", i, block.Labels[i]),
						}
					}

					prefixRng := rng
					tokenRange, err := d.labelTokenRangeAtPos(labelRange.Filename, pos)
					if err == nil {
						rng, prefixRng = tokenRange, tokenRange
					}
					prefixRng.End = pos

					return d.labelCandidatesFromDependentSchema(i, bSchema.DependentBody, prefixRng, rng)
				}
			}

			if isPosOutsideBody(block, pos) {
				return lang.ZeroCandidates(), &PositionalError{
					Filename: filename,
					Pos:      pos,
					Msg:      fmt.Sprintf("position outside of %q body", block.Type),
				}
			}

			if block.Body != nil && block.Body.Range().ContainsPos(pos) {
				mergedSchema, err := mergeBlockBodySchemas(block, bSchema)
				if err != nil {
					return lang.ZeroCandidates(), err
				}

				return d.candidatesAtPos(block.Body, mergedSchema, pos)
			}
		}
	}

	tokenRng, err := d.nameTokenRangeAtPos(body.Range().Filename, pos)
	if err == nil {
		rng = tokenRng
	}

	return d.bodySchemaCandidates(body, bodySchema, rng, rng), nil
}

func (d *Decoder) nameTokenRangeAtPos(filename string, pos hcl.Pos) (hcl.Range, error) {
	rng := hcl.Range{
		Filename: filename,
		Start:    pos,
		End:      pos,
	}

	f, err := d.fileByName(filename)
	if err != nil {
		return rng, err
	}
	tokens, diags := hclsyntax.LexConfig(f.Bytes, filename, hcl.InitialPos)
	if diags.HasErrors() {
		return rng, diags
	}

	return nameTokenRangeAtPos(tokens, pos)
}

func nameTokenRangeAtPos(tokens hclsyntax.Tokens, pos hcl.Pos) (hcl.Range, error) {
	for i, t := range tokens {
		if t.Range.ContainsPos(pos) {
			if t.Type == hclsyntax.TokenIdent {
				return t.Range, nil
			}
			if t.Type == hclsyntax.TokenNewline && i > 0 {
				// end of line
				previousToken := tokens[i-1]
				if previousToken.Type == hclsyntax.TokenIdent {
					return previousToken.Range, nil
				}
			}
			return hcl.Range{}, fmt.Errorf("token is %s, not Ident", t.Type.String())
		}

		// EOF token has zero length
		// so we just compare start/end position
		if t.Type == hclsyntax.TokenEOF && t.Range.Start == pos && t.Range.End == pos && i > 0 {
			previousToken := tokens[i-1]
			if previousToken.Type == hclsyntax.TokenIdent {
				return previousToken.Range, nil
			}
		}
	}
	return hcl.Range{}, fmt.Errorf("no token found at %s", stringPos(pos))
}

func (d *Decoder) labelTokenRangeAtPos(filename string, pos hcl.Pos) (hcl.Range, error) {
	rng := hcl.Range{
		Filename: filename,
		Start:    pos,
		End:      pos,
	}

	f, err := d.fileByName(filename)
	if err != nil {
		return rng, err
	}
	tokens, diags := hclsyntax.LexConfig(f.Bytes, filename, hcl.InitialPos)
	if diags.HasErrors() {
		return rng, diags
	}

	prefixRng, err := labelTokenRangeAtPos(tokens, pos)
	if err != nil {
		return rng, err
	}

	return prefixRng, nil
}

func labelTokenRangeAtPos(tokens hclsyntax.Tokens, pos hcl.Pos) (hcl.Range, error) {
	for i, t := range tokens {
		if t.Range.ContainsPos(pos) {
			if t.Type == hclsyntax.TokenQuotedLit || t.Type == hclsyntax.TokenIdent {
				return t.Range, nil
			}
			if t.Type == hclsyntax.TokenCQuote && i > 0 {
				// end of label
				if tokens[i-1].Type == hclsyntax.TokenQuotedLit {
					return tokens[i-1].Range, nil
				}
			}
		}
	}
	return hcl.Range{}, fmt.Errorf("no valid token found at %s", stringPos(pos))
}

func isPosOutsideBody(block *hclsyntax.Block, pos hcl.Pos) bool {
	if block.OpenBraceRange.ContainsPos(pos) {
		return true
	}
	if block.CloseBraceRange.ContainsPos(pos) {
		return true
	}

	if block.DefRange().ContainsPos(pos) {
		return true
	}

	return false
}
