package decoder

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (d *PathDecoder) attrValueCandidatesAtPos(attr *hclsyntax.Attribute, schema *schema.AttributeSchema, outerBodyRng hcl.Range, pos hcl.Pos) (lang.Candidates, error) {
	constraints, editRng := constraintsAtPos(attr.Expr, ExprConstraints(schema.Expr), pos)
	if len(constraints) > 0 {
		prefixRng := editRng
		prefixRng.End = pos
		return d.expressionCandidatesAtPos(constraints, outerBodyRng, prefixRng, editRng)
	}
	return lang.ZeroCandidates(), nil
}

func constraintsAtPos(expr hcl.Expression, constraints ExprConstraints, pos hcl.Pos) (ExprConstraints, hcl.Range) {
	// TODO: Support middle-of-expression completion

	// Ideally the edit range should always match the expression range
	// but expression range in many cases includes the last newline character
	// so we'd have to reinsert that character in the text edit.
	// More importantly LSP does not support multi-line text edits.
	// Overall this is not an issue (yet) as we don't support more complex
	// completion in nested expressions

	switch eType := expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		if !eType.Val.IsWhollyKnown() {
			return constraints, hcl.Range{
				Start:    eType.Range().Start,
				End:      eType.Range().Start,
				Filename: eType.Range().Filename,
			}
		}
	case *hclsyntax.ScopeTraversalExpr:
		matchedConstraints := make(ExprConstraints, 0)
		ke, ok := constraints.KeywordExpr()
		if ok {
			matchedConstraints = append(matchedConstraints, ke)
		}
		tes, ok := constraints.TraversalExprs()
		if ok {
			matchedConstraints = append(matchedConstraints, tes.AsConstraints()...)
		}

		if len(matchedConstraints) > 0 {
			endPos := eType.Range().End
			if pos.Byte-endPos.Byte == 1 {
				endPos = pos
			}
			return matchedConstraints, hcl.Range{
				Filename: eType.Range().Filename,
				Start:    eType.Range().Start,
				End:      endPos,
			}
		}
	case *hclsyntax.TupleConsExpr:
		rng := eType.Range()
		tupleConsBody := hcl.Range{
			Start: hcl.Pos{
				Line:   rng.Start.Line,
				Column: rng.Start.Column + 1,
				Byte:   rng.Start.Byte + 1,
			},
			End: hcl.Pos{
				Line:   rng.End.Line,
				Column: rng.End.Column - 1,
				Byte:   rng.End.Byte - 1,
			},
			Filename: rng.Filename,
		}

		tc, ok := constraints.TupleConsExpr()
		if ok && len(eType.Exprs) == 0 && tupleConsBody.ContainsPos(pos) {
			return ExprConstraints(tc.AnyElem), hcl.Range{
				Start:    pos,
				End:      pos,
				Filename: eType.Range().Filename,
			}
		}

		se, ok := constraints.SetExpr()
		if ok && len(eType.Exprs) == 0 && tupleConsBody.ContainsPos(pos) {
			return ExprConstraints(se.Elem), hcl.Range{
				Start:    pos,
				End:      pos,
				Filename: eType.Range().Filename,
			}
		}
		le, ok := constraints.ListExpr()
		if ok && len(eType.Exprs) == 0 && tupleConsBody.ContainsPos(pos) {
			return ExprConstraints(le.Elem), hcl.Range{
				Start:    pos,
				End:      pos,
				Filename: eType.Range().Filename,
			}
		}
		te, ok := constraints.TupleExpr()
		if ok && len(eType.Exprs) == 0 && tupleConsBody.ContainsPos(pos) {
			return ExprConstraints(te.Elems[0]), hcl.Range{
				Start:    pos,
				End:      pos,
				Filename: eType.Range().Filename,
			}
		}
	case *hclsyntax.ObjectConsExpr:
		oe, ok := constraints.ObjectExpr()
		if ok {
			undeclaredAttributes := oe.Attributes
			for _, item := range eType.Items {
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
				delete(undeclaredAttributes, key.AsString())

				itemRng := hcl.RangeBetween(item.KeyExpr.Range(), item.ValueExpr.Range())
				if item.ValueExpr.Range().ContainsPos(pos) {
					return constraintsAtPos(item.ValueExpr, ExprConstraints(attr.Expr), pos)
				} else if itemRng.ContainsPos(pos) {
					// middle of attribute name or equal sign
					return ExprConstraints{}, expr.Range()
				}
			}

			return ExprConstraints{undeclaredAttributes}, hcl.Range{
				Start:    pos,
				End:      pos,
				Filename: eType.Range().Filename,
			}
		}
	}

	return ExprConstraints{}, expr.Range()
}

