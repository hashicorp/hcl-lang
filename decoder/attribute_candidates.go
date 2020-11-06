package decoder

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

func attributeSchemaToCandidate(name string, attr *schema.AttributeSchema, rng hcl.Range) lang.Candidate {
	return lang.Candidate{
		Label:        name,
		Detail:       detailForAttribute(attr),
		Description:  attr.Description,
		IsDeprecated: attr.IsDeprecated,
		Kind:         lang.AttributeCandidateKind,
		TextEdit: lang.TextEdit{
			NewText: name,
			Snippet: snippetForAttribute(name, attr),
			Range:   rng,
		},
	}
}

func detailForAttribute(attr *schema.AttributeSchema) string {
	var detail string
	if attr.IsRequired {
		detail = "Required"
	} else {
		detail = "Optional"
	}

	if len(attr.ValueTypes) > 0 {
		detail += fmt.Sprintf(", %s", strings.Join(attr.ValueTypes.FriendlyNames(), " or "))
	} else {
		detail += fmt.Sprintf(", %s", attr.ValueType.FriendlyName())
	}

	return detail
}

func snippetForAttribute(name string, attr *schema.AttributeSchema) string {
	if len(attr.ValueTypes) > 0 {
		return fmt.Sprintf("%s %s", name, snippetForAttrValue(1, true, attr.ValueTypes[0]))
	}
	return fmt.Sprintf("%s %s", name, snippetForAttrValue(1, true, attr.ValueType))
}

func snippetForAttrValue(placeholder uint, equalSign bool, attrType cty.Type) string {
	eq := ""
	if equalSign {
		eq = "= "
	}

	switch attrType {
	case cty.String:
		return fmt.Sprintf(`%s"${%d:value}"`, eq, placeholder)
	case cty.Bool:
		return fmt.Sprintf(`%s${%d:false}`, eq, placeholder)
	case cty.Number:
		return fmt.Sprintf(`%s${%d:1}`, eq, placeholder)
	case cty.DynamicPseudoType:
		return fmt.Sprintf(`%s${%d}`, eq, placeholder)
	}

	if attrType.IsMapType() {
		return fmt.Sprintf("%s{\n"+`  "${1:key}" %s`+"\n}", eq,
			snippetForAttrValue(placeholder+1, true, *attrType.MapElementType()))
	}

	if attrType.IsListType() || attrType.IsSetType() {
		elType := attrType.ElementType()
		if elType.IsPrimitiveType() || elType == cty.DynamicPseudoType {
			return fmt.Sprintf("%s[ %s ]", eq, snippetForAttrValue(placeholder, false, elType))
		}

		return snippetForAttrValue(placeholder, true, elType)
	}

	if attrType.IsObjectType() {
		objSnippet := ""
		for _, name := range sortedObjectAttrNames(attrType) {
			valType := attrType.AttributeType(name)

			objSnippet += fmt.Sprintf("  %s %s\n", name,
				snippetForAttrValue(placeholder, true, valType))
			placeholder++
		}
		return fmt.Sprintf("{\n%s}", objSnippet)
	}

	if attrType.IsTupleType() {
		elTypes := attrType.TupleElementTypes()
		if len(elTypes) == 1 {
			return fmt.Sprintf("%s[ %s ]", eq, snippetForAttrValue(placeholder, false, elTypes[0]))
		}

		tupleSnippet := ""
		for _, elType := range elTypes {
			placeholder++
			tupleSnippet += snippetForAttrValue(placeholder, false, elType)
		}
		return fmt.Sprintf("%s[\n%s]", eq, tupleSnippet)
	}

	return eq
}

func sortedObjectAttrNames(obj cty.Type) []string {
	if !obj.IsObjectType() {
		return []string{}
	}

	types := obj.AttributeTypes()
	names := make([]string, len(types))
	i := 0
	for name := range types {
		names[i] = name
		i++
	}

	sort.Strings(names)
	return names
}
