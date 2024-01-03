// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reference

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
)

func TraversalToLocalOrigin(traversal hcl.Traversal, cons OriginConstraints, allowSelfRefs bool) (LocalOrigin, bool) {
	// traversal should not be relative here, since the input to this
	// function `expr.Variables()` only returns absolute traversals
	if !traversal.IsRelative() && traversal.RootName() == "self" && !allowSelfRefs {
		// Only if a block allows the usage of self.* we create a origin,
		// else just continue
		return LocalOrigin{}, false
	}

	addr, err := lang.TraversalToAddress(traversal)
	if err != nil {
		return LocalOrigin{}, false
	}

	return LocalOrigin{
		Addr:        addr,
		Range:       traversal.SourceRange(),
		Constraints: cons,
	}, true
}