func (d *PathDecoder) expressionCandidatesAtPos(constraints ExprConstraints, outerBodyRng, prefixRng, editRng hcl.Range) (lang.Candidates, error) {
	candidates := lang.NewCandidates()

	for _, c := range constraints {
		candidates.List = append(candidates.List, d.constraintToCandidates(c, outerBodyRng, prefixRng, editRng)...)
	}

	candidates.IsComplete = true
	return candidates, nil
}

func (d *PathDecoder) constraintToCandidates(constraint schema.ExprConstraint, outerBodyRng, prefixRng, editRng hcl.Range) []lang.Candidate {
	candidates := make([]lang.Candidate, 0)

	switch c := constraint.(type) {
	case schema.LiteralTypeExpr:
		candidates = append(candidates, typeToCandidates(c.Type, editRng)...)
	case schema.LiteralValue:
		if c, ok := valueToCandidate(c.Val, c.Description, c.IsDeprecated, editRng); ok {
			candidates = append(candidates, c)
		}
	case schema.KeywordExpr:
		candidates = append(candidates, lang.Candidate{
			Label:       c.Keyword,
			Detail:      c.FriendlyName(),
			Description: c.Description,
			Kind:        lang.KeywordCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: c.Keyword,
				Snippet: c.Keyword,
				Range:   editRng,
			},
		})
	case schema.TraversalExpr:
		candidates = append(candidates, d.candidatesForTraversalConstraint(c, outerBodyRng, prefixRng, editRng)...)
	case schema.TupleConsExpr:
		candidates = append(candidates, lang.Candidate{
			Label:       fmt.Sprintf(`[%s]`, labelForConstraints(c.AnyElem)),
			Detail:      c.Name,
			Description: c.Description,
			Kind:        lang.TupleCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: `[ ]`,
				Snippet: `[ ${0} ]`,
				Range:   editRng,
			},
			TriggerSuggest: triggerSuggestForExprConstraints(c.AnyElem),
		})
	case schema.ListExpr:
		candidates = append(candidates, lang.Candidate{
			Label:       fmt.Sprintf(`[%s]`, labelForConstraints(c.Elem)),
			Detail:      c.FriendlyName(),
			Description: c.Description,
			Kind:        lang.ListCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: `[ ]`,
				Snippet: `[ ${0} ]`,
				Range:   editRng,
			},
			TriggerSuggest: triggerSuggestForExprConstraints(c.Elem),
		})
	case schema.SetExpr:
		candidates = append(candidates, lang.Candidate{
			Label:       fmt.Sprintf(`[%s]`, labelForConstraints(c.Elem)),
			Detail:      c.FriendlyName(),
			Description: c.Description,
			Kind:        lang.SetCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: `[ ]`,
				Snippet: `[ ${0} ]`,
				Range:   editRng,
			},
			TriggerSuggest: triggerSuggestForExprConstraints(c.Elem),
		})
	case schema.TupleExpr:
		triggerSuggest := false
		if len(c.Elems) > 0 {
			triggerSuggest = triggerSuggestForExprConstraints(c.Elems[0])
		}
		candidates = append(candidates, lang.Candidate{
			Label:       fmt.Sprintf(`[%s]`, labelForConstraints(c.Elems[0])),
			Detail:      c.FriendlyName(),
			Description: c.Description,
			Kind:        lang.TupleCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: `[ ]`,
				Snippet: `[ ${0} ]`,
				Range:   editRng,
			},
			TriggerSuggest: triggerSuggest,
		})
	case schema.MapExpr:
		candidates = append(candidates, lang.Candidate{
			Label:       fmt.Sprintf(`{ key =%s}`, labelForConstraints(c.Elem)),
			Detail:      c.FriendlyName(),
			Description: c.Description,
			Kind:        lang.MapCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: fmt.Sprintf("{\n  name = %s\n}",
					newTextForConstraints(c.Elem, true)),
				Snippet: fmt.Sprintf("{\n  ${1:name} = %s\n}",
					snippetForConstraints(1, c.Elem, true)),
				Range: editRng,
			},
			TriggerSuggest: triggerSuggestForExprConstraints(c.Elem),
		})
	case schema.ObjectExpr:
		candidates = append(candidates, lang.Candidate{
			Label:       `{ }`,
			Detail:      c.FriendlyName(),
			Description: c.Description,
			Kind:        lang.ObjectCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: "{\n  \n}",
				Snippet: "{\n  ${1}\n}",
				Range:   editRng,
			},
			TriggerSuggest: len(c.Attributes) > 0,
		})
	case schema.ObjectExprAttributes:
		attrNames := sortedObjectExprAttrNames(c)
		for _, name := range attrNames {
			attr := c[name]
			candidates = append(candidates, lang.Candidate{
				Label:        name,
				Detail:       detailForAttribute(attr),
				IsDeprecated: attr.IsDeprecated,
				Description:  attr.Description,
				Kind:         lang.AttributeCandidateKind,
				TextEdit: lang.TextEdit{
					NewText: fmt.Sprintf("%s = %s", name, newTextForConstraints(attr.Expr, true)),
					Snippet: fmt.Sprintf("%s = %s", name, snippetForConstraints(1, attr.Expr, true)),
					Range:   editRng,
				},
			})
		}
	case schema.TypeDeclarationExpr:
		for _, t := range []string{
			"bool",
			"number",
			"string",
			"list()",
			"set()",
			"tuple()",
			"map()",
			"object({})",
		} {
			candidates = append(candidates, lang.Candidate{
				Label:  t,
				Detail: t,
				Kind:   lang.AttributeCandidateKind,
				TextEdit: lang.TextEdit{
					NewText: t,
					Snippet: snippetForTypeDeclaration(t),
					Range:   editRng,
				},
			})
		}
	}

	return candidates
}

