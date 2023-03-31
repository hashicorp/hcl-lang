// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/zclconf/go-cty/cty"
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

func (o OneOf) EmptyCompletionData(ctx context.Context, nextPlaceholder int, nestingLevel int) CompletionData {
	if len(o) == 0 {
		return CompletionData{
			NextPlaceholder: nextPlaceholder,
		}
	}

	cData := o[0].EmptyCompletionData(ctx, nextPlaceholder, nestingLevel)

	return CompletionData{
		NewText:         cData.NewText,
		Snippet:         cData.Snippet,
		NextPlaceholder: cData.NextPlaceholder,
		TriggerSuggest:  cData.TriggerSuggest,
	}
}

func (o OneOf) ConstraintType() (cty.Type, bool) {
	for _, cons := range o {
		c, ok := cons.(TypeAwareConstraint)
		if !ok {
			continue
		}
		typ, ok := c.ConstraintType()
		if !ok {
			continue
		}

		// Picking first type-aware constraint may not always be
		// appropriate since we cannot match it against configuration,
		// but it is mostly a pragmatic choice to mimic existing behaviours
		// based on common schema, such as OneOf{Reference{}, LiteralType{}}.
		// TODO: Revisit when AnyExpression{} is implemented & rolled out
		return typ, true
	}

	return cty.NilType, false
}
