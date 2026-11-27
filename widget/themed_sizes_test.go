package widget

import (
	"image/color"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
	. "github.com/gascoigne-ai/uitoolkit/view"
)

// themedCtx is a Ctx under a theme that pins the named sizes over metrics.
func themedCtx(t *testing.T, metrics core.Metrics, pins map[string]int) *core.Ctx {
	t.Helper()
	for name, pixels := range pins {
		pinnedMetrics, err := metrics.Pin(name, pixels)
		if err != nil {
			t.Fatal(err)
		}
		metrics = pinnedMetrics
	}
	ctx := testCtx(t)
	theme := ctx.Th
	theme.Metrics = metrics
	if err := ctx.SetTheme(theme); err != nil {
		t.Fatal(err)
	}
	return ctx
}

// themedPins give every spacing role another value than any other, so a
// widget that reads the wrong one shows.
var themedPins = map[string]int{
	"tabLead": 5, "caretW": 3, "nearGap": 7, "tightGap": 11, "tabGap": 9, "ruleInset": 13,
	"edgeGrab": 15, "looseGap": 14, "cellPad": 17, "labelPad": 19, "tabSepInset": 3,
	"gutter": 21, "dialogPad": 23, "toolbarH": 51, "tabBarH": 53, "dialogTitleH": 55,
	"dialogToolbarH": 57, "dialogW": 500, "dialogButtonMinW": 150,
	"scrollbarW": 20, "minThumbLen": 60, "thumbInset": 6, "hscrollArrowW": 31,
	"menuMinW": 205, "menuMaxW": 610, "menuSepH": 27, "menuPad": 25, "menuHintH": 29,
	"menuRowPadX": 33, "menuIconW": 35, "menuShortcutGap": 37,
	"tipPadX": 39, "tipPadY": 41, "tipOffset": 43,
	"toastW": 441, "toastH": 45, "toastGap": 47, "toastMargin": 49,
	"buttonPadX": 59, "fieldPadX": 61, "selectionInset": 8, "revealMargin": 63,
	"sliderW": 330, "sliderH": 71, "sliderTrack": 10, "sliderThumbW": 22, "sliderThumbH": 46,
	"sliderTick": 34, "sliderPad": 54, "dividerGrab": 12, "edgeReach": 24,
	"tableActionW": 62, "tableGripW": 48, "minColumnW": 130,
	"foldColW": 40, "monoTextLead": 28, "monoLinePad": 18, "switchInset": 4,
}

func laidOut(ctx *core.Ctx, root *El) *El {
	Layout(ctx, root, 800, 600)
	return root
}

// leftOf is how far child starts right of parent: the parent's left padding
// for a first child.
func leftOf(parent, child *El) int { return child.Rect().X - parent.Rect().X }

// between is the space between two neighbours in a row: its gap.
func between(first, second *El) int { return second.Rect().X - first.Rect().X - first.Rect().W }

func leftBox() *El { return Row(W(10), H(10)) }

// filledBoxes draws root and returns the boxes filled in clr, flat, rounded
// or as a gradient from it.
func filledBoxes(ctx *core.Ctx, root *El, clr color.RGBA) (boxes []core.Rect) {
	canvas := rendertest.NewCanvas(800, 600)
	Draw(ctx, root, canvas)
	for _, op := range canvas.Ops() {
		var box render.Box
		var fill color.RGBA
		switch drawnOp := op.(type) {
		case rendertest.FillRect:
			box, fill = drawnOp.Box, drawnOp.Color
		case rendertest.FillRoundRect:
			box, fill = drawnOp.Box, drawnOp.Top
		case rendertest.FillRectGradient:
			box, fill = drawnOp.Box, drawnOp.Top
		default:
			continue
		}
		if fill == clr {
			boxes = append(boxes, core.Rect{X: int(box.X0), Y: int(box.Y0), W: int(box.X1 - box.X0), H: int(box.Y1 - box.Y0)})
		}
	}
	return boxes
}

// nth is the box at index, a zero one when fewer were drawn.
func nth(boxes []core.Rect, index int) core.Rect {
	if index < len(boxes) {
		return boxes[index]
	}
	return core.Rect{}
}

// monoBars is what the two scrollbars of a CodeView or a TextArea show under
// the themed pins: how far the vertical thumb is off the right edge, its
// width and length, how far the horizontal one is off the bottom, its height,
// and whether a press beyond the default track and thumb grabs each.
func monoBars(ctx *core.Ctx, element *El, grabbed func() bool) []int {
	root := laidOut(ctx, Row(element))
	rect := element.Rect()
	thumbs := filledBoxes(ctx, root, ctx.Color(core.Muted))
	vertical, horizontal := nth(thumbs, 0), nth(thumbs, 1)
	got := []int{rect.X + rect.W - vertical.X, vertical.W, vertical.H, rect.Y + rect.H - horizontal.Y, horizontal.H}
	for _, at := range [][2]int{{rect.X + rect.W - 18, rect.Y + 40}, {rect.X + 40, rect.Y + rect.H - 18}} {
		ctx.In = core.Input{X: at[0], Y: at[1], Click: true, Held: true, Clicks: 1}
		Update(ctx, root)
		got = append(got, btoi(grabbed()))
		ctx.In = core.Input{X: at[0], Y: at[1]}
		Update(ctx, root)
	}
	return got
}

// longText is more and wider lines than any view of these tests shows.
func longText() string {
	return strings.Repeat("m", 400) + strings.Repeat("\nline", 500)
}

// drawnOps are the calls a Draw of root makes.
func drawnOps(ctx *core.Ctx, root *El) []rendertest.Op {
	canvas := rendertest.NewCanvas(800, 600)
	Draw(ctx, root, canvas)
	return canvas.Ops()
}

