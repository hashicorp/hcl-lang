package decoder

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (d *Decoder) HoverAtPos(filename string, pos hcl.Pos) (*lang.HoverData, error) {
	f, err := d.fileByName(filename)
	if err != nil {
		return nil, err
	}

	rootBody, err := d.bodyForFileAndPos(filename, f, pos)
	if err != nil {
		return nil, err
	}

	d.rootSchemaMu.RLock()
	defer d.rootSchemaMu.RUnlock()

	if d.rootSchema == nil {
		return nil, &NoSchemaError{}
	}

	data, err := d.hoverAtPos(rootBody, d.rootSchema, pos)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *Decoder) hoverAtPos(body *hclsyntax.Body, bodySchema *schema.BodySchema, pos hcl.Pos) (*lang.HoverData, error) {
	if bodySchema == nil {
		return nil, nil
	}

	filename := body.Range().Filename

	for name, attr := range body.Attributes {
		if attr.Range().ContainsPos(pos) {
			aSchema, ok := bodySchema.Attributes[attr.Name]
			if !ok {
				if bodySchema.AnyAttribute == nil {
					return nil, &PositionalError{
						Filename: filename,
						Pos:      pos,
						Msg:      fmt.Sprintf("unknown attribute %q", attr.Name),
					}
				}
				aSchema = bodySchema.AnyAttribute
			}

			if attr.NameRange.ContainsPos(pos) {
				return &lang.HoverData{
					Content: hoverContentForAttribute(name, aSchema),
					Range:   attr.Range(),
				}, nil
			}

			if attr.Expr.Range().ContainsPos(pos) {
				exprCons := ExprConstraints(aSchema.Expr)
				data, err := hoverDataForExpr(attr.Expr, exprCons, pos)
				if err != nil {
					return nil, &PositionalError{
						Filename: filename,
						Pos:      pos,
						Msg:      err.Error(),
					}
				}
				return data, nil
			}
		}
	}

	for _, block := range body.Blocks {
		if block.Range().ContainsPos(pos) {
			bSchema, ok := bodySchema.Blocks[block.Type]
			if !ok {
				return nil, &PositionalError{
					Filename: filename,
					Pos:      pos,
					Msg:      fmt.Sprintf("unknown block type %q", block.Type),
				}
			}

			if block.TypeRange.ContainsPos(pos) {
				return &lang.HoverData{
					Content: hoverContentForBlock(block.Type, bSchema),
					Range:   block.TypeRange,
				}, nil
			}

			for i, labelRange := range block.LabelRanges {
				if labelRange.ContainsPos(pos) {
					if i+1 > len(bSchema.Labels) {
						return nil, &PositionalError{
							Filename: filename,
							Pos:      pos,
							Msg:      fmt.Sprintf("unexpected label (%d) %q", i, block.Labels[i]),
						}
					}

					return &lang.HoverData{
						Content: d.hoverContentForLabel(i, block, bSchema),
						Range:   labelRange,
					}, nil
				}
			}

			if isPosOutsideBody(block, pos) {
				return nil, &PositionalError{
					Filename: filename,
					Pos:      pos,
					Msg:      fmt.Sprintf("position outside of %q body", block.Type),
				}
			}

			if block.Body != nil && block.Body.Range().ContainsPos(pos) {
				mergedSchema, err := mergeBlockBodySchemas(block, bSchema)
				if err != nil {
					return nil, err
				}

				return d.hoverAtPos(block.Body, mergedSchema, pos)
			}
		}
	}

	// Position outside of any attribute or block
	return nil, &PositionalError{
		Filename: filename,
		Pos:      pos,
		Msg:      "position outside of any attribute name, value or block",
	}
}

func (d *Decoder) hoverContentForLabel(i int, block *hclsyntax.Block, bSchema *schema.BlockSchema) lang.MarkupContent {
	value := block.Labels[i]
	labelSchema := bSchema.Labels[i]

	if labelSchema.IsDepKey {
		dk := dependencyKeysFromBlock(block, bSchema)
		bs, ok := bSchema.DependentBodySchema(dk)
		if ok {
			content := fmt.Sprintf("`%s`", value)
			if bs.Detail != "" {
				content += " " + bs.Detail
			} else if labelSchema.Name != "" {
				content += " " + labelSchema.Name
			}
			if bs.Description.Value != "" {
				content += "\n\n" + bs.Description.Value
			} else if labelSchema.Description.Value != "" {
				content += "\n\n" + labelSchema.Description.Value
			}

			if bs.DocsLink != nil {
				link := bs.DocsLink
				u, err := d.docsURL(link.URL, "documentHover")
				if err == nil {
					content += fmt.Sprintf("\n\n[`%s` on %s](%s)",
						value, u.Hostname(), u.String())
				}
			}

			return lang.Markdown(content)
		}
	}

	content := fmt.Sprintf("%q", value)
	if labelSchema.Name != "" {
		content += fmt.Sprintf(" (%s)", labelSchema.Name)
	}
	content = strings.TrimSpace(content)
	if labelSchema.Description.Value != "" {
		content += "\n\n" + labelSchema.Description.Value
	}

	return lang.Markdown(content)
}

