package schema

// TypeDeclaration represents a type declaration as
// interpreted by HCL's ext/typeexpr package,
// i.e. declaration of cty.Type in HCL
type TypeDeclaration struct{}

func (TypeDeclaration) isConstraintImpl() constraintSigil {
	return constraintSigil{}
}

func (td TypeDeclaration) FriendlyName() string {
	return "type"
}

func (td TypeDeclaration) Copy() Constraint {
	return TypeDeclaration{}
}
