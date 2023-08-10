// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl-lang/decoder/internal/schemahelper"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (d *PathDecoder) HoverAtPos(ctx context.Context, filename string, pos hcl.Pos) (*lang.HoverData, error) {
	f, err := d.fileByName(filename)
	if err != nil {
		return nil, err
	}

	rootBody, err := d.bodyForFileAndPos(filename, f, pos)
	if err != nil {
		return nil, err
	}

	if d.pathCtx.Schema == nil {
		return nil, &NoSchemaError{}
	}

	data, err := d.hoverAtPos(ctx, rootBody, d.pathCtx.Schema, pos)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *PathDecoder) hoverAtPos(ctx context.Context, body *hclsyntax.Body, bodySchema *schema.BodySchema, pos hcl.Pos) (*lang.HoverData, error) {
	if bodySchema == nil {
		return nil, nil
	}

	filename := body.Range().Filename

	for name, attr := range body.Attributes {
		if attr.Range().ContainsPos(pos) {
			var aSchema *schema.AttributeSchema
			if bodySchema.Extensions != nil && bodySchema.Extensions.SelfRefs {
				ctx = schema.WithActiveSelfRefs(ctx)
			}

			if bodySchema.Extensions != nil && bodySchema.Extensions.Count && name == "count" {
				aSchema = schemahelper.CountAttributeSchema()
			} else if bodySchema.Extensions != nil && bodySchema.Extensions.ForEach && name == "for_each" {
				aSchema = schemahelper.ForEachAttributeSchema()
			} else {
				var ok bool
				aSchema, ok = bodySchema.Attributes[attr.Name]
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
			}

			if attr.NameRange.ContainsPos(pos) {
				return &lang.HoverData{
					Content: hoverContentForAttribute(name, aSchema),
					Range:   attr.Range(),
				}, nil
			}

			if attr.Expr.Range().ContainsPos(pos) {
				if aSchema.Constraint != nil {
					return d.newExpression(attr.Expr, aSchema.Constraint).HoverAtPos(ctx, pos), nil
				}

				exprCons := ExprConstraints(aSchema.Expr)
				data, err := d.hoverDataForExpr(ctx, attr.Expr, exprCons, 0, pos)
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
			blockSchema, ok := bodySchema.Blocks[block.Type]
			if !ok {
				return nil, &PositionalError{
					Filename: filename,
					Pos:      pos,
					Msg:      fmt.Sprintf("unknown block type %q", block.Type),
				}
			}

			if block.TypeRange.ContainsPos(pos) {
				return &lang.HoverData{
					Content: d.hoverContentForBlock(block.Type, blockSchema),
					Range:   block.TypeRange,
				}, nil
			}

			for i, labelRange := range block.LabelRanges {
				if labelRange.ContainsPos(pos) {
					if i+1 > len(blockSchema.Labels) {
						return nil, &PositionalError{
							Filename: filename,
							Pos:      pos,
							Msg:      fmt.Sprintf("unexpected label (%d) %q", i, block.Labels[i]),
						}
					}

					return &lang.HoverData{
						Content: d.hoverContentForLabel(i, block, blockSchema),
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
				mergedSchema, _ := schemahelper.MergeBlockBodySchemas(block.AsHCLBlock(), blockSchema)
				return d.hoverAtPos(ctx, block.Body, mergedSchema, pos)
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

func (d *PathDecoder) hoverContentForLabel(i int, block *hclsyntax.Block, bSchema *schema.BlockSchema) lang.MarkupContent {
	value := block.Labels[i]
	labelSchema := bSchema.Labels[i]

	if labelSchema.IsDepKey {
		bs, _, ok := schemahelper.NewBlockSchema(bSchema).DependentBodySchema(block.AsHCLBlock())
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

			if bs.HoverURL != "" {
				u, err := d.docsURL(bs.HoverURL, "documentHover")
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

func (d *PathDecoder) hoverContentForBlock(bType string, schema *schema.BlockSchema) lang.MarkupContent {
	value := fmt.Sprintf("**%s** _%s_", bType, detailForBlock(schema))
	if schema.Description.Value != "" {
		value += fmt.Sprintf("\n\n%s", schema.Description.Value)
	}

	if schema.Body != nil && schema.Body.HoverURL != "" {
		u, err := d.docsURL(schema.Body.HoverURL, "documentHover")
		if err == nil {
			value += fmt.Sprintf("\n\n[`%s` on %s](%s)",
				bType, u.Hostname(), u.String())
		}
	}

	return lang.MarkupContent{
		Kind:  lang.MarkdownKind,
		Value: value,
	}
}

func (d *PathDecoder) hoverDataForExpr(ctx context.Context, expr hcl.Expression, constraints ExprConstraints, nestingLvl int, pos hcl.Pos) (*lang.HoverData, error) {
	switch e := expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		kw, ok := constraints.KeywordExpr()
		if ok && len(e.Traversal) == 1 {
			if nestingLvl > 0 {
				return &lang.HoverData{
					Content: lang.Markdown(kw.FriendlyName()),
					Range:   expr.Range(),
				}, nil
			}
			return &lang.HoverData{
				Content: lang.Markdown(fmt.Sprintf("`%s` _%s_", kw.Keyword, kw.FriendlyName())),
				Range:   expr.Range(),
			}, nil
		}

		tes, ok := constraints.TraversalExprs()
		if ok {
			content, err := d.legacyHoverContentForTraversalExpr(ctx, e.AsTraversal(), tes, pos)
			if err != nil {
				return nil, err
			}
			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   expr.Range(),
			}, nil
		}

		_, ok = constraints.TypeDeclarationExpr()
		if ok {
			return &lang.HoverData{
				Content: lang.Markdown("Type declaration"),
				Range:   expr.Range(),
			}, nil
		}
	case *hclsyntax.FunctionCallExpr:
		_, ok := constraints.TypeDeclarationExpr()
		if ok {
			return &lang.HoverData{
				Content: lang.Markdown("Type declaration"),
				Range:   expr.Range(),
			}, nil
		}
	case *hclsyntax.TemplateExpr:
		if e.IsStringLiteral() {
			data, err := d.hoverDataForExpr(ctx, e.Parts[0], constraints, nestingLvl, pos)
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
		data, err := d.hoverDataForExpr(ctx, e.Wrapped, constraints, nestingLvl, pos)
		if err != nil {
			return nil, err
		}
		// Account for the enclosing quotes
		return &lang.HoverData{
			Content: data.Content,
			Range:   expr.Range(),
		}, nil
	case *hclsyntax.TupleConsExpr:
		se, ok := constraints.SetExpr()
		if ok {
			for _, elemExpr := range e.Exprs {
				if elemExpr.Range().ContainsPos(pos) {
					return d.hoverDataForExpr(ctx, elemExpr, ExprConstraints(se.Elem), nestingLvl, pos)
				}
			}
			content := fmt.Sprintf("_%s_", se.FriendlyName())
			if se.Description.Value != "" {
				content += "\n\n" + se.Description.Value
			}
			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   expr.Range(),
			}, nil
		}
		le, ok := constraints.ListExpr()
		if ok {
			for _, elemExpr := range e.Exprs {
				if elemExpr.Range().ContainsPos(pos) {
					return d.hoverDataForExpr(ctx, elemExpr, ExprConstraints(le.Elem), nestingLvl, pos)
				}
			}
			content := fmt.Sprintf("_%s_", le.FriendlyName())
			if le.Description.Value != "" {
				content += "\n\n" + le.Description.Value
			}
			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   expr.Range(),
			}, nil
		}
		te, ok := constraints.TupleExpr()
		if ok {
			for i, elemExpr := range e.Exprs {
				if elemExpr.Range().ContainsPos(pos) {
					if i >= len(te.Elems) {
						return nil, &ConstraintMismatch{elemExpr}
					}
					ec := ExprConstraints(te.Elems[i])
					return d.hoverDataForExpr(ctx, elemExpr, ec, nestingLvl, pos)
				}
			}
			content := fmt.Sprintf("_%s_", te.FriendlyName())
			if te.Description.Value != "" {
				content += "\n\n" + te.Description.Value
			}
			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   expr.Range(),
			}, nil
		}
		lt, ok := constraints.LiteralTypeOfTupleExpr()
		if ok {
			content, err := hoverContentForType(lt.Type, nestingLvl)
			if err != nil {
				return nil, err
			}
			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   expr.Range(),
			}, nil
		}
		lv, ok := constraints.LiteralValueOfTupleExpr(e)
		if ok {
			content, err := hoverContentForValue(lv.Val, nestingLvl)
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
			return d.hoverDataForObjectExpr(ctx, e, objExpr, nestingLvl, pos)
		}
		mapExpr, ok := constraints.MapExpr()
		if ok {
			content := mapExpr.FriendlyName()
			if nestingLvl == 0 {
				content = fmt.Sprintf("_%s_", mapExpr.FriendlyName())
				if mapExpr.Description.Value != "" {
					content += "\n\n" + mapExpr.Description.Value
				}
			}
			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   expr.Range(),
			}, nil
		}
		lt, ok := constraints.LiteralTypeOfObjectConsExpr()
		if ok {
			content, err := hoverContentForType(lt.Type, nestingLvl)
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
			content, err := hoverContentForValue(litVal.Val, nestingLvl)
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
			content := ""
			if nestingLvl == 0 {
				valContent, err := hoverContentForValue(e.Val, nestingLvl)
				if err != nil {
					return nil, err
				}
				content = valContent
			} else {
				typeContent, err := hoverContentForType(e.Val.Type(), nestingLvl)
				if err != nil {
					return nil, err
				}
				content = typeContent
			}
			return &lang.HoverData{
				Content: lang.Markdown(content),
				Range:   expr.Range(),
			}, nil
		}
		lv, ok := constraints.LiteralValueOf(e.Val)
		if ok {
			content, err := hoverContentForValue(lv.Val, nestingLvl)
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

func (d *PathDecoder) hoverDataForObjectExpr(ctx context.Context, objExpr *hclsyntax.ObjectConsExpr, oe schema.ObjectExpr, nestingLvl int, pos hcl.Pos) (*lang.HoverData, error) {
	declaredAttributes := make(map[string]hclsyntax.Expression, 0)
	for _, item := range objExpr.Items {
		key, _ := item.KeyExpr.Value(nil)
		if key.IsNull() || !key.IsWhollyKnown() || key.Type() != cty.String {
			// skip items keys that can't be interpolated
			// without further context
			continue
		}
		attr, ok := oe.Attributes[key.AsString()]
		if !ok {
			// unknown attribute
			continue
		}

		if item.ValueExpr.Range().ContainsPos(pos) {
			return d.hoverDataForExpr(ctx, item.ValueExpr, ExprConstraints(attr.Expr), nestingLvl+1, pos)
		}

		itemRng := hcl.RangeBetween(item.KeyExpr.Range(), item.ValueExpr.Range())
		if itemRng.ContainsPos(pos) {
			content := hoverContentForAttribute(key.AsString(), attr)
			return &lang.HoverData{
				Content: content,
				Range:   itemRng,
			}, nil
		}

		declaredAttributes[key.AsString()] = item.ValueExpr
	}

	if len(oe.Attributes) == 0 {
		content := oe.FriendlyName()
		if nestingLvl == 0 {
			content := fmt.Sprintf("_%s_", oe.FriendlyName())
			if oe.Description.Value != "" {
				content += "\n\n" + oe.Description.Value
			}
		}
		return &lang.HoverData{
			Content: lang.Markdown(content),
			Range:   objExpr.Range(),
		}, nil
	}

	attrNames := sortedObjectExprAttrNames(oe.Attributes)
	content := ""
	if nestingLvl == 0 {
		content += "```\n"
	}
	content += "{\n"
	insideNesting := strings.Repeat("  ", nestingLvl+1)
	for _, name := range attrNames {
		ec := oe.Attributes[name].Expr
		attrData := ec.FriendlyName()

		if attrExpr, ok := declaredAttributes[name]; ok {
			data, err := d.hoverDataForExpr(ctx, attrExpr, ExprConstraints(ec), nestingLvl+1, pos)
			if err == nil && data.Content.Value != "" {
				attrData = data.Content.Value
			}
		}

		content += fmt.Sprintf("%s%s = %s\n", insideNesting, name, attrData)
	}
	endBraceNesting := strings.Repeat("  ", nestingLvl)
	content += fmt.Sprintf("%s}", endBraceNesting)
	if nestingLvl == 0 {
		content += fmt.Sprintf("\n```\n_%s_", oe.FriendlyName())
		if oe.Description.Value != "" {
			content += "\n\n" + oe.Description.Value
		}
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

func (d *PathDecoder) legacyHoverContentForTraversalExpr(ctx context.Context, traversal hcl.Traversal, tes []schema.TraversalExpr, pos hcl.Pos) (string, error) {
	origins, ok := d.pathCtx.ReferenceOrigins.AtPos(traversal.SourceRange().Filename, pos)
	if !ok {
		return "", &reference.NoOriginFound{}
	}

	for _, origin := range origins {
		matchableOrigin, ok := origin.(reference.MatchableOrigin)
		if !ok {
			continue
		}
		targets, ok := d.pathCtx.ReferenceTargets.Match(matchableOrigin)
		if !ok {
			// target not found
			continue
		}

		// TODO: Reflect additional found targets here?
		return hoverContentForReferenceTarget(ctx, targets[0], pos)
	}

	return "", &reference.NoTargetFound{}
}

func hoverContentForReferenceTarget(ctx context.Context, ref reference.Target, pos hcl.Pos) (string, error) {
	content := fmt.Sprintf("`%s`", ref.Address(ctx, pos))

	var friendlyName string
	if ref.Type != cty.NilType {
		typeContent, err := hoverContentForType(ref.Type, 0)
		if err == nil {
			friendlyName = "\n" + typeContent
		}
	}
	if friendlyName == "" {
		friendlyName = " " + ref.FriendlyName()
	}
	content += friendlyName

	if ref.Description.Value != "" {
		content += fmt.Sprintf("\n\n%s", ref.Description.Value)
	}

	return content, nil
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

func hoverContentForType(attrType cty.Type, nestingLvl int) (string, error) {
	if attrType.IsPrimitiveType() || attrType == cty.DynamicPseudoType {
		if nestingLvl > 0 {
			return attrType.FriendlyName(), nil
		}
		return fmt.Sprintf(`_%s_`, attrType.FriendlyName()), nil
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
		insideNesting := strings.Repeat("  ", nestingLvl+1)
		for _, name := range attrNames {
			valType := attrType.AttributeType(name)
			valData := valType.FriendlyNameForConstraint()

			data, err := hoverContentForType(valType, nestingLvl+1)
			if err == nil {
				valData = data
			}

			if attrType.AttributeOptional(name) {
				valData = fmt.Sprintf("optional, %s", valData)
			}

			value += fmt.Sprintf("%s%s = %s\n", insideNesting, name, valData)
		}
		endBraceNesting := strings.Repeat("  ", nestingLvl)
		value += fmt.Sprintf("%s}", endBraceNesting)
		if nestingLvl == 0 {
			value += "\n```\n_object_"
		}

		return value, nil
	}

	if attrType.IsMapType() || attrType.IsListType() || attrType.IsSetType() || attrType.IsTupleType() {
		if nestingLvl > 0 {
			return attrType.FriendlyName(), nil
		}
		value := fmt.Sprintf(`_%s_`, attrType.FriendlyName())
		return value, nil
	}

	return "", fmt.Errorf("unsupported type: %q", attrType.FriendlyName())
}
