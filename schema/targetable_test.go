package schema

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestNestedTargetablesForValue(t *testing.T) {
	testCases := []struct {
		name                string
		addr                lang.Address
		scopeId             lang.ScopeId
		val                 cty.Value
		expectedTargetables Targetables
	}{
		{
			"primitive type",
			lang.Address{
				lang.RootStep{Name: "foo"},
			},
			lang.ScopeId("test"),
			cty.StringVal("test"),
			nil,
		},
		{
			"set type",
			lang.Address{
				lang.RootStep{Name: "foo"},
			},
			lang.ScopeId("test"),
			cty.SetVal([]cty.Value{
				cty.StringVal("test"),
			}),
			nil,
		},
		{
			"list type",
			lang.Address{
				lang.RootStep{Name: "foo"},
			},
			lang.ScopeId("test"),
			cty.ListVal([]cty.Value{
				cty.StringVal("test"),
			}),
			Targetables{
				{
					Address: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.IndexStep{Key: cty.NumberIntVal(0)},
					},
					ScopeId: lang.ScopeId("test"),
					AsType:  cty.String,
				},
			},
		},
		{
			"tuple type",
			lang.Address{
				lang.RootStep{Name: "foo"},
			},
			lang.ScopeId("test"),
			cty.TupleVal([]cty.Value{
				cty.StringVal("test"),
			}),
			Targetables{
				{
					Address: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.IndexStep{Key: cty.NumberIntVal(0)},
					},
					ScopeId: lang.ScopeId("test"),
					AsType:  cty.String,
				},
			},
		},
		{
			"object type",
			lang.Address{
				lang.RootStep{Name: "foo"},
			},
			lang.ScopeId("test"),
			cty.ObjectVal(map[string]cty.Value{
				"attr1": cty.StringVal("test1"),
				"attr2": cty.StringVal("test2"),
				"attr3": cty.StringVal("test3"),
			}),
			Targetables{
				{
					Address: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.AttrStep{Name: "attr1"},
					},
					ScopeId: lang.ScopeId("test"),
					AsType:  cty.String,
				},
				{
					Address: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.AttrStep{Name: "attr2"},
					},
					ScopeId: lang.ScopeId("test"),
					AsType:  cty.String,
				},
				{
					Address: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.AttrStep{Name: "attr3"},
					},
					ScopeId: lang.ScopeId("test"),
					AsType:  cty.String,
				},
			},
		},
		{
			"map type",
			lang.Address{
				lang.RootStep{Name: "foo"},
			},
			lang.ScopeId("test"),
			cty.MapVal(map[string]cty.Value{
				"key": cty.StringVal("test"),
			}),
			Targetables{
				{
					Address: lang.Address{
						lang.RootStep{Name: "foo"},
						lang.IndexStep{Key: cty.StringVal("key")},
					},
					ScopeId: lang.ScopeId("test"),
					AsType:  cty.String,
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			targetables := NestedTargetablesForValue(tc.addr, tc.scopeId, tc.val)
			if diff := cmp.Diff(tc.expectedTargetables, targetables, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("mismatch of targetables: %s", diff)
			}
		})
	}
}
