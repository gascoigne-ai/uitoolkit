package ebitenrender_test

import (
	"runtime/debug"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/renderer/ebiten"
)

// The report names the renderer and the version of the library it was built
// with, so the system report needs no knowledge of Ebitengine.
func TestReadGraphicsNamesEbitengineAndItsVersion(t *testing.T) {
	g := ebitenrender.ReadGraphics()
	if g.Renderer != "Ebitengine" {
		t.Errorf("Renderer = %q, want Ebitengine", g.Renderer)
	}
	want := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/hajimehoshi/ebiten/v2" {
				want = dep.Version
			}
		}
	}
	if g.RendererVersion != want {
		t.Errorf("RendererVersion = %q, want %q", g.RendererVersion, want)
	}
	if g.Library == "" {
		t.Error("Library is empty, want the graphics library of the run loop")
	}
}
