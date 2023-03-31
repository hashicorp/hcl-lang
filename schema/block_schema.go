// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl-lang/lang"
)

// AttributeSchema describes schema for a block
// e.g. "resource" or "provider" in Terraform
type BlockSchema struct {
	Labels []*LabelSchema
	Type   BlockType

	// SemanticTokenModifiers represents the semantic token modifiers
	// to report for the block's type and labels
	// (in addition to any modifiers of any parent blocks)
	SemanticTokenModifiers lang.SemanticTokenModifiers

	// Body represents the body within block
	// such as attributes and nested blocks
	Body *BodySchema

	// DependentBody represents any "dynamic parts" of the body
	// depending on SchemaKey (labels or attributes)
	DependentBody map[SchemaKey]*BodySchema

	Description  lang.MarkupContent
	IsDeprecated bool
	MinItems     uint64
	MaxItems     uint64

	Address *BlockAddrSchema
}

type BlockAddrSchema struct {
	// Steps describes address steps used to describe the attribute as whole.
	// The last step would typically be LabelStep{}.
	Steps Address

	// FriendlyName is (optional) human-readable name of the block as whole
	// interpreted as reference target.
	//
	// The name is used in completion item and in hover data.
	FriendlyName string

	// ScopeId defines scope of a reference to allow for more granular
	// filtering in completion and accurate matching, which is especially
	// important for type-less reference targets (i.e. AsReference: true).
	ScopeId lang.ScopeId

	// AsReference defines whether the block itself
	// is addressable as a type-less reference
	AsReference bool

	// BodyAsData defines whether the data in the block body
	// is addressable as cty.Object or cty.List(cty.Object),
	// cty.Set(cty.Object) etc. depending on block type
	BodyAsData bool

	// InferBody defines whether (static) Body's
	// blocks and attributes are also walked
	// and their addresses inferred as data
	InferBody bool

	// AsTypeOf makes the block addressable based on type
	// of an attribute
	AsTypeOf *BlockAsTypeOf

	// DependentBodyAsData defines whether the data in
	// the dependent block body is addressable as cty.Object
	// or cty.List(cty.Object), cty.Set(cty.Object) etc.
	// depending on block type
	DependentBodyAsData bool

	// InferDependentBody defines whether DependentBody's
	// blocks and attributes are also walked
	// and their addresses inferred as data
	InferDependentBody bool

	// DependentBodySelfRef instructs collection of reference
	// targets with an additional self.* LocalAddr and
	// makes those targetable by origins within the block body
	// via reference.Target.TargetableFromRangePtr.
	//
	// The targetting (matching w/ origins) is further limited by
	// BodySchema.Extensions.SelfRef, where only self.* origins
	// within a body w/ SelfRef:true will be collected.
	DependentBodySelfRef bool
}

type BlockAsTypeOf struct {
	// AttributeExpr defines whether the block
	// is addressable as a particular type declared
	// directly as expression of the attribute
	AttributeExpr string

	// AttributeValue defines whether the block
	// is addressable as a type of the attribute value.
	//
	// This will be used as a fallback if AttributeExpr
	// is also defined, or when the attribute defined there
	// is of cty.DynamicPseudoType.
	AttributeValue string
}

func (bas *BlockAddrSchema) Validate() error {
	if err := bas.Steps.BlockValidate(); err != nil {
		return err
	}

	if bas.InferBody && !bas.BodyAsData {
		return errors.New("InferBody requires BodyAsData")
	}

	if bas.InferDependentBody && !bas.DependentBodyAsData {
		return errors.New("InferDependentBody requires DependentBodyAsData")
	}

	if bas.DependentBodySelfRef && !bas.InferDependentBody {
		return errors.New("DependentBodySelfRef requires InferDependentBody")
	}

	return nil
}

func (bas *BlockAddrSchema) Copy() *BlockAddrSchema {
	if bas == nil {
		return nil
	}

	newBas := &BlockAddrSchema{
		FriendlyName:         bas.FriendlyName,
		ScopeId:              bas.ScopeId,
		AsReference:          bas.AsReference,
		AsTypeOf:             bas.AsTypeOf.Copy(),
		BodyAsData:           bas.BodyAsData,
		InferBody:            bas.InferBody,
		DependentBodyAsData:  bas.DependentBodyAsData,
		InferDependentBody:   bas.InferDependentBody,
		DependentBodySelfRef: bas.DependentBodySelfRef,
		Steps:                bas.Steps.Copy(),
	}

	return newBas
}

func (bato *BlockAsTypeOf) Copy() *BlockAsTypeOf {
	if bato == nil {
		return nil
	}

	return &BlockAsTypeOf{
		AttributeExpr:  bato.AttributeExpr,
		AttributeValue: bato.AttributeValue,
	}
}

func (*BlockSchema) isSchemaImpl() schemaImplSigil {
	return schemaImplSigil{}
}

func (bSchema *BlockSchema) Validate() error {
	var errs *multierror.Error

	if bSchema.Address != nil {
		err := bSchema.Address.Validate()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("Address: %w", err))
		}
	}

	if bSchema.Body != nil {
		err := bSchema.Body.Validate()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("Body: %w", err))
		}
	}

	if errs != nil && len(errs.Errors) == 1 {
		return errs.Errors[0]
	}

	return errs.ErrorOrNil()
}

func (bs *BlockSchema) Copy() *BlockSchema {
	if bs == nil {
		return nil
	}

	newBs := &BlockSchema{
		Type:                   bs.Type,
		SemanticTokenModifiers: bs.SemanticTokenModifiers.Copy(),
		IsDeprecated:           bs.IsDeprecated,
		MinItems:               bs.MinItems,
		MaxItems:               bs.MaxItems,
		Description:            bs.Description,
		Body:                   bs.Body.Copy(),
		Address:                bs.Address.Copy(),
	}

	if bs.Labels != nil {
		newBs.Labels = make([]*LabelSchema, len(bs.Labels))
		for i, label := range bs.Labels {
			newBs.Labels[i] = label.Copy()
		}
	}

	if bs.DependentBody != nil {
		newBs.DependentBody = make(map[SchemaKey]*BodySchema, 0)
		for key, depSchema := range bs.DependentBody {
			newBs.DependentBody[key] = depSchema.Copy()
		}
	}

	return newBs
}
