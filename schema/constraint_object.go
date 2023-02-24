package schema

import (
	"fmt"
	"sort"
	"strings"

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
	if len(o.Attributes) == 0 {
		return nil
	}

	attrNames := sortedObjectExprAttrNames(o.Attributes)

	data := ""
	if nestingLevel == 0 {
		data += "```\n"
	}

	data += "{\n"
	for _, name := range attrNames {
		attr := o.Attributes[name]

		cons, ok := attr.Constraint.(ConstraintWithHoverData)
		if !ok {
			return nil
		}

		hoverData := cons.EmptyHoverData(nestingLevel + 1)
		if hoverData == nil {
			return nil
		}

		attrFlags := []string{}
		if attr.IsOptional {
			attrFlags = append(attrFlags, "optional")
		}
		if attr.IsSensitive {
			attrFlags = append(attrFlags, "sensitive")
		}
		attrComment := ""
		if len(attrFlags) > 0 {
			attrComment = fmt.Sprintf(" # %s", strings.Join(attrFlags, ", "))
		}

		data += fmt.Sprintf("%s%s = %s%s\n",
			strings.Repeat("  ", nestingLevel+1),
			name, hoverData.Content.Value, attrComment)
	}
	data += fmt.Sprintf("%s}", strings.Repeat("  ", nestingLevel))
	if nestingLevel == 0 {
		data += "\n```\n"
	}

	return &HoverData{
		Content: lang.Markdown(data),
	}
}

func sortedObjectExprAttrNames(attributes ObjectAttributes) []string {
	if len(attributes) == 0 {
		return []string{}
	}

	constraints := attributes
	names := make([]string, len(constraints))
	i := 0
	for name := range constraints {
		names[i] = name
		i++
	}

	sort.Strings(names)
	return names
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