func (d *PathDecoder) candidatesForTraversalConstraint(tc schema.TraversalExpr, outerBodyRng, prefixRng, editRng hcl.Range) []lang.Candidate {
	candidates := make([]lang.Candidate, 0)

	if d.pathCtx.ReferenceTargets == nil {
		return candidates
	}

	if tc.Address != nil {
		// no candidates if traversal itself is addressable
		return candidates
	}

	prefix, _ := d.bytesFromRange(prefixRng)

	d.pathCtx.ReferenceTargets.MatchWalk(tc, string(prefix), func(ref reference.Target) error {
		// avoid suggesting references to block's own fields from within (for now)
		if ref.RangePtr != nil &&
			(outerBodyRng.ContainsPos(ref.RangePtr.Start) ||
				posEqual(outerBodyRng.End, ref.RangePtr.End)) {
			return nil
		}

		candidates = append(candidates, lang.Candidate{
			Label:       ref.Addr.String(),
			Detail:      ref.FriendlyName(),
			Description: ref.Description,
			Kind:        lang.TraversalCandidateKind,
			TextEdit: lang.TextEdit{
				NewText: ref.Addr.String(),
				Snippet: ref.Addr.String(),
				Range:   editRng,
			},
		})
		return nil
	})

	return candidates
}

func newTextForConstraints(cons schema.ExprConstraints, isNested bool) string {
	for _, constraint := range cons {
		switch c := constraint.(type) {
		case schema.LiteralTypeExpr:
			return newTextForLiteralType(c.Type)
		case schema.LiteralValue:
			return newTextForLiteralValue(c.Val)
		case schema.KeywordExpr:
			return c.Keyword
		case schema.TupleConsExpr:
			if isNested {
				return "[  ]"
			}
			return fmt.Sprintf("[\n  %s\n]", newTextForConstraints(c.AnyElem, true))
		case schema.ListExpr:
			if isNested {
				return "[  ]"
			}
			return fmt.Sprintf("[\n  %s\n]", newTextForConstraints(c.Elem, true))
		case schema.SetExpr:
			if isNested {
				return "[  ]"
			}
			return fmt.Sprintf("[\n  %s\n]", newTextForConstraints(c.Elem, true))
		case schema.TupleExpr:
			if isNested {
				return "[  ]"
			}
			return fmt.Sprintf("[\n  %s\n]", newTextForConstraints(c.Elems[0], true))
		case schema.MapExpr:
			return fmt.Sprintf("{\n  %s\n}", newTextForConstraints(c.Elem, true))
		case schema.ObjectExpr:
			return "{\n  \n}"
		}
	}
	return ""
}