func hoverContentForAttribute(name string, schema *schema.AttributeSchema) lang.MarkupContent {
	value := fmt.Sprintf("**%s** _%s_", name, detailForAttribute(schema))
	if schema.Description.Value != "" {
		value += fmt.Sprintf("\n\n%s", schema.Description.Value)
	}
	return lang.MarkupContent{
		Kind:  lang.MarkdownKind,
		Value: value,
	}
}

func hoverContentForBlock(bType string, schema *schema.BlockSchema) lang.MarkupContent {
	value := fmt.Sprintf("**%s** _%s_", bType, detailForBlock(schema))
	if schema.Description.Value != "" {
		value += fmt.Sprintf("\n\n%s", schema.Description.Value)
	}
	return lang.MarkupContent{
		Kind:  lang.MarkdownKind,
		Value: value,
	}
}

func hoverDataForExpr(expr hcl.Expression, constraints ExprConstraints, pos hcl.Pos) (*lang.HoverData, error) {
	switch e := expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		kw, ok := constraints.KeywordExpr()
		if ok && len(e.Traversal) == 1 {
			return &lang.HoverData{
				Content: lang.Markdown(fmt.Sprintf("`%s` _%s_", kw.Keyword, kw.FriendlyName())),
				Range:   expr.Range(),
			}, nil
		}
	case *hclsyntax.TemplateExpr:
		if e.IsStringLiteral() {
			data, err := hoverDataForExpr(e.Parts[0], constraints, pos)
			if err != nil {
				return nil, err
			}
			// Account for the enclosing quotes
			return &lang.HoverData{
				Content: data.Content,
				Range:   expr.Range(),
			}, nil
		}
		if v, ok := stringValFromTemplateExpr(e); ok {
			if constraints.HasLiteralTypeOf(cty.String) {
				content, err := hoverContentForValue(v, 0)
				if err != nil {
					return nil, err
				}
				return &lang.HoverData{
					Content: lang.Markdown(content),
					Range:   expr.Range(),
				}, nil
			}
			lv, ok := constraints.LiteralValueOf(v)
			if ok {
				content, err := hoverContentForValue(lv.Val, 0)
				if err != nil {
					return nil, err
				}
				return &lang.HoverData{
					Content: lang.Markdown(content),
					Range:   expr.Range(),
				}, nil
			}
		}
	case *hclsyntax.TemplateWrapExpr:
		data, err := hoverDataForExpr(e.Wrapped, constraints, pos)
		if err != nil {
			return nil, err
		}
		// Account for the enclosing quotes
		return &lang.HoverData{
			Content: data.Content,
			Range:   expr.Range(),
		}, nil
	case *hclsyntax.TupleConsExpr:
		tupleCons, ok := constraints.TupleConsExpr()
		if ok {
			content := fmt.Sprintf("_%s_", tupleCons.FriendlyName())
			if tupleCons.Description.Value != "" {
				content += "\n\n" + tupleCons.Description.Value
			}
			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   expr.Range(),
			}, nil
		}

		lt, ok := constraints.LiteralTypeOfTupleExpr()
		if ok {
			content, err := hoverContentForType(lt.Type)
			if err != nil {
				return nil, err
			}
			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   expr.Range(),
			}, nil
		}
		litVal, ok := constraints.LiteralValueOfTupleExpr(e)
		if ok {
			content, err := hoverContentForValue(litVal.Val, 0)
			if err != nil {
				return nil, err
			}
			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   expr.Range(),
			}, nil
		}
	case *hclsyntax.ObjectConsExpr:
		objExpr, ok := constraints.ObjectExpr()
		if ok {
			return hoverDataForObjectExpr(e, objExpr, pos)
		}
		mapExpr, ok := constraints.MapExpr()
		if ok {
			content := fmt.Sprintf("_%s_", mapExpr.FriendlyName())
			if mapExpr.Description.Value != "" {
				content += "\n\n" + mapExpr.Description.Value
			}
			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   expr.Range(),
			}, nil
		}
		lt, ok := constraints.LiteralTypeOfObjectConsExpr()
		if ok {
			content, err := hoverContentForType(lt.Type)
			if err != nil {
				return nil, err
			}
			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   expr.Range(),
			}, nil
		}
		litVal, ok := constraints.LiteralValueOfObjectConsExpr(e)
		if ok {
			content, err := hoverContentForValue(litVal.Val, 0)
			if err != nil {
				return nil, err
			}
			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   expr.Range(),
			}, nil
		}
	case *hclsyntax.LiteralValueExpr:
		if constraints.HasLiteralTypeOf(e.Val.Type()) {
			content, err := hoverContentForValue(e.Val, 0)
			if err != nil {
				return nil, err
			}
			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   expr.Range(),
			}, nil
		}
		lv, ok := constraints.LiteralValueOf(e.Val)
		if ok {
			content, err := hoverContentForValue(lv.Val, 0)
			if err != nil {
				return nil, err
			}
			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   expr.Range(),
			}, nil
		}
		return nil, &ConstraintMismatch{e}
	}

	return nil, fmt.Errorf("unsupported expression (%T)", expr)
}

