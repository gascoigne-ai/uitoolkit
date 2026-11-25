package core

import (
	"fmt"
	"image"
	"image/color"
	"reflect"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/ebitenrender"
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
	"github.com/hajimehoshi/ebiten/v2"
)

// otherDevice is the device of another renderer.
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

// A pixel cannot be read here, so the canvas the bridge made for the image is
// replaced by a recording one: what Fill sends it must be what it sends a
// canvas it is handed.
func TestFillOnAnEbitenImageAndOnACanvasAreTheSameOps(t *testing.T) {
	ctx := newTestCtx(t, ebitenrender.NewDevice())
	rect := Rect{X: 3, Y: 5, W: 17, H: 11}
	direct := rendertest.NewCanvas(40, 40)
	ctx.Fill(direct, rect, color.RGBA{255, 0, 0, 255})
	picture := ebiten.NewImage(40, 40)
	bridged := rendertest.NewCanvas(40, 40)
	ctx.bridge = ebitenBridge{dst: picture, canvas: bridged}
	ctx.Fill(picture, rect, color.RGBA{255, 0, 0, 255})
	if len(direct.Ops()) != 1 || !reflect.DeepEqual(direct.Ops(), bridged.Ops()) {
		t.Fatalf("canvas ops %v, image ops %v, want one equal op each", direct.Ops(), bridged.Ops())
	}
}

func TestCanvasOfDrawsOnTheImageItIsGiven(t *testing.T) {
	ctx := newTestCtx(t, ebitenrender.NewDevice())
	first, second := ebiten.NewImage(8, 8), ebiten.NewImage(8, 8)
	canvas := ctx.CanvasOf(first)
	if got, ok := ctx.EbitenImage(canvas); !ok || got != first {
		t.Fatalf("EbitenImage(CanvasOf(first)) = %p, %v, want %p", got, ok, first)
	}
	if ctx.CanvasOf(first) != canvas {
		t.Error("a second call on the same image made another canvas")
	}
	if got, ok := ctx.EbitenImage(ctx.CanvasOf(second)); !ok || got != second {
		t.Errorf("EbitenImage(CanvasOf(second)) = %p, %v, want %p", got, ok, second)
	}
}

func TestCanvasOfACanvasIsThatCanvasWithAnyDevice(t *testing.T) {
	ctx := newTestCtx(t, &otherDevice{})
	canvas := rendertest.NewCanvas(8, 8)
	if got := ctx.CanvasOf(canvas); got != render.Canvas(canvas) {
		t.Errorf("CanvasOf = %v, want the canvas itself", got)
	}
	if got, ok := ctx.EbitenImage(canvas); ok || got != nil {
		t.Errorf("EbitenImage of another renderer's canvas = %v, %v, want nil, false", got, ok)
	}
}

func TestCanvasOfAnEbitenImagePanicsWithAnotherDevice(t *testing.T) {
	ctx := newTestCtx(t, &otherDevice{})
	defer func() {
		const want = "core: drawing on an *ebiten.Image needs the device of ebitenrender, not *core.otherDevice"
		if got := fmt.Sprint(recover()); got != want {
			t.Errorf("panic = %q, want %q", got, want)
		}
	}()
	ctx.CanvasOf(ebiten.NewImage(8, 8))
}

// ClipCanvas covers the device pixels Clip does.
func TestClipCanvasCoversThePixelsOfClip(t *testing.T) {
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
			ctx := newTestCtx(t, ebitenrender.NewDevice())
			if err := ctx.SetScale(test.scale); err != nil {
				t.Fatal(err)
			}
			got := ctx.ClipCanvas(rendertest.NewCanvas(40, 40), test.rect).Bounds()
			if got != test.want && !(got.Empty() && test.want.Empty()) {
				t.Errorf("ClipCanvas bounds = %v, want %v", got, test.want)
			}
			old := ctx.Clip(ebiten.NewImage(40, 40), test.rect).Bounds()
			if old != got && !(old.Empty() && got.Empty()) {
				t.Errorf("Clip bounds = %v, ClipCanvas bounds = %v", old, got)
			}
		})
	}
}
