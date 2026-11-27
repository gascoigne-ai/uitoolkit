package widget

import (
	"image"
	"strings"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
	. "github.com/gascoigne-ai/uitoolkit/view"
)

// inWindow lays a dialog out over a window of that size and returns the root
// and the dialog's panel.
func inWindow(ctx *core.Ctx, frag Node, width, height int) (root, panel *El) {
	root = Stack(Row(Grow(1)), frag)
	Layout(ctx, root, width, height)
	return root, root.Children()[2]
}

// clicked presses and releases at the middle of rect.
func clicked(ctx *core.Ctx, root *El, rect core.Rect) {
	tapAt(ctx, root, rect.X+rect.W/2, rect.Y+rect.H/2)
}

func inside(inner, outer core.Rect) bool {
	return inner.X >= outer.X && inner.Y >= outer.Y &&
		inner.X+inner.W <= outer.X+outer.W && inner.Y+inner.H <= outer.Y+outer.H
}

// A dialog stays dialogMargin off the window's edges and is at most
// dialogMaxH high, also when the app gives it a height; MaxH and Margin after
// that opt out.
func TestDialogKeepsItsMarginAndMaxHeight(t *testing.T) {
	tests := []struct {
		name          string
		pins          map[string]int
		options       []Item
		width, height int
		want          core.Rect
	}{
		{"tall content, the margin bounds it", nil, nil, 800, 600, core.Rect{X: 250, Y: 40, W: 300, H: 520}},
		{"tall content, the maximum bounds it", nil, nil, 1024, 1200, core.Rect{X: 362, Y: 320, W: 300, H: 560}},
		{"themed margin and maximum", map[string]int{"dialogMargin": 10, "dialogMaxH": 200}, nil, 800, 600,
			core.Rect{X: 250, Y: 200, W: 300, H: 200}},
		{"themed margin in a low window", map[string]int{"dialogMargin": 10, "dialogMaxH": 200}, nil, 800, 150,
			core.Rect{X: 250, Y: 10, W: 300, H: 130}},
		{"too wide", nil, []Item{W(2000)}, 800, 600, core.Rect{X: 40, Y: 40, W: 720, H: 520}},
		{"the app's height is capped", nil, []Item{H(700)}, 800, 1200, core.Rect{X: 250, Y: 320, W: 300, H: 560}},
		{"MaxH(0) lifts the cap", nil, []Item{H(700), MaxH(0)}, 800, 1200, core.Rect{X: 250, Y: 250, W: 300, H: 700}},
		{"another MaxH replaces it", nil, []Item{H(700), MaxH(650)}, 800, 1200, core.Rect{X: 250, Y: 275, W: 300, H: 650}},
		{"MaxH(0) keeps the margin", nil, []Item{MaxH(0)}, 800, 600, core.Rect{X: 250, Y: 40, W: 300, H: 520}},
		{"Margin(0) lifts the margin", nil, []Item{Margin(0)}, 800, 500, core.Rect{X: 250, Y: 0, W: 300, H: 500}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := themedCtx(t, core.Metrics{}, test.pins)
			items := append([]Item{W(300), Row(H(2000))}, test.options...)
			_, panel := inWindow(ctx, Dialog(items...), test.width, test.height)
			if panel.Rect() != test.want {
				t.Errorf("panel = %+v, want %+v", panel.Rect(), test.want)
			}
		})
	}
}

func TestADraggedDialogStopsAtItsMargin(t *testing.T) {
	ctx := testCtx(t)
	move := &MoveState{X: 5000, Y: -5000}
	_, panel := inWindow(ctx, Dialog(W(300), H(200), Moved(move)), 800, 600)
	if want := (core.Rect{X: 460, Y: 40, W: 300, H: 200}); panel.Rect() != want {
		t.Errorf("panel = %+v, want %+v", panel.Rect(), want)
	}
	if move.X != 210 || move.Y != -160 {
		t.Errorf("moved = %d,%d, want 210,-160", move.X, move.Y)
	}
}

