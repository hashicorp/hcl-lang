// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"
	"errors"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
)

// Reference represents a reference (equivalent of hcl.Traversal),
// i.e. the dot-separated address such as var.foobar
// of a given scope (type-less) or type (type-aware).
type Reference struct {
	// OfScopeId defines scope of a type-less reference
	OfScopeId lang.ScopeId

	// OfType defines the type of a type-aware reference
	OfType cty.Type

	// Name overrides friendly name of the constraint
	Name string

	// Address (if not nil) makes the reference
	// itself addressable and provides scope
	// for the decoded reference.
	//
	// Only one of Address or OfScopeId/OfType can be declared
	Address *ReferenceAddrSchema
}

type ReferenceAddrSchema struct {
	ScopeId lang.ScopeId
}

func (ras *ReferenceAddrSchema) Copy() *ReferenceAddrSchema {
	if ras == nil {
		return nil
	}
	return &ReferenceAddrSchema{
		ScopeId: ras.ScopeId,
	}
}

func (Reference) isConstraintImpl() constraintSigil {
	return constraintSigil{}
}

func (ref Reference) FriendlyName() string {
	if ref.Name != "" {
		return ref.Name
	}
	if ref.OfType != cty.NilType {
		return ref.OfType.FriendlyNameForConstraint()
	}

	return "reference"
}

func (ref Reference) Copy() Constraint {
	return Reference{
		OfScopeId: ref.OfScopeId,
		OfType:    ref.OfType,
		Name:      ref.Name,
		Address:   ref.Address.Copy(),
	}
}

func (ref Reference) EmptyCompletionData(ctx context.Context, nextPlaceholder int, nestingLevel int) CompletionData {
	return CompletionData{
		NewText:        "",
		Snippet:        "",
		TriggerSuggest: true,
	}
}

func (ref Reference) Validate() error {
	if ref.Address != nil && (ref.OfType != cty.NilType || ref.OfScopeId != "") {
		return errors.New("cannot have both Address and OfType/OfScopeId set")
	}
	if ref.Address != nil && ref.Address.ScopeId == "" {
		return errors.New("Address requires non-empty ScopeId")
	}
	if ref.OfType == cty.NilType && ref.OfScopeId == "" && ref.Address == nil {
		return errors.New("one of OfType, OfScopeId and Address is required")
	}
	return nil
}
