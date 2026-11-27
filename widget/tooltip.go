package widget

import (
	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
	. "github.com/gascoigne-ai/uitoolkit/view"
)

// TooltipState is the delay bookkeeping of the tooltip layer. The window
// owns it, so it can keep frames coming while a tooltip waits to show.
type TooltipState struct {
	shown, waiting bool
	text           string // the text the delay is counted for
	since          int    // tick at which the pointer came to rest on it
}

// Shown reports whether the last frame drew a tooltip.
func (s *TooltipState) Shown() bool { return s.shown }

// Waiting reports whether the pointer rests on an element with a tooltip
// that is not due yet: the window must draw again to show it.
func (s *TooltipState) Waiting() bool { return s.waiting }

// TooltipLayer draws the tooltip of the element under the pointer (set
// with view.Tooltip) over everything drawn before it, kept inside the
// window. Mount it last in the window's Stack.
func TooltipLayer(state *TooltipState) *El { return Custom(tooltipB{state: state}) }

type tooltipB struct {
	Base
	state *TooltipState
}

func (b tooltipB) Draw(element *El, ctx *core.Ctx, dst render.Canvas) {
	b.state.draw(ctx, dst, element.Rect())
}

// draw updates the delay bookkeeping for the tip of the element under the
// pointer and, once it is due, draws it inside screen.
func (s *TooltipState) draw(ctx *core.Ctx, dst render.Canvas, screen core.Rect) {
	s.shown, s.waiting = false, false
	if ctx.TipText == "" {
		s.text = ""
		return
	}
	if ctx.TipText != s.text || ctx.In.Click {
		s.text, s.since = ctx.TipText, ctx.Tick // a new element, or a click: wait again
	}
	// The delay is read each frame: a new theme moves a waiting tip too.
	if ctx.Tick-s.since < ctx.In.Ticks(ctx.Feel().TooltipDelay) {
		s.waiting = true
		return
	}
	s.shown = true
	padX, offset := ctx.Sizes.TipPadX, ctx.Sizes.TipOffset
	w := ctx.Width(ctx.F.Small, ctx.TipText) + 2*padX
	h := ctx.LineH(ctx.F.Small) + 2*ctx.Sizes.TipPadY
	x := max(screen.X, min(ctx.In.X+offset, screen.X+screen.W-w))
	y := ctx.In.Y + offset
	if y+h > screen.Y+screen.H {
		y = ctx.In.Y - offset - h // no room below: above the pointer
	}
	box := core.Rect{X: x, Y: max(screen.Y, y), W: w, H: h}
	st := ctx.Th.Style("tooltip")
	ctx.FillStyled(dst, box, ctx.Color(core.Frame), ctx.Color(core.Frame), st.Radius)
	border := st.Border
	if border == "" {
		border = core.Border
	}
	ctx.OutlineStyled(dst, box, ctx.Color(border), st.Radius)
	ctx.TextMid(dst, ctx.F.Small, ctx.TipText, box.X+padX, box, ctx.Color(core.Text))
}
