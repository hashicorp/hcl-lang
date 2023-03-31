// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"fmt"
)

// BlockType tells a decoder how to interpret instance(s) of a block
//
// Types reflect hcldec.Block*Spec types and terraform-json's SchemaNestingMode
type BlockType uint

const (
	BlockTypeNil BlockType = iota
	BlockTypeList
	BlockTypeMap
	BlockTypeObject
	BlockTypeSet
)

func (t BlockType) String() string {
	switch t {
	case BlockTypeList:
		return "list"
	case BlockTypeMap:
		return "map"
	case BlockTypeObject:
		return "object"
	case BlockTypeSet:
		return "set"
	}
	return ""
}

func (t BlockType) GoString() string {
	switch t {
	case BlockTypeList:
		return "BlockTypeList"
	case BlockTypeMap:
		return "BlockTypeMap"
	case BlockTypeObject:
		return "BlockTypeObject"
	case BlockTypeSet:
		return "BlockTypeSet"
	}
	return fmt.Sprintf("BlockType(%d)", t)
}
