package schema

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
)

// OneOf represents multiple constraints where any one of them is acceptable.
type OneOf []Constraint

func (OneOf) isConstraintImpl() constraintSigil {
	return constraintSigil{}
}

func (o OneOf) FriendlyName() string {
	names := make([]string, 0)
	for _, constraint := range o {
		if name := constraint.FriendlyName(); name != "" &&
			!namesContain(names, name) {
			names = append(names, name)
		}
	}
	if len(names) > 0 {
		return strings.Join(names, " or ")
	}
	return ""
}

func (o OneOf) Copy() Constraint {
	if o == nil {
		return make(OneOf, 0)
	}

	newCons := make(OneOf, len(o))
	for i, c := range o {
		newCons[i] = c.Copy()
	}

	return newCons
}

func (o OneOf) Validate() error {
	if len(o) == 0 {
		return nil
	}

	var errs *multierror.Error

	for i, constraint := range o {
		if c, ok := constraint.(Validatable); ok {
			err := c.Validate()
			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("(%d: %T) %w", i, constraint, err))
			}
		}
	}

	if errs != nil && len(errs.Errors) == 1 {
		return errs.Errors[0]
	}

	return errs.ErrorOrNil()
}

func namesContain(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

func (o OneOf) EmptyCompletionData(nextPlaceholder int) CompletionData {
	// TODO
	return CompletionData{}
}
