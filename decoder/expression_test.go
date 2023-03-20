package decoder

import (
	"bytes"
	"fmt"
	"testing"
	"unicode"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

var (
	_ Expression = Any{}
	_ Expression = Keyword{}
	_ Expression = List{}
	_ Expression = LiteralType{}
	_ Expression = LiteralValue{}
	_ Expression = Map{}
	_ Expression = Object{}
	_ Expression = Set{}
	_ Expression = Reference{}
	_ Expression = Tuple{}
	_ Expression = TypeDeclaration{}

	_ ReferenceOriginsExpression = Any{}
	_ ReferenceOriginsExpression = List{}
	_ ReferenceOriginsExpression = Map{}
	_ ReferenceOriginsExpression = Object{}
	_ ReferenceOriginsExpression = Set{}
	_ ReferenceOriginsExpression = Reference{}
	_ ReferenceOriginsExpression = Tuple{}

	_ ReferenceTargetsExpression = Any{}
	_ ReferenceTargetsExpression = List{}
	_ ReferenceTargetsExpression = LiteralType{}
	_ ReferenceTargetsExpression = Map{}
	_ ReferenceTargetsExpression = Object{}
	_ ReferenceTargetsExpression = Reference{}
	_ ReferenceTargetsExpression = Tuple{}
)

func TestRecoverLeftBytes(t *testing.T) {
	testCases := []struct {
		b             []byte
		pos           hcl.Pos
		f             func(int, rune) bool
		expectedBytes []byte
	}{
		{
			[]byte(`toot  foobar`),
			hcl.Pos{Line: 1, Column: 13, Byte: 12},
			func(i int, r rune) bool {
				return unicode.IsSpace(r)
			},
			[]byte(` foobar`),
		},
		{
			[]byte(`hello👋world and other planets`),
			hcl.Pos{Line: 1, Column: 15, Byte: 14},
			func(i int, r rune) bool {
				return r == '👋'
			},
			[]byte(`👋world`),
		},
		{
			[]byte(`hello world👋and other planets`),
			hcl.Pos{Line: 1, Column: 16, Byte: 15},
			func(i int, r rune) bool {
				return r == '👋'
			},
			[]byte(`👋`),
		},
		{
			[]byte(`attr = {
  foo = "foo",
  bar = 
}
`),
			hcl.Pos{Line: 3, Column: 9, Byte: 32},
			func(i int, r rune) bool {
				return r == '\n' || r == ','
			},
			[]byte("\n  bar = "),
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			recoveredBytes := recoverLeftBytes(tc.b, tc.pos, tc.f)
			if !bytes.Equal(tc.expectedBytes, recoveredBytes) {
				t.Fatalf("mismatch!\nexpected:  %q\nrecovered: %q\n", string(tc.expectedBytes), string(recoveredBytes))
			}
		})
	}
}

func TestRawObjectKey(t *testing.T) {
	testCases := []struct {
		cfg           string
		expectedKey   string
		expectedRange *hcl.Range
		expectedOk    bool
	}{
		{
			`attr = { foo = "bar" }`,
			"foo",
			&hcl.Range{
				Filename: "test.hcl",
				Start:    hcl.Pos{Line: 1, Column: 10, Byte: 9},
				End:      hcl.Pos{Line: 1, Column: 13, Byte: 12},
			},
			true,
		},
		{
			`attr = { 42 = "bar" }`,
			"",
			nil,
			false,
		},
		{
			`attr = { foo.x = "bar" }`,
			"",
			nil,
			false,
		},
		{
			`attr = { (foo) = "bar" }`,
			"",
			nil,
			false,
		},
		{
			`attr = { (var.foo) = "bar" }`,
			"",
			nil,
			false,
		},
		{
			`attr = { "foo" = "bar" }`,
			"foo",
			&hcl.Range{
				Filename: "test.hcl",
				Start:    hcl.Pos{Line: 1, Column: 11, Byte: 10},
				End:      hcl.Pos{Line: 1, Column: 14, Byte: 13},
			},
			true,
		},
		{
			`attr = { "${foo}" = "bar" }`,
			"",
			nil,
			false,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			f, diags := hclsyntax.ParseConfig([]byte(tc.cfg), "test.hcl", hcl.InitialPos)
			if len(diags) > 0 {
				t.Fatal(diags)
			}

			attrs, diags := f.Body.JustAttributes()
			if len(diags) > 0 {
				t.Fatal(diags)
			}

			if len(attrs) != 1 {
				t.Fatalf("expected exactly 1 attribute, %d given", len(attrs))
			}

			attr, ok := attrs["attr"]
			if !ok {
				t.Fatalf("expected to find attribute %q", "attr")
			}

			objConsExpr, ok := attr.Expr.(*hclsyntax.ObjectConsExpr)
			if !ok {
				t.Fatalf("expected expression to be ObjectConsExpr, %T given", attr.Expr)
			}

			if len(objConsExpr.Items) != 1 {
				t.Fatalf("expected exactly 1 object item, %d given", len(objConsExpr.Items))
			}

			rawKey, rng, ok := rawObjectKey(objConsExpr.Items[0].KeyExpr)
			if !tc.expectedOk && ok {
				t.Fatal("expected parsing to fail")
			}
			if tc.expectedOk && !ok {
				t.Fatal("expected parsing to succeed")
			}
			if tc.expectedKey != rawKey {
				t.Fatalf("extracted key mismatch\nexpected: %q\ngiven: %q", tc.expectedKey, rawKey)
			}
			if diff := cmp.Diff(tc.expectedRange, rng); diff != "" {
				t.Fatalf("unexpected range: %s", diff)
			}
		})
	}
}
