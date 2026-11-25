package core

import (
	"image"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
	"github.com/gascoigne-ai/uitoolkit/render/text"
)

// testDevice is a device with no GPU: faces measured by render/text, and
// offscreens and images that only record.
type testDevice struct{}

func (testDevice) NewFontSource(data []byte) (render.FontSource, error) {
	return text.NewFontSource(data)
}

func (testDevice) NewOffscreen(width, height int) render.Offscreen {
	return rendertest.NewOffscreen(width, height)
}

func (testDevice) NewImage(source image.Image) render.Image {
	return &rendertest.Image{Rect: source.Bounds()}
}

// otherDevice is the device of another renderer: its faces are the fixed
// ones of rendertest.
type otherDevice struct{ font rendertest.FixedFont }

func (d *otherDevice) NewFontSource([]byte) (render.FontSource, error) { return &d.font, nil }

func (*otherDevice) NewOffscreen(width, height int) render.Offscreen {
	return rendertest.NewOffscreen(width, height)
}

func (*otherDevice) NewImage(source image.Image) render.Image {
	return &rendertest.Image{Rect: source.Bounds()}
}

func newTestCtx(t *testing.T, device render.Device) *Ctx {
	t.Helper()
	ctx, err := NewCtx(testTheme(), device)
	if err != nil {
		t.Fatal(err)
	}
	return ctx
}

// Clip covers the device pixels of the rect it is given, rounded out at the
// far edge and cut to the canvas.
func TestClipBoundsAreDevicePixelsCutToTheCanvas(t *testing.T) {
	tests := []struct {
		name  string
		scale float64
		rect  Rect
		want  image.Rectangle
	}{
		{"inside", 1, Rect{X: 3, Y: 5, W: 10, H: 7}, image.Rect(3, 5, 13, 12)},
		{"the near edge is cut and the far edge rounded", 1.25, Rect{X: 3, Y: 5, W: 10, H: 7}, image.Rect(3, 6, 16, 15)},
		{"past the surface", 1, Rect{X: 30, Y: 30, W: 50, H: 50}, image.Rect(30, 30, 40, 40)},
		{"before the surface", 1, Rect{X: -5, Y: -5, W: 10, H: 10}, image.Rect(0, 0, 5, 5)},
		{"outside", 1, Rect{X: 50, Y: 50, W: 5, H: 5}, image.Rectangle{}},
		{"negative size", 1, Rect{X: 20, Y: 20, W: -5, H: -5}, image.Rectangle{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := newTestCtx(t, testDevice{})
			if err := ctx.SetScale(test.scale); err != nil {
				t.Fatal(err)
			}
			got := ctx.Clip(rendertest.NewCanvas(40, 40), test.rect).Bounds()
			if got != test.want && !(got.Empty() && test.want.Empty()) {
				t.Errorf("Clip bounds = %v, want %v", got, test.want)
			}
		})
	}
}
