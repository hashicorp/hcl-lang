// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"sort"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type PathDecoder struct {
	path       lang.Path
	pathCtx    *PathContext
	decoderCtx DecoderContext

	// maxCandidates defines maximum number of completion candidates returned
	maxCandidates uint

	// PrefillRequiredFields enriches label-based completion candidates
	// with required attributes and blocks
	// TODO: Move under DecoderContext
	PrefillRequiredFields bool
}

func (d *Decoder) Path(path lang.Path) (*PathDecoder, error) {
	pathCtx, err := d.pathReader.PathContext(path)

	return &PathDecoder{
		path:          path,
		pathCtx:       pathCtx,
		decoderCtx:    d.ctx,
		maxCandidates: 100,
	}, err
}

func (d *PathDecoder) bytesForFile(file string) ([]byte, error) {
	f, ok := d.pathCtx.Files[file]
	if !ok {
		return nil, &FileNotFoundError{Filename: file}
	}

	return f.Bytes, nil
}

// filenames returns a slice of filenames already loaded via LoadFile
func (d *PathDecoder) filenames() []string {
	var files []string

	for filename := range d.pathCtx.Files {
		files = append(files, filename)
	}

	sort.Strings(files)

	return files
}

func (d *PathDecoder) bytesFromRange(rng hcl.Range) ([]byte, error) {
	b, err := d.bytesForFile(rng.Filename)
	if err != nil {
		return nil, err
	}

	return rng.SliceBytes(b), nil
}

func (d *PathDecoder) fileByName(name string) (*hcl.File, error) {
	f, ok := d.pathCtx.Files[name]
	if !ok {
		return nil, &FileNotFoundError{Filename: name}
	}
	return f, nil
}

func (d *PathDecoder) bodyForFileAndPos(name string, f *hcl.File, pos hcl.Pos) (*hclsyntax.Body, error) {
	body, isHcl := f.Body.(*hclsyntax.Body)
	if !isHcl {
		return nil, &UnknownFileFormatError{Filename: name}
	}

	if !body.Range().ContainsPos(pos) &&
		!posEqual(body.Range().Start, pos) &&
		!posEqual(body.Range().End, pos) {

		return nil, &PosOutOfRangeError{
			Filename: name,
			Pos:      pos,
			Range:    body.Range(),
		}
	}

	return body, nil
}
