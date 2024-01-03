// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/reference"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func (d *PathDecoder) attrValueCompletionAtPos(ctx context.Context, attr *hclsyntax.Attribute, schema *schema.AttributeSchema, outerBodyRng hcl.Range, pos hcl.Pos) (lang.Candidates, error) {
	candidates := lang.NewCandidates()
	candidates.IsComplete = true

	if len(schema.CompletionHooks) > 0 {
		candidates.IsComplete = false
		candidates.List = append(candidates.List, d.candidatesFromHooks(ctx, attr, schema, outerBodyRng, pos)...)
	}
	count := len(candidates.List)

	if uint(count) < d.maxCandidates {
		expr := d.newExpression(attr.Expr, schema.Constraint)
		for _, candidate := range expr.CompletionAtPos(ctx, pos) {
			if uint(count) >= d.maxCandidates {
				return candidates, nil
			}

			candidates.List = append(candidates.List, candidate)
			count++
		}
	}

	return candidates, nil
}

type pathKey struct{}

// WithPath is not intended to be used outside this package
// except for testing hooks downstream.
func WithPath(ctx context.Context, path lang.Path) context.Context {
	return context.WithValue(ctx, pathKey{}, path)
}

func PathFromContext(ctx context.Context) (lang.Path, bool) {
	p, ok := ctx.Value(pathKey{}).(lang.Path)
	return p, ok
}

type posKey struct{}

// WithPos is not intended to be used outside this package
// except for testing hooks downstream.
func WithPos(ctx context.Context, pos hcl.Pos) context.Context {
	return context.WithValue(ctx, posKey{}, pos)
}

func PosFromContext(ctx context.Context) (hcl.Pos, bool) {
	p, ok := ctx.Value(posKey{}).(hcl.Pos)
	return p, ok
}

type filenameKey struct{}

// WithFilename is not intended to be used outside this package
// except for testing hooks downstream.
func WithFilename(ctx context.Context, filename string) context.Context {
	return context.WithValue(ctx, filenameKey{}, filename)
}

func FilenameFromContext(ctx context.Context) (string, bool) {
	f, ok := ctx.Value(filenameKey{}).(string)
	return f, ok
}

type maxCandidatesKey struct{}

// WithMaxCandidates is not intended to be used outside this package
// except for testing hooks downstream.
func WithMaxCandidates(ctx context.Context, maxCandidates uint) context.Context {
	return context.WithValue(ctx, maxCandidatesKey{}, maxCandidates)
}

func MaxCandidatesFromContext(ctx context.Context) (uint, bool) {
	mc, ok := ctx.Value(maxCandidatesKey{}).(uint)
	return mc, ok
}

func isMultilineTemplateExpr(expr hclsyntax.Expression) bool {
	t, ok := expr.(*hclsyntax.TemplateExpr)
	if !ok {
		return false
	}
	return t.Range().Start.Line != t.Range().End.Line
}

func (d *PathDecoder) candidatesFromHooks(ctx context.Context, attr *hclsyntax.Attribute, aSchema *schema.AttributeSchema, outerBodyRng hcl.Range, pos hcl.Pos) []lang.Candidate {
	candidates := make([]lang.Candidate, 0)
	con, ok := aSchema.Constraint.(schema.TypeAwareConstraint)
	if !ok {
		// Return early as we only support string values for now
		return candidates
	}
	typ, ok := con.ConstraintType()
	if !ok || typ != cty.String {
		// Return early as we only support string values for now
		return candidates
	}

	editRng := attr.Expr.Range()
	if isEmptyExpression(attr.Expr) || isMultilineTemplateExpr(attr.Expr) {
		// An empty expression or a string without a closing quote will lead to
		// an attribute expression spanning multiple lines.
		// Since text edits only support a single line, we're resetting the End
		// position here.
		editRng.End = pos
	}
	prefixRng := attr.Expr.Range()
	prefixRng.End = pos
	prefixBytes, _ := d.bytesFromRange(prefixRng)
	prefix := string(prefixBytes)
	prefix = strings.TrimLeft(prefix, `"`)

	ctx = WithPath(ctx, d.path)
	ctx = WithFilename(ctx, attr.Expr.Range().Filename)
	ctx = WithPos(ctx, pos)
	ctx = WithMaxCandidates(ctx, d.maxCandidates)

	count := 0
	for _, hook := range aSchema.CompletionHooks {
		if completionFunc, ok := d.decoderCtx.CompletionHooks[hook.Name]; ok {
			res, _ := completionFunc(ctx, cty.StringVal(prefix))

			for _, c := range res {
				if uint(count) >= d.maxCandidates {
					return candidates
				}

				candidates = append(candidates, lang.Candidate{
					Label:        c.Label,
					Detail:       c.Detail,
					Description:  c.Description,
					Kind:         c.Kind,
					IsDeprecated: c.IsDeprecated,
					TextEdit: lang.TextEdit{
						NewText: c.RawInsertText,
						Snippet: c.RawInsertText,
						Range:   editRng,
					},
					ResolveHook: c.ResolveHook,
					SortText:    c.SortText,
				})
				count++
			}

		}
	}

	return candidates
}

func referenceTargetIsInRange(target reference.Target, bodyRange hcl.Range) bool {
	return target.RangePtr != nil &&
		bodyRange.Filename == target.RangePtr.Filename &&
		(bodyRange.ContainsPos(target.RangePtr.Start) ||
			posEqual(bodyRange.End, target.RangePtr.End))
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

type labelSet []string

func (ls labelSet) String() string {
	if len(ls) > 10 {
		return strings.Join(ls[:9], " | ") + " …"
	}

	return strings.Join(ls, " | ")
}

func (ls labelSet) AddLabelIfNotPresent(label string) labelSet {
	if !ls.HasLabel(label) {
		return append(ls, label)
	}
	return ls
}

func (ls labelSet) HasLabel(label string) bool {
	for _, l := range ls {
		if l == label {
			return true
		}
	}
	return false
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
