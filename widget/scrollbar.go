package widget

import (
	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
)

// clampScroll keeps pos inside [0, contentLen-viewLen].
func clampScroll(pos, contentLen, viewLen int) int {
	return min(max(pos, 0), max(contentLen-viewLen, 0))
}

// thumb returns the offset and length of a scrollbar thumb inside a
// track of trackLen pixels, for content of contentLen shown in a view of
// viewLen scrolled to pos. The thumb is at least minLen long, as far as the
// track allows. ok is false when everything fits.
func thumb(trackLen, contentLen, viewLen, pos, minLen int) (off, length int, ok bool) {
	if trackLen <= 0 || contentLen <= viewLen {
		return 0, 0, false
	}
	length = min(max(trackLen*viewLen/contentLen, minLen), trackLen)
	pos = clampScroll(pos, contentLen, viewLen)
	return (trackLen - length) * pos / (contentLen - viewLen), length, true
}

// posFromOffset is the scroll position that puts the thumb at off, so a
// dragged thumb can drive the content.
func posFromOffset(trackLen, contentLen, viewLen, off, minLen int) int {
	_, length, ok := thumb(trackLen, contentLen, viewLen, 0, minLen)
	if !ok || trackLen-length <= 0 {
		return 0
	}
	maxOff := trackLen - length
	off = min(max(off, 0), maxOff)
	return off * (contentLen - viewLen) / maxOff
}

// drawScrollbar draws the thumb in track for the given scroll state. The
// caller says which way the track runs: a themed width can be more than the
// length of a short view's track. active brightens it (hover or drag). It
// draws nothing when everything fits.
func drawScrollbar(ctx *core.Ctx, dst render.Canvas, track core.Rect, vertical bool, contentLen, viewLen, pos int, active bool) {
	trackLen := track.W
	if vertical {
		trackLen = track.H
	}
	off, length, ok := thumb(trackLen, contentLen, viewLen, pos, ctx.Sizes.MinThumbLen)
	if !ok {
		return
	}
	clr := ctx.Color(core.Muted)
	if active {
		clr = ctx.Color(core.Text)
	}
	inset := ctx.Sizes.ThumbInset
	bar := core.Rect{X: track.X + inset, Y: track.Y + off, W: max(track.W-2*inset, 1), H: length}
	if !vertical {
		bar = core.Rect{X: track.X + off, Y: track.Y + inset, W: length, H: max(track.H-2*inset, 1)}
	}
	ctx.FillStyled(dst, bar, clr, clr, ctx.Radius("scrollbar"))
}

// pressThumb is a press at `at` in a scrollbar track for content shown
// through a view scrolled to pos: it either grabs the thumb (grabbed, with
// the offset inside it) or pages toward the press (the new pos).
func pressThumb(track, content, view, pos, at, minLen int) (grabbed bool, off, newPos int) {
	start, length, ok := thumb(track, content, view, pos, minLen)
	switch {
	case !ok:
		return false, 0, pos
	case at >= start && at < start+length:
		return true, at - start, pos
	case at < start:
		return false, 0, clampScroll(pos-view, content, view)
	}
	return false, 0, clampScroll(pos+view, content, view)
}
