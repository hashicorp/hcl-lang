package reference

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

func TraversalsToLocalOrigins(traversals []hcl.Traversal, tes schema.TraversalExprs, allowSelfRefs bool) Origins {
	origins := make(Origins, 0)
	for _, traversal := range traversals {
		// traversal should not be relative here, since the input to this
		// function `expr.Variables()` only returns absolute traversals
		if !traversal.IsRelative() && traversal.RootName() == "self" && !allowSelfRefs {
			// Only if a block allows the usage of self.* we create a origin,
			// else just continue
			continue
		}
		origin, err := TraversalToLocalOrigin(traversal, tes)
		if err != nil {
			continue
		}
		origins = append(origins, origin)
	}

	return origins
}

func TraversalToLocalOrigin(traversal hcl.Traversal, tes schema.TraversalExprs) (LocalOrigin, error) {
	addr, err := lang.TraversalToAddress(traversal)
	if err != nil {
		return LocalOrigin{}, err
	}

	return LocalOrigin{
		Addr:        addr,
		Range:       traversal.SourceRange(),
		Constraints: traversalExpressionsToOriginConstraints(tes),
	}, nil
}

func traversalExpressionsToOriginConstraints(tes []schema.TraversalExpr) OriginConstraints {
	if len(tes) == 0 {
		return nil
	}

	roc := make(OriginConstraints, 0)
	for _, te := range tes {
		if te.Address != nil {
			// skip traversals which are targets by themselves (not origins)
			continue
		}
		roc = append(roc, OriginConstraint{
			OfType:    te.OfType,
			OfScopeId: te.OfScopeId,
		})
	}
	return roc
}
