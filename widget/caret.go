package widget

import (
	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
)

// caretGlyph is the character the caret stands on: the rune at cursor, or a
// space at the end of the text and before a line break or tab, so the block
// still has a cell to cover and the underline a width.
func caretGlyph(runes []rune, cursor int) string {
	if cursor < 0 || cursor >= len(runes) || runes[cursor] == '\n' || runes[cursor] == '\t' {
		return " "
	}
	return string(runes[cursor])
}

// caretBlock is the block's rectangle: w wide at x, one text line high and
// centered in row (or the row's own height when it is shorter than a line).
func caretBlock(row core.Rect, posX, width, lineH int) core.Rect {
	height := min(lineH, row.H)
	return core.Rect{X: posX, Y: row.Y + (row.H-height)/2, W: width, H: height}
}

// caretStyle is how one shape is drawn: the part of the block it fills, and
// whether the character is drawn again on top of it.
type caretStyle struct {
	rect  func(block core.Rect, glyphW int) core.Rect
	glyph bool
}

// caretStyles is indexed by core.CaretShape.
var caretStyles = [...]caretStyle{
	core.CaretBlock:     {rect: wholeBlock, glyph: true},
	core.CaretBar:       {rect: barOf},
	core.CaretUnderline: {rect: underlineOf},
}

func wholeBlock(block core.Rect, _ int) core.Rect { return block }

func barOf(block core.Rect, _ int) core.Rect {
	return core.Rect{X: block.X, Y: block.Y, W: core.CaretW, H: block.H}
}

func underlineOf(block core.Rect, glyphW int) core.Rect {
	return core.Rect{X: block.X, Y: block.Y + block.H - core.CaretW, W: glyphW, H: core.CaretW}
}

// styleOf is the one lookup of caretStyles.
func styleOf(shape core.CaretShape) caretStyle {
	if int(shape) >= len(caretStyles) {
		return caretStyles[core.CaretBlock]
	}
	return caretStyles[shape]
}

// drawCaret paints the caret of ctx.Caret.Shape at posX in the accent color,
// sized for glyph; the block also redraws glyph on itself.
func drawCaret(ctx *core.Ctx, dst render.Canvas, face render.Face, glyph string, posX int, row core.Rect) {
	glyphW := ctx.Width(face, glyph)
	style := styleOf(ctx.Caret.Shape)
	ctx.Fill(dst, style.rect(caretBlock(row, posX, glyphW, ctx.LineH(face)), glyphW), ctx.Color(core.Accent))
	if style.glyph {
		ctx.TextMid(dst, face, glyph, posX, row, ctx.Color(core.AccentText))
	}
}
