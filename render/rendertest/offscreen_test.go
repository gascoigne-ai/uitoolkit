package rendertest_test

import (
	"image"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
)

func TestReleasedOffscreenPanicsOnUse(t *testing.T) {
	var path render.Path
	box := render.Box{X1: 1, Y1: 1}
	picture := &rendertest.Image{}
	// sub was taken while the offscreen was live: it shares the surface.
	cases := []struct {
		name string
		use  func(offscreen *rendertest.Offscreen, sub render.Canvas)
	}{
		{"FillRect", func(offscreen *rendertest.Offscreen, _ render.Canvas) { offscreen.FillRect(box, red) }},
		{"FillRectGradient", func(offscreen *rendertest.Offscreen, _ render.Canvas) { offscreen.FillRectGradient(box, red, blue) }},
		{"FillRoundRect", func(offscreen *rendertest.Offscreen, _ render.Canvas) {
			offscreen.FillRoundRect(box, 1, render.CornersAll, red, blue)
		}},
		{"StrokeRect", func(offscreen *rendertest.Offscreen, _ render.Canvas) { offscreen.StrokeRect(box, 1, red) }},
		{"StrokeRoundRect", func(offscreen *rendertest.Offscreen, _ render.Canvas) { offscreen.StrokeRoundRect(box, 1, 1, red) }},
		{"StrokeLine", func(offscreen *rendertest.Offscreen, _ render.Canvas) { offscreen.StrokeLine(0, 0, 1, 1, 1, red) }},
		{"FillPath", func(offscreen *rendertest.Offscreen, _ render.Canvas) { offscreen.FillPath(&path, red) }},
		{"StrokePath", func(offscreen *rendertest.Offscreen, _ render.Canvas) {
			offscreen.StrokePath(&path, render.Stroke{}, red)
		}},
		{"DrawImage", func(offscreen *rendertest.Offscreen, _ render.Canvas) {
			offscreen.DrawImage(picture, 0, 0, 1, render.FilterNearest)
		}},
		{"DrawMask", func(offscreen *rendertest.Offscreen, _ render.Canvas) { offscreen.DrawMask(picture, 0, 0, red) }},
		{"DrawText", func(offscreen *rendertest.Offscreen, _ render.Canvas) {
			offscreen.DrawText(render.Face{}, "a", 0, 0, red)
		}},
		{"Clear", func(offscreen *rendertest.Offscreen, _ render.Canvas) { offscreen.Clear() }},
		{"Release", func(offscreen *rendertest.Offscreen, _ render.Canvas) { offscreen.Release() }},
		{"ReadPixels", func(offscreen *rendertest.Offscreen, _ render.Canvas) { offscreen.ReadPixels() }},
		{"Sub", func(offscreen *rendertest.Offscreen, _ render.Canvas) { offscreen.Sub(image.Rect(0, 0, 1, 1)) }},
		{"draw on an earlier Sub", func(_ *rendertest.Offscreen, sub render.Canvas) { sub.FillRect(box, red) }},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			offscreen := rendertest.NewOffscreen(4, 4)
			sub := offscreen.Sub(offscreen.Bounds())
			offscreen.Release()
			defer func() {
				if recover() == nil {
					t.Errorf("%s after Release did not panic", testCase.name)
				}
			}()
			testCase.use(offscreen, sub)
		})
	}
}

func TestClearIsRecordedInPaintOrder(t *testing.T) {
	offscreen := rendertest.NewOffscreen(4, 3)
	offscreen.FillRect(render.Box{X1: 1, Y1: 1}, red)
	offscreen.Clear()
	whole := image.Rect(0, 0, 4, 3)
	assertOps(t, offscreen.Ops(), []rendertest.Op{
		rendertest.FillRect{Clip: whole, Box: render.Box{X1: 1, Y1: 1}, Color: red},
		rendertest.Clear{Clip: whole},
	})
}

func TestReadPixelsMatchesBounds(t *testing.T) {
	offscreen := rendertest.NewOffscreen(4, 3)
	pixels := offscreen.ReadPixels()
	if want := image.Rect(0, 0, 4, 3); pixels.Rect != want || offscreen.Bounds() != want {
		t.Fatalf("ReadPixels().Rect = %v, Bounds() = %v, want both %v", pixels.Rect, offscreen.Bounds(), want)
	}
	for y := range 3 {
		for x := range 4 {
			if got := pixels.RGBAAt(x, y); got.A != 0 {
				t.Errorf("pixel (%d,%d) = %v, want transparent", x, y, got)
			}
		}
	}
}
