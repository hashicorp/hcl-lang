package decoder

import (
	"bytes"
	"fmt"
	"testing"
	"unicode"

	"github.com/hashicorp/hcl/v2"
)

var (
	_ Expression = Any{}
	_ Expression = Keyword{}
	_ Expression = List{}
	_ Expression = LiteralType{}
	_ Expression = LiteralValue{}
	_ Expression = Map{}
	_ Expression = ObjectAttributes{}
	_ Expression = Object{}
	_ Expression = Set{}
	_ Expression = Reference{}
	_ Expression = Tuple{}
	_ Expression = TypeDeclaration{}

	_ ReferenceOriginsExpression = Any{}
	_ ReferenceOriginsExpression = List{}
	_ ReferenceOriginsExpression = LiteralType{}
	_ ReferenceOriginsExpression = Map{}
	_ ReferenceOriginsExpression = ObjectAttributes{}
	_ ReferenceOriginsExpression = Object{}
	_ ReferenceOriginsExpression = Set{}
	_ ReferenceOriginsExpression = Reference{}
	_ ReferenceOriginsExpression = Tuple{}

	_ ReferenceTargetsExpression = Any{}
	_ ReferenceTargetsExpression = List{}
	_ ReferenceTargetsExpression = LiteralType{}
	_ ReferenceTargetsExpression = Map{}
	_ ReferenceTargetsExpression = ObjectAttributes{}
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
