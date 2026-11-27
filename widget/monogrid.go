package widget

import (
	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
)

// monoGrid is the pixel geometry that CodeView and TextArea share: rows of
// lineH pixels with text at textX in a monospace font. They embed it so the
// column math, wrapping width, wheel steps and span drawing are written
// once and cannot drift apart.
type monoGrid struct {
	lineH, charW, textX int
	adv                 float64 // the exact advance of one character
	gutter              int     // Sizes.Gutter, the room around the text and the numbers
	lead                int     // Sizes.MonoTextLead, the room left of the text that its clip includes
	barW                int     // Sizes.ScrollbarW, the room the vertical scrollbar takes
}

// newMonoGrid measures the mono font; textX is set by the widget, which
// knows its gutter. A row is the measured line plus Sizes.MonoLinePad, which
// includes the theme's leading unless the theme pins it, so drawing, clicks,
// the wheel and paging all step by the same rows.
func newMonoGrid(ctx *core.Ctx) monoGrid {
	return monoGrid{lineH: ctx.LineH(ctx.F.Mono) + ctx.Sizes.MonoLinePad, charW: ctx.Width(ctx.F.Mono, "0"), adv: ctx.Advance(ctx.F.Mono), gutter: ctx.Sizes.Gutter, lead: ctx.Sizes.MonoTextLead, barW: ctx.Sizes.ScrollbarW}
}

// colX is the pixel offset of the start of column col from the start of the
// text, by the exact advance, so the caret, selection and matches line up
// with the text however long the line. Not col*charW: Width rounds to whole
// pixels, and that error would add up along the line.
func (g monoGrid) colX(col int) int { return int(float64(col)*g.adv + 0.5) }

// colAt is the column nearest to offset x from the start of the text. It
// can be negative left of the text; TextArea clamps it in its own colAt.
func (g monoGrid) colAt(posX int) int { return int((float64(posX) + g.adv/2) / g.adv) }

// wrapWidth is the room in pixels for the text of a row, 0 when the text
// does not wrap. The vertical bar is assumed present, so the rows do not
// change when it appears.
func (g monoGrid) wrapWidth(wrap bool, rect core.Rect) float64 {
	if !wrap {
		return 0
	}
	return float64(max(rect.W-g.barW-(g.textX-rect.X)-g.gutter, 1))
}

// wheelDX and wheelDY are the pixels a wheel reading scrolls sideways and
// down, at columns or lines to a notch (core.Feel).
func (g monoGrid) wheelDX(wheel float64, columns int) int {
	return int(-wheel * float64(g.charW) * float64(columns))
}

func (g monoGrid) wheelDY(wheel float64, lines int) int {
	return int(-wheel * float64(g.lineH) * float64(lines))
}

// keepTop is the ScrollY that shows, in the layout after, the line that was
// at the top in the layout before; a switch of Wrap would otherwise jump.
func (g monoGrid) keepTop(scrollY int, before, after wrapLayout) int {
	top := before.rowStart(scrollY / g.lineH)
	if row := after.rowOf(top); row >= 0 {
		return row * g.lineH
	}
	return scrollY
}

// selectionRect is the strip a selection fills on a row of cell height
// height at posY, from column from to column to of the row.
func (g monoGrid) selectionRect(scrollX, from, to, posY, height int) core.Rect {
	x0 := g.textX - scrollX + g.colX(from)
	x1 := g.textX - scrollX + g.colX(to)
	return core.Rect{X: x0, Y: posY, W: max(x1-x0, 0), H: height}
}

// drawSpans draws spans, icons included, along a row of a CodeView or a
// TextArea, placed by column so rounding does not add up along the line. A
// tab is drawn as the spaces up to its stop, counted along the whole line
// (row.col), so a continuation row keeps the line's stops.
func (g monoGrid) drawSpans(ctx *core.Ctx, dst render.Canvas, scrollX int, cell core.Rect, row visualRow, spans []Span) {
	col := row.col
	for _, sp := range spans {
		next := monoColumns(g.adv).spanString(sp.Text, col)
		spanX := g.textX - scrollX + g.colX(col-row.col)
		if sp.Icon != "" {
			drawIconAt(ctx, dst, sp.Icon, core.Rect{X: spanX, Y: cell.Y, W: g.colX(next-row.col) - g.colX(col-row.col), H: cell.H}, sp.Style)
		} else {
			ctx.TextMid(dst, ctx.F.Mono, core.ExpandTabs(sp.Text, col), spanX, cell, ctx.Color(sp.Style))
		}
		col = next
	}
}