// The size the user drags a ResizableDialog to never runs past the height it
// is drawn at, and follows the height limit the app chose.
func TestResizableDialogStateStopsAtItsHeightLimit(t *testing.T) {
	tests := []struct {
		name    string
		pins    map[string]int
		options []Item
		wantH   int
	}{
		{"dialogMaxH", nil, nil, 560},
		{"themed dialogMaxH", map[string]int{"dialogMaxH": 450}, nil, 450},
		{"MaxH(0) leaves the resize bounds", nil, []Item{MaxH(0)}, 900},
		{"another MaxH", nil, []Item{MaxH(700)}, 700},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := themedCtx(t, core.Metrics{}, test.pins)
			state := &ResizeState{}
			items := append([]Item{W(300), H(300)}, test.options...)
			root, panel := inWindow(ctx, ResizableDialog(state, 100, 100, 500, 900, items...), 800, 1200)
			rect := panel.Rect()
			x, y := rect.X+rect.W/2, rect.Y+rect.H-2
			ctx.In = core.Input{X: x, Y: y, Click: true, Held: true, Clicks: 1}
			Update(ctx, root)
			ctx.In = core.Input{X: x, Y: y + 1000, Held: true}
			Update(ctx, root)
			if state.H != test.wantH {
				t.Errorf("state height = %d, want %d", state.H, test.wantH)
			}
		})
	}
}

// A message dialog whose content fits lays out as it did before it could
// scroll: these are the rects it had then.
func TestAMessageDialogThatFitsLaysOutAsBefore(t *testing.T) {
	tests := []struct {
		name                        string
		body, content, input, empty bool
		panelY, panelH, row, field  int
	}{
		{name: "buttons only", panelY: 247, panelH: 106, row: 308},
		{name: "body", body: true, panelY: 218, panelH: 164, row: 337},
		{name: "content", content: true, panelY: 218, panelH: 164, row: 337},
		{name: "body and content", body: true, content: true, panelY: 189, panelH: 222, row: 366},
		{name: "input", input: true, panelY: 226, panelH: 147, row: 328, field: 287},
		{name: "body and input", body: true, input: true, panelY: 197, panelH: 205, row: 357, field: 316},
		{name: "content and input", content: true, input: true, panelY: 197, panelH: 205, row: 357, field: 316},
		{name: "all", body: true, content: true, input: true, panelY: 168, panelH: 263, row: 386, field: 345},
		{name: "content that builds to nothing", empty: true, panelY: 247, panelH: 106, row: 308},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := testCtx(t)
			field := NewTextField("")
			state := &DialogState{Title: "T", Buttons: []DialogButton{{Label: "OK"}}}
			var content []Item
			if test.body {
				state.Body = "one\ntwo"
			}
			if test.content {
				content = []Item{Text("extra"), Text("more")}
			}
			if test.input {
				state.Input = &field
			}
			if test.empty {
				content = []Item{If(false, Text("never")), For([]string{}, func(_ int, line string) Node { return Text(line) })}
			}
			root, panel := inWindow(ctx, MessageDialog(state, content...), 800, 600)
			if want := (core.Rect{X: 210, Y: test.panelY, W: 380, H: test.panelH}); panel.Rect() != want {
				t.Errorf("panel = %+v, want %+v", panel.Rect(), want)
			}
			body := panel.Children()[1].Children()
			if want := (core.Rect{X: 226, Y: test.row, W: 348, H: 29}); body[len(body)-1].Rect() != want {
				t.Errorf("button row = %+v, want %+v", body[len(body)-1].Rect(), want)
			}
			if want := (core.Rect{X: 226, Y: test.field, W: 348, H: 29}); test.input && body[len(body)-2].Rect() != want {
				t.Errorf("input = %+v, want %+v", body[len(body)-2].Rect(), want)
			}
			// No scrollbar: the text is as wide as the body's inside.
			if thumbs := filledBoxes(ctx, root, ctx.Color(core.Muted)); len(thumbs) != 0 {
				t.Errorf("a dialog that fits drew a thumb: %+v", thumbs)
			}
		})
	}
}

// tallMessage is a message dialog with far more text than the window holds.
func tallMessage(ran *string) (*DialogState, *TextField) {
	field := NewTextField("")
	return &DialogState{
		Title: "T", Body: strings.Repeat("a line of the message\n", 60) + "the last line", Input: &field,
		Buttons: []DialogButton{{Label: "OK", Run: func() { *ran = "ok" }}},
	}, &field
}

