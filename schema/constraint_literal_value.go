// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty/cty"
)

// LiteralValue represents a literal value, as defined by Value
// with additional metadata.
type LiteralValue struct {
	Value cty.Value

	// IsDeprecated defines whether the value is deprecated
	IsDeprecated bool

	// Description defines description of the value
	Description lang.MarkupContent
}

func (LiteralValue) isConstraintImpl() constraintSigil {
	return constraintSigil{}
}

func (lv LiteralValue) FriendlyName() string {
	return lv.Value.Type().FriendlyNameForConstraint()
}

func (lv LiteralValue) Copy() Constraint {
	return LiteralValue{
		Value:        lv.Value,
		IsDeprecated: lv.IsDeprecated,
		Description:  lv.Description,
	}
}

func (lv LiteralValue) EmptyCompletionData(ctx context.Context, nextPlaceholder int, nestingLevel int) CompletionData {
	if lv.Value.Type().IsPrimitiveType() {
		var value string
		switch lv.Value.Type() {
		case cty.Bool:
			value = fmt.Sprintf("%t", lv.Value.True())
		case cty.String:
			if strings.ContainsAny(lv.Value.AsString(), "\n") && nestingLevel == 0 {
				// avoid double newline
				strValue := strings.TrimSuffix(lv.Value.AsString(), "\n")
				value = fmt.Sprintf("<<<STRING\n%s\nSTRING", strValue)
				if nestingLevel == 0 {
					value += "\n"
				}
			} else {
				value = fmt.Sprintf("%q", lv.Value.AsString())
			}
		case cty.Number:
			value = formatNumberVal(lv.Value)
		}

		return CompletionData{
			NewText:         value,
			Snippet:         value,
			NextPlaceholder: nextPlaceholder,
		}
	}
	if lv.Value.Type().IsListType() {
		vals := lv.Value.AsValueSlice()
		elemNewText := make([]string, len(vals))
		elemSnippets := make([]string, len(vals))
		lastPlaceholder := nextPlaceholder

		for i, val := range vals {
			c := LiteralValue{
				Value: val,
			}
			cData := c.EmptyCompletionData(ctx, lastPlaceholder, nestingLevel)
			if cData.NewText == "" || cData.Snippet == "" {
				return CompletionData{
					NextPlaceholder: lastPlaceholder,
				}
			}
			elemNewText[i] = cData.NewText
			elemSnippets[i] = cData.Snippet
			lastPlaceholder = cData.NextPlaceholder
		}

		return CompletionData{
			// TODO: consider wrapping this in tolist()
			NewText:         fmt.Sprintf("[%s]", strings.Join(elemNewText, ", ")),
			Snippet:         fmt.Sprintf("[%s]", strings.Join(elemSnippets, ", ")),
			NextPlaceholder: lastPlaceholder,
		}
	}
	if lv.Value.Type().IsSetType() {
		vals := lv.Value.AsValueSlice()
		elemNewText := make([]string, len(vals))
		elemSnippets := make([]string, len(vals))
		lastPlaceholder := nextPlaceholder

		for i, val := range vals {
			c := LiteralValue{
				Value: val,
			}
			cData := c.EmptyCompletionData(ctx, lastPlaceholder, nestingLevel)
			if cData.NewText == "" || cData.Snippet == "" {
				return CompletionData{
					NextPlaceholder: lastPlaceholder,
				}
			}
			elemNewText[i] = cData.NewText
			elemSnippets[i] = cData.Snippet
			lastPlaceholder = cData.NextPlaceholder
		}

		return CompletionData{
			// TODO: consider wrapping this in toset()
			NewText:         fmt.Sprintf("[%s]", strings.Join(elemNewText, ", ")),
			Snippet:         fmt.Sprintf("[%s]", strings.Join(elemSnippets, ", ")),
			NextPlaceholder: lastPlaceholder,
		}
	}
	if lv.Value.Type().IsTupleType() {
		vals := lv.Value.AsValueSlice()
		elemNewText := make([]string, len(vals))
		elemSnippets := make([]string, len(vals))
		lastPlaceholder := nextPlaceholder

		for i, val := range vals {
			c := LiteralValue{
				Value: val,
			}
			cData := c.EmptyCompletionData(ctx, lastPlaceholder, nestingLevel)
			if cData.NewText == "" || cData.Snippet == "" {
				return CompletionData{
					NextPlaceholder: lastPlaceholder,
				}
			}
			elemNewText[i] = cData.NewText
			elemSnippets[i] = cData.Snippet
			lastPlaceholder = cData.NextPlaceholder
		}

		return CompletionData{
			NewText:         fmt.Sprintf("[%s]", strings.Join(elemNewText, ", ")),
			Snippet:         fmt.Sprintf("[%s]", strings.Join(elemSnippets, ", ")),
			NextPlaceholder: lastPlaceholder,
		}
	}
	if lv.Value.Type().IsMapType() {
		valueMap := lv.Value.AsValueMap()

		attrNames := sortedValueMap(valueMap)

		// TODO: consider wrapping in tomap()
		newText, snippet := "{\n", "{\n"
		lastPlaceholder := nextPlaceholder
		for _, name := range attrNames {
			val := valueMap[name]

			cons := LiteralValue{
				Value: val,
			}

			cData := cons.EmptyCompletionData(ctx, lastPlaceholder, nestingLevel+1)
			if cData.NewText == "" || cData.Snippet == "" {
				return CompletionData{
					NextPlaceholder: lastPlaceholder,
				}
			}

			newText += fmt.Sprintf("%s%q = %s\n",
				strings.Repeat("  ", nestingLevel+1),
				name, cData.NewText)
			snippet += fmt.Sprintf("%s%q = %s\n",
				strings.Repeat("  ", nestingLevel+1),
				name, cData.Snippet)
			lastPlaceholder = cData.NextPlaceholder
		}
		newText += fmt.Sprintf("%s}", strings.Repeat("  ", nestingLevel))
		snippet += fmt.Sprintf("%s}", strings.Repeat("  ", nestingLevel))

		return CompletionData{
			NewText:         newText,
			Snippet:         snippet,
			NextPlaceholder: lastPlaceholder,
		}
	}
	if lv.Value.Type().IsObjectType() {
		valueMap := lv.Value.AsValueMap()
		attrs := make(ObjectAttributes, 0)
		for name, attrValue := range valueMap {
			aSchema := &AttributeSchema{
				Constraint: LiteralValue{
					Value: attrValue,
				},
			}
			if lv.Value.Type().AttributeOptional(name) {
				aSchema.IsOptional = true
			} else {
				aSchema.IsRequired = true
			}
			attrs[name] = aSchema
		}
		cons := Object{
			Attributes: attrs,
		}
		return cons.EmptyCompletionData(ctx, nextPlaceholder, nestingLevel)
	}

	return CompletionData{
		NextPlaceholder: nextPlaceholder,
	}
}

