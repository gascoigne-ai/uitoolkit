package ebitenrender_test

import (
	"fmt"
	"image/color"
	"math"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/theming"
	"github.com/hajimehoshi/ebiten/v2"
)

// The parity tests draw one shape twice in this process, once with core.Ctx
// and once through a canvas given the device pixels core computes, and require
// the same bytes. There are no golden files: pixels differ between GPUs.

// The scales every parity test runs at: whole and fractional.
var parityScales = []float64{1, 1.5}

// device turns logical values into the device pixels core draws at.
type device struct{ ctx *core.Ctx }

// snap is core.Ctx.snap: a logical edge on a whole device pixel.
func (d device) snap(pixels int) float32 { return float32(math.Round(float64(pixels) * d.ctx.Scale)) }

// scaled is core.Ctx.dev: a logical value in device pixels, not snapped. The
// multiply is in float32, as core's is.
func (d device) scaled(value float32) float32 { return value * float32(d.ctx.Scale) }

func (d device) snapped(rect core.Rect) render.Box {
	return render.Box{X0: d.snap(rect.X), Y0: d.snap(rect.Y), X1: d.snap(rect.X + rect.W), Y1: d.snap(rect.Y + rect.H)}
}

func (d device) unsnapped(rect core.Rect) render.Box {
	return render.Box{
		X0: d.scaled(float32(rect.X)), Y0: d.scaled(float32(rect.Y)),
		X1: d.scaled(float32(rect.X + rect.W)), Y1: d.scaled(float32(rect.Y + rect.H)),
	}
}

// parity runs one drawing at every parity scale.
func parity(t *testing.T, name string, withCore func(ctx *core.Ctx, dst *ebiten.Image), through func(canvas render.Canvas, at device)) {
	t.Helper()
	for _, scale := range parityScales {
		t.Run(fmt.Sprintf("%s at %v", name, scale), func(t *testing.T) {
			ctx, err := core.NewCtx(theming.Dark())
			if err != nil {
				t.Fatal(err)
			}
			if err := ctx.SetScale(scale); err != nil {
				t.Fatal(err)
			}
			assertMatchesEbiten(t,
				func(dst *ebiten.Image) { withCore(ctx, dst) },
				func(canvas render.Canvas) { through(canvas, device{ctx}) })
		})
	}
}

func TestFillRectMatchesCoreFill(t *testing.T) {
	rect := core.Rect{X: 3, Y: 5, W: 17, H: 11}
	for _, clr := range []color.RGBA{red, glass} {
		parity(t, "fill",
			func(ctx *core.Ctx, dst *ebiten.Image) { ctx.Fill(dst, rect, clr) },
			func(canvas render.Canvas, at device) { canvas.FillRect(at.snapped(rect), clr) })
		parity(t, "horizontal line",
			func(ctx *core.Ctx, dst *ebiten.Image) { ctx.HLine(dst, 3, 5, 17, clr) },
			func(canvas render.Canvas, at device) {
				top := at.snap(5)
				canvas.FillRect(render.Box{X0: at.snap(3), Y0: top, X1: at.snap(3 + 17), Y1: top + at.ctx.LineWidth()}, clr)
			})
		parity(t, "vertical line",
			func(ctx *core.Ctx, dst *ebiten.Image) { ctx.VLine(dst, 3, 5, 17, clr) },
			func(canvas render.Canvas, at device) {
				left := at.snap(3)
				canvas.FillRect(render.Box{X0: left, Y0: at.snap(5), X1: left + at.ctx.LineWidth(), Y1: at.snap(5 + 17)}, clr)
			})
	}
}

func TestFillRectGradientMatchesCoreSquareGradient(t *testing.T) {
	// Even edges are whole device pixels at 1.5 too. On other boxes core
	// fills unsnapped edges without anti-aliasing and a canvas smooths them.
	rect := core.Rect{X: 2, Y: 4, W: 30, H: 22}
	for _, colors := range [][2]color.RGBA{{red, deepGlass}, {glass, deepGlass}} {
		parity(t, "gradient",
			func(ctx *core.Ctx, dst *ebiten.Image) { ctx.FillStyled(dst, rect, colors[0], colors[1], 0) },
			func(canvas render.Canvas, at device) {
				canvas.FillRectGradient(at.unsnapped(rect), colors[0], colors[1])
			})
	}
}

func TestFillRoundRectMatchesCoreFillStyled(t *testing.T) {
	rect := core.Rect{X: 3, Y: 5, W: 31, H: 19}
	fills := []struct {
		name        string
		top, bottom color.RGBA
		radius      float64
	}{
		{"flat", glass, glass, 4},
		{"gradient", red, deepGlass, 4},
		{"half a pixel radius rounds up", glass, deepGlass, 5},
		{"radius clamps to the height", glass, deepGlass, 40},
		{"radius of one", red, red, 1},
	}
	for _, fill := range fills {
		parity(t, fill.name,
			func(ctx *core.Ctx, dst *ebiten.Image) { ctx.FillStyled(dst, rect, fill.top, fill.bottom, fill.radius) },
			func(canvas render.Canvas, at device) {
				canvas.FillRoundRect(at.snapped(rect), at.scaled(float32(fill.radius)), render.CornersAll, fill.top, fill.bottom)
			})
	}
	corners := []struct {
		name    string
		core    core.Corners
		through render.Corners
	}{
		{"top corners", core.CornersTop, render.CornersTop},
		{"bottom left corner", core.CornerBL, render.CornerBL},
	}
	for _, set := range corners {
		parity(t, set.name,
			func(ctx *core.Ctx, dst *ebiten.Image) { ctx.FillStyledCorners(dst, rect, red, deepGlass, 6, set.core) },
			func(canvas render.Canvas, at device) {
				canvas.FillRoundRect(at.snapped(rect), at.scaled(6), set.through, red, deepGlass)
			})
	}
}

