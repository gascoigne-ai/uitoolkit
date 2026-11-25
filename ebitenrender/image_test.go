package ebitenrender_test

import (
	"testing"

	"github.com/gascoigne-ai/uitoolkit/ebitenrender"
	"github.com/hajimehoshi/ebiten/v2"
)

// A bare *ebiten.Image has Bounds, so it satisfies render.Image and compiles
// where one is wanted, but it is not an image this renderer made: drawing it
// panics instead of silently drawing something.
func TestACanvasPanicsOnABareEbitenImage(t *testing.T) {
	off := ebitenrender.NewDevice().NewOffscreen(8, 8)
	defer off.Release()
	defer func() {
		const want = "render: image from another renderer"
		if got := recover(); got != want {
			t.Errorf("panic = %v, want %q", got, want)
		}
	}()
	off.DrawImage(ebiten.NewImage(4, 4), 2, 2, 1, 0)
}
