// Copyright IBM Corp. 2020, 2025
// SPDX-License-Identifier: MPL-2.0

package decoder

var (
	_ Symbol = &AttributeSymbol{}
	_ Symbol = &BlockSymbol{}
	_ Symbol = &ExprSymbol{}
)
