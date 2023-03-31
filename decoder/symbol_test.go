// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

var (
	_ Symbol = &AttributeSymbol{}
	_ Symbol = &BlockSymbol{}
	_ Symbol = &ExprSymbol{}
)