// textOf is the first call that draws text, a zero one without any.
func textOf(ops []rendertest.Op, text string) rendertest.DrawText {
	for _, op := range ops {
		if drawnText, ok := op.(rendertest.DrawText); ok && drawnText.Text == text {
			return drawnText
		}
	}
	return rendertest.DrawText{}
}

// cursorAt is the cursor shape root asks for with the pointer there.
func cursorAt(ctx *core.Ctx, root *El, posX, posY int) core.Cursor {
	ctx.In, ctx.Cursor = core.Input{X: posX, Y: posY}, core.CursorDefault
	Update(ctx, root)
	return ctx.Cursor
}

// tapAt puts the pointer down there and releases it in the frame after.
func tapAt(ctx *core.Ctx, root *El, posX, posY int) {
	ctx.In = core.Input{X: posX, Y: posY, Click: true, Held: true, Clicks: 1}
	Update(ctx, root)
	ctx.In = core.Input{X: posX, Y: posY}
	Update(ctx, root)
}

// buttonSizes are where a button draws its label, how much wider than it the
// button is, and whether a press at the far end of the right pad clicks it.
func buttonSizes(ctx *core.Ctx, button func(label string, opts ...Option) *El) []int {
	const label = "a label"
	clicks := 0
	element := button(label, OnClick(func() { clicks++ }))
	root := laidOut(ctx, Row(AlignItems(AlignStart), element))
	rect := element.Rect()
	text := textOf(drawnOps(ctx, root), label)
	// Right of a button with the default pads.
	tapAt(ctx, root, rect.X+ctx.Width(ctx.F.UI, label)+2*core.DefaultSizes().ButtonPadX+1, rect.Y+2)
	return []int{int(text.Left) - rect.X, rect.W - ctx.Width(ctx.F.UI, label), clicks}
}

// monoRows is where a CodeView or a TextArea puts its rows under the theme:
// the step from one row to the next, and the row a press just inside the
// third one selects or moves to.
func monoRows(ctx *core.Ctx, element *El, rowAt func() int) []int {
	root := laidOut(ctx, Row(element))
	ops := drawnOps(ctx, root)
	step := textOf(ops, "second").Top - textOf(ops, "first").Top
	pitch := ctx.LineH(ctx.F.Mono) + ctx.Sizes.MonoLinePad
	tapAt(ctx, root, element.Rect().X+element.Rect().W/2, element.Rect().Y+2*pitch+1)
	return []int{int(step) - ctx.LineH(ctx.F.Mono), rowAt()}
}

func namedLines() []Line {
	var lines []Line
	for _, text := range []string{"first", "second", "third", "fourth"} {
		lines = append(lines, Line{Spans: []Span{{Text: text}}})
	}
	return lines
}

// pickedAt clicks the point and returns the row the menu reports, -1 for none.
func pickedAt(ctx *core.Ctx, root *El, picked *int, posX, posY int) int {
	*picked = -1
	click(ctx, posX, posY)
	Update(ctx, root)
	return *picked
}

func themedMenu(ctx *core.Ctx, picked *int, items []MenuItem) (root, menu *El) {
	state := &MenuState{Items: items}
	root = laidOut(ctx, Stack(Menu(ctx, state, OnSelect(func(index int) { *picked = index }))))
	return root, root.Children()[0]
}

