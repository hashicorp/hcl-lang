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
	} else if attr.IsOptional {
		detail = "Optional"
	}

	friendlyName := attr.Expr.FriendlyName()
	if friendlyName != "" {
		if detail != "" {
			detail = strings.Join([]string{detail, friendlyName}, ", ")
		} else {
			detail = friendlyName
		}
	}

	return detail
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
