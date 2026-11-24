package render_test

import (
	"go/build"
	"slices"
	"strings"
	"testing"
)

func TestRenderImportsOnlyStandardLibrary(t *testing.T) {
	cases := []struct {
		name    string
		dir     string
		allowed []string
	}{
		{"render", ".", nil},
		{"rendertest", "rendertest", []string{"github.com/gascoigne-ai/uitoolkit/render"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			pkg, err := build.ImportDir(testCase.dir, 0)
			if err != nil {
				t.Fatalf("import %s: %v", testCase.dir, err)
			}
			for _, path := range pkg.Imports {
				first, _, _ := strings.Cut(path, "/")
				if strings.Contains(first, ".") && !slices.Contains(testCase.allowed, path) {
					t.Errorf("%s imports %s, which is not in the standard library", testCase.name, path)
				}
			}
		})
	}
}
