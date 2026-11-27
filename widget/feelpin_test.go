package widget

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gascoigne-ai/uitoolkit/core"
	. "github.com/gascoigne-ai/uitoolkit/view"
)

func TestBuiltInTooltipShowsAtTick30(t *testing.T) {
	state := &TooltipState{}
	root := Stack(Row(W(100), Tooltip("tip")), TooltipLayer(state))
	ctx := testCtx(t)
	ctx.In = core.Input{X: 50, Y: 10, TPS: 60}
	for tick, want := range map[int]bool{29: false, 30: true} {
		state, ctx.Tick = &TooltipState{}, 100
		root = Stack(Row(W(100), Tooltip("tip")), TooltipLayer(state))
		drawn(t, ctx, root, 200, 50)
		ctx.Tick = 100 + tick
		drawn(t, ctx, root, 200, 50)
		if state.Shown() != want {
			t.Errorf("after %d ticks: shown = %v, want %v", tick, state.Shown(), want)
		}
	}
}

// holdSteps presses the right arrow of a long strip, holds it for the given
// number of further ticks and returns the hold ticks that scrolled it.
func holdSteps(t *testing.T, ctx *core.Ctx, ticks int) []int {
	t.Helper()
	const tps = 60
	var hits []int
	state := &ScrollState{}
	root := Col(strip(state, 100, &hits))
	Layout(ctx, root, 300, 100)
	posX := 100 - ctx.Sizes.HScrollArrowW + 2
	ctx.In = core.Input{X: posX, Y: 10, Click: true, Held: true, TPS: tps}
	Update(ctx, root)
	var steps []int
	for held := 1; held <= ticks; held++ {
		before := state.X
		Layout(ctx, root, 300, 100)
		ctx.In = core.Input{X: posX, Y: 10, Held: true, TPS: tps}
		Update(ctx, root)
		if state.X != before {
			steps = append(steps, held)
		}
	}
	return steps
}

func TestBuiltInHoldStepsAtTicks20Then23Then26(t *testing.T) {
	got := holdSteps(t, testCtx(t), 27)
	if want := []int{20, 23, 26}; !slices.Equal(got, want) {
		t.Fatalf("hold stepped at %v, want %v", got, want)
	}
}

// scrollNotch is how far one wheel notch down moves a Scroll and an HScroll.
func scrollNotch(t *testing.T, ctx *core.Ctx) (down, sideways int) {
	t.Helper()
	var hits []int
	vertical := &ScrollState{}
	root := Col(scroller(vertical, 100, rows(200, &hits)...))
	Layout(ctx, root, 200, 100)
	wheel(ctx, 50, 50, -1)
	Update(ctx, root)
	horizontal := &ScrollState{}
	root = Col(strip(horizontal, 100, &hits))
	Layout(ctx, root, 300, 100)
	wheel(ctx, 50, 10, -1)
	Update(ctx, root)
	return vertical.Y, horizontal.X
}

func TestBuiltInWheelNotchScrolls48px(t *testing.T) {
	if down, sideways := scrollNotch(t, testCtx(t)); down != 48 || sideways != 48 {
		t.Fatalf("one notch moved %d and %d px, want 48 and 48", down, sideways)
	}
}

// wheelRows is how many rows and columns one wheel notch moves a CodeView and
// a TextArea, as pixels.
func wheelRows(t *testing.T, ctx *core.Ctx, lines, columns int) {
	t.Helper()
	long := strings.Repeat(strings.Repeat("0", 500)+"\n", 200)
	rect := core.Rect{W: 300, H: 200}

	code := &CodeView{}
	code.SetLines(textLines(strings.Split(long, "\n")))
	area := NewTextArea(long)
	area.Cursor = 0
	for name, widget := range map[string]struct {
		update func()
		scroll func() (int, int)
		grid   monoGrid
	}{
		"code": {func() { code.Update(ctx, rect) }, func() (int, int) { return code.ScrollX, code.ScrollY }, code.geometry(ctx, rect).monoGrid},
		"area": {func() { area.Update(ctx, rect) }, func() (int, int) { return area.ScrollX, area.ScrollY }, area.geometry(ctx, rect).monoGrid},
	} {
		frameInput(ctx, core.Input{})
		widget.update() // the first update settles the reveal of the cursor
		startX, startY := widget.scroll()
		frameInput(ctx, core.Input{X: 100, Y: 100, Wheel: -1})
		widget.update()
		frameInput(ctx, core.Input{X: 100, Y: 100, Wheel: -1, Shift: true})
		widget.update()
		gotX, gotY := widget.scroll()
		wantX, wantY := int(float64(widget.grid.charW*columns)), widget.grid.lineH*lines
		if gotX-startX != wantX || gotY-startY != wantY {
			t.Errorf("%s: one notch moved %d, %d px, want %d columns = %d and %d lines = %d",
				name, gotX-startX, gotY-startY, columns, wantX, lines, wantY)
		}
	}
}

func textLines(texts []string) []Line {
	lines := make([]Line, len(texts))
	for index, text := range texts {
		lines[index] = Line{Spans: []Span{{Text: text}}}
	}
	return lines
}

func TestBuiltInWheelNotchScrolls3LinesAnd6Columns(t *testing.T) {
	wheelRows(t, testCtx(t), 3, 6)
}

// panAfter drags a strip by distance pixels and returns how far it scrolled.
func panAfter(t *testing.T, ctx *core.Ctx, distance int) int {
	t.Helper()
	var hits []int
	state := &ScrollState{X: 50}
	root := Col(strip(state, 5, &hits))
	Layout(ctx, root, 300, 100)
	ctx.In = core.Input{X: 60, Y: 10, Click: true, Held: true}
	Update(ctx, root)
	Layout(ctx, root, 300, 100)
	ctx.In = core.Input{X: 60 - distance, Y: 10, Held: true}
	Update(ctx, root)
	return state.X - 50
}

func TestBuiltInPanStartsAt4px(t *testing.T) {
	ctx := testCtx(t)
	if got := panAfter(t, ctx, 3); got != 0 {
		t.Errorf("3px panned %d", got)
	}
	if got := panAfter(t, ctx, 4); got != 4 {
		t.Errorf("4px panned %d, want 4", got)
	}
}

func TestBuiltInInfoToastLasts3s(t *testing.T) {
	var stack ToastStack
	start := time.Unix(100, 0)
	stack.Show("saved", ToastInfo, start, core.Feel{})
	if stack.Prune(start.Add(2999 * time.Millisecond)) {
		t.Fatal("gone before 3s")
	}
	if !stack.Prune(start.Add(3 * time.Second)) {
		t.Fatal("still there at 3s")
	}
}
