package decoder

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/hcl-lang/lang"
)

type testPathReader struct {
	paths map[string]*PathContext
}

func (r *testPathReader) Paths(ctx context.Context) []lang.Path {
	paths := make([]lang.Path, len(r.paths))

	i := 0
	for path := range r.paths {
		paths[i] = lang.Path{Path: path}
		i++
	}

	return paths
}

func (r *testPathReader) PathContext(path lang.Path) (*PathContext, error) {
	if ctx, ok := r.paths[path.Path]; ok {
		return ctx, nil
	}

	return nil, fmt.Errorf("path not found: %q", path.Path)
}

func testPathDecoder(t *testing.T, pathCtx *PathContext) *PathDecoder {
	dirPath := t.TempDir()
	dirs := map[string]*PathContext{
		dirPath: pathCtx,
	}

	d := NewDecoder(&testPathReader{
		paths: dirs,
	})

	pathDecoder, err := d.Path(lang.Path{Path: dirPath})
	if err != nil {
		t.Fatal(err)
	}

	return pathDecoder
}
