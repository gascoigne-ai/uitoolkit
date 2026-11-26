//go:build !race

package ebitenrender_test

import (
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/renderer/ebiten"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Every label of a frame is drawn and measured through the font source, which
// made a new face for each call. A caller that keeps its text/v2 face
// allocates none, and a font source must not either.
//
// It runs without the race detector, which makes sync.Pool drop items at
// random: text/v2 pools its draw buffers, so a call's allocation count then
// varies from run to run. The Makefile gives it a pass of its own.
func TestTextCallsAllocateNoMoreThanTextV2WithAKeptFace(t *testing.T) {
	source := newFont(t)
	face := render.Face{Source: source, Size: 13}
	kept := textV2Face(t, 13)
	off := ebitenrender.NewDevice().NewOffscreen(surface, surface)
	defer off.Release()
	direct := ebiten.NewImage(surface, surface)
	calls := []struct {
		name           string
		through, plain func()
	}{
		{"measure",
			func() { face.Measure("Send request") },
			func() { text.Measure("Send request", kept, 0) }},
		{"draw",
			func() { off.DrawText(face, "Send request", 1, 2, red) },
			func() {
				options := &text.DrawOptions{}
				options.GeoM.Translate(1, 2)
				options.ColorScale.ScaleWithColor(red)
				text.Draw(direct, "Send request", kept, options)
			}},
	}
	for _, call := range calls {
		t.Run(call.name, func(t *testing.T) {
			got, want := testing.AllocsPerRun(100, call.through), testing.AllocsPerRun(100, call.plain)
			if got > want {
				t.Errorf("%v allocations a call, text/v2 with a kept face makes %v", got, want)
			}
		})
	}
}
