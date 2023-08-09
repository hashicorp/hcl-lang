// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
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

func stringPos(pos hcl.Pos) string {
	return fmt.Sprintf("%d,%d", pos.Line, pos.Column)
}
