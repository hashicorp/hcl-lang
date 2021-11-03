package reference

import (
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl-lang/schema"
	"github.com/hashicorp/hcl/v2"
)

func TraversalsToLocalOrigins(traversals []hcl.Traversal, tes schema.TraversalExprs) Origins {
	origins := make(Origins, 0)
	for _, traversal := range traversals {
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