func hoverDataForObjectExpr(objExpr *hclsyntax.ObjectConsExpr, oe schema.ObjectExpr, pos hcl.Pos) (*lang.HoverData, error) {
	for _, item := range objExpr.Items {
		key, _ := item.KeyExpr.Value(nil)
		if !key.IsWhollyKnown() || key.Type() != cty.String {
			continue
		}
		attr, ok := oe.Attributes[key.AsString()]
		if !ok {
			// unknown attribute
			continue
		}

		if item.ValueExpr.Range().ContainsPos(pos) {
			return hoverDataForExpr(item.ValueExpr, ExprConstraints(attr.Expr), pos)
		}

		itemRng := hcl.RangeBetween(item.KeyExpr.Range(), item.ValueExpr.Range())
		if itemRng.ContainsPos(pos) {
			content := fmt.Sprintf(`**%s** _%s_`, key.AsString(), attr.FriendlyName())
			if attr.Description.Value != "" {
				content += fmt.Sprintf("\n\n%s", attr.Description.Value)
			}

			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   itemRng,
			}, nil
		}
	}

	if len(oe.Attributes) == 0 {
		content := fmt.Sprintf("_%s_", oe.FriendlyName())
		if oe.Description.Value != "" {
			content += "\n\n" + oe.Description.Value
		}
		return &lang.HoverData{
			Content: lang.Markdown(content),
			Range:   objExpr.Range(),
		}, nil
	}

	attrNames := sortedObjectExprAttrNames(oe.Attributes)
	content := "```\n{\n"
	for _, name := range attrNames {
		content += fmt.Sprintf("  %s = %s\n", name, oe.Attributes[name].FriendlyName())
	}
	content += fmt.Sprintf("}\n```\n_%s_", oe.FriendlyName())
	if oe.Description.Value != "" {
		content += "\n\n" + oe.Description.Value
	}
	return &lang.HoverData{
		Content: lang.Markdown(content),
		Range:   objExpr.Range(),
	}, nil
}

