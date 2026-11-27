package widget

import (
	"github.com/gascoigne-ai/uitoolkit/core"
	. "github.com/gascoigne-ai/uitoolkit/view"
)

// ResizeState is the size the user gave a Resizable element; 0 means the
// element's default. The application owns it, like a split or a scroll
// state, so it survives the per-frame rebuild and can be saved.
type ResizeState struct {
	W, H int

	active                         bool
	left, right, top, bottom       bool // the edges being dragged
	startX, startY, startW, startH int
}

type resizableB struct {
	Linear     // lays its children out top to bottom, like Col
	state      *ResizeState
	minW, minH int
	maxW, maxH int
}

// Resizable lets the user drag the edges and corners of the element it
// wraps, which is laid out and centered by its parent (a floating dialog):
// the size changes by twice the pointer's movement, so the edge stays under
// the pointer. The size is kept in state within the given minimum and
// maximum, and state.H also within the element's own MaxH, so it never runs
// past the height that is drawn; the caller applies state.W and state.H to
// the element's own size, and a double click on an edge puts both back to 0, the default. It has
// no element of its own in templates, because only a dialog uses it:
// <resizable-dialog> is the template form.
func Resizable(state *ResizeState, minW, minH, maxW, maxH int, items ...Item) *El {
	return Custom(&resizableB{state: state, minW: minW, minH: minH, maxW: maxW, maxH: maxH}, append([]Item{DismissMarginOf(core.EdgeReachOf)}, items...)...)
}

// edges reports which edges of r the point is on, if it is on the border:
// grab pixels inside r and reach pixels (Sizes.EdgeReach) outside it, where
// a modal element does not dismiss either (view.DismissMarginOf).
func edges(rect core.Rect, posX, posY, grab, reach int) (left, right, top, bottom bool) {
	if posX < rect.X-reach || posX >= rect.X+rect.W+reach || posY < rect.Y-reach || posY >= rect.Y+rect.H+reach {
		return
	}
	left, right = posX < rect.X+grab, posX >= rect.X+rect.W-grab
	top, bottom = posY < rect.Y+grab, posY >= rect.Y+rect.H-grab
	return
}

// cursorFor is the resize cursor of a set of edges.
func cursorFor(left, right, top, bottom bool) core.Cursor {
	switch {
	case (left && top) || (right && bottom):
		return core.CursorResizeNWSE
	case (right && top) || (left && bottom):
		return core.CursorResizeNESW
	case left || right:
		return core.CursorResizeH
	}
	return core.CursorResizeV
}

func (b *resizableB) Update(element *El, ctx *core.Ctx) {
	in, st := ctx.In, b.state
	if st.active {
		if !in.Held {
			st.active = false
			return
		}
		ctx.Cursor = cursorFor(st.left, st.right, st.top, st.bottom)
		dx, dy := ctx.PointerX-st.startX, ctx.PointerY-st.startY
		if st.left || st.right {
			grow := 2 * dx
			if st.left {
				grow = -grow
			}
			st.W = min(max(st.startW+grow, b.minW), b.maxW)
		}
		if st.top || st.bottom {
			grow := 2 * dy
			if st.top {
				grow = -grow
			}
			limit := b.maxH
			if own := element.MaxHeight(); own > 0 {
				limit = min(limit, own)
			}
			st.H = min(max(st.startH+grow, b.minH), limit)
		}
		ctx.Fired++ // the app builds the dialog at its state's size
		return
	}
	left, right, top, bottom := edges(element.Rect(), in.X, in.Y, ctx.Sizes.EdgeGrab, ctx.Sizes.EdgeReach)
	if !(left || right || top || bottom) {
		return
	}
	ctx.Cursor = cursorFor(left, right, top, bottom)
	if !in.Click {
		return
	}
	if in.Clicks >= 2 {
		st.W, st.H = 0, 0
		ctx.Fired++
		return
	}
	st.active, st.left, st.right, st.top, st.bottom = true, left, right, top, bottom
	st.startX, st.startY, st.startW, st.startH = in.X, in.Y, element.Rect().W, element.Rect().H
	st.W, st.H = element.Rect().W, element.Rect().H // both are the user's size from now on, whichever edge moves
	ctx.Fired++
}
