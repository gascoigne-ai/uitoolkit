package widget

import (
	"image/color"
	"testing"
	"time"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
	. "github.com/gascoigne-ai/uitoolkit/view"
)

func TestCaretRectPerShape(t *testing.T) {
	const glyphW, lineH = 8, 16
	rows := []struct {
		name  string
		row   core.Rect
		block core.Rect
	}{
		{"tall row", core.Rect{Y: 10, W: 200, H: 40}, core.Rect{X: 30, Y: 22, W: 8, H: 16}},
		{"short table cell", core.Rect{Y: 10, W: 200, H: 20}, core.Rect{X: 30, Y: 12, W: 8, H: 16}},
		{"row shorter than the line", core.Rect{Y: 10, W: 200, H: 12}, core.Rect{X: 30, Y: 10, W: 8, H: 12}},
	}
	for _, row := range rows {
		block := row.block
		shapes := []struct {
			name  string
			shape core.CaretShape
			want  core.Rect
		}{
			{"block", core.CaretBlock, block},
			{"bar", core.CaretBar, core.Rect{X: 30, Y: block.Y, W: core.CaretW, H: block.H}},
			{"underline", core.CaretUnderline, core.Rect{X: 30, Y: block.Y + block.H - core.CaretW, W: glyphW, H: core.CaretW}},
			{"no such shape draws the block", core.CaretShape(7), block},
		}
		for _, shape := range shapes {
			t.Run(row.name+"/"+shape.name, func(t *testing.T) {
				got := styleOf(shape.shape).rect(caretBlock(row.row, 30, glyphW, lineH), glyphW)
				if got != shape.want {
					t.Fatalf("rect = %+v, want %+v", got, shape.want)
				}
			})
		}
	}
}

// paintedOps lays root out and returns what one draw of it asks for.
func paintedOps(ctx *core.Ctx, root *El) []rendertest.Op {
	Layout(ctx, root, 300, 200)
	canvas := rendertest.NewCanvas(600, 400)
	Draw(ctx, root, canvas)
	return canvas.Ops()
}

// fillsOf are the plain fills in clr, with the index of each in ops.
func fillsOf(ops []rendertest.Op, clr color.RGBA) (boxes []render.Box, at []int) {
	for i, op := range ops {
		if fill, ok := op.(rendertest.FillRect); ok && fill.Color == clr {
			boxes, at = append(boxes, fill.Box), append(at, i)
		}
	}
	return boxes, at
}

// caretFill is the one fill in the accent color: the caret. An outline is a
// stroke, so no other op of a focused input matches.
func caretFill(t *testing.T, ctx *core.Ctx, ops []rendertest.Op) (render.Box, int) {
	t.Helper()
	boxes, at := fillsOf(ops, ctx.Color(core.Accent))
	if len(boxes) != 1 {
		t.Fatalf("%d accent fills, want the caret alone", len(boxes))
	}
	return boxes[0], at[0]
}

// glyphsIn counts the texts drawn in exactly clr.
func glyphsIn(ops []rendertest.Op, clr color.RGBA) (count int) {
	for _, op := range ops {
		if text, ok := op.(rendertest.DrawText); ok && text.Color == clr {
			count++
		}
	}
	return count
}

func textsIn(ops []rendertest.Op) (count int) {
	for _, op := range ops {
		if _, ok := op.(rendertest.DrawText); ok {
			count++
		}
	}
	return count
}

func typingField(value string, cursor int) *TextField {
	field := NewTextField(value)
	field.Focused, field.Cursor = true, cursor
	return &field
}

func TestEachCaretShapeDrawsItsOwnOps(t *testing.T) {
	for _, scale := range []float64{1, 1.5} {
		ctx := testCtx(t)
		if err := ctx.SetScale(scale); err != nil {
			t.Fatal(err)
		}
		field := typingField("abc", 1)
		draw := func(shape core.CaretShape) (render.Box, []rendertest.Op, []rendertest.Op) {
			ctx.Caret.Shape = shape
			ops := paintedOps(ctx, Field(field))
			box, at := caretFill(t, ctx, ops)
			return box, ops, ops[at+1:]
		}
		block, blockOps, _ := draw(core.CaretBlock)
		if glyphsIn(blockOps, ctx.Color(core.AccentText)) != 1 {
			t.Errorf("scale %v: the block does not redraw its glyph in AccentText", scale)
		}

		bar, barOps, afterBar := draw(core.CaretBar)
		if width := bar.X1 - bar.X0; width != 2 && width != 3 {
			t.Errorf("scale %v: bar is %v px wide, want 2 or 3", scale, width)
		}
		if bar.X0 != block.X0 || bar.Y0 != block.Y0 || bar.Y1 != block.Y1 {
			t.Errorf("scale %v: bar %+v does not stand at the left of the block %+v", scale, bar, block)
		}
		if glyphsIn(barOps, ctx.Color(core.AccentText)) != 0 || textsIn(afterBar) != 0 {
			t.Errorf("scale %v: the bar redraws the glyph", scale)
		}

		under, underOps, afterUnder := draw(core.CaretUnderline)
		if height := under.Y1 - under.Y0; height != 2 && height != 3 {
			t.Errorf("scale %v: underline is %v px high, want 2 or 3", scale, height)
		}
		if under.X0 != block.X0 || under.X1 != block.X1 || under.Y1 != block.Y1 {
			t.Errorf("scale %v: underline %+v is not the bottom of the block %+v", scale, under, block)
		}
		if glyphsIn(underOps, ctx.Color(core.AccentText)) != 0 || textsIn(afterUnder) != 0 {
			t.Errorf("scale %v: the underline redraws the glyph", scale)
		}
	}
}

