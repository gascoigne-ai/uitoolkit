package ebitenrender

import (
	"go/build"
	"strings"
	"testing"
)

func TestEbitenRenderImportsOnlyRenderAndEbiten(t *testing.T) {
	pkg, err := build.ImportDir(".", 0)
	if err != nil {
		t.Fatalf("import .: %v", err)
	}
	allowed := func(path string) bool {
		first, _, _ := strings.Cut(path, "/")
		return !strings.Contains(first, ".") ||
			path == "github.com/gascoigne-ai/uitoolkit/render" ||
			path == "github.com/hajimehoshi/ebiten/v2" ||
			strings.HasPrefix(path, "github.com/hajimehoshi/ebiten/v2/") ||
			strings.HasPrefix(path, "golang.org/x/image/")
	}
	for _, path := range pkg.Imports {
		if !allowed(path) {
			t.Errorf("ebitenrender imports %s, want only the standard library, render, ebiten and x/image", path)
		}
	}
}
