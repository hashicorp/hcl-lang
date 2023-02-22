package schema

import (
	"github.com/zclconf/go-cty/cty"
)

// LiteralType represents literal type constraint
// e.g. any literal string ("foo"), number (42), etc.
//
// Non-literal expressions (even if these evaluate to the given
// type) are excluded.
//
// Complex types are supported, but dedicated List,
// Set, Map and other types are preferred, as these can
// convey more details, such as description, unlike
// e.g. LiteralType{Type: cty.List(...)}.
type LiteralType struct {
	Type cty.Type
}

func (LiteralType) isConstraintImpl() constraintSigil {
	return constraintSigil{}
}

func (lt LiteralType) FriendlyName() string {
	return lt.Type.FriendlyNameForConstraint()
}

func (lt LiteralType) Copy() Constraint {
	return LiteralType{
		Type: lt.Type,
	}
}

func (lt LiteralType) EmptyCompletionData(nextPlaceholder int, nestingLevel int) CompletionData {
	// TODO
	return CompletionData{}
}

func (lt LiteralType) EmptyHoverData(nestingLevel int) *HoverData {
	// TODO
	return nil
}

func (lt LiteralType) ConstraintType() (cty.Type, bool) {
	return lt.Type, true
}
