package schema

type constraintSigil struct{}

type Constraint interface {
	isConstraintImpl() constraintSigil
	FriendlyName() string
	Copy() Constraint
}

type Validatable interface {
	Validate() error
}