func snippetForTypeDeclaration(td string) string {
	switch td {
	case "list()":
		return "list(${0})"
	case "set()":
		return "set(${0})"
	case "tuple()":
		return "tuple(${0})"
	case "map()":
		return "map(${0})"
	case "object({})":
		return "object({\n ${1:name} = ${2}\n})"
	default:
		return td
	}
}

func snippetForConstraints(placeholder uint, cons schema.ExprConstraints, isNested bool) string {
	for _, constraint := range cons {
		switch c := constraint.(type) {
		case schema.LiteralTypeExpr:
			return snippetForLiteralType(placeholder, c.Type)
		case schema.LiteralValue:
			return snippetForLiteralValue(placeholder, c.Val)
		case schema.KeywordExpr:
			return fmt.Sprintf("${%d:%s}", placeholder, c.Keyword)
		case schema.TupleConsExpr:
			if isNested {
				return fmt.Sprintf("[ ${%d} ]", placeholder+1)
			}
			return fmt.Sprintf("[\n  %s\n]", snippetForConstraints(placeholder+1, c.AnyElem, true))
		case schema.ListExpr:
			if isNested {
				return fmt.Sprintf("[ ${%d} ]", placeholder+1)
			}
			return fmt.Sprintf("[\n  %s\n]", snippetForConstraints(placeholder+1, c.Elem, true))
		case schema.SetExpr:
			if isNested {
				return fmt.Sprintf("[ ${%d} ]", placeholder+1)
			}
			return fmt.Sprintf("[\n  %s\n]", snippetForConstraints(placeholder+1, c.Elem, true))
		case schema.TupleExpr:
			if isNested {
				return fmt.Sprintf("[ ${%d} ]", placeholder+1)
			}
			return fmt.Sprintf("[\n  %s\n]", snippetForConstraints(placeholder+1, c.Elems[0], true))
		case schema.MapExpr:
			return fmt.Sprintf("{\n  %s\n}", snippetForConstraints(placeholder+1, c.Elem, true))
		case schema.ObjectExpr:
			return fmt.Sprintf("{\n  ${%d}\n}", placeholder+1)
		}
	}
	return ""
}

func labelForConstraints(cons schema.ExprConstraints) string {
	labels := " "
	labelsAdded := 0
	for _, constraint := range cons {
		if len(labels) > 10 {
			labels += "…"
			break
		}
		if labelsAdded > 0 {
			labels += "| "
		}
		switch c := constraint.(type) {
		case schema.LiteralTypeExpr:
			labels += labelForLiteralType(c.Type)
		case schema.LiteralValue:
			continue
		case schema.KeywordExpr:
			labels += c.FriendlyName()
		case schema.TraversalExpr:
			labels += c.FriendlyName()
		case schema.TupleConsExpr:
			labels += fmt.Sprintf("[%s]", labelForConstraints(c.AnyElem))
		case schema.ListExpr:
			labels += fmt.Sprintf("[%s]", labelForConstraints(c.Elem))
		case schema.SetExpr:
			labels += fmt.Sprintf("[%s]", labelForConstraints(c.Elem))
		case schema.TupleExpr:
			labels += fmt.Sprintf("[%s]", labelForConstraints(c.Elems[0]))
		}
		labelsAdded++
	}
	labels += " "

	return labels
}

func typeToCandidates(ofType cty.Type, editRng hcl.Range) []lang.Candidate {
	candidates := make([]lang.Candidate, 0)

	// TODO: Ensure TextEdit is always single-line, otherwise use AdditionalTextEdit
	// See https://github.com/microsoft/language-server-protocol/issues/92

	if ofType == cty.Bool {
		if c, ok := valueToCandidate(cty.True, lang.MarkupContent{}, false, editRng); ok {
			candidates = append(candidates, c)
		}
		if c, ok := valueToCandidate(cty.False, lang.MarkupContent{}, false, editRng); ok {
			candidates = append(candidates, c)
		}
		return candidates
	}

	if ofType.IsPrimitiveType() || ofType == cty.DynamicPseudoType {
		// Nothing to complete for these types
		return candidates
	}

	candidates = append(candidates, lang.Candidate{
		Label:  labelForLiteralType(ofType),
		Detail: ofType.FriendlyNameForConstraint(),
		Kind:   candidateKindForType(ofType),
		TextEdit: lang.TextEdit{
			NewText: newTextForLiteralType(ofType),
			Snippet: snippetForLiteralType(1, ofType),
			Range:   editRng,
		},
	})

	return candidates
}