func TestTheAreaCaretStandsWhereCaretAtSays(t *testing.T) {
	long := "alpha bravo charlie delta echo foxtrot golf hotel india juliett kilo lima"
	cases := []struct {
		name, text string
		cursor     int
		wrapped    bool // the cursor goes to the start of the second row
	}{
		{"a tab is one space wide at the tab's start", "a\tb", 1, false},
		{"end of a line", "abc\ndef", 3, false},
		{"empty text", "", 0, false},
		{"the start of the next soft-wrapped row", long, 0, true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			ctx := testCtx(t)
			area := NewTextArea(test.text)
			area.Focused, area.Wrap = true, true
			rect := core.Rect{W: 300, H: 200}
			glyph := " " // whose width the underline has
			if test.wrapped {
				rows := area.geometry(ctx, rect).layout.rows
				if len(rows) < 2 {
					t.Fatalf("%d rows, the text does not wrap", len(rows))
				}
				test.cursor = rows[1].from
				glyph = string([]rune(test.text)[test.cursor])
			}
			area.Cursor = test.cursor
			area.Update(ctx, rect)
			x, rowBottom, ok := area.CaretAt()
			if !ok {
				t.Fatal("CaretAt does not know the caret")
			}
			// the caret keeps to the text line, which a row of the editor
			// centers with the same room above and below
			margin := (newMonoGrid(ctx).lineH - ctx.LineH(ctx.F.Mono)) / 2
			bottom := rowBottom - margin
			if test.wrapped {
				if row, offset := area.caretRow(area.geometry(ctx, rect)); row != 1 || offset != 0 {
					t.Fatalf("caret at row %d column %d, want the start of row 1", row, offset)
				}
			}
			draw := func(shape core.CaretShape) render.Box {
				ctx.Caret.Shape = shape
				canvas := rendertest.NewCanvas(rect.W, rect.H)
				area.Draw(ctx, canvas, rect)
				box, _ := caretFill(t, ctx, canvas.Ops())
				return box
			}
			width := ctx.Width(ctx.F.Mono, glyph)
			wantUnder := render.Box{X0: float32(x), Y0: float32(bottom - core.CaretW), X1: float32(x + width), Y1: float32(bottom)}
			if got := draw(core.CaretUnderline); got != wantUnder {
				t.Errorf("underline = %+v, want %+v", got, wantUnder)
			}
			wantBar := render.Box{X0: float32(x), Y0: float32(bottom - ctx.LineH(ctx.F.Mono)), X1: float32(x + core.CaretW), Y1: float32(bottom)}
			if got := draw(core.CaretBar); got != wantBar {
				t.Errorf("bar = %+v, want %+v", got, wantBar)
			}
		})
	}
}

func TestASteadyCaretIsDrawnInTheOffPhase(t *testing.T) {
	ctx := testCtx(t)
	start := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	now := start
	ctx.Now = func() time.Time { return now }
	ctx.Caret.Steady = true
	ctx.RestartCaret()
	now = start.Add(700 * time.Millisecond) // hidden, were it blinking
	area := NewTextArea("abc")
	area.Focused = true
	for name, root := range map[string]*El{"field": Field(typingField("abc", 1)), "area": Area(&area)} {
		if boxes, _ := fillsOf(paintedOps(ctx, root), ctx.Color(core.Accent)); len(boxes) != 1 {
			t.Errorf("%s: %d caret fills in the off phase, want 1", name, len(boxes))
		}
	}
}

// A bar or underline leaves the selected text as it is, so it must lie over
// the selection fill to be seen.
func TestABarOrUnderlineLiesOverTheSelection(t *testing.T) {
	for name, shape := range map[string]core.CaretShape{"bar": core.CaretBar, "underline": core.CaretUnderline} {
		t.Run(name, func(t *testing.T) {
			ctx := testCtx(t)
			ctx.Caret.Shape = shape
			field := typingField("abcdef", 0)
			field.PressN(1, false, 1)
			field.Drag(4)
			ops := paintedOps(ctx, Field(field))
			_, selection := fillsOf(ops, ctx.Color(core.Selection))
			_, caret := caretFill(t, ctx, ops)
			if len(selection) != 1 || selection[0] > caret {
				t.Errorf("selection fills at %v, caret fill at %d: want one before the caret", selection, caret)
			}
			if glyphsIn(ops, ctx.Color(core.AccentText)) != 0 {
				t.Error("a glyph is redrawn in AccentText")
			}
		})
	}
}
