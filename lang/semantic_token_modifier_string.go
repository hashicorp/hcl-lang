// Code generated by "stringer -type=SemanticTokenModifier -output=semantic_token_modifier_string.go"; DO NOT EDIT.

package lang

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TokenModifierNil-0]
	_ = x[TokenModifierDependent-1]
}

const _SemanticTokenModifier_name = "TokenModifierNilTokenModifierDependent"

var _SemanticTokenModifier_index = [...]uint8{0, 16, 38}

func (i SemanticTokenModifier) String() string {
	if i >= SemanticTokenModifier(len(_SemanticTokenModifier_index)-1) {
		return "SemanticTokenModifier(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _SemanticTokenModifier_name[_SemanticTokenModifier_index[i]:_SemanticTokenModifier_index[i+1]]
}
