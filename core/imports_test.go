package core

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Only the temporary bridge file may name the GPU library or its renderer, so
// that removing that one file frees core of both.
func TestOnlyTheBridgeImportsEbiten(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") || name == "ebitenbridge.go" {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range parsed.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(path, "hajimehoshi") || strings.HasSuffix(path, "/ebitenrender") {
				t.Errorf("%s imports %s; only ebitenbridge.go may", name, path)
			}
		}
	}
}