func TestFillRoundRectTooSmallFallsBackLikeCore(t *testing.T) {
	// Even edges are whole device pixels at 1.5 too, where core's square
	// gradient does not snap.
	square := core.Rect{X: 2, Y: 4, W: 30, H: 22}
	thin := core.Rect{X: 2, Y: 3, W: 20, H: 1}
	fills := []struct {
		name        string
		rect        core.Rect
		top, bottom color.RGBA
		radius      float64
		corners     core.Corners
	}{
		{"one pixel high, flat", thin, glass, glass, 4, core.CornersAll},
		{"one pixel high, gradient", thin, red, deepGlass, 4, core.CornersAll},
		{"no radius, flat", square, glass, glass, 0, core.CornersAll},
		{"no radius, gradient", square, red, deepGlass, 0, core.CornersAll},
		{"radius below a pixel, flat", square, glass, glass, 0.5, core.CornersAll},
		{"radius below a pixel, gradient", square, red, deepGlass, 0.5, core.CornersAll},
		{"no corners, flat", square, glass, glass, 4, 0},
		{"no corners, gradient", square, red, deepGlass, 4, 0},
	}
	for _, fill := range fills {
		parity(t, fill.name,
			func(ctx *core.Ctx, dst *ebiten.Image) {
				ctx.FillStyledCorners(dst, fill.rect, fill.top, fill.bottom, fill.radius, fill.corners)
			},
			func(canvas render.Canvas, at device) {
				canvas.FillRoundRect(at.snapped(fill.rect), at.scaled(float32(fill.radius)), render.Corners(fill.corners), fill.top, fill.bottom)
			})
	}
}

func TestStrokeRectMatchesCoreOutline(t *testing.T) {
	for _, rect := range []core.Rect{{X: 3, Y: 5, W: 31, H: 19}, {X: 3, Y: 5, W: 20, H: 1}} {
		parity(t, fmt.Sprintf("outline %dx%d", rect.W, rect.H),
			func(ctx *core.Ctx, dst *ebiten.Image) { ctx.Outline(dst, rect, glass) },
			func(canvas render.Canvas, at device) { canvas.StrokeRect(at.snapped(rect), at.ctx.LineWidth(), glass) })
	}
}

func TestStrokeRoundRectMatchesCoreOutlineStyled(t *testing.T) {
	outlines := []struct {
		name   string
		rect   core.Rect
		radius float64
	}{
		{"rounded", core.Rect{X: 3, Y: 5, W: 31, H: 19}, 4},
		{"half a pixel radius rounds up", core.Rect{X: 3, Y: 5, W: 31, H: 19}, 5},
		{"radius clamps to the height", core.Rect{X: 3, Y: 5, W: 31, H: 19}, 40},
		{"line clamps to the radius", core.Rect{X: 3, Y: 5, W: 31, H: 3}, 4},
		{"one pixel high is square", core.Rect{X: 3, Y: 5, W: 20, H: 1}, 4},
		{"no radius is square", core.Rect{X: 3, Y: 5, W: 31, H: 19}, 0},
		{"radius below a pixel is square", core.Rect{X: 3, Y: 5, W: 31, H: 19}, 0.5},
	}
	for _, outline := range outlines {
		parity(t, outline.name,
			func(ctx *core.Ctx, dst *ebiten.Image) { ctx.OutlineStyled(dst, outline.rect, glass, outline.radius) },
			func(canvas render.Canvas, at device) {
				canvas.StrokeRoundRect(at.snapped(outline.rect), at.scaled(float32(outline.radius)), at.ctx.LineWidth(), glass)
			})
	}
}

func TestStrokeLineMatchesCoreLine(t *testing.T) {
	lines := []struct {
		name                  string
		x0, y0, x1, y1, width float32
	}{
		{"diagonal", 2.5, 3, 27, 19.5, 1.5},
		{"horizontal", 4, 10, 30, 10, 1},
	}
	for _, line := range lines {
		parity(t, line.name,
			func(ctx *core.Ctx, dst *ebiten.Image) {
				ctx.Line(dst, line.x0, line.y0, line.x1, line.y1, line.width, glass)
			},
			func(canvas render.Canvas, at device) {
				canvas.StrokeLine(at.scaled(line.x0), at.scaled(line.y0), at.scaled(line.x1), at.scaled(line.y1), at.scaled(line.width), glass)
			})
	}
}
