package markup

import (
	goparser "go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// No file of this package, a test
// included, imports Ebitengine or the adapter of it.
func TestMarkupImportsNoEbiten(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range files {
		parsed, err := goparser.ParseFile(token.NewFileSet(), name, nil, goparser.ImportsOnly)
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
