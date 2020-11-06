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
		return fmt.Sprintf("%s = %s", name, snippetForAttrValue(1, attr.ValueTypes[0]))
	}
	return fmt.Sprintf("%s = %s", name, snippetForAttrValue(1, attr.ValueType))
}

func snippetForAttrValue(placeholder uint, attrType cty.Type) string {
	switch attrType {
	case cty.String:
		return fmt.Sprintf(`"${%d:value}"`, placeholder)
	case cty.Bool:
		return fmt.Sprintf(`${%d:false}`, placeholder)
	case cty.Number:
		return fmt.Sprintf(`${%d:1}`, placeholder)
	case cty.DynamicPseudoType:
		return fmt.Sprintf(`${%d}`, placeholder)
	}

	if attrType.IsMapType() {
		return fmt.Sprintf("{\n"+`  "${1:key}" = %s`+"\n}",
			snippetForAttrValue(placeholder+1, *attrType.MapElementType()))
	}

	if attrType.IsListType() || attrType.IsSetType() {
		elType := attrType.ElementType()
		return fmt.Sprintf("[ %s ]", snippetForAttrValue(placeholder, elType))
	}

	if attrType.IsObjectType() {
		objSnippet := ""
		for _, name := range sortedObjectAttrNames(attrType) {
			valType := attrType.AttributeType(name)

			objSnippet += fmt.Sprintf("  %s = %s\n", name,
				snippetForAttrValue(placeholder, valType))
			placeholder++
		}
		return fmt.Sprintf("{\n%s}", objSnippet)
	}

	if attrType.IsTupleType() {
		elTypes := attrType.TupleElementTypes()
		if len(elTypes) == 1 {
			return fmt.Sprintf("[ %s ]", snippetForAttrValue(placeholder, elTypes[0]))
		}

		tupleSnippet := ""
		for _, elType := range elTypes {
			placeholder++
			tupleSnippet += snippetForAttrValue(placeholder, elType)
		}
		return fmt.Sprintf("[\n%s]", tupleSnippet)
	}

	return ""
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