func (lv LiteralValue) EmptyHoverData(nestingLevel int) *HoverData {
	if lv.Value.Type().IsPrimitiveType() {
		var value string
		switch lv.Value.Type() {
		case cty.Bool:
			value = fmt.Sprintf("%t", lv.Value.True())
		case cty.String:
			if strings.ContainsAny(lv.Value.AsString(), "\n") && nestingLevel == 0 {
				// avoid double newline
				strValue := strings.TrimSuffix(lv.Value.AsString(), "\n")
				value = fmt.Sprintf("```\n%s\n```\n", strValue)
			} else {
				value = fmt.Sprintf("%q", lv.Value.AsString())
			}
		case cty.Number:
			value = formatNumberVal(lv.Value)
		}

		return &HoverData{
			Content: lang.Markdown(value),
		}
	}
	if lv.Value.Type().IsListType() {
		vals := lv.Value.AsValueSlice()
		elemData := make([]string, len(vals))
		for i, val := range vals {
			c := LiteralValue{
				Value: val,
			}
			hoverData := c.EmptyHoverData(nestingLevel)
			if hoverData == nil {
				return nil
			}
			elemData[i] = hoverData.Content.Value
		}

		return &HoverData{
			Content: lang.Markdown(fmt.Sprintf(`tolist([%s])`, strings.Join(elemData, ", "))),
		}
	}
	if lv.Value.Type().IsSetType() {
		vals := lv.Value.AsValueSlice()
		elemData := make([]string, len(vals))
		for i, val := range vals {
			c := LiteralValue{
				Value: val,
			}
			hoverData := c.EmptyHoverData(nestingLevel)
			if hoverData == nil {
				return nil
			}
			elemData[i] = hoverData.Content.Value
		}

		return &HoverData{
			Content: lang.Markdown(fmt.Sprintf(`toset([%s])`, strings.Join(elemData, ", "))),
		}
	}
	if lv.Value.Type().IsTupleType() {
		vals := lv.Value.AsValueSlice()
		elemData := make([]string, len(vals))
		for i, val := range vals {
			c := LiteralValue{
				Value: val,
			}
			hoverData := c.EmptyHoverData(nestingLevel)
			if hoverData == nil {
				return nil
			}
			elemData[i] = hoverData.Content.Value
		}

		return &HoverData{
			Content: lang.Markdown(fmt.Sprintf(`[%s]`, strings.Join(elemData, ", "))),
		}
	}
	if lv.Value.Type().IsMapType() {
		valueMap := lv.Value.AsValueMap()

		attrNames := sortedValueMap(valueMap)

		data := ""
		if nestingLevel == 0 {
			data += "```\n"
		}

		data += "tomap({\n"
		for _, name := range attrNames {
			val := valueMap[name]

			cons := LiteralValue{
				Value: val,
			}

			hoverData := cons.EmptyHoverData(nestingLevel + 1)
			if hoverData == nil {
				return nil
			}

			data += fmt.Sprintf("%s%q = %s\n",
				strings.Repeat("  ", nestingLevel+1),
				name, hoverData.Content.Value)
		}
		data += fmt.Sprintf("%s})", strings.Repeat("  ", nestingLevel))
		if nestingLevel == 0 {
			data += "\n```\n"
		}

		return &HoverData{
			Content: lang.Markdown(data),
		}
	}
	if lv.Value.Type().IsObjectType() {
		valueMap := lv.Value.AsValueMap()
		attrs := make(ObjectAttributes, 0)
		for name, attrValue := range valueMap {
			aSchema := &AttributeSchema{
				Constraint: LiteralValue{
					Value: attrValue,
				},
			}
			if lv.Value.Type().AttributeOptional(name) {
				aSchema.IsOptional = true
			} else {
				aSchema.IsRequired = true
			}
			attrs[name] = aSchema
		}
		cons := Object{
			Attributes: attrs,
		}
		return cons.EmptyHoverData(nestingLevel)
	}

	return nil
}

func sortedValueMap(valueMap map[string]cty.Value) []string {
	if len(valueMap) == 0 {
		return []string{}
	}

	constraints := valueMap
	names := make([]string, len(constraints))
	i := 0
	for name := range constraints {
		names[i] = name
		i++
	}

	sort.Strings(names)
	return names
}

func formatNumberVal(val cty.Value) string {
	bf := val.AsBigFloat()

	if bf.IsInt() {
		intNum, _ := bf.Int64()
		return fmt.Sprintf("%d", intNum)
	}

	fNum, _ := bf.Float64()
	return strconv.FormatFloat(fNum, 'f', -1, 64)
}

func (lv LiteralValue) ConstraintType() (cty.Type, bool) {
	return lv.Value.Type(), true
}