func valueToCandidate(val cty.Value, desc lang.MarkupContent, isDeprecated bool, editRng hcl.Range) (lang.Candidate, bool) {
	if !val.IsWhollyKnown() {
		// Avoid unknown values
		return lang.Candidate{}, false
	}

	detail := val.Type().FriendlyNameForConstraint()

	// shorten types which may have longer friendly names
	if val.Type().IsObjectType() {
		detail = "object"
	}
	if val.Type().IsMapType() {
		detail = "map"
	}
	if val.Type().IsListType() {
		detail = "list"
	}
	if val.Type().IsSetType() {
		detail = "set"
	}
	if val.Type().IsTupleType() {
		detail = "tuple"
	}

	return lang.Candidate{
		Label:        labelForLiteralValue(val, false),
		Detail:       detail,
		Description:  desc,
		IsDeprecated: isDeprecated,
		Kind:         candidateKindForType(val.Type()),
		TextEdit: lang.TextEdit{
			NewText: newTextForLiteralValue(val),
			Snippet: snippetForLiteralValue(1, val),
			Range:   editRng,
		},
	}, true
}

func candidateKindForType(t cty.Type) lang.CandidateKind {
	if t == cty.Bool {
		return lang.BoolCandidateKind
	}
	if t == cty.String {
		return lang.StringCandidateKind
	}
	if t == cty.Number {
		return lang.NumberCandidateKind
	}
	if t.IsListType() {
		return lang.ListCandidateKind
	}
	if t.IsSetType() {
		return lang.SetCandidateKind
	}
	if t.IsTupleType() {
		return lang.TupleCandidateKind
	}
	if t.IsMapType() {
		return lang.MapCandidateKind
	}
	if t.IsObjectType() {
		return lang.ObjectCandidateKind
	}

	return lang.NilCandidateKind
}

func triggerSuggestForExprConstraints(ec schema.ExprConstraints) bool {
	if len(ec) > 0 {
		expr := ec[0]
		switch et := expr.(type) {
		case schema.LiteralTypeExpr:
			if et.Type == cty.Bool {
				return true
			}
		case schema.LiteralValue:
			if len(ec) > 1 {
				return true
			}
		case schema.KeywordExpr:
			return true
		case schema.TraversalExpr:
			return true
		case schema.TupleConsExpr:
			return triggerSuggestForExprConstraints(et.AnyElem)
		case schema.ListExpr:
			return triggerSuggestForExprConstraints(et.Elem)
		case schema.SetExpr:
			return triggerSuggestForExprConstraints(et.Elem)
		case schema.TupleExpr:
			if len(et.Elems) > 0 {
				return triggerSuggestForExprConstraints(et.Elems[0])
			}
		case schema.MapExpr:
			return triggerSuggestForExprConstraints(et.Elem)
		case schema.ObjectExpr:
			if len(et.Attributes) > 0 {
				return true
			}
		}
	}
	return false
}

func snippetForExprContraints(placeholder uint, ec schema.ExprConstraints) string {
	if len(ec) > 0 {
		expr := ec[0]

		switch et := expr.(type) {
		case schema.LiteralTypeExpr:
			return snippetForLiteralType(placeholder, et.Type)
		case schema.LiteralValue:
			if len(ec) == 1 {
				return snippetForLiteralValue(placeholder, et.Val)
			}
			return ""
		case schema.TupleConsExpr:
			ec := ExprConstraints(et.AnyElem)
			if ec.HasKeywordsOnly() {
				return "[ ${0} ]"
			}
			return "[\n  ${0}\n]"
		case schema.ListExpr:
			ec := ExprConstraints(et.Elem)
			if ec.HasKeywordsOnly() {
				return "[ ${0} ]"
			}
			return "[\n  ${0}\n]"
		case schema.SetExpr:
			ec := ExprConstraints(et.Elem)
			if ec.HasKeywordsOnly() {
				return "[ ${0} ]"
			}
			return "[\n  ${0}\n]"
		case schema.TupleExpr:
			// TODO: multiple constraints?
			ec := ExprConstraints(et.Elems[0])
			if ec.HasKeywordsOnly() {
				return "[ ${0} ]"
			}
			return "[\n  ${0}\n]"
		case schema.MapExpr:
			return fmt.Sprintf("{\n  ${%d:name} = %s\n }",
				placeholder,
				snippetForExprContraints(placeholder+1, et.Elem))
		case schema.ObjectExpr:
			return fmt.Sprintf("{\n  ${%d}\n }", placeholder+1)
		}
		return ""
	}
	return ""
}

