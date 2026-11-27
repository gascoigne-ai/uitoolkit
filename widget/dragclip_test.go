package widget

import (
	"image"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/core"
	. "github.com/gascoigne-ai/uitoolkit/view"
)

// hiddenAt is the coordinate a clipping element hands its children while
// the pointer is outside it.
const hiddenAt = -1 << 20

// inClip puts owner in a clipping box at 100,100, 300 by 200, inside a
// larger window, so that the pointer can leave the box on every side.
func inClip(items ...Item) *El {
	box := append([]Item{W(300), H(200), Clip()}, items...)
	return Col(PadL(100), PadT(100), AlignItems(AlignStart), Col(box...))
}

// clipFrame runs one frame of a 600x500 window over a tree built anew.
func clipFrame(ctx *core.Ctx, build func() *El, in core.Input) {
	ctx.In = in
	root := build()
	Layout(ctx, root, 600, 500)
	Update(ctx, root)
}

func pressAt(posX, posY int) core.Input {
	return core.Input{X: posX, Y: posY, Click: true, Held: true, Clicks: 1}
}

func heldAt(posX, posY int) core.Input { return core.Input{X: posX, Y: posY, Held: true} }

func freeAt(posX, posY int) core.Input { return core.Input{X: posX, Y: posY} }

// pointerLog records the pointer it is given each frame.
type pointerLog struct {
	Base
	seen *[]image.Point
}

func (p pointerLog) Update(_ *El, ctx *core.Ctx) {
	*p.seen = append(*p.seen, image.Pt(ctx.In.X, ctx.In.Y))
}

// dragCase is one widget whose drag starts inside the clip and goes on
// outside it. setup returns the tree and a check of the dragged state.
type dragCase struct {
	name     string
	from, to image.Point
	setup    func(ctx *core.Ctx) (build func() *El, check func(t *testing.T))
}

func wantInt(t *testing.T, what string, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %d, want %d", what, got, want)
	}
}

func runDragCases(t *testing.T, tests []dragCase) {
	t.Helper()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := testCtx(t)
			build, check := test.setup(ctx)
			clipFrame(ctx, build, pressAt(test.from.X, test.from.Y))
			clipFrame(ctx, build, heldAt(test.to.X, test.to.Y))
			check(t)
		})
	}
}

// A clipping parent hid the pointer from the owner of a drag too: dragging a
// Resizable's edge into a bar above its parent dropped the width to the
// minimum, a Split collapsed a pane, thumbs and selections jumped to the
// start.
func TestADragThatLeavesAClipKeepsFollowingThePointer(t *testing.T) {
	runDragCases(t, []dragCase{
		{"resizable edge above", image.Pt(398, 200), image.Pt(418, 50), func(*core.Ctx) (func() *El, func(*testing.T)) {
			state := &ResizeState{}
			return func() *El { return inClip(Resizable(state, 120, 80, 500, 400, Grow(1), Row())) },
				func(t *testing.T) { wantInt(t, "width", state.W, 340) }
		}},
		{"split divider above", image.Pt(200, 200), image.Pt(350, 50), func(*core.Ctx) (func() *El, func(*testing.T)) {
			state := &SplitState{}
			return func() *El { return inClip(Split(state, 100, Grow(1), Row(), Row())) },
				func(t *testing.T) { wantInt(t, "first size", int(state.Size), 250) }
		}},
		{"vertical split divider right", image.Pt(200, 200), image.Pt(450, 250), func(*core.Ctx) (func() *El, func(*testing.T)) {
			state := &SplitState{}
			return func() *El { return inClip(Split(state, 100, Vertical(), Grow(1), Row(), Row())) },
				func(t *testing.T) { wantInt(t, "first size", int(state.Size), 150) }
		}},
		{"scroll thumb below", image.Pt(395, 110), image.Pt(395, 450), func(*core.Ctx) (func() *El, func(*testing.T)) {
			state := &ScrollState{}
			return func() *El { return inClip(Scroll(state, Col(H(800)))) },
				func(t *testing.T) { wantInt(t, "scroll", state.Y, 600) }
		}},
		{"field selection right", image.Pt(111, 115), image.Pt(500, 115), func(*core.Ctx) (func() *El, func(*testing.T)) {
			field := NewTextField("hello world")
			field.Cursor = 0
			return func() *El { return inClip(Field(&field, H(30))) },
				func(t *testing.T) {
					wantInt(t, "anchor", field.Anchor, 0)
					wantInt(t, "cursor", field.Cursor, 11)
				}
		}},
		{"text area selection below", image.Pt(101, 105), image.Pt(500, 450), func(*core.Ctx) (func() *El, func(*testing.T)) {
			area := NewTextArea("hello\nworld\nagain")
			return func() *El { return inClip(Area(&area, Grow(1))) },
				func(t *testing.T) { wantInt(t, "cursor", area.Cursor, 17) }
		}},
		{"code selection below", image.Pt(200, 105), image.Pt(2000, 2000), func(ctx *core.Ctx) (func() *El, func(*testing.T)) {
			code := &CodeView{}
			code.SetLines(sampleLines())
			return func() *El { return inClip(Code(code, Grow(1))) },
				func(t *testing.T) {
					last := len(sampleLines()) - 1
					wantInt(t, "selection end line", code.selB.line, last)
					wantInt(t, "selection end column", code.selB.col, len([]rune(code.lineText(last))))
				}
		}},
		{"table column grip above", image.Pt(250, 110), image.Pt(310, 50), func(*core.Ctx) (func() *El, func(*testing.T)) {
			state := &TableState{}
			return func() *El {
					return inClip(Table(TableSpec{Columns: []string{"a", "b"}, Weights: []float64{1, 1}, State: state}))
				}, func(t *testing.T) {
					if len(state.Weights) != 2 {
						t.Fatalf("weights %v, want two", state.Weights)
					}
					wantInt(t, "left share in percent", int(100*state.Weights[0]/(state.Weights[0]+state.Weights[1])+0.5), 70)
				}
		}},
		// The two below passed before the fix: they read Ctx.PointerX already,
		// and pin the precedent.
		{"slider right", image.Pt(250, 110), image.Pt(550, 50), func(*core.Ctx) (func() *El, func(*testing.T)) {
			state, value := &SliderState{}, 0
			return func() *El {
					return inClip(Row(Slider(state, 0, 100, 1, value, OnSlide(func(next int) { value = next }))))
				},
				func(t *testing.T) { wantInt(t, "value", value, 100) }
		}},
		{"hscroll pan above", image.Pt(250, 115), image.Pt(150, 50), func(*core.Ctx) (func() *El, func(*testing.T)) {
			state := &ScrollState{}
			return func() *El { return inClip(HScroll(state, H(30), Col(W(800)))) },
				func(t *testing.T) { wantInt(t, "scroll", state.X, 100) }
		}},
	})
}

