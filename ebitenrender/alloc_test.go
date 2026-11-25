package ebitenrender_test

import (
	"testing"

	"github.com/gascoigne-ai/uitoolkit/ebitenrender"
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Every label of a frame is drawn and measured through the font source, which
// made a new face for each call. A caller that keeps its text/v2 face
// allocates none, and a font source must not either.
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

// A program that animates a text size must not keep every face it ever used.
func TestAFontSourceKeepsMeasuringRightPastItsFaceCache(t *testing.T) {
	source := newFont(t)
	for step := range 100 {
		size := 10 + float64(step)/4
		want, _ := text.Measure("Hg", textV2Face(t, size), 0)
		if got := source.Measure("Hg", size); got != want {
			t.Fatalf("size %v: Measure = %v, want %v", size, got, want)
		}
	}
}
