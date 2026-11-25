package widget

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// The widget package draws through render.Canvas alone: no file of it, a test
// included, imports Ebitengine or the adapter of it.
func TestWidgetImportsNoEbiten(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range files {
		parsed, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range parsed.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(path, "hajimehoshi") || strings.HasPrefix(path, "github.com/gascoigne-ai/uitoolkit/renderer/") {
				t.Errorf("%s imports %s", name, path)
			}
		}
	}
}
