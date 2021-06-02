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
		TriggerSuggest: triggerSuggestForExprConstraints(attr.Expr),
	}
}

func detailForAttribute(attr *schema.AttributeSchema) string {
	details := []string{}

	if attr.IsRequired {
		details = append(details, "required")
	} else if attr.IsOptional {
		details = append(details, "optional")
	}

	if attr.IsSensitive {
		details = append(details, "sensitive")
	}

	friendlyName := attr.Expr.FriendlyName()
	if friendlyName != "" {
		details = append(details, friendlyName)
	}

	return strings.Join(details[:], ", ")
}

func snippetForAttribute(name string, attr *schema.AttributeSchema) string {
	return fmt.Sprintf("%s = %s", name, snippetForExprContraints(1, attr.Expr))
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
