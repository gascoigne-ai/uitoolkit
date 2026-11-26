package widget

import (
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gascoigne-ai/uitoolkit/core"
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
			return []int{geometry.textX - geometry.numW - foldColW - 8,
				code.contentWidthOf(geometry, rect) - geometry.textX - geometry.colX(code.maxCols),
				400 - scrollbarW - geometry.textX - int(geometry.wrapWidth(true, rect))}
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
		{"resize edge grab", func(ctx *core.Ctx) []int {
			root := laidOut(ctx, Stack(Resizable(&ResizeState{}, 10, 10, 400, 400, W(200), H(100), FloatCenter())))
			rect := root.Children()[0].Rect()
			ctx.In = core.Input{X: rect.X + 12, Y: rect.Y + 50}
			ctx.Cursor = core.CursorDefault
			Update(ctx, root)
			return []int{int(ctx.Cursor)}
		}, []int{int(core.CursorResizeH)}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.got(themedCtx(t, core.Metrics{}, themedPins)); !slices.Equal(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}

// The switch's track is inset by the hair step, which no named size holds.
func TestSwitchInsetFollowsTheHairStep(t *testing.T) {
	ctx := themedCtx(t, core.Metrics{Steps: core.Steps{Hair: 3}}, nil)
	toggle := Switch(true)
	laidOut(ctx, Row(AlignItems(AlignStart), toggle))
	if got, want := toggle.Rect().H, ctx.Sizes.Checkbox+2*6; got != want {
		t.Errorf("switch height = %d, want %d", got, want)
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
