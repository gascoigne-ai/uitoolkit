package ebitenrender_test

import (
	"image"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/ebitenrender"
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestImageOfIsTheEbitenImageACanvasDrawsOn(t *testing.T) {
	device := ebitenrender.NewDevice()
	picture := ebiten.NewImage(8, 8)
	if got, ok := ebitenrender.ImageOf(device.NewCanvas(picture)); !ok || got != picture {
		t.Errorf("ImageOf(canvas) = %p, %v, want %p, true", got, ok, picture)
	}
	clip := image.Rect(1, 2, 5, 6)
	if got, ok := ebitenrender.ImageOf(device.NewCanvas(picture).Sub(clip)); !ok || got.Bounds() != clip {
		t.Errorf("ImageOf(sub) = %v, %v, want an image with bounds %v", got, ok, clip)
	}
	off := device.NewOffscreen(4, 3)
	defer off.Release()
	if got, ok := ebitenrender.ImageOf(off); !ok || got.Bounds() != off.Bounds() {
		t.Errorf("ImageOf(offscreen) = %v, %v, want an image with bounds %v", got, ok, off.Bounds())
	}
	if got, ok := ebitenrender.ImageOf(rendertest.NewCanvas(4, 4)); ok || got != nil {
		t.Errorf("ImageOf(a canvas of another renderer) = %v, %v, want nil, false", got, ok)
	}
}

// While the toolkit still hands ebiten images around, a canvas draws one as a
// picture.
func TestACanvasDrawsABareEbitenImage(t *testing.T) {
	picture := ebiten.NewImage(4, 4)
	picture.Fill(red)
	off := ebitenrender.NewDevice().NewOffscreen(8, 8)
	defer off.Release()
	off.DrawImage(picture, 2, 2, 1, render.FilterNearest)
	if got := off.ReadPixels().RGBAAt(3, 3); got != red {
		t.Errorf("pixel = %v, want %v", got, red)
	}
}
