package widget

import (
	"image"
	"image/color"
	"strconv"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/theming"
)

// The CodeView benchmarks time the cost of a large response body (the view
// limit is 2 MiB, about 40k lines of pretty-printed JSON) on the frames that
// change nothing: an Update with no input and a Draw of the same rows.

const benchLines = 40000

func benchCodeView(b *testing.B, wrap bool) (*CodeView, *core.Ctx, core.Rect) {
	b.Helper()
	ctx, err := core.NewCtx(theming.Dark(), testDevice{})
	if err != nil {
		b.Fatal(err)
	}
	lines := make([]Line, benchLines)
	for i := range lines {
		lines[i] = Line{Spans: []Span{
			{Text: "    \"field" + strconv.Itoa(i) + "\": "},
			{Text: "\"a value of some length that wraps in a narrow view " + strconv.Itoa(i) + "\","},
		}}
	}
	view := &CodeView{Wrap: wrap}
	view.SetLines(lines)
	rect := core.Rect{W: 700, H: 500}
	view.Update(ctx, rect)
	view.ScrollY = 200000
	view.Update(ctx, rect)
	b.ReportAllocs()
	b.ResetTimer()
	return view, ctx, rect
}

func BenchmarkCodeViewUpdate(b *testing.B) {
	for _, wrap := range []bool{false, true} {
		b.Run("wrap="+strconv.FormatBool(wrap), func(b *testing.B) {
			view, ctx, rect := benchCodeView(b, wrap)
			for i := 0; i < b.N; i++ {
				view.Update(ctx, rect)
			}
		})
	}
}

func BenchmarkCodeViewDraw(b *testing.B) {
	for _, wrap := range []bool{false, true} {
		b.Run("wrap="+strconv.FormatBool(wrap), func(b *testing.B) {
			view, ctx, rect := benchCodeView(b, wrap)
			screen := nopCanvas{bounds: image.Rect(0, 0, rect.W, rect.H)}
			for i := 0; i < b.N; i++ {
				view.Draw(ctx, screen, rect)
			}
		})
	}
}

// nopCanvas draws nothing and records nothing, so a benchmark times the
// toolkit and not a recorder.
type nopCanvas struct{ bounds image.Rectangle }

func (c nopCanvas) Bounds() image.Rectangle { return c.bounds }
func (c nopCanvas) Sub(clip image.Rectangle) render.Canvas {
	return nopCanvas{bounds: c.bounds.Intersect(clip)}
}
func (nopCanvas) FillRect(render.Box, color.RGBA)                                           {}
func (nopCanvas) FillRectGradient(render.Box, color.RGBA, color.RGBA)                       {}
func (nopCanvas) FillRoundRect(render.Box, float32, render.Corners, color.RGBA, color.RGBA) {}
func (nopCanvas) StrokeRect(render.Box, float32, color.RGBA)                                {}
func (nopCanvas) StrokeRoundRect(render.Box, float32, float32, color.RGBA)                  {}
func (nopCanvas) StrokeLine(float32, float32, float32, float32, float32, color.RGBA)        {}
func (nopCanvas) FillPath(*render.Path, color.RGBA)                                         {}
func (nopCanvas) StrokePath(*render.Path, render.Stroke, color.RGBA)                        {}
func (nopCanvas) DrawImage(render.Image, float64, float64, float64, render.Filter)          {}
func (nopCanvas) DrawMask(render.Image, float64, float64, color.RGBA)                       {}
func (nopCanvas) DrawText(render.Face, string, float64, float64, color.RGBA)                {}
