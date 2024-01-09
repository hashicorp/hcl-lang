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