func snippetForExprContraint(placeholder uint, ec schema.ExprConstraints) string {
	e := ExprConstraints(ec)

	// TODO: implement rest of these
	if t, ok := e.LiteralType(); ok {
		return snippetForLiteralType(placeholder, t)
	}
	// case schema.LiteralValue:
	// 	if len(ec) == 1 {
	// 		return snippetForLiteralValue(placeholder, et.Val)
	// 	}
	// 	return ""
	if t, ok := e.TupleConsExpr(); ok {
		ec := ExprConstraints(t.AnyElem)
		if ec.HasKeywordsOnly() {
			return "[ ${0} ]"
		}
		return "[\n  ${0}\n]"
	}
	if t, ok := e.ListExpr(); ok {
		ec := ExprConstraints(t.Elem)
		if ec.HasKeywordsOnly() {
			return "[ ${0} ]"
		}
		return "[\n  ${0}\n]"
	}
	if t, ok := e.SetExpr(); ok {
		ec := ExprConstraints(t.Elem)
		if ec.HasKeywordsOnly() {
			return "[ ${0} ]"
		}
		return "[\n  ${0}\n]"
	}
	if t, ok := e.TupleExpr(); ok {
		// TODO: multiple constraints?
		ec := ExprConstraints(t.Elems[0])
		if ec.HasKeywordsOnly() {
			return "[ ${0} ]"
		}
		return "[\n  ${0}\n]"
	}
	if t, ok := e.MapExpr(); ok {
		return fmt.Sprintf("{\n  ${%d:name} = %s\n }",
			placeholder,
			snippetForExprContraints(placeholder+1, t.Elem))
	}
	if _, ok := e.ObjectExpr(); ok {
		return fmt.Sprintf("{\n  ${%d}\n }", placeholder+1)
	}

	return ""
}

type snippetGenerator struct {
	placeholder uint
}

func snippetForLiteralType(placeholder uint, attrType cty.Type) string {
	sg := &snippetGenerator{placeholder: placeholder}
	return sg.forLiteralType(attrType, 0)
}

func (sg *snippetGenerator) forLiteralType(attrType cty.Type, nestingLvl int) string {
	switch attrType {
	case cty.String:
		sg.placeholder++
		return fmt.Sprintf(`"${%d:value}"`, sg.placeholder-1)
	case cty.Bool:
		sg.placeholder++
		return fmt.Sprintf(`${%d:false}`, sg.placeholder-1)
	case cty.Number:
		sg.placeholder++
		return fmt.Sprintf(`${%d:1}`, sg.placeholder-1)
	case cty.DynamicPseudoType:
		sg.placeholder++
		return fmt.Sprintf(`${%d}`, sg.placeholder-1)
	}

	nesting := strings.Repeat("  ", nestingLvl+1)
	endBraceNesting := strings.Repeat("  ", nestingLvl)

	if attrType.IsMapType() {
		mapSnippet := "{\n"
		mapSnippet += fmt.Sprintf(`%s"${%d:key}" = `, nesting, sg.placeholder)
		sg.placeholder++
		mapSnippet += sg.forLiteralType(*attrType.MapElementType(), nestingLvl+1)
		mapSnippet += fmt.Sprintf("\n%s}", endBraceNesting)
		return mapSnippet
	}

	if attrType.IsListType() || attrType.IsSetType() {
		elType := attrType.ElementType()
		return fmt.Sprintf("[ %s ]", sg.forLiteralType(elType, nestingLvl))
	}

	if attrType.IsObjectType() {
		objSnippet := ""
		for _, name := range sortedObjectAttrNames(attrType) {
			valType := attrType.AttributeType(name)

			objSnippet += fmt.Sprintf("%s%s = %s\n",
				nesting, name, sg.forLiteralType(valType, nestingLvl+1))
		}
		return fmt.Sprintf("{\n%s%s}", objSnippet, endBraceNesting)
	}

	if attrType.IsTupleType() {
		elTypes := attrType.TupleElementTypes()
		if len(elTypes) == 1 {
			return fmt.Sprintf("[ %s ]", sg.forLiteralType(elTypes[0], nestingLvl))
		}

		tupleSnippet := ""
		for _, elType := range elTypes {
			tupleSnippet += sg.forLiteralType(elType, nestingLvl+1)
		}
		return fmt.Sprintf("[\n%s]", tupleSnippet)
	}

	return ""
}

