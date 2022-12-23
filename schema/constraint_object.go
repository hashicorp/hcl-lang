package schema

import "github.com/hashicorp/hcl-lang/lang"

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