// With two clipping parents between the window and the owner, the drag
// follows the pointer just the same. It failed before the fix.
func TestADragThatLeavesNestedClipsKeepsFollowingThePointer(t *testing.T) {
	nested := func(state *ScrollState, content Node) *El {
		return inClip(Col(Clip(), Grow(1), Scroll(state, content)))
	}
	runDragCases(t, []dragCase{
		{"scroll thumb below", image.Pt(395, 110), image.Pt(395, 450), func(*core.Ctx) (func() *El, func(*testing.T)) {
			state := &ScrollState{}
			return func() *El { return nested(state, Col(H(800))) },
				func(t *testing.T) { wantInt(t, "scroll", state.Y, 600) }
		}},
		{"field in a scroll right", image.Pt(111, 115), image.Pt(500, 115), func(*core.Ctx) (func() *El, func(*testing.T)) {
			state, field := &ScrollState{}, NewTextField("hello world")
			field.Cursor = 0
			return func() *El { return nested(state, Field(&field, H(30))) },
				func(t *testing.T) { wantInt(t, "cursor", field.Cursor, 11) }
		}},
	})
}

// Guard, passes before and after the fix: a drag belongs to the element that
// took the press. One started on a slider outside the clip gives nothing
// inside the clip the pointer, here over the hidden half of a field.
func TestADragFromOutsideAClipDoesNotReachItsHiddenParts(t *testing.T) {
	ctx := testCtx(t)
	state, field, seen := &ScrollState{Y: 15}, NewTextField("hello world"), []image.Point{}
	slider, value := &SliderState{}, 0
	build := func() *El {
		above := Slider(slider, 0, 100, 1, value, H(100), OnSlide(func(next int) { value = next }))
		clipped := Col(W(300), H(200), Clip(),
			Scroll(state, Field(&field, H(30)), Custom(pointerLog{seen: &seen}, H(800))))
		return Col(PadL(100), AlignItems(AlignStart), Row(W(300), above), clipped)
	}
	clipFrame(ctx, build, pressAt(200, 50))
	pressed := value
	for _, in := range []core.Input{heldAt(200, 90), heldAt(300, 90)} {
		clipFrame(ctx, build, in)
	}
	if !slider.Dragging || value <= pressed {
		t.Fatalf("slider dragging %v, value %d after %d at the press, want it to follow", slider.Dragging, value, pressed)
	}
	if field.Dragging || field.HasSel || field.Focused || field.Cursor != 11 {
		t.Errorf("field = dragging %v, selection %v, focused %v, cursor %d, want untouched",
			field.Dragging, field.HasSel, field.Focused, field.Cursor)
	}
	for frame, at := range seen {
		if at != image.Pt(hiddenAt, hiddenAt) {
			t.Errorf("frame %d: sibling in the clip saw %v, want it hidden", frame, at)
		}
	}
}

