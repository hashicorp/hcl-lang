// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package walker

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl-lang/schemacontext"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func TestWalk_basic(t *testing.T) {
	ctx := context.Background()
	src := []byte(`
first {
  nested1 {}
  nested2 {}
}
second {
  foo = "bar"
}
`)
	file, diags := hclsyntax.ParseConfig(src, "test.hcl", hcl.InitialPos)
	if len(diags) > 0 {
		t.Fatalf("unexpected parser diagnostics: %s", diags)
	}

	rootSchema := schema.NewBodySchema()
	nodeCount := 0
	tw := NewTestWalker(func(ctx context.Context, node hclsyntax.Node, nodeSchema schema.Schema) (context.Context, hcl.Diagnostics) {
		nodeCount++
		return ctx, nil
	})
	diags = Walk(ctx, file.Body.(*hclsyntax.Body), rootSchema, tw)
	if len(diags) > 0 {
		t.Fatalf("unexpected diagnostics: %s", diags)
	}
	// root body + first(block+body) + nested1(block+body) + nested2(block+body) + second(block+body) + foo
	expectedNodeCount := 10
	if nodeCount != expectedNodeCount {
		t.Fatalf("unexpected node count: %d, expected: %d", nodeCount, expectedNodeCount)
	}
}

func TestWalk_nestingLevel(t *testing.T) {
	ctx := context.Background()
	src := []byte(`
first {
  nested1 {}
  nested2 {}
}
rootattr = "foo"
second {
  foo = "bar"
}
`)
	file, pDiags := hclsyntax.ParseConfig(src, "test.hcl", hcl.InitialPos)
	if len(pDiags) > 0 {
		t.Fatalf("unexpected parser diagnostics: %s", pDiags)
	}

	rootSchema := schema.NewBodySchema()

	firstBlockFound := false
	nestedBlockFound := false
	rootAttrFound := false
	nestedAttrFound := false

	tw := NewTestWalker(func(ctx context.Context, node hclsyntax.Node, nodeSchema schema.Schema) (context.Context, hcl.Diagnostics) {
		var diags hcl.Diagnostics

		nestingLvl, ok := schemacontext.BlockNestingLevel(ctx)
		if !ok {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "nesting level not available",
			})
		}

		if block, ok := node.(*hclsyntax.Block); ok && block.Type == "first" {
			firstBlockFound = true
			if nestingLvl != 0 {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("%q: unexpected nesting level %d, expected 0", "first", nestingLvl),
				})
			}
		}
		if block, ok := node.(*hclsyntax.Block); ok && block.Type == "nested2" {
			nestedBlockFound = true
			if nestingLvl != 1 {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("%q: unexpected nesting level %d, expected 1", "nested2", nestingLvl),
				})
			}
		}

		if attr, ok := node.(*hclsyntax.Attribute); ok && attr.Name == "rootattr" {
			rootAttrFound = true
			if nestingLvl != 0 {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("%q: unexpected nesting level %d, expected 0", "rootattr", nestingLvl),
				})
			}
		}

		if attr, ok := node.(*hclsyntax.Attribute); ok && attr.Name == "foo" {
			nestedAttrFound = true
			if nestingLvl != 1 {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("%q: unexpected nesting level %d, expected 1", "foo", nestingLvl),
				})
			}
		}

		return ctx, diags
	})

	diags := Walk(ctx, file.Body.(*hclsyntax.Body), rootSchema, tw)
	if len(diags) > 0 {
		t.Fatalf("unexpected diagnostics: %s", diags)
	}

	if !firstBlockFound {
		t.Error("expected first block to be found")
	}
	if !nestedBlockFound {
		t.Error("expected nested block to be found")
	}
	if !rootAttrFound {
		t.Error("expected root attribute to be found")
	}
	if !nestedAttrFound {
		t.Error("expected nested attribute to be found")
	}
}

func NewTestWalker(visitFunc visitFunc) Walker {
	return testWalker{
		visitFunc: visitFunc,
	}
}

type visitFunc func(ctx context.Context, node hclsyntax.Node, nodeSchema schema.Schema) (context.Context, hcl.Diagnostics)

type testWalker struct {
	visitFunc visitFunc
}

func (tw testWalker) Visit(ctx context.Context, node hclsyntax.Node, nodeSchema schema.Schema) (context.Context, hcl.Diagnostics) {
	return tw.visitFunc(ctx, node, nodeSchema)
}
