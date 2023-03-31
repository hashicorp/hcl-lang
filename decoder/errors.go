// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

type NoSchemaError struct{}

func (*NoSchemaError) Error() string {
	return fmt.Sprintf("no schema available")
}

type ConstraintMismatch struct {
	Expr hcl.Expression
}

func (cm *ConstraintMismatch) Error() string {
	if cm.Expr != nil {
		return fmt.Sprintf("%T does not match any constraint", cm.Expr)
	}
	return fmt.Sprintf("expression does not match any constraint")
}

type FileNotFoundError struct {
	Filename string
}

func (e *FileNotFoundError) Error() string {
	return fmt.Sprintf("%s: file not found", e.Filename)
}

type UnknownFileFormatError struct {
	Filename string
}

func (e *UnknownFileFormatError) Error() string {
	return fmt.Sprintf("%s: unknown file format", e.Filename)
}

type PosOutOfRangeError struct {
	Filename string
	Pos      hcl.Pos
	Range    hcl.Range
}

func (e *PosOutOfRangeError) Error() string {
	return fmt.Sprintf("%s: position %s is out of range %s", e.Filename, stringPos(e.Pos), e.Range)
}

type PositionalError struct {
	Filename string
	Pos      hcl.Pos
	Msg      string
}

func (e *PositionalError) Error() string {
	return fmt.Sprintf("%s (%s): %s", e.Filename, stringPos(e.Pos), e.Msg)
}