// Guard, passes before and after the fix: OnDragOver is not the owner of a
// drag, so a row scrolled out of a Scroll is not dragged over through the
// clip, with either button.
func TestOnDragOverDoesNotFireThroughAClip(t *testing.T) {
	buttons := []struct {
		name string
		in   core.Input
	}{
		{"left", core.Input{X: 200, Y: 350, Held: true}},
		{"middle", core.Input{X: 200, Y: 350, MiddleHeld: true}},
	}
	for _, button := range buttons {
		t.Run(button.name, func(t *testing.T) {
			ctx := testCtx(t)
			state, fired := &ScrollState{}, 0
			build := func() *El {
				return inClip(Scroll(state, Col(H(220)), Row(H(60), OnDragOver(func() { fired++ })), Col(H(400))))
			}
			clipFrame(ctx, build, button.in)
			wantInt(t, "calls", fired, 0)
		})
	}
}

// Guard, passes before and after the fix: a row grip drag that started
// inside a Scroll does not run OnDragOverRow on a row that is scrolled out
// of view under the pointer.
func TestOnDragOverRowDoesNotFireThroughAClip(t *testing.T) {
	ctx := testCtx(t)
	state, grabbed, over := &ScrollState{}, []int{}, []int{}
	rows := make([]TableRow, 20)
	for i := range rows {
		rows[i] = TableRow{Cells: []string{"row"}}
	}
	build := func() *El {
		return inClip(Scroll(state, Table(TableSpec{
			Columns: []string{"a"}, Weights: []float64{1}, Rows: rows,
			OnGrabRow:     func(row int) { grabbed = append(grabbed, row) },
			OnDragOverRow: func(row int) { over = append(over, row) },
		})))
	}
	firstRow := 100 + core.RowOf(ctx.Sizes) + core.RowOf(ctx.Sizes)/2
	clipFrame(ctx, build, pressAt(105, firstRow))
	if len(grabbed) != 1 || grabbed[0] != 0 {
		t.Fatalf("grabbed %v, want row 0", grabbed)
	}
	over = over[:0]
	clipFrame(ctx, build, heldAt(200, 370))
	clipFrame(ctx, build, core.Input{X: 200, Y: 370, MiddleHeld: true})
	if len(over) != 0 {
		t.Errorf("rows %v were dragged over while the pointer was outside the clip", over)
	}
}

// Guard, passes before and after the fix: a release outside the clip ends
// the drag. Coming back without a button moves nothing and clicks nothing.
func TestAReleaseOutsideAClipEndsTheDrag(t *testing.T) {
	tests := []struct {
		name  string
		from  image.Point
		setup func() (owner func(onClick Option) *El, value func() int)
	}{
		{"split", image.Pt(200, 200), func() (func(Option) *El, func() int) {
			state := &SplitState{}
			return func(onClick Option) *El { return Split(state, 100, Grow(1), Row(onClick), Row(onClick)) },
				func() int { return int(state.Size) }
		}},
		{"resizable", image.Pt(398, 200), func() (func(Option) *El, func() int) {
			state := &ResizeState{}
			return func(onClick Option) *El {
					return Resizable(state, 120, 80, 500, 400, Grow(1), Row(onClick, Grow(1)))
				},
				func() int { return state.W }
		}},
		{"field", image.Pt(111, 115), func() (func(Option) *El, func() int) {
			field := NewTextField("hello world")
			return func(onClick Option) *El { return Field(&field, H(30), onClick) },
				func() int { return field.Cursor*1000 + field.Anchor }
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := testCtx(t)
			owner, value := test.setup()
			clicks := 0
			build := func() *El { return inClip(owner(OnClick(func() { clicks++ }))) }
			clipFrame(ctx, build, pressAt(test.from.X, test.from.Y))
			clipFrame(ctx, build, heldAt(450, 50))
			clipFrame(ctx, build, freeAt(450, 50))
			released, clicksAtRelease := value(), clicks
			clipFrame(ctx, build, freeAt(150, 115))
			clipFrame(ctx, build, freeAt(300, 200))
			wantInt(t, "state after the release", value(), released)
			wantInt(t, "clicks after the release", clicks, clicksAtRelease)
		})
	}
}

// Guard, passes before and after the fix: when the owner leaves the tree
// mid-drag nothing else takes the drag over, and the owner does not resume it
// when it is back after the release.
func TestADragWhoseOwnerLeftTheTreeReachesNobody(t *testing.T) {
	ctx := testCtx(t)
	state, shown, seen := &SplitState{}, true, []image.Point{}
	build := func() *El {
		return inClip(Custom(pointerLog{seen: &seen}, H(20)), If(shown, Split(state, 100, Grow(1), Row(), Row())))
	}
	clipFrame(ctx, build, pressAt(200, 200))
	shown, seen = false, nil
	clipFrame(ctx, build, heldAt(350, 50))
	if len(seen) != 1 || seen[0] != image.Pt(hiddenAt, hiddenAt) {
		t.Errorf("sibling in the clip saw %v, want the pointer hidden", seen)
	}
	clipFrame(ctx, build, freeAt(350, 50))
	shown = true
	before := state.Size
	clipFrame(ctx, build, freeAt(350, 200))
	clipFrame(ctx, build, heldAt(300, 200))
	if state.Size != before {
		t.Errorf("size = %v, want %v: the drag must not resume", state.Size, before)
	}
}
