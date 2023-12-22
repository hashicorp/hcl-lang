// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

func attributeSchemaToCandidate(ctx context.Context, name string, attr *schema.AttributeSchema, rng hcl.Range) lang.Candidate {
	var snippet string
	var triggerSuggest bool
	cData := attr.Constraint.EmptyCompletionData(ctx, 1, 0)
	snippet = fmt.Sprintf("%s = %s", name, cData.Snippet)
	triggerSuggest = cData.TriggerSuggest

	return lang.Candidate{
		Label:        name,
		Detail:       detailForAttribute(attr),
		Description:  attr.Description,
		IsDeprecated: attr.IsDeprecated,
		Kind:         lang.AttributeCandidateKind,
		TextEdit: lang.TextEdit{
			NewText: name,
			Snippet: snippet,
			Range:   rng,
		},
		TriggerSuggest: triggerSuggest,
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

	friendlyName := attr.Constraint.FriendlyName()

	if friendlyName != "" {
		details = append(details, friendlyName)
	}

	return strings.Join(details[:], ", ")
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
