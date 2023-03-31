// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lang

import (
	"github.com/zclconf/go-cty/cty"
)

type exprKindSigil struct{}

type SymbolExprKind interface {
	isSymbolExprKindSigil() exprKindSigil
}

type LiteralTypeKind struct {
	Type cty.Type
}

func (LiteralTypeKind) isSymbolExprKindSigil() exprKindSigil {
	return exprKindSigil{}
}

type TupleConsExprKind struct{}

func (TupleConsExprKind) isSymbolExprKindSigil() exprKindSigil {
	return exprKindSigil{}
}

type ObjectConsExprKind struct{}

func (ObjectConsExprKind) isSymbolExprKindSigil() exprKindSigil {
	return exprKindSigil{}
}

type TraversalExprKind struct{}

func (TraversalExprKind) isSymbolExprKindSigil() exprKindSigil {
	return exprKindSigil{}
}
