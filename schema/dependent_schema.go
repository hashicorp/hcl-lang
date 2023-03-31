// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// DependencyKeys represent values of labels or attributes
// on which BodySchema depends on.
//
// e.g. resource or data block in Terraform
type DependencyKeys struct {
	Labels     []LabelDependent     `json:"labels,omitempty"`
	Attributes []AttributeDependent `json:"attrs,omitempty"`
}

func (dk DependencyKeys) MarshalJSON() ([]byte, error) {
	type sortedKeys DependencyKeys
	var sk = sortedKeys{}

	if len(dk.Labels) > 0 {
		sk.Labels = dk.Labels
		sort.SliceStable(sk.Labels, func(i, j int) bool {
			return sk.Labels[i].Index < sk.Labels[j].Index
		})
	}

	if len(dk.Attributes) > 0 {
		sk.Attributes = dk.Attributes
		sort.SliceStable(sk.Attributes, func(i, j int) bool {
			return sk.Attributes[i].Name < sk.Attributes[j].Name
		})
	}

	return json.Marshal(sk)
}

// SchemaKey represents marshalled DependencyKeys
// which can be created using NewSchemaKey()
type SchemaKey string

// NewSchemaKey creates a marshalled form of DependencyKeys
// to be used inside a map of BlockSchema's DependentBody
func NewSchemaKey(keys DependencyKeys) SchemaKey {
	b, err := keys.MarshalJSON()
	if err != nil {
		msg := fmt.Sprintf(`{"error": %q}`, err.Error())
		return SchemaKey(msg)
	}
	return SchemaKey(string(b))
}

type depKeySigil struct{}

// DependencyKey represents a key used to find a dependent body schema
type DependencyKey interface {
	isDependencyKeyImpl() depKeySigil
}

// LabelDependent represents a pair of label index and value
// used to find a dependent body schema
type LabelDependent struct {
	Index int    `json:"index"`
	Value string `json:"value"`
}

func (ld LabelDependent) isDependencyKeyImpl() depKeySigil {
	return depKeySigil{}
}

// AttributeDependent represents a pair of attribute name and value
// used to find a dependent body schema
type AttributeDependent struct {
	Name string          `json:"name"`
	Expr ExpressionValue `json:"expr"`
}

// ExpressionValue represents static value or a reference
// used to find a dependent body schema
type ExpressionValue struct {
	Static  cty.Value
	Address lang.Address
}

type exprVal struct {
	Static  interface{} `json:"static,omitempty"`
	Address string      `json:"addr,omitempty"`
}

func (ev ExpressionValue) MarshalJSON() ([]byte, error) {
	var val exprVal

	if ev.Static.Type() != cty.NilType {
		val.Static = ctyjson.SimpleJSONValue{Value: ev.Static}
	}

	v, err := ev.Address.Marshal()
	if err != nil {
		return nil, err
	}
	val.Address = string(v)

	return json.Marshal(val)
}

func (ad AttributeDependent) isDependencyKeyImpl() depKeySigil {
	return depKeySigil{}
}
