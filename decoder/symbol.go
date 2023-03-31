// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
)

type symbolImplSigil struct{}

// Symbol represents any attribute, or block (and its nested blocks or attributes)
type Symbol interface {
	Path() lang.Path
	Name() string
	NestedSymbols() []Symbol
	Range() hcl.Range

	isSymbolImpl() symbolImplSigil
}

// BlockSymbol is Symbol implementation representing a block
type BlockSymbol struct {
	Type   string
	Labels []string

	path          lang.Path
	rng           hcl.Range
	nestedSymbols []Symbol
}

func (*BlockSymbol) isSymbolImpl() symbolImplSigil {
	return symbolImplSigil{}
}

func (bs *BlockSymbol) Equal(other Symbol) bool {
	obs, ok := other.(*BlockSymbol)
	if !ok {
		return false
	}
	if bs == nil || obs == nil {
		return bs == obs
	}

	return reflect.DeepEqual(*bs, *obs)
}

func (bs *BlockSymbol) Name() string {
	name := bs.Type
	for _, label := range bs.Labels {
		name += fmt.Sprintf(" %q", label)
	}
	return name
}

func (bs *BlockSymbol) NestedSymbols() []Symbol {
	return bs.nestedSymbols
}

func (bs *BlockSymbol) Range() hcl.Range {
	return bs.rng
}

func (bs *BlockSymbol) Path() lang.Path {
	return bs.path
}

// AttributeSymbol is Symbol implementation representing an attribute
type AttributeSymbol struct {
	AttrName string
	ExprKind lang.SymbolExprKind

	path          lang.Path
	rng           hcl.Range
	nestedSymbols []Symbol
}

func (*AttributeSymbol) isSymbolImpl() symbolImplSigil {
	return symbolImplSigil{}
}

func (as *AttributeSymbol) Equal(other Symbol) bool {
	oas, ok := other.(*AttributeSymbol)
	if !ok {
		return false
	}
	if as == nil || oas == nil {
		return as == oas
	}

	return reflect.DeepEqual(*as, *oas)
}

func (as *AttributeSymbol) Name() string {
	return as.AttrName
}

func (as *AttributeSymbol) NestedSymbols() []Symbol {
	return as.nestedSymbols
}

func (as *AttributeSymbol) Range() hcl.Range {
	return as.rng
}

func (as *AttributeSymbol) Path() lang.Path {
	return as.path
}

type ExprSymbol struct {
	ExprName string
	ExprKind lang.SymbolExprKind

	path          lang.Path
	rng           hcl.Range
	nestedSymbols []Symbol
}

func (*ExprSymbol) isSymbolImpl() symbolImplSigil {
	return symbolImplSigil{}
}

func (as *ExprSymbol) Equal(other Symbol) bool {
	oas, ok := other.(*ExprSymbol)
	if !ok {
		return false
	}
	if as == nil || oas == nil {
		return as == oas
	}

	return reflect.DeepEqual(*as, *oas)
}

func (as *ExprSymbol) Name() string {
	return as.ExprName
}

func (as *ExprSymbol) NestedSymbols() []Symbol {
	return as.nestedSymbols
}

func (as *ExprSymbol) Range() hcl.Range {
	return as.rng
}

func (as *ExprSymbol) Path() lang.Path {
	return as.path
}
