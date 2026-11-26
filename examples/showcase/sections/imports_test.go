package sections

import (
	"os/exec"
	"strings"
	"testing"
)

// The sections run without a display, so nothing they build with may pull in
// Ebitengine or a renderer.
func TestSectionsImportNoEbiten(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "-test", ".").Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, dependency := range strings.Fields(string(out)) {
		if strings.Contains(dependency, "hajimehoshi") || strings.Contains(dependency, "uitoolkit/renderer/") {
			t.Errorf("sections depend on %s", dependency)
		}
	}
}
