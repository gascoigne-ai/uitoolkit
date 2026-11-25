package core

import (
	"image/color"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
)

func TestRoundedFillIsOneMaskDraw(t *testing.T) {
	ctx := newTestCtx(t, &otherDevice{font: rendertest.FixedFont{AdvanceEm: 0.5, AscentEm: 0.8, DescentEm: 0.2}})
	ctx.Stats = &DrawStats{}
	canvas := rendertest.NewCanvas(20, 20)
	ctx.FillStyled(canvas, Rect{X: 2, Y: 2, W: 16, H: 16}, color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255}, 4)
	if ctx.Stats.MaskFills != 1 || len(canvas.Ops()) != 1 {
		t.Fatalf("MaskFills = %d with ops %v, want 1 and one op", ctx.Stats.MaskFills, canvas.Ops())
	}
}

func TestRoundedOutlineIsOneRingDraw(t *testing.T) {
	ctx := newTestCtx(t, &otherDevice{font: rendertest.FixedFont{AdvanceEm: 0.5, AscentEm: 0.8, DescentEm: 0.2}})
	ctx.Stats = &DrawStats{}
	canvas := rendertest.NewCanvas(20, 20)
	ctx.OutlineStyled(canvas, Rect{X: 2, Y: 2, W: 16, H: 16}, color.RGBA{255, 0, 0, 255}, 4)
	if ctx.Stats.RingFills != 1 || len(canvas.Ops()) != 1 {
		t.Fatalf("RingFills = %d with ops %v, want 1 and one op", ctx.Stats.RingFills, canvas.Ops())
	}
}

// A radius that rounds below one device pixel is a square outline, as before.
func TestOutlineStyledWithTinyRadiusIsTheSquareOutline(t *testing.T) {
	ctx := newTestCtx(t, &otherDevice{font: rendertest.FixedFont{AdvanceEm: 0.5, AscentEm: 0.8, DescentEm: 0.2}})
	ctx.Stats = &DrawStats{}
	canvas := rendertest.NewCanvas(20, 20)
	ctx.OutlineStyled(canvas, Rect{X: 2, Y: 2, W: 16, H: 16}, color.RGBA{255, 0, 0, 255}, 0.5)
	if ctx.Stats.RingFills != 0 {
		t.Fatalf("RingFills = %d, want 0", ctx.Stats.RingFills)
	}
	if len(canvas.Ops()) != 1 {
		t.Fatalf("ops = %v, want one", canvas.Ops())
	}
	if _, ok := canvas.Ops()[0].(rendertest.StrokeRect); !ok {
		t.Fatalf("op = %T, want a square outline", canvas.Ops()[0])
	}
}