func sortedObjectExprAttrNames(attributes schema.ObjectExprAttributes) []string {
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

func isMultilineStringLiteral(tplExpr *hclsyntax.TemplateExpr) bool {
	if len(tplExpr.Parts) < 1 {
		return false
	}
	for _, part := range tplExpr.Parts {
		if _, ok := part.(*hclsyntax.LiteralValueExpr); !ok {
			return false
		}
	}
	return true
}

func stringValFromTemplateExpr(tplExpr *hclsyntax.TemplateExpr) (cty.Value, bool) {
	value := ""
	for _, part := range tplExpr.Parts {
		if lv, ok := part.(*hclsyntax.LiteralValueExpr); ok {
			v, _ := lv.Value(nil)
			if !v.IsWhollyKnown() || v.Type() != cty.String {
				return cty.NilVal, false
			}
			value += v.AsString()
		} else {
			return cty.NilVal, false
		}
	}
	return cty.StringVal(value), true
}

func hoverContentForValue(val cty.Value, nestingLvl int) (string, error) {
	if !val.IsWhollyKnown() {
		if nestingLvl > 0 {
			return "", nil
		}
		return fmt.Sprintf("_%s_", val.Type().FriendlyName()), nil
	}

	attrType := val.Type()
	if attrType.IsPrimitiveType() {
		var value string
		switch attrType {
		case cty.Bool:
			value = fmt.Sprintf("%t", val.True())
		case cty.String:
			if strings.ContainsAny(val.AsString(), "\n") && nestingLvl == 0 {
				// avoid double newline
				strValue := strings.TrimSuffix(val.AsString(), "\n")
				return fmt.Sprintf("```\n%s\n```\n_string_",
					strValue), nil
			}
			value = fmt.Sprintf("%q", val.AsString())
		case cty.Number:
			value = formatNumberVal(val)
		}

		if nestingLvl > 0 {
			return value, nil
		}
		return fmt.Sprintf("`%s` _%s_",
			value, attrType.FriendlyName()), nil
	}

	if attrType.IsObjectType() {
		attrNames := sortedObjectAttrNames(attrType)
		if len(attrNames) == 0 {
			return attrType.FriendlyName(), nil
		}
		value := ""
		if nestingLvl == 0 {
			value += "```\n"
		}
		value += "{\n"
		for _, name := range attrNames {
			whitespace := strings.Repeat("  ", nestingLvl+1)
			val, err := hoverContentForValue(val.GetAttr(name), nestingLvl+1)
			if err == nil {
				value += fmt.Sprintf("%s%s = %s\n",
					whitespace, name, val)
			}
		}
		value += fmt.Sprintf("%s}", strings.Repeat("  ", nestingLvl))
		if nestingLvl == 0 {
			value += "\n```\n_object_"
		}

		return value, nil
	}

	if attrType.IsMapType() {
		elems := val.AsValueMap()
		if len(elems) == 0 {
			return attrType.FriendlyName(), nil
		}
		value := ""
		if nestingLvl == 0 {
			value += "```\n"
		}
		value += "{\n"
		mapKeys := sortedKeysOfValueMap(elems)
		for _, key := range mapKeys {
			val := elems[key]
			elHover, err := hoverContentForValue(val, nestingLvl+1)
			if err == nil {
				whitespace := strings.Repeat("  ", nestingLvl+1)
				value += fmt.Sprintf("%s%q = %s\n",
					whitespace, key, elHover)
			}
		}
		value += fmt.Sprintf("%s}", strings.Repeat("  ", nestingLvl))
		if nestingLvl == 0 {
			value += fmt.Sprintf("\n```\n_%s_", attrType.FriendlyName())
		}

		return value, nil
	}

	if attrType.IsListType() || attrType.IsSetType() || attrType.IsTupleType() {
		elems := val.AsValueSlice()
		if len(elems) == 0 {
			return fmt.Sprintf(`_%s_`, attrType.FriendlyName()), nil
		}
		value := ""
		if nestingLvl == 0 {
			value += "```\n"
		}

		value += "[\n"
		for _, elem := range elems {
			whitespace := strings.Repeat("  ", nestingLvl+1)
			elHover, err := hoverContentForValue(elem, nestingLvl+1)
			if err == nil {
				value += fmt.Sprintf("%s%s,\n", whitespace, elHover)
			}
		}
		value += fmt.Sprintf("%s]", strings.Repeat("  ", nestingLvl))
		if nestingLvl == 0 {
			value += fmt.Sprintf("\n```\n_%s_", attrType.FriendlyName())
		}

		return value, nil
	}

	return "", fmt.Errorf("unsupported type: %q", attrType.FriendlyName())
}

func hoverContentForType(attrType cty.Type) (string, error) {
	if attrType.IsPrimitiveType() {
		return fmt.Sprintf(`_%s_`, attrType.FriendlyName()), nil
	}

	if attrType.IsObjectType() {
		attrNames := sortedObjectAttrNames(attrType)
		if len(attrNames) == 0 {
			return attrType.FriendlyName(), nil
		}
		value := "```\n{\n"
		for _, name := range attrNames {
			valType := attrType.AttributeType(name)
			value += fmt.Sprintf("  %s = %s\n", name,
				valType.FriendlyName())
		}
		value += "}\n```\n_object_"

		return value, nil
	}

	if attrType.IsMapType() || attrType.IsListType() || attrType.IsSetType() || attrType.IsTupleType() {
		value := fmt.Sprintf(`_%s_`, attrType.FriendlyName())
		return value, nil
	}

	return "", fmt.Errorf("unsupported type: %q", attrType.FriendlyName())
}
