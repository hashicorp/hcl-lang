package schema

import "context"

// TypeDeclaration represents a type declaration as
// interpreted by HCL's ext/typeexpr package,
// i.e. declaration of cty.Type in HCL
type TypeDeclaration struct {
	// TODO: optional object attribute mode
}

func (TypeDeclaration) isConstraintImpl() constraintSigil {
	return constraintSigil{}
}

func (td TypeDeclaration) FriendlyName() string {
	return "type"
}

func (td TypeDeclaration) Copy() Constraint {
	return TypeDeclaration{}
}

func (td TypeDeclaration) EmptyCompletionData(ctx context.Context, nextPlaceholder int, nestingLevel int) CompletionData {
	return CompletionData{
		TriggerSuggest:  true,
		NextPlaceholder: nextPlaceholder,
	}
}
