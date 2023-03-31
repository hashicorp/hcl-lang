// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"sort"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
)

type Targetable struct {
	Address      lang.Address
	ScopeId      lang.ScopeId
	AsType       cty.Type
	IsSensitive  bool
	FriendlyName string
	Description  lang.MarkupContent

	NestedTargetables Targetables
}

type Targetables []*Targetable

func (ts Targetables) Len() int {
	return len(ts)
}

func (ts Targetables) Less(i, j int) bool {
	return ts[i].Address.String() < ts[j].Address.String()
}

func (ts Targetables) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}

func (tb *Targetable) Copy() *Targetable {
	newTb := &Targetable{
		Address:      tb.Address,
		ScopeId:      tb.ScopeId,
		AsType:       tb.AsType,
		IsSensitive:  tb.IsSensitive,
		FriendlyName: tb.FriendlyName,
		Description:  tb.Description,
	}

	if tb.NestedTargetables != nil {
		for i, ntb := range tb.NestedTargetables {
			newTb.NestedTargetables[i] = ntb.NestedTargetables[i].Copy()
		}
	}

	return newTb
}

func NestedTargetablesForValue(address lang.Address, scopeId lang.ScopeId, val cty.Value) Targetables {
	if val.IsNull() {
		return nil
	}
	typ := val.Type()

	if typ.IsPrimitiveType() || typ == cty.DynamicPseudoType {
		return nil
	}

	if typ.IsSetType() {
		// set elements are not addressable
		return nil
	}

	nestedTargetables := make(Targetables, 0)

	if typ.IsObjectType() {
		for key := range typ.AttributeTypes() {
			elAddr := address.Copy()
			elAddr = append(elAddr, lang.AttrStep{Name: key})

			nestedTargetables = append(nestedTargetables,
				targetableForValue(elAddr, scopeId, val.GetAttr(key)))
		}
	}

	if typ.IsMapType() {
		for key, val := range val.AsValueMap() {
			elAddr := address.Copy()
			elAddr = append(elAddr, lang.IndexStep{Key: cty.StringVal(key)})

			nestedTargetables = append(nestedTargetables,
				targetableForValue(elAddr, scopeId, val))
		}
	}

	if typ.IsListType() || typ.IsTupleType() {
		for i, val := range val.AsValueSlice() {
			elAddr := address.Copy()
			elAddr = append(elAddr, lang.IndexStep{Key: cty.NumberIntVal(int64(i))})

			nestedTargetables = append(nestedTargetables,
				targetableForValue(elAddr, scopeId, val))
		}
	}

	sort.Sort(nestedTargetables)

	return nestedTargetables
}

func targetableForValue(addr lang.Address, scopeId lang.ScopeId, val cty.Value) *Targetable {
	typ := cty.DynamicPseudoType
	if !val.IsNull() {
		typ = val.Type()
	}

	return &Targetable{
		Address:           addr,
		ScopeId:           scopeId,
		AsType:            typ,
		NestedTargetables: NestedTargetablesForValue(addr, scopeId, val),
	}
}