func TestWidgetsFollowThemedSizes(t *testing.T) {
	barSpacing := func(bar *El) []int {
		return []int{bar.Rect().H, leftOf(bar, bar.Children()[0]), between(bar.Children()[0], bar.Children()[1])}
	}
	tests := []struct {
		name string
		got  func(ctx *core.Ctx) []int
		want []int
	}{
		{"toolbar height, padding, gap", func(ctx *core.Ctx) []int {
			return barSpacing(laidOut(ctx, Col(Toolbar(leftBox(), leftBox()))).Children()[0])
		}, []int{51, 21, 11}},
		{"tab bar height and top gap", func(ctx *core.Ctx) []int {
			bar := laidOut(ctx, Col(TabBar(Row(Grow(1))))).Children()[0]
			return []int{bar.Rect().H, bar.Children()[0].Rect().Y - bar.Rect().Y}
		}, []int{53, 9}},
		{"dialog title height, padding, gap", func(ctx *core.Ctx) []int {
			bar := laidOut(ctx, Col(DialogTitle("T", leftBox()))).Children()[0]
			return []int{bar.Rect().H, leftOf(bar, bar.Children()[0]), between(bar.Children()[0], bar.Children()[1])}
		}, []int{55, 23, 14}},
		{"dialog toolbar height, padding, gap", func(ctx *core.Ctx) []int {
			return barSpacing(laidOut(ctx, Col(DialogToolbar(leftBox(), leftBox()))).Children()[0])
		}, []int{57, 23, 14}},
		{"message dialog width, body pad and gap, button row gap, button min width", func(ctx *core.Ctx) []int {
			state := &DialogState{Title: "T", Body: "B", Buttons: []DialogButton{{Label: "a"}, {Label: "b"}}}
			panel := laidOut(ctx, Stack(MessageDialog(state))).Children()[1]
			body := panel.Children()[1]
			text, row := body.Children()[0], body.Children()[1]
			first, second := row.Children()[1], row.Children()[2]
			return []int{panel.Rect().W, leftOf(body, text), row.Rect().Y - text.Rect().Y - text.Rect().H,
				between(first, second), first.Rect().W}
		}, []int{500, 23, 21, 14, 150}},
		{"form row gap", func(ctx *core.Ctx) []int {
			row := laidOut(ctx, Col(FormRow("L", 40, leftBox()))).Children()[0]
			return []int{between(row.Children()[0], row.Children()[1])}
		}, []int{21}},
		{"check row gap", func(ctx *core.Ctx) []int {
			row := laidOut(ctx, Col(CheckRow(true, leftBox()))).Children()[0]
			return []int{between(row.Children()[0], row.Children()[1])}
		}, []int{14}},
		{"menu separator inset and row gap", func(ctx *core.Ctx) []int {
			state := &MenuState{Items: []MenuItem{{Label: "a", Icon: "cross"}, {Separator: true}}}
			menu := laidOut(ctx, Stack(Menu(ctx, state))).Children()[0]
			item, separator := menu.Children()[0], menu.Children()[1]
			return []int{leftOf(separator, separator.Children()[0]), between(item.Children()[0], item.Children()[1])}
		}, []int{13, 14}},
		{"tab lead, label pad, badge gap, separator inset", func(ctx *core.Ctx) []int {
			strip := laidOut(ctx, Col(Tabs([]TabItem{{Label: "a", Badge: 1}, {Label: "b"}}, 0))).Children()[0]
			first := strip.Children()[0]
			_, inset, _ := strings.Cut(Dump(strip.Children()[1]), "borderInset=")
			value, _ := strconv.Atoi(strings.Fields(inset + " 0")[0])
			return []int{leftOf(strip, first), leftOf(first, first.Children()[0]),
				between(first.Children()[0], first.Children()[1]), value}
		}, []int{5, 19, 7, 3}},
		{"scrolling tab lead", func(ctx *core.Ctx) []int {
			strip := laidOut(ctx, Row(ScrollTabs(&ScrollState{}, []TabItem{{Label: "a"}}, 0))).Children()[0]
			return []int{leftOf(strip, strip.Children()[0])}
		}, []int{5}},
		{"key chips gap and padding", func(ctx *core.Ctx) []int {
			chips := laidOut(ctx, Row(KeyChips("Ctrl+A"))).Children()[0]
			first := chips.Children()[0]
			return []int{leftOf(first, first.Children()[0]), between(first, chips.Children()[1])}
		}, []int{7, 7}},
		{"link button side padding", func(ctx *core.Ctx) []int {
			button := laidOut(ctx, Row(LinkButton("go"))).Children()[0]
			return []int{leftOf(button, button.Children()[0])}
		}, []int{14}},
		{"table header and cell padding", func(ctx *core.Ctx) []int {
			table := laidOut(ctx, Col(Table(TableSpec{Columns: []string{"a"}, Weights: []float64{1}, Rows: []TableRow{{Cells: []string{"x"}}}}))).Children()[0]
			header, cell := table.Children()[0].Children()[0], table.Children()[1].Children()[0]
			return []int{leftOf(header, header.Children()[0]), leftOf(cell, cell.Children()[0])}
		}, []int{17, 17}},
		{"table edit field keeps the text where the cell had it", func(ctx *core.Ctx) []int {
			field := NewTextField("x")
			spec := TableSpec{Columns: []string{"a"}, Weights: []float64{1}, Rows: []TableRow{{Cells: []string{"x"}}}, Editing: &TableEdit{Field: &field}}
			table := laidOut(ctx, Col(Table(spec))).Children()[0]
			cell := table.Children()[1].Children()[0]
			edit := cell.Children()[0]
			return []int{leftOf(cell, edit), leftOf(cell, edit) + edit.Behavior().(*fieldB).textInset(ctx.Sizes)}
		}, []int{5, 17}},
		{"toast gutter at the window edge and inside", func(ctx *core.Ctx) []int {
			stack := &ToastStack{}
			stack.Show("saved", ToastInfo, time.Unix(0, 0))
			toasts := laidOut(ctx, Stack(Toasts(stack, 800, 600, 0))).Children()[0]
			toast := toasts.Children()[0]
			return []int{800 - toasts.Rect().X - toasts.Rect().W, leftOf(toast, toast.Children()[0])}
		}, []int{21, 21}},
		{"code view gutter before the numbers and after the widest line", func(ctx *core.Ctx) []int {
			code := &CodeView{}
			code.SetLines(sampleLines())
			rect := core.Rect{W: 400, H: 300}
			geometry := code.geometry(ctx, rect)
			return []int{geometry.textX - geometry.numW - geometry.foldW - 2*geometry.lead,
				code.contentWidthOf(geometry, rect) - geometry.textX - geometry.colX(code.maxCols),
				400 - ctx.Sizes.ScrollbarW - geometry.textX - int(geometry.wrapWidth(true, rect))}
		}, []int{21, 21, 21}},
		{"text area gutter around the numbers and after the widest line", func(ctx *core.Ctx) []int {
			area := &TextArea{Text: "one\ntwo"}
			rect := core.Rect{W: 400, H: 300}
			geometry := area.geometry(ctx, rect)
			return []int{geometry.textX - geometry.charW,
				geometry.contentW(rect) - geometry.textX - geometry.colX(geometry.longest)}
		}, []int{42, 21}},
		{"bar caret width", func(ctx *core.Ctx) []int {
			ctx.Caret.Shape = core.CaretBar
			canvas := rendertest.NewCanvas(100, 40)
			drawCaret(ctx, canvas, ctx.F.UI, "a", 30, core.Rect{W: 100, H: 40})
			box, _ := caretFill(t, ctx, canvas.Ops())
			return []int{int(box.X1 - box.X0)}
		}, []int{3}},
		{"underline caret height", func(ctx *core.Ctx) []int {
			ctx.Caret.Shape = core.CaretUnderline
			canvas := rendertest.NewCanvas(100, 40)
			drawCaret(ctx, canvas, ctx.F.UI, "a", 30, core.Rect{W: 100, H: 40})
			box, _ := caretFill(t, ctx, canvas.Ops())
			return []int{int(box.Y1 - box.Y0)}
		}, []int{3}},
		{"scrollbarW, thumbInset, minThumbLen", func(ctx *core.Ctx) []int {
			state := &ScrollState{}
			root := laidOut(ctx, Col(Scroll(state, Row(H(40000)))))
			content := root.Children()[0].Children()[0]
			bar := nth(filledBoxes(ctx, root, ctx.Color(core.Muted)), 0)
			got := []int{content.Rect().W, bar.X, bar.W, bar.H}
			// Left of an 8px track and below a 24px thumb: a grab, not a page.
			ctx.In = core.Input{X: 782, Y: 40, Click: true, Held: true}
			Update(ctx, root)
			got = append(got, state.Y)
			ctx.In = core.Input{X: 782, Y: 94, Held: true}
			Update(ctx, root)
			return append(got, state.Y)
		}, []int{780, 786, 8, 60, 0, 54 * (40000 - 600) / (600 - 60)}},
		{"code view scrollbars", func(ctx *core.Ctx) []int {
			code := &CodeView{}
			var lines []Line
			for _, text := range strings.Split(longText(), "\n") {
				lines = append(lines, Line{Spans: []Span{{Text: text}}})
			}
			code.SetLines(lines)
			return monoBars(ctx, Code(code, Grow(1)), code.drag.active)
		}, []int{14, 8, 60, 14, 8, 1, 1}},
		{"text area scrollbars", func(ctx *core.Ctx) []int {
			area := NewTextArea(longText())
			return monoBars(ctx, Area(&area, Grow(1)), area.drag.active)
		}, []int{14, 8, 60, 14, 8, 1, 1}},
		{"hscrollArrowW", func(ctx *core.Ctx) []int {
			var hits []int
			state := &ScrollState{}
			items := []Item{W(200), H(20)}
			for range 20 {
				items = append(items, Row(W(40), OnClick(func() { hits = append(hits, 1) })))
			}
			root := laidOut(ctx, Col(AlignItems(AlignStart), HScroll(state, items...)))
			first := root.Children()[0].Children()[0]
			arrow := nth(filledBoxes(ctx, root, ctx.Color(core.Frame)), 0)
			got := []int{first.Rect().X, arrow.X, arrow.W}
			// Inside a 31px arrow, left of an 18px one.
			ctx.In = core.Input{X: 171, Y: 10, Click: true, Held: true}
			Update(ctx, root)
			return append(got, state.X, len(hits))
		}, []int{31, 169, 31, 100, 0}},
		{"menuMinW, menuPad, menuSepH, menuHintH, menuRowPadX, menuIconW", func(ctx *core.Ctx) []int {
			picked := -1
			root, menu := themedMenu(ctx, &picked, []MenuItem{{Label: "a", Icon: "cross", Hint: "x\ny"}, {Separator: true}, {Label: "b"}})
			row, separator, last := menu.Children()[0], menu.Children()[1], menu.Children()[2]
			icon, label := row.Children()[0], row.Children()[1]
			hot := nth(filledBoxes(ctx, root, ctx.Color(core.Selected)), 0)
			rect := row.Rect()
			return []int{menu.Rect().W, leftOf(menu, row), rect.Y - menu.Rect().Y, rect.H - ctx.Sizes.Compact,
				separator.Rect().H, leftOf(row, icon), icon.Rect().W, between(icon, label),
				hot.X - menu.Rect().X, hot.W, hot.H - ctx.Sizes.Compact,
				pickedAt(ctx, root, &picked, menu.Rect().X+10, rect.Y+5), // in the pad
				pickedAt(ctx, root, &picked, menu.Rect().X+170, rect.Y+rect.H-1),
				pickedAt(ctx, root, &picked, rect.X+1, separator.Rect().Y+separator.Rect().H-1),
				pickedAt(ctx, root, &picked, rect.X+1, last.Rect().Y+last.Rect().H-1),
			}
		}, []int{205, 25, 25, 58, 27, 33, 35, 14, 25, 155, 58, -1, 0, -1, 2}},
		{"menuMaxW", func(ctx *core.Ctx) []int {
			picked := -1
			root, menu := themedMenu(ctx, &picked, []MenuItem{{Label: strings.Repeat("m", 200)}})
			background := nth(filledBoxes(ctx, root, ctx.Color(core.Bg)), 0)
			row := menu.Children()[0].Rect()
			return []int{menu.Rect().W, background.W, pickedAt(ctx, root, &picked, menu.Rect().X+500, row.Y+2)}
		}, []int{610, 610, 0}},
		{"menuShortcutGap and the width's parts", func(ctx *core.Ctx) []int {
			picked := -1
			label := "mmmmmmmmmm"
			items := []MenuItem{{Label: label, Shortcut: "Ctrl+K", Icon: "cross"}}
			root, menu := themedMenu(ctx, &picked, items)
			rest := ctx.Width(ctx.F.UI, label) + ctx.Width(ctx.F.Small, core.PlatformKeyText(ctx.HostOS, "Ctrl+K"))
			background := nth(filledBoxes(ctx, root, ctx.Color(core.Bg)), 0)
			row := menu.Children()[0].Rect()
			return []int{menu.Rect().W - rest, background.W - rest,
				pickedAt(ctx, root, &picked, menu.Rect().X+menu.Rect().W-26, row.Y+2)}
		}, []int{37 + 35 + 14 + 2*33 + 2*25, 37 + 35 + 14 + 2*33 + 2*25, 0}},
		// A tooltip takes no input: what places it is the pointer.
		{"tipPadX, tipPadY, tipOffset", func(ctx *core.Ctx) []int {
			var got []int
			for _, pointerY := range []int{100, 590} {
				// A state whose delay for the tip has passed.
				root := laidOut(ctx, Stack(Col(Tooltip("tip")), TooltipLayer(&TooltipState{text: "tip"})))
				ctx.In, ctx.Tick = core.Input{X: 100, Y: pointerY}, tipDelayTicks
				Update(ctx, root)
				canvas := rendertest.NewCanvas(800, 600)
				Draw(ctx, root, canvas)
				box := nth(filledBoxes(ctx, root, ctx.Color(core.Frame)), 0)
				if pointerY == 590 {
					got = append(got, pointerY-box.Y-box.H)
					continue
				}
				textLeft := 0
				for _, op := range canvas.Ops() {
					if drawnText, ok := op.(rendertest.DrawText); ok {
						textLeft = int(drawnText.Left)
					}
				}
				got = append(got, box.X, box.Y, box.W-ctx.Width(ctx.F.Small, "tip"), box.H-ctx.LineH(ctx.F.Small), textLeft-box.X)
			}
			return got
		}, []int{143, 143, 78, 82, 39, 43}},
		{"toastW, toastH, toastGap, toastMargin", func(ctx *core.Ctx) []int {
			stack := &ToastStack{}
			stack.Show("a", ToastInfo, time.Unix(0, 0))
			stack.Show("b", ToastError, time.Unix(0, 0))
			root := laidOut(ctx, Stack(Toasts(stack, 800, 600, 10)))
			toasts := root.Children()[0]
			first, second := toasts.Children()[0].Rect(), toasts.Children()[1].Rect()
			boxes := filledBoxes(ctx, root, ctx.Color(core.Panel))
			got := []int{toasts.Rect().W, first.H, second.Y - first.Y - first.H, 590 - second.Y - second.H,
				len(boxes), nth(boxes, 0).W, nth(boxes, 0).H, nth(boxes, 1).Y - nth(boxes, 0).Y}
			for _, posY := range []int{first.Y + first.H + 1, first.Y + 1} { // the gap, then a toast
				// Left of a 320px toast.
				click(ctx, 800-21-441+2, posY)
				Update(ctx, root)
				got = append(got, len(*stack))
			}
			return got
		}, []int{441, 45, 47, 49, 2, 441, 45, 92, 2, 1}},
		{"resize edge grab", func(ctx *core.Ctx) []int {
			root := laidOut(ctx, Stack(Resizable(&ResizeState{}, 10, 10, 400, 400, W(200), H(100), FloatCenter())))
			rect := root.Children()[0].Rect()
			ctx.In = core.Input{X: rect.X + 12, Y: rect.Y + 50}
			ctx.Cursor = core.CursorDefault
			Update(ctx, root)
			return []int{int(ctx.Cursor)}
		}, []int{int(core.CursorResizeH)}},
		{"buttonPadX", func(ctx *core.Ctx) []int { return buttonSizes(ctx, Button) }, []int{59, 2 * 59, 1}},
		{"buttonPadX of a primary button", func(ctx *core.Ctx) []int { return buttonSizes(ctx, PrimaryButton) }, []int{59, 2 * 59, 1}},
		{"fieldPadX", func(ctx *core.Ctx) []int {
			field := NewTextField("mmmmmmmm")
			element := Field(&field, W(600))
			root := laidOut(ctx, Row(element))
			text := textOf(drawnOps(ctx, root), field.Text)
			tapAt(ctx, root, element.Rect().X+61+ctx.Width(ctx.F.UI, "mmm"), element.Rect().Y+2)
			return []int{int(text.Left) - element.Rect().X, field.Cursor}
		}, []int{61, 3}},
		// The selection band takes no input of its own.
		{"selectionInset", func(ctx *core.Ctx) []int {
			field := NewTextField("mm")
			field.SelectAll()
			element := Field(&field, W(600))
			root := laidOut(ctx, Row(element))
			band := nth(filledBoxes(ctx, root, ctx.Color(core.Selection)), 0)
			return []int{band.Y - element.Rect().Y, element.Rect().H - band.H}
		}, []int{8, 16}},
		{"revealMargin", func(ctx *core.Ctx) []int {
			field := NewTextField(strings.Repeat("m", 60))
			field.Focused = true
			element := Field(&field, W(300))
			root := laidOut(ctx, Row(element))
			viewW := 300 - 2*61
			ctx.In = core.Input{}
			Update(ctx, root)
			// The text is drawn left of the pad by the scroll.
			textShift := int(textOf(drawnOps(ctx, root), field.Text).Left) - element.Rect().X - 61 + field.ScrollX
			past := field.ScrollX - (ctx.Width(ctx.F.UI, field.Text) - viewW)
			// The end of the text is the margin left of the view's end.
			tapAt(ctx, root, element.Rect().X+61+viewW-63, element.Rect().Y+2)
			end := field.Cursor
			// A cursor left of the view comes back to the margin.
			field.Cursor, field.HasSel = 30, false
			field.ScrollX = ctx.Width(ctx.F.UI, field.Text[:30])
			ctx.In = core.Input{}
			Update(ctx, root)
			return []int{past, textShift, end, ctx.Width(ctx.F.UI, field.Text[:30]) - field.ScrollX}
		}, []int{63, 0, 60, 63}},
		{"sliderW, sliderH", func(ctx *core.Ctx) []int {
			var values []int
			slider := Slider(&SliderState{}, 0, 100, 1, 0, H(0), Grow(0), OnSlide(func(value int) { values = append(values, value) }))
			root := laidOut(ctx, Col(AlignItems(AlignStart), slider))
			rect := slider.Rect()
			// Outside a 120 by 20 slider.
			tapAt(ctx, root, rect.X+300, rect.Y+60)
			return []int{rect.W, rect.H, len(values)}
		}, []int{330, 71, 1}},
		// A slider takes a press anywhere in its rect, so the track, the thumb
		// and the tick are drawn only.
		{"sliderTrack, sliderThumbW, sliderThumbH, sliderTick, sliderPad", func(ctx *core.Ctx) []int {
			var values []int
			// A track of 330-2*54 = 222 pixels, one per value.
			slider := Slider(&SliderState{}, 0, 222, 1, 111, W(330), H(71), Grow(0), SliderMid(111),
				OnSlide(func(value int) { values = append(values, value) }))
			root := laidOut(ctx, Col(AlignItems(AlignStart), slider))
			rect := slider.Rect()
			middle := rect.Y + rect.H/2
			track := nth(filledBoxes(ctx, root, ctx.Color(core.Chip)), 0)
			tick := nth(filledBoxes(ctx, root, ctx.Color(core.Rule)), 0)
			accents := filledBoxes(ctx, root, ctx.Color(core.Accent))
			filled, thumb := nth(accents, 0), nth(accents, 1)
			tapAt(ctx, root, rect.X+54, rect.Y+2)
			tapAt(ctx, root, rect.X+54+60, rect.Y+2)
			return append([]int{track.X - rect.X, track.W, middle - track.Y, track.H,
				middle - tick.Y, tick.H, filled.H, filled.W,
				filled.X + filled.W - thumb.X, thumb.W, middle - thumb.Y, thumb.H}, values...)
		}, []int{54, 222, 5, 10, 17, 34, 10, 111, 11, 22, 23, 46, 0, 60}},
		// A divider stays a one pixel line; the grab is hit only.
		{"dividerGrab of a split", func(ctx *core.Ctx) []int {
			var got []int
			for _, vertical := range []bool{false, true} {
				split := Split(&SplitState{}, 200, Row(), Row())
				if vertical {
					Vertical()(split)
				}
				root := laidOut(ctx, split)
				for _, offset := range []int{-13, -12, 12, 13} {
					posX, posY := 200+offset, 100
					if vertical {
						posX, posY = 100, 200+offset
					}
					got = append(got, btoi(cursorAt(ctx, root, posX, posY) != core.CursorDefault))
				}
			}
			return got
		}, []int{0, 1, 1, 0, 0, 1, 1, 0}},
		{"dividerGrab of a table column", func(ctx *core.Ctx) []int {
			spec := TableSpec{Columns: []string{"a", "b"}, Weights: []float64{1, 1}, State: &TableState{}}
			root := laidOut(ctx, Col(Table(spec)))
			var got []int
			for _, offset := range []int{-13, -12, 12, 13} {
				got = append(got, btoi(cursorAt(ctx, root, 400+offset, 5) == core.CursorResizeH))
			}
			return got
		}, []int{0, 1, 1, 0}},
		// Nothing is drawn: the reach is where a press grabs the edge of a modal
		// element and does not dismiss it.
		{"edgeReach", func(ctx *core.Ctx) []int {
			var got []int
			for _, outside := range []int{24, 25} {
				dismissed := 0
				state := &ResizeState{}
				root := laidOut(ctx, Stack(Resizable(state, 10, 10, 400, 400, W(200), H(100), FloatCenter(), Modal(), OnDismiss(func() { dismissed++ }))))
				rect := root.Children()[0].Rect()
				ctx.In = core.Input{X: rect.X - outside, Y: rect.Y + 50, Click: true, Held: true, Clicks: 1}
				Update(ctx, root)
				got = append(got, btoi(state.active), dismissed)
			}
			return got
		}, []int{1, 0, 0, 1}},
		{"tableGripW, tableActionW", func(ctx *core.Ctx) []int {
			var grabbed, toggled, deleted []int
			spec := TableSpec{Columns: []string{"a"}, Weights: []float64{1}, Rows: []TableRow{{Cells: []string{"x"}}},
				Editable: true, State: &TableState{},
				RowLead:   func(int) Node { return nil },
				OnGrabRow: func(row int) { grabbed = append(grabbed, row) },
				OnToggle:  func(row int) { toggled = append(toggled, row) },
				OnDelete:  func(row int) { deleted = append(deleted, row) },
			}
			root := laidOut(ctx, Col(Table(spec)))
			table := root.Children()[0]
			header := table.Children()[0].Children()[0]
			row, placeholder := table.Children()[1], table.Children()[2]
			var got []int
			for _, line := range []*El{header, row, placeholder} {
				cells := line.Children()
				for _, index := range []int{0, 1, 2, len(cells) - 1} {
					got = append(got, cells[index].Rect().W)
				}
			}
			posY := row.Rect().Y + 2
			for _, posX := range []int{47, 48, 48 + 61, 48 + 62, 800 - 63, 800 - 62} {
				tapAt(ctx, root, posX, posY)
				got = append(got, len(grabbed)*100+len(toggled)*10+len(deleted))
			}
			return got
		}, []int{48, 62, 62, 62, 48, 62, 62, 62, 48, 62, 62, 62, 100, 110, 120, 120, 120, 121}},
		{"minColumnW", func(ctx *core.Ctx) []int {
			spec := TableSpec{Columns: []string{"a", "b"}, Weights: []float64{1, 1}, State: &TableState{}}
			root := laidOut(ctx, Col(Table(spec)))
			ctx.In = core.Input{X: 400, Y: 5, Click: true, Held: true, Clicks: 1}
			Update(ctx, root)
			ctx.In = core.Input{X: 0, Y: 5, Held: true}
			Update(ctx, root)
			ctx.In = core.Input{} // released
			Update(ctx, root)
			root = laidOut(ctx, Col(Table(spec)))
			header := root.Children()[0].Children()[0].Children()[0]
			return []int{header.Children()[0].Rect().W, header.Children()[1].Rect().W,
				btoi(cursorAt(ctx, root, 130, 5) == core.CursorResizeH)}
		}, []int{130, 670, 1}},
		{"foldColW, monoTextLead of a code view", func(ctx *core.Ctx) []int {
			code := &CodeView{}
			code.SetLines([]Line{{Spans: []Span{{Text: "first"}}, Foldable: true}, {Spans: []Span{{Text: "  second"}}}})
			element := Code(code, Grow(1))
			root := laidOut(ctx, Row(element))
			ops := drawnOps(ctx, root)
			text := textOf(ops, "first")
			// Where the fold column starts: after the gutter, the numbers and a lead.
			foldX := element.Rect().X + 21 + ctx.Width(ctx.F.Mono, "0") + 28
			arrow := 0
			for _, op := range ops {
				if mask, ok := op.(rendertest.DrawMask); ok {
					arrow = int(mask.Left) + mask.Mask.Bounds().Dx()/2 - foldX
					break
				}
			}
			got := []int{int(text.Left) - foldX, int(text.Left) - text.Clip.Min.X, arrow}
			for _, offset := range []int{-1, 0, 39, 40} {
				code.Folded, code.HasSel = map[int]bool{}, false
				tapAt(ctx, root, foldX+offset, element.Rect().Y+2)
				got = append(got, btoi(code.Folded[0])*10+btoi(code.HasSel))
			}
			return got
		}, []int{40 + 28, 28, 20, 0, 10, 10, 1}},
		// A press anywhere in a text area places the cursor, so its lead is
		// drawn only: it is what the text's clip reaches left of the text.
		{"monoTextLead of a text area", func(ctx *core.Ctx) []int {
			area := NewTextArea("first")
			root := laidOut(ctx, Row(Area(&area, Grow(1))))
			text := textOf(drawnOps(ctx, root), "first")
			return []int{int(text.Left) - text.Clip.Min.X}
		}, []int{28}},
		{"monoLinePad of a code view", func(ctx *core.Ctx) []int {
			code := &CodeView{}
			code.SetLines(namedLines())
			return monoRows(ctx, Code(code, Grow(1)), func() int { return code.selA.line })
		}, []int{18, 2}},
		{"monoLinePad of a text area", func(ctx *core.Ctx) []int {
			area := NewTextArea("first\nsecond\nthird\nfourth")
			return monoRows(ctx, Area(&area, Grow(1)), func() int { line, _ := area.LineCol(area.Cursor); return line })
		}, []int{18, 2}},
		{"monoLinePad moves a wheel notch", func(ctx *core.Ctx) []int {
			area := NewTextArea(longText())
			root := laidOut(ctx, Row(Area(&area, Grow(1))))
			ctx.In = core.Input{X: 400, Y: 100, Wheel: -1}
			Update(ctx, root)
			return []int{area.ScrollY - 3*ctx.LineH(ctx.F.Mono)}
		}, []int{3 * 18}},
		{"switchInset", func(ctx *core.Ctx) []int {
			var got []int
			for _, on := range []bool{false, true} {
				changes := 0
				toggle := Switch(on, OnChange(func() { changes++ }))
				root := laidOut(ctx, Row(AlignItems(AlignStart), toggle))
				track := toggle.Rect()
				knobColor := core.Muted
				if on {
					knobColor = core.AccentText
				}
				knob := nth(filledBoxes(ctx, root, ctx.Color(knobColor)), 0)
				// Below a track with the default inset.
				tapAt(ctx, root, track.X+2, track.Y+ctx.Sizes.Checkbox+2*core.DefaultSizes().SwitchInset)
				got = append(got, track.H-ctx.Sizes.Checkbox, knob.X-track.X, track.X+track.W-knob.X-knob.W, changes)
			}
			return got
		}, []int{8, 4, 39 - 15 - 4, 1, 8, 39 - 15 - 4, 4, 1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.got(themedCtx(t, core.Metrics{}, themedPins)); !slices.Equal(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}

// Without a pin the switch's inset is the hair step.
func TestSwitchInsetFollowsTheHairStep(t *testing.T) {
	ctx := themedCtx(t, core.Metrics{Steps: core.Steps{Hair: 3}}, nil)
	toggle := Switch(true)
	laidOut(ctx, Row(AlignItems(AlignStart), toggle))
	if got, want := toggle.Rect().H, ctx.Sizes.Checkbox+2*6; got != want {
		t.Errorf("switch height = %d, want %d", got, want)
	}
}

// The theme's leading spaces the rows of Code and TextArea, and a press
// follows the rows. A pin of monoLinePad wins over it.
func TestLeadingSpacesMonoLines(t *testing.T) {
	tests := []struct {
		name    string
		leading int
		pins    map[string]int
		pad     int
	}{
		{"no leading", 0, nil, 4},
		{"leading", 6, nil, 4 + 6},
		{"a pin wins over the leading", 6, map[string]int{"monoLinePad": 3}, 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := themedCtx(t, core.Metrics{Leading: test.leading}, test.pins)
			// A press just inside the third row of this case's rows.
			pitch := ctx.LineH(ctx.F.Mono) + test.pad
			code := &CodeView{}
			code.SetLines(namedLines())
			area := NewTextArea("first\nsecond\nthird\nfourth")
			for name, view := range map[string]struct {
				element *El
				row     func() int
			}{
				"code": {Code(code, Grow(1)), func() int { return code.selA.line }},
				"area": {Area(&area, Grow(1)), func() int { line, _ := area.LineCol(area.Cursor); return line }},
			} {
				root := laidOut(ctx, Row(view.element))
				ops := drawnOps(ctx, root)
				if got := textOf(ops, "second").Top - textOf(ops, "first").Top; got != float64(pitch) {
					t.Errorf("%s: rows are %v apart, want %d", name, got, pitch)
				}
				for _, at := range []struct{ posY, row int }{{2*pitch - 1, 1}, {2 * pitch, 2}} {
					tapAt(ctx, root, view.element.Rect().X+view.element.Rect().W/2, view.element.Rect().Y+at.posY)
					if got := view.row(); got != at.row {
						t.Errorf("%s: a press %d down hits row %d, want %d", name, at.posY, got, at.row)
					}
				}
			}
		})
	}
}

func TestParagraphLeadingSpacesLines(t *testing.T) {
	ctx := themedCtx(t, core.Metrics{Leading: 5}, nil)
	line := ctx.LineH(ctx.F.Small)
	paragraph := Paragraph(paragraphText)
	root := Col(AlignItems(AlignStart), paragraph)
	Layout(ctx, root, 150, 600)
	lines := wrapParagraph(ctx, ctx.F.Small, paragraphText, paragraph.Rect().W)
	if got, want := paragraph.Rect().H, len(lines)*line+(len(lines)-1)*5; len(lines) < 3 || got != want {
		t.Fatalf("height of %d lines = %d, want %d", len(lines), got, want)
	}
	canvas := rendertest.NewCanvas(150, 600)
	Draw(ctx, root, canvas)
	var tops []float64
	for _, op := range canvas.Ops() {
		// A glyph run may be drawn more than once (stem weight), at one top.
		if drawnText, ok := op.(rendertest.DrawText); ok && !slices.Contains(tops, drawnText.Top) {
			tops = append(tops, drawnText.Top)
		}
	}
	for i := 1; i < len(tops); i++ {
		if step := tops[i] - tops[i-1]; step != float64(line+5) {
			t.Fatalf("line %d is %v below the one before, want %d", i, step, line+5)
		}
	}
	if len(tops) != len(lines) {
		t.Fatalf("%d lines drawn, want %d", len(tops), len(lines))
	}

	empty := Paragraph("")
	Layout(ctx, Col(empty), 150, 600)
	if got := empty.Rect().H; got != 0 {
		t.Errorf("an empty paragraph is %d high, want 0", got)
	}
}

// Bounds hold each size alone, so a theme can still cross two of them: a
// thumb inset wider than its track, a menu's least width above its most, a
// menu's insets wider than the menu, a scrollbar wider than its view is high.
func TestCrossedPinsDrawNoNegativeRects(t *testing.T) {
	tests := []struct {
		name  string
		pins  map[string]int
		menuW int
	}{
		{"thumb inset wider than the track, menu minimum above its maximum",
			map[string]int{"thumbInset": 8, "scrollbarW": 2, "menuMinW": 800, "menuMaxW": 40}, 800},
		{"menu insets wider than the menu, bar wider than the view is high",
			map[string]int{"menuMinW": 40, "menuMaxW": 40, "menuPad": 64, "menuRowPadX": 64, "scrollbarW": 32}, 40},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := themedCtx(t, core.Metrics{}, test.pins)
			state := &MenuState{Items: []MenuItem{{Label: "a", Hint: "x", Shortcut: "Ctrl+K"}, {Separator: true}}}
			scroll := Scroll(&ScrollState{}, H(20), Row(H(400)))
			root := laidOut(ctx, Stack(Col(AlignItems(AlignStretch), scroll), Menu(ctx, state)))
			checkNoNegativeBoxes(t, drawnOps(ctx, root))
			if got := root.Children()[1].Rect().W; got != test.menuW {
				t.Errorf("menu width = %d, want %d", got, test.menuW)
			}
			// A track wider than it is long is still vertical: the thumb is
			// as wide as the track less the insets, and stays in the view.
			thumbBox := nth(filledBoxes(ctx, root, ctx.Color(core.Muted)), 0)
			track := scroll.Rect()
			if want := max(ctx.Sizes.ScrollbarW-2*ctx.Sizes.ThumbInset, 1); thumbBox.W != want || thumbBox.Y < track.Y || thumbBox.Y+thumbBox.H > track.Y+track.H {
				t.Errorf("thumb %+v in the view %+v, want %d wide inside it", thumbBox, track, want)
			}
		})
	}
}

func checkNoNegativeBoxes(t *testing.T, ops []rendertest.Op) {
	t.Helper()
	for _, op := range ops {
		var box render.Box
		switch drawnOp := op.(type) {
		case rendertest.FillRect:
			box = drawnOp.Box
		case rendertest.FillRoundRect:
			box = drawnOp.Box
		case rendertest.FillRectGradient:
			box = drawnOp.Box
		case rendertest.StrokeRect:
			box = drawnOp.Box
		case rendertest.StrokeRoundRect:
			box = drawnOp.Box
		}
		if box.X1 < box.X0 || box.Y1 < box.Y0 {
			t.Errorf("negative box in %+v", op)
		}
	}
}

// The controls' crossed pins: a selection inset of more than half the field,
// slider ends wider than the slider, a thumb taller than it, a switch inset
// that leaves the knob no room, and a mono lead wider than its view.
func TestCrossedControlPinsDrawNoNegativeRects(t *testing.T) {
	ctx := themedCtx(t, core.Metrics{}, map[string]int{
		"selectionInset": 32, "sliderPad": 64, "sliderW": 16, "sliderThumbH": 64, "sliderH": 8,
		"switchInset": 16, "monoTextLead": 32, "foldColW": 0, "monoLinePad": 0, "gutter": 0,
	})
	field := NewTextField("mm")
	field.SelectAll()
	slider := Slider(&SliderState{}, 0, 10, 1, 5, H(0), Grow(0), SliderMid(5))
	code := &CodeView{}
	code.SetLines(namedLines())
	code.SelectLines(0, 1)
	area := NewTextArea("first\nsecond " + strings.Repeat("m", 40))
	area.ScrollX = 30
	areaElement := Area(&area, W(60), H(40))
	root := laidOut(ctx, Col(AlignItems(AlignStart), PadL(100), Field(&field, W(60)), slider, Switch(true),
		Code(code, W(20), H(40)), areaElement))
	ops := drawnOps(ctx, root)
	checkNoNegativeBoxes(t, ops)
	// A lead wider than what is left of the text does not let the text out of
	// the editor.
	if text := textOf(ops, "second "+strings.Repeat("m", 40)); text.Clip.Min.X < areaElement.Rect().X || text.Text == "" {
		t.Errorf("text clip %v, want it inside the editor at %+v", text.Clip, areaElement.Rect())
	}
	if rect := slider.Rect(); rect.W != 16 || rect.H != 8 {
		t.Errorf("slider = %+v, want 16 by 8", rect)
	}
	thumb := nth(filledBoxes(ctx, root, ctx.Color(core.Accent)), 1)
	if thumb.H != 64 {
		t.Errorf("thumb = %+v, want 64 high", thumb)
	}
	// A track with no room between its ends keeps one pixel.
	if track := nth(filledBoxes(ctx, root, ctx.Color(core.Chip)), 0); track.W != 1 {
		t.Errorf("track = %+v, want 1 wide", track)
	}
}
