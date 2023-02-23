package schema

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
)

// Object represents an object, equivalent of hclsyntax.ObjectConsExpr
// interpreted as object, i.e. with items of known keys
// and different value types.
type Object struct {
	// Attributes defines names and constraints of attributes within the object
	Attributes ObjectAttributes

	// Name overrides friendly name of the constraint
	Name string

	// Description defines description of the whole object (affects hover)
	Description lang.MarkupContent
}

type ObjectAttributes map[string]*AttributeSchema

func (Object) isConstraintImpl() constraintSigil {
	return constraintSigil{}
}

func (o Object) FriendlyName() string {
	if o.Name == "" {
		return "object"
	}
	return o.Name
}

func (o Object) Copy() Constraint {
	return Object{
		Attributes:  o.Attributes.Copy().(ObjectAttributes),
		Name:        o.Name,
		Description: o.Description,
	}
}

func (o Object) EmptyCompletionData(placeholder int, nestingLevel int) CompletionData {
	return CompletionData{}
}

func (o Object) EmptyHoverData(nestingLevel int) *HoverData {
	// TODO
	return nil
}

func (o Object) ConstraintType() (cty.Type, bool) {
	objAttributes := make(map[string]cty.Type)

	for name, attr := range o.Attributes {
		cons, ok := attr.Constraint.(TypeAwareConstraint)
		if !ok {
			return cty.NilType, false
		}
		attrType, ok := cons.ConstraintType()
		if !ok {
			return cty.NilType, false
		}

		objAttributes[name] = attrType
	}

	return cty.Object(objAttributes), true
}

func (ObjectAttributes) isConstraintImpl() constraintSigil {
	return constraintSigil{}
}

func (oa ObjectAttributes) FriendlyName() string {
	return "attributes"
}

func (oa ObjectAttributes) Copy() Constraint {
	m := make(ObjectAttributes, 0)
	for name, aSchema := range oa {
		m[name] = aSchema.Copy()
	}
	return m
}

func (oa ObjectAttributes) EmptyCompletionData(nextPlaceholder int, nestingLevel int) CompletionData {
	// TODO
	return CompletionData{}
}

func (oa ObjectAttributes) EmptyHoverData(nestingLevel int) *HoverData {
	// TODO
	return nil
}