// A message too tall for the window scrolls its body text and content, and
// keeps its input and buttons in view and working.
func TestATallMessageDialogScrollsItsBodyAndKeepsInputAndButtons(t *testing.T) {
	ctx := testCtx(t)
	ran := ""
	state, field := tallMessage(&ran)
	build := func() Node { return MessageDialog(state, Text("the extra content")) }
	root, panel := inWindow(ctx, build(), 800, 600)
	if want := (core.Rect{X: 210, Y: 40, W: 380, H: 520}); panel.Rect() != want {
		t.Fatalf("panel = %+v, want %+v", panel.Rect(), want)
	}
	kids := panel.Children()[1].Children()
	if len(kids) != 3 {
		t.Fatalf("the body has %d children, want a scroll, the input and the buttons", len(kids))
	}
	scroll, input, buttons := kids[0].Rect(), kids[1].Rect(), kids[2].Rect()
	for name, rect := range map[string]core.Rect{"scroll": scroll, "input": input, "buttons": buttons} {
		if !inside(rect, panel.Rect()) || rect.H <= 0 {
			t.Errorf("%s %+v is not in the panel %+v", name, rect, panel.Rect())
		}
	}
	if scroll.Y+scroll.H > input.Y || input.Y+input.H > buttons.Y {
		t.Errorf("scroll %+v, input %+v and buttons %+v overlap", scroll, input, buttons)
	}
	clicked(ctx, root, input)
	if !field.Focused {
		t.Error("a click in the input did not focus it")
	}
	row := kids[2].Children()
	clicked(ctx, root, row[len(row)-1].Rect())
	if ran != "ok" {
		t.Errorf("a click on the button ran %q", ran)
	}
	ctx.In = core.Input{X: scroll.X + 10, Y: scroll.Y + 10, Wheel: -1000}
	Update(ctx, root)
	if state.Scroll.Y <= 0 {
		t.Fatalf("the wheel left the scroll at %d", state.Scroll.Y)
	}
	_, panel = inWindow(ctx, build(), 800, 600)
	content := panel.Children()[1].Children()[0].Children()
	if last := content[len(content)-1].Rect(); !inside(last, panel.Children()[1].Children()[0].Rect()) {
		t.Errorf("scrolled to the end, the content %+v is not in view", last)
	}
}

// A dialog clips what it holds to its panel: what overflows is neither drawn
// nor clicked outside it. Other dialogs scroll only with a Scroll of the app.
func TestADialogClipsItsContentToThePanel(t *testing.T) {
	ctx := testCtx(t)
	clicks := 0
	root, panel := inWindow(ctx, Dialog(W(300), H(100),
		Row(H(150)), Text("below the panel", H(20), OnClick(func() { clicks++ })),
	), 800, 600)
	rect := panel.Rect()
	bounds := image.Rect(rect.X, rect.Y, rect.X+rect.W, rect.Y+rect.H)
	for _, op := range drawnOps(ctx, root) {
		if drawnText, ok := op.(rendertest.DrawText); ok && !drawnText.Clip.In(bounds) {
			t.Errorf("text drawn with the clip %v, outside the panel %v", drawnText.Clip, bounds)
		}
	}
	clicked(ctx, root, panel.Children()[1].Rect())
	if clicks != 0 {
		t.Errorf("a click on content cut off below the panel ran %d times", clicks)
	}
}

// The panel's own shadow, and what opens from inside a dialog (the text menu,
// a tooltip), are drawn outside the panel, not clipped by it.
func TestADialogDoesNotClipItsShadowOrPopups(t *testing.T) {
	ctx := testCtx(t)
	field := NewTextField("text")
	tips := &TooltipState{}
	build := func() (*El, *El) {
		frag := Dialog(W(300), H(60), Field(&field, H(30), TextMenu(&field)), Row(H(30), Tooltip("a tip below the dialog")))
		root := Stack(Row(Grow(1)), frag, PopupMenuLayer(ctx), TooltipLayer(tips))
		Layout(ctx, root, 800, 600)
		return root, frag.Children()[1]
	}
	whole := image.Rect(0, 0, 800, 600)
	root, panel := build()
	rect := panel.Rect()
	shadow := ctx.Color(core.Shadow)
	offset := ctx.Sizes.ShadowOffset
	if boxes := filledBoxes(ctx, root, shadow); !containsRect(boxes, core.Rect{X: rect.X + offset, Y: rect.Y + offset, W: rect.W, H: rect.H}) {
		t.Errorf("no whole shadow of %+v in %+v", rect, boxes)
	}

	input := panel.Children()[0].Rect()
	ctx.In = core.Input{X: input.X + 10, Y: input.Y + 10, RightClick: true}
	Update(ctx, root)
	root, _ = build()
	pasted := textOf(drawnOps(ctx, root), "Select all")
	if pasted.Text == "" || pasted.Clip != whole || int(pasted.Top) <= rect.Y+rect.H {
		t.Errorf("the text menu's last row drew as %+v, want it unclipped below the panel %+v", pasted, rect)
	}

	CloseMenu(ctx)
	root, panel = build()
	hovered := panel.Children()[1].Rect()
	ctx.In = core.Input{X: hovered.X + 5, Y: hovered.Y + hovered.H - 2}
	ctx.PointerX, ctx.PointerY = ctx.In.X, ctx.In.Y
	var tip rendertest.DrawText
	for _, tick := range []int{0, 1000} {
		ctx.Tick = tick
		tip = textOf(drawnOps(ctx, root), "a tip below the dialog")
	}
	if tip.Text == "" || tip.Clip != whole || int(tip.Top) < rect.Y+rect.H {
		t.Errorf("the tooltip drew as %+v, want it unclipped below the panel %+v", tip, rect)
	}
}

