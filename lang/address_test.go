package lang

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestAddress_Equals_basic(t *testing.T) {
	originalAddr := Address{
		RootStep{Name: "provider"},
		AttrStep{Name: "aws"},
	}

	matchingAddr := Address{
		RootStep{Name: "provider"},
		AttrStep{Name: "aws"},
	}
	if !originalAddr.Equals(Address(matchingAddr)) {
		t.Fatalf("expected %q to match %q", originalAddr, matchingAddr)
	}

	mismatchingAddr := Address{
		RootStep{Name: "provider"},
		AttrStep{Name: "aaa"},
	}
	if originalAddr.Equals(Address(mismatchingAddr)) {
		t.Fatalf("expected %q not to match %q", originalAddr, mismatchingAddr)
	}
}

func TestAddress_Equals_numericIndexStep(t *testing.T) {
	originalAddr := Address{
		RootStep{Name: "aws_alb"},
		AttrStep{Name: "example"},
		IndexStep{Key: cty.NumberIntVal(0)},
	}

	matchingAddr := Address{
		RootStep{Name: "aws_alb"},
		AttrStep{Name: "example"},
		IndexStep{Key: cty.NumberIntVal(0)},
	}
	if !originalAddr.Equals(Address(matchingAddr)) {
		t.Fatalf("expected %q to match %q", originalAddr, matchingAddr)
	}

	mismatchingAddr := Address{
		RootStep{Name: "aws_alb"},
		AttrStep{Name: "example"},
		IndexStep{Key: cty.NumberIntVal(4)},
	}
	if originalAddr.Equals(Address(mismatchingAddr)) {
		t.Fatalf("expected %q not to match %q", originalAddr, mismatchingAddr)
	}
}

func TestAddress_Equals_stringIndexStep(t *testing.T) {
	originalAddr := Address{
		RootStep{Name: "aws_alb"},
		AttrStep{Name: "example"},
		IndexStep{Key: cty.StringVal("first")},
	}

	matchingAddr := Address{
		RootStep{Name: "aws_alb"},
		AttrStep{Name: "example"},
		IndexStep{Key: cty.StringVal("first")},
	}
	if !originalAddr.Equals(Address(matchingAddr)) {
		t.Fatalf("expected %q to match %q", originalAddr, matchingAddr)
	}

	mismatchingAddr := Address{
		RootStep{Name: "aws_alb"},
		AttrStep{Name: "example"},
		IndexStep{Key: cty.StringVal("second")},
	}
	if originalAddr.Equals(Address(mismatchingAddr)) {
		t.Fatalf("expected %q not to match %q", originalAddr, mismatchingAddr)
	}
}

func TestTraversalToAddress(t *testing.T) {
	testCases := []struct {
		rawTraversal string
		expectedAddr Address
	}{
		{
			"one",
			Address{
				RootStep{Name: "one"},
			},
		},
		{
			"first.second",
			Address{
				RootStep{Name: "first"},
				AttrStep{Name: "second"},
			},
		},
		{
			"foo[2]",
			Address{
				RootStep{Name: "foo"},
				IndexStep{Key: cty.NumberIntVal(2)},
			},
		},
		{
			`foo["bar"]`,
			Address{
				RootStep{Name: "foo"},
				IndexStep{Key: cty.StringVal("bar")},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			traversal, diags := hclsyntax.ParseTraversalAbs([]byte(tc.rawTraversal), "test.tf", hcl.InitialPos)
			if len(diags) > 0 {
				t.Fatal(diags)
			}

			addr, err := TraversalToAddress(traversal)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedAddr, addr, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("address mismatch: %s", diff)
			}
		})
	}
}
