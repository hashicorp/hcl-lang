// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reference

type NoTargetFound struct{}

func (*NoTargetFound) Error() string {
	return "no reference target found"
}

type NoOriginFound struct{}

func (*NoOriginFound) Error() string {
	return "no reference origin found"
}
