// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

type schemaImplSigil struct{}

// Schema represents any schema (e.g. attribute, label, or a block)
type Schema interface {
	isSchemaImpl() schemaImplSigil
}