func containsRect(rects []core.Rect, want core.Rect) bool {
	for _, rect := range rects {
		if rect == want {
			return true
		}
	}
	return false
}

// The panel's clip does not stop a drag that leaves it: a scrollbar thumb and
// a text selection keep following the pointer, as without the clip. The same
// goes for what a message dialog scrolls.
func TestADragThatLeavesADialogKeepsFollowing(t *testing.T) {
	drag := func(ctx *core.Ctx, root *El, fromX, fromY, toX, toY int) {
		ctx.In = core.Input{X: fromX, Y: fromY, Click: true, Held: true, Clicks: 1}
		Update(ctx, root)
		ctx.In = core.Input{X: toX, Y: toY, Held: true}
		Update(ctx, root)
	}
	t.Run("scrollbar thumb", func(t *testing.T) {
		ctx := testCtx(t)
		state := &ScrollState{}
		root, panel := inWindow(ctx, Dialog(W(300), H(300), Scroll(state, Row(H(2000)))), 800, 600)
		rect := panel.Children()[0].Rect()
		drag(ctx, root, rect.X+rect.W-3, rect.Y+3, rect.X+rect.W-3, 590)
		if state.Y < 1000 {
			t.Errorf("scrolled to %d with the pointer below the panel, want the end", state.Y)
		}
	})
	t.Run("text selection", func(t *testing.T) {
		ctx := testCtx(t)
		field := NewTextField("select me")
		root, panel := inWindow(ctx, Dialog(W(300), Field(&field, H(30))), 800, 600)
		rect := panel.Children()[0].Rect()
		drag(ctx, root, rect.X+12, rect.Y+15, 790, rect.Y+15)
		if got := field.SelectedText(); !strings.HasSuffix(got, "me") {
			t.Errorf("selected %q with the pointer right of the panel, want up to the end", got)
		}
	})
	t.Run("text selection in a message dialog's content", func(t *testing.T) {
		ctx := testCtx(t)
		field := NewTextField("select me")
		root, panel := inWindow(ctx, MessageDialog(&DialogState{Title: "T"}, Field(&field, H(30))), 800, 600)
		scroll := panel.Children()[1].Children()[0]
		if len(scroll.Children()) == 0 {
			t.Fatal("the content is not in a scroll")
		}
		rect := scroll.Children()[0].Rect()
		drag(ctx, root, rect.X+12, rect.Y+15, 790, rect.Y+15)
		if got := field.SelectedText(); !strings.HasSuffix(got, "me") {
			t.Errorf("selected %q with the pointer right of the scroll, want up to the end", got)
		}
	})
}

// A window too small even for the input and the buttons draws no negative box
// and does not panic.
func TestAMessageDialogInATinyWindowDrawsNoNegativeRects(t *testing.T) {
	sizes := []struct{ width, height int }{{0, 0}, {1, 1}, {79, 79}, {85, 85}, {120, 100}, {200, 150}, {400, 190}, {500, 230}}
	for _, size := range sizes {
		ctx := testCtx(t)
		ran := ""
		state, _ := tallMessage(&ran)
		state.Scroll.Y = 500
		root, panel := inWindow(ctx, MessageDialog(state), size.width, size.height)
		ctx.In = core.Input{X: size.width / 2, Y: size.height / 2, Click: true, Held: true, Wheel: -1}
		Update(ctx, root)
		ctx.In = core.Input{X: size.width, Y: size.height, Held: true}
		Update(ctx, root)
		Layout(ctx, root, size.width, size.height)
		checkNoNegativeBoxes(t, drawnOps(ctx, root))
		var check func(element *El)
		check = func(element *El) {
			if rect := element.Rect(); rect.W < 0 || rect.H < 0 {
				t.Errorf("%dx%d: negative rect %+v", size.width, size.height, rect)
			}
			for _, kid := range element.Children() {
				check(kid)
			}
		}
		check(root)
		if rect := panel.Rect(); rect.W > max(size.width-80, 0) || rect.H > max(size.height-80, 0) {
			t.Errorf("%dx%d: panel %+v is not inside the margins", size.width, size.height, rect)
		}
	}
}
