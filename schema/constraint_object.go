package schema

import (
	"context"
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
		Attributes:  o.Attributes.Copy(),
		Name:        o.Name,
		Description: o.Description,
	}
}

type prefillRequiredFieldsCtxKey struct{}

func WithPrefillRequiredFields(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, prefillRequiredFieldsCtxKey{}, enabled)
}

func prefillRequiredFields(ctx context.Context) bool {
	enabled, ok := ctx.Value(prefillRequiredFieldsCtxKey{}).(bool)
	if !ok {
		return false
	}
	return enabled
}

func (o Object) EmptyCompletionData(ctx context.Context, placeholder int, nestingLevel int) CompletionData {
	if len(o.Attributes) == 0 {
		return CompletionData{
			NewText:         "{}",
			Snippet:         fmt.Sprintf("{ ${%d} }", placeholder),
			NextPlaceholder: placeholder + 1,
		}
	}

	newText := "{\n"
	snippet := "{\n"

	nesting := strings.Repeat("  ", nestingLevel+1)
	lastPlaceholder := placeholder

	attrNames := sortedObjectExprAttrNames(o.Attributes)

	for _, name := range attrNames {
		attr := o.Attributes[name]
		cData := attr.Constraint.EmptyCompletionData(ctx, lastPlaceholder, nestingLevel+1)
		if cData.NewText == "" || cData.Snippet == "" {
			return CompletionData{
				NewText:         "{}",
				Snippet:         fmt.Sprintf("{ ${%d} }", placeholder),
				TriggerSuggest:  cData.TriggerSuggest,
				NextPlaceholder: placeholder + 1,
			}
		}

		newText += fmt.Sprintf("%s%s = %s\n", nesting, name, cData.NewText)
		snippet += fmt.Sprintf("%s%s = %s\n", nesting, name, cData.Snippet)
		lastPlaceholder = cData.NextPlaceholder
	}

	if nestingLevel > 0 {
		newText += fmt.Sprintf("%s}", strings.Repeat("  ", nestingLevel))
		snippet += fmt.Sprintf("%s}", strings.Repeat("  ", nestingLevel))
	} else {
		newText += "}"
		snippet += "}"
	}

	return CompletionData{
		NewText:         newText,
		Snippet:         snippet,
		NextPlaceholder: lastPlaceholder,
	}
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

func (oa ObjectAttributes) Copy() ObjectAttributes {
	m := make(ObjectAttributes, 0)
	for name, aSchema := range oa {
		m[name] = aSchema.Copy()
	}
	return m
}