func labelForLiteralValue(val cty.Value, isNested bool) string {
	if !val.IsWhollyKnown() {
		return ""
	}

	switch val.Type() {
	case cty.Bool:
		return fmt.Sprintf("%t", val.True())
	case cty.String:
		if isNested {
			return fmt.Sprintf("%q", val.AsString())
		}
		return val.AsString()
	case cty.Number:
		return formatNumberVal(val)
	}

	if val.Type().IsMapType() {
		label := `{ `
		valueMap := val.AsValueMap()
		mapKeys := sortedKeysOfValueMap(valueMap)
		i := 0
		for _, key := range mapKeys {
			if i > 0 {
				label += ", "
			}
			if len(label) > 10 {
				label += "…"
				break
			}

			label += fmt.Sprintf("%q = %s",
				key, labelForLiteralValue(valueMap[key], true))
			i++
		}
		label += ` }`
		return label
	}

	if val.Type().IsListType() || val.Type().IsSetType() || val.Type().IsTupleType() {
		label := `[ `
		for i, elem := range val.AsValueSlice() {
			if i > 0 {
				label += ", "
			}
			if len(label) > 10 {
				label += "…"
				break
			}

			label += labelForLiteralValue(elem, true)

		}
		label += ` ]`
		return label
	}

	if val.Type().IsObjectType() {
		label := `{ `
		attrNames := sortedObjectAttrNames(val.Type())
		i := 0
		for _, name := range attrNames {
			if i > 0 {
				label += ", "
			}
			if len(label) > 10 {
				label += "…"
				break
			}
			val := val.GetAttr(name)

			label += fmt.Sprintf("%s = %s", name, labelForLiteralValue(val, true))
			i++
		}

		label += ` }`
		return label
	}

	return ""
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

func labelForLiteralType(attrType cty.Type) string {
	if attrType.IsMapType() {
		elType := *attrType.MapElementType()
		return fmt.Sprintf(`{ "key" = %s }`,
			labelForLiteralType(elType))
	}

	if attrType.IsListType() || attrType.IsSetType() {
		elType := attrType.ElementType()
		return fmt.Sprintf(`[ %s ]`,
			labelForLiteralType(elType))
	}

	if attrType.IsTupleType() {
		elTypes := attrType.TupleElementTypes()
		if len(elTypes) > 2 {
			return fmt.Sprintf("[ %s , %s , … ]",
				labelForLiteralType(elTypes[0]),
				labelForLiteralType(elTypes[1]))
		}
		if len(elTypes) == 2 {
			return fmt.Sprintf("[ %s , %s ]",
				labelForLiteralType(elTypes[0]),
				labelForLiteralType(elTypes[1]))
		}
		if len(elTypes) == 1 {
			return fmt.Sprintf("[ %s ]", labelForLiteralType(elTypes[0]))
		}
		return "[ ]"
	}

	if attrType.IsObjectType() {
		attrNames := sortedObjectAttrNames(attrType)
		label := "{ "
		for i, attrName := range attrNames {
			if i > 0 {
				label += ", "
			}
			if len(label) > 10 {
				label += "…"
				break
			}

			label += fmt.Sprintf("%s = %s",
				attrName,
				labelForLiteralType(attrType.AttributeType(attrName)))
		}
		label += " }"
		return label
	}

	return attrType.FriendlyNameForConstraint()
}

func newTextForLiteralValue(val cty.Value) string {
	switch val.Type() {
	case cty.String:
		return fmt.Sprintf("%q", val.AsString())
	case cty.Bool:
		return fmt.Sprintf("%t", val.True())
	case cty.Number:
		return formatNumberVal(val)
	case cty.DynamicPseudoType:
		return ""
	}

	if val.Type().IsMapType() {
		newText := "{\n"
		valueMap := val.AsValueMap()
		mapKeys := sortedKeysOfValueMap(valueMap)
		for _, key := range mapKeys {
			newText += fmt.Sprintf("  %q = %s\n",
				key, newTextForLiteralValue(valueMap[key]))
		}
		newText += "}"
		return newText
	}

	if val.Type().IsListType() || val.Type().IsSetType() || val.Type().IsTupleType() {
		newText := "[\n"
		for _, elem := range val.AsValueSlice() {
			newText += fmt.Sprintf("  %s,\n", newTextForLiteralValue(elem))
		}
		newText += "]"
		return newText
	}

	if val.Type().IsObjectType() {
		newText := "{\n"
		attrNames := sortedObjectAttrNames(val.Type())
		for _, name := range attrNames {
			v := val.GetAttr(name)
			newText += fmt.Sprintf("  %s = %s\n", name, newTextForLiteralValue(v))
		}
		newText += "}"
		return newText
	}

	return ""
}

func snippetForLiteralValue(placeholder uint, val cty.Value) string {
	sg := &snippetGenerator{placeholder: placeholder}
	return sg.forLiteralValue(val, 0)
}

func (sg *snippetGenerator) forLiteralValue(val cty.Value, nestingLvl int) string {
	switch val.Type() {
	case cty.String:
		sg.placeholder++
		return fmt.Sprintf(`"${%d:%s}"`, sg.placeholder-1, val.AsString())
	case cty.Bool:
		sg.placeholder++
		return fmt.Sprintf(`${%d:%t}`, sg.placeholder-1, val.True())
	case cty.Number:
		sg.placeholder++
		return fmt.Sprintf(`${%d:%s}`, sg.placeholder-1, formatNumberVal(val))
	case cty.DynamicPseudoType:
		sg.placeholder++
		return fmt.Sprintf(`${%d}`, sg.placeholder-1)
	}

	nesting := strings.Repeat("  ", nestingLvl+1)
	endBraceNesting := strings.Repeat("  ", nestingLvl)

	if val.Type().IsMapType() {
		mapSnippet := "{\n"
		valueMap := val.AsValueMap()
		mapKeys := sortedKeysOfValueMap(valueMap)
		for _, key := range mapKeys {
			mapSnippet += fmt.Sprintf(`%s"${%d:%s}" = `, nesting, sg.placeholder, key)
			sg.placeholder++
			mapSnippet += sg.forLiteralValue(valueMap[key], nestingLvl+1)
			mapSnippet += "\n"
		}
		mapSnippet += fmt.Sprintf("%s}", endBraceNesting)
		return mapSnippet
	}

	if val.Type().IsListType() || val.Type().IsSetType() || val.Type().IsTupleType() {
		snippet := "[\n"
		for _, elem := range val.AsValueSlice() {
			snippet += fmt.Sprintf("%s%s,\n", nesting, sg.forLiteralValue(elem, nestingLvl+1))
		}
		snippet += fmt.Sprintf("%s]", endBraceNesting)
		return snippet
	}

	if val.Type().IsObjectType() {
		snippet := "{\n"
		for _, name := range sortedObjectAttrNames(val.Type()) {
			v := val.GetAttr(name)
			snippet += fmt.Sprintf("%s%s = %s\n",
				nesting, name, sg.forLiteralValue(v, nestingLvl+1))
		}
		snippet += fmt.Sprintf("%s}", endBraceNesting)
		return snippet
	}

	return ""
}

func sortedKeysOfValueMap(m map[string]cty.Value) []string {
	keys := make([]string, 0)
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func newTextForLiteralType(attrType cty.Type) string {
	switch attrType {
	case cty.String:
		return `""`
	case cty.Bool:
		return `false`
	case cty.Number:
		return `1`
	case cty.DynamicPseudoType:
		return ``
	}

	if attrType.IsMapType() {
		elType := *attrType.MapElementType()
		return fmt.Sprintf("{\n"+`  "key" = %s`+"\n}",
			newTextForLiteralType(elType))
	}

	if attrType.IsListType() || attrType.IsSetType() {
		elType := attrType.ElementType()
		return fmt.Sprintf("[ %s ]", newTextForLiteralType(elType))
	}

	if attrType.IsObjectType() {
		objSnippet := ""
		attrNames := sortedObjectAttrNames(attrType)
		for _, name := range attrNames {
			valType := attrType.AttributeType(name)

			objSnippet += fmt.Sprintf("  %s = %s\n", name,
				newTextForLiteralType(valType))
		}
		return fmt.Sprintf("{\n%s}", objSnippet)
	}

	if attrType.IsTupleType() {
		elTypes := attrType.TupleElementTypes()
		if len(elTypes) == 1 {
			return fmt.Sprintf("[ %s ]", newTextForLiteralType(elTypes[0]))
		}

		tupleSnippet := ""
		for _, elType := range elTypes {
			tupleSnippet += newTextForLiteralType(elType)
		}
		return fmt.Sprintf("[\n%s]", tupleSnippet)
	}

	return ""
}
