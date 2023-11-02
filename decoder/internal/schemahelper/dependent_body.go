// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schemahelper

import (
	"github.com/hashicorp/hcl-lang/decoder/internal/ast"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type LookupResult int

const (
	LookupFailed LookupResult = iota
	LookupSuccessful
	LookupPartiallySuccessful
	NoDependentKeys
)

type blockSchema struct {
	*schema.BlockSchema
	seenNestedDepKeys bool
}

func NewBlockSchema(bs *schema.BlockSchema) blockSchema {
	return blockSchema{BlockSchema: bs}
}

// DependentBodySchema finds relevant BodySchema based on dependency keys
// such as a label or an attribute (or combination of both).
func (bs blockSchema) DependentBodySchema(block *hcl.Block) (*schema.BodySchema, schema.DependencyKeys, LookupResult) {
	result := LookupFailed

	dks := dependencyKeysFromBlock(block, bs)
	b, err := dks.MarshalJSON()
	if err != nil {
		return nil, schema.DependencyKeys{}, result
	}

	if len(dks.Labels) == 0 && len(dks.Attributes) == 0 {
		return bs.Body, schema.DependencyKeys{}, NoDependentKeys
	}

	key := schema.SchemaKey(string(b))
	depBodySchema, ok := bs.DependentBody[key]
	if ok {
		result = LookupSuccessful
		hasDepKeys := false
		for _, attr := range depBodySchema.Attributes {
			if attr.IsDepKey {
				hasDepKeys = true
			}
		}

		if hasDepKeys && !bs.seenNestedDepKeys {
			mergedBlockSchema := NewBlockSchema(bs.Copy())
			mergedBlockSchema.seenNestedDepKeys = true
			mergedBlockSchema.Body = depBodySchema
			if depBodySchema, dks, nestedOk := mergedBlockSchema.DependentBodySchema(block); nestedOk == LookupSuccessful {
				return depBodySchema, dks, LookupSuccessful
			} else {
				// Ensure we report lookup failure overall if we couldn't
				// lookup nested dependent body
				result = LookupPartiallySuccessful
			}
		}
	}

	return depBodySchema, dks, result
}

func dependencyKeysFromBlock(block *hcl.Block, blockSchema blockSchema) schema.DependencyKeys {
	dk := schema.DependencyKeys{
		Labels:     []schema.LabelDependent{},
		Attributes: []schema.AttributeDependent{},
	}
	for i, labelSchema := range blockSchema.Labels {
		if labelSchema.IsDepKey {
			if i+1 > len(block.Labels) {
				// mismatching label schema
				return dk
			}

			dk.Labels = append(dk.Labels, schema.LabelDependent{
				Index: i,
				Value: block.Labels[i],
			})
		}
	}

	if blockSchema.Body == nil {
		return dk
	}

	if block.Body == nil {
		// no attributes to find
		return dk
	}

	content := ast.DecodeBody(block.Body, blockSchema.Body)

	for name, attrSchema := range blockSchema.Body.Attributes {
		if attrSchema.IsDepKey {
			var value cty.Value
			attr, ok := content.Attributes[name]
			if ok {
				st, ok := attr.Expr.(*hclsyntax.ScopeTraversalExpr)
				if ok {
					addr, err := lang.TraversalToAddress(st.AsTraversal())
					if err != nil {
						// skip unparsable traversal
						continue
					}
					dk.Attributes = append(dk.Attributes, schema.AttributeDependent{
						Name: name,
						Expr: schema.ExpressionValue{
							Address: addr,
						},
					})
					continue
				}

				var diags hcl.Diagnostics
				value, diags = attr.Expr.Value(nil)
				if len(diags) > 0 && value.IsNull() {
					// skip attribute if we can't get the value
					continue
				}
			} else if attrSchema.DefaultValue != nil {
				defaultValue, ok := attrSchema.DefaultValue.(schema.DefaultValue)
				if !ok {
					// TODO: DefaultKeyword
					// TODO: DefaultTypeDeclaration
					continue
				}
				value = defaultValue.Value
			} else {
				// dependent attribute not present
				continue
			}

			dk.Attributes = append(dk.Attributes, schema.AttributeDependent{
				Name: name,
				Expr: schema.ExpressionValue{
					Static: value,
				},
			})
		}
	}
	return dk
}
