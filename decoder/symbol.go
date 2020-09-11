package decoder

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type symbolImplSigil struct{}

// Symbol represents any attribute, or block (and its nested blocks or attributes)
type Symbol interface {
	Name() string
	NestedSymbols() []Symbol
	Range() hcl.Range
	Kind() lang.SymbolKind

	isSymbolImpl() symbolImplSigil
}

// BlockSymbol is Symbol implementation representing a block
type BlockSymbol struct {
	Type   string
	Labels []string

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

func (*BlockSymbol) Kind() lang.SymbolKind {
	return lang.BlockSymbolKind
}

func (bs *BlockSymbol) NestedSymbols() []Symbol {
	return bs.nestedSymbols
}

func (bs *BlockSymbol) Range() hcl.Range {
	return bs.rng
}

// BlockSymbol is Symbol implementation representing an attribute
type AttributeSymbol struct {
	AttrName string
	Type     cty.Type

	rng hcl.Range
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

func (*AttributeSymbol) Kind() lang.SymbolKind {
	return lang.AttributeSymbolKind
}

func (as *AttributeSymbol) NestedSymbols() []Symbol {
	return []Symbol{}
}

func (as *AttributeSymbol) Range() hcl.Range {
	return as.rng
}
