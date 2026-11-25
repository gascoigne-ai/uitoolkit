package core

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"reflect"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
)

// What each Ctx draw call asks of a canvas. The geometry is what the canvases
// of ebitenrender were compared with, pixel for pixel, while Ctx still drew
// with ebiten itself.
func TestDrawCallsAskTheCanvasForDevicePixels(t *testing.T) {
	red, glass := color.RGBA{255, 0, 0, 255}, color.RGBA{120, 40, 10, 160}
	rect := Rect{X: 3, Y: 5, W: 31, H: 19}
	thin := Rect{X: 3, Y: 5, W: 20, H: 1}
	clip := image.Rect(0, 0, 100, 100)
	for _, scale := range []float64{1, 1.5} {
		ctx := newTestCtx(t, &otherDevice{font: rendertest.FixedFont{AdvanceEm: 0.5, AscentEm: 0.8, DescentEm: 0.2}})
		if err := ctx.SetScale(scale); err != nil {
			t.Fatal(err)
		}
		snap := func(pixels int) float32 { return float32(math.Round(float64(pixels) * scale)) }
		scaled := func(value float32) float32 { return value * float32(scale) }
		snapped := func(of Rect) render.Box {
			return render.Box{X0: snap(of.X), Y0: snap(of.Y), X1: snap(of.X + of.W), Y1: snap(of.Y + of.H)}
		}
		unsnapped := render.Box{X0: scaled(3), Y0: scaled(5), X1: scaled(34), Y1: scaled(24)}
		lineWidth := ctx.LineWidth()
		tests := []struct {
			name string
			draw func(dst Surface)
			want rendertest.Op
		}{
			{"fill", func(dst Surface) { ctx.Fill(dst, rect, red) },
				rendertest.FillRect{Clip: clip, Box: snapped(rect), Color: red}},
			{"horizontal line", func(dst Surface) { ctx.HLine(dst, 3, 5, 17, red) },
				rendertest.FillRect{Clip: clip, Box: render.Box{X0: snap(3), Y0: snap(5), X1: snap(20), Y1: snap(5) + lineWidth}, Color: red}},
			{"vertical line", func(dst Surface) { ctx.VLine(dst, 3, 5, 17, red) },
				rendertest.FillRect{Clip: clip, Box: render.Box{X0: snap(3), Y0: snap(5), X1: snap(3) + lineWidth, Y1: snap(22)}, Color: red}},
			{"outline", func(dst Surface) { ctx.Outline(dst, rect, glass) },
				rendertest.StrokeRect{Clip: clip, Box: snapped(rect), Width: lineWidth, Color: glass}},
			{"line", func(dst Surface) { ctx.Line(dst, 2.5, 3, 27, 19.5, 1.5, glass) },
				rendertest.StrokeLine{Clip: clip, X0: scaled(2.5), Y0: scaled(3), X1: scaled(27), Y1: scaled(19.5), Width: scaled(1.5), Color: glass}},
			{"square flat style", func(dst Surface) { ctx.FillStyled(dst, rect, glass, glass, 0) },
				rendertest.FillRect{Clip: clip, Box: snapped(rect), Color: glass}},
			{"square gradient is not snapped", func(dst Surface) { ctx.FillStyled(dst, rect, red, glass, 0.5) },
				rendertest.FillRectGradient{Clip: clip, Box: unsnapped, Top: red, Bottom: glass}},
			{"rounded style", func(dst Surface) { ctx.FillStyled(dst, rect, red, glass, 5) },
				rendertest.FillRoundRect{Clip: clip, Box: snapped(rect), Radius: scaled(5), Corners: CornersAll, Top: red, Bottom: glass}},
			{"rounded corners", func(dst Surface) { ctx.FillStyledCorners(dst, rect, red, glass, 6, CornersTop) },
				rendertest.FillRoundRect{Clip: clip, Box: snapped(rect), Radius: scaled(6), Corners: CornersTop, Top: red, Bottom: glass}},
			{"no corners is square", func(dst Surface) { ctx.FillStyledCorners(dst, rect, glass, glass, 6, 0) },
				rendertest.FillRect{Clip: clip, Box: snapped(rect), Color: glass}},
			{"corners with a radius below a pixel are square", func(dst Surface) { ctx.FillStyledCorners(dst, rect, glass, glass, 0.5, CornersAll) },
				rendertest.FillRect{Clip: clip, Box: snapped(rect), Color: glass}},
			// A box too thin to round is the canvas's to square off.
			{"a thin rounded fill is still asked for rounded", func(dst Surface) { ctx.FillStyled(dst, thin, glass, glass, 4) },
				rendertest.FillRoundRect{Clip: clip, Box: snapped(thin), Radius: scaled(4), Corners: CornersAll, Top: glass, Bottom: glass}},
			{"rounded outline", func(dst Surface) { ctx.OutlineStyled(dst, rect, glass, 5) },
				rendertest.StrokeRoundRect{Clip: clip, Box: snapped(rect), Radius: scaled(5), Width: lineWidth, Color: glass}},
		}
		for _, test := range tests {
			t.Run(fmt.Sprintf("%s at %v", test.name, scale), func(t *testing.T) {
				canvas := rendertest.NewCanvas(100, 100)
				test.draw(canvas)
				if want := []rendertest.Op{test.want}; !reflect.DeepEqual(canvas.Ops(), want) {
					t.Errorf("ops = %+v, want %+v", canvas.Ops(), want)
				}
			})
		}
	}
}

// Text is the same glyphs twice, the second pass in the color scaled by
// StemBoost, with each tab drawn as a gap.
func TestTextDrawsTwoPassesTheSecondAtTheStemBoost(t *testing.T) {
	ctx := newTestCtx(t, &otherDevice{font: rendertest.FixedFont{AdvanceEm: 0.5, AscentEm: 0.8, DescentEm: 0.2}})
	if err := ctx.SetScale(1.5); err != nil {
		t.Fatal(err)
	}
	canvas := rendertest.NewCanvas(100, 100)
	ctx.Text(canvas, ctx.F.UI, "a\tb", 3, 5, color.RGBA{200, 100, 50, 255})
	pass := rendertest.DrawText{Clip: canvas.Bounds(), Face: ctx.F.UI, Text: "a    b", Left: 4.5, Top: 7.5}
	full, boost := pass, pass
	full.Color, boost.Color = color.RGBA{200, 100, 50, 255}, color.RGBA{100, 50, 25, 128}
	if want := []rendertest.Op{full, boost}; !reflect.DeepEqual(canvas.Ops(), want) {
		t.Errorf("ops = %+v, want %+v", canvas.Ops(), want)
	}
}
