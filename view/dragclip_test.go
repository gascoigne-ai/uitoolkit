package view

import (
	"image"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/core"
)

// inClip puts the items in a clipping box at 100,100, 300 by 200, inside a
// larger window, so that the pointer can leave the box.
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

// A moved element in a clipping parent jumped away once the pointer that
// dragged its handle left the parent.
func TestAMovedHandleInAClipFollowsTheDragOutside(t *testing.T) {
	ctx, move := testCtx(t), &MoveState{}
	build := func() *El {
		dialog := Col(W(100), H(40), FloatCenter(), Moved(move), Row(H(20), DragHandle()))
		return inClip(Stack(Grow(1), box(), dialog))
	}
	clipFrame(ctx, build, core.Input{X: 250, Y: 190, Click: true, Held: true})
	clipFrame(ctx, build, core.Input{X: 280, Y: 50, Held: true})
	if move.X != 30 || move.Y != -140 {
		t.Fatalf("offset %d,%d, want 30,-140", move.X, move.Y)
	}
}

// pointerLog records what a component inside a clip is given each frame.
type pointerLog struct {
	Base
	seen   *[]core.Input
	actual *[]image.Point
}

func (p pointerLog) Update(_ *El, ctx *core.Ctx) {
	*p.seen = append(*p.seen, ctx.In)
	*p.actual = append(*p.actual, image.Pt(ctx.PointerX, ctx.PointerY))
}

// Guard, passes before and after the fix: while the pointer is outside a
// clip, a child under it (its part past the box) gets no pointer, wheel or
// click, whether pressed, hovered or held. Ctx.PointerX and PointerY stay
// real.
func TestAPointerOutsideAClipDoesNotHoverOrClickItsChildren(t *testing.T) {
	const hidden = -1 << 20
	ctx := testCtx(t)
	seen, actual, clicks := []core.Input{}, []image.Point{}, 0
	build := func() *El {
		return inClip(Custom(pointerLog{seen: &seen, actual: &actual}, H(300), OnClick(func() { clicks++ })))
	}
	frames := []core.Input{
		{X: 200, Y: 350, Click: true, Held: true, Wheel: 1, WheelX: 1},
		{X: 200, Y: 350, Wheel: 1, WheelX: 1},
		{X: 200, Y: 350, Held: true},
	}
	for _, in := range frames {
		clipFrame(ctx, build, in)
	}
	if len(seen) != len(frames) {
		t.Fatalf("%d frames seen, want %d", len(seen), len(frames))
	}
	for frame, in := range seen {
		if in.X != hidden || in.Y != hidden || in.Wheel != 0 || in.WheelX != 0 {
			t.Errorf("frame %d: child saw pointer %d,%d wheel %v,%v, want them hidden", frame, in.X, in.Y, in.Wheel, in.WheelX)
		}
		if actual[frame] != image.Pt(200, 350) {
			t.Errorf("frame %d: PointerX,Y = %v, want 200,350", frame, actual[frame])
		}
	}
	if clicks != 0 {
		t.Errorf("%d clicks, want none", clicks)
	}
}
