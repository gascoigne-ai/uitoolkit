package view

import (
	"testing"

	"github.com/gascoigne-ai/uitoolkit/core"
)

// A floating child's margins are kept free between it and its Stack: its room
// is the Stack less those margins, and every way of floating is placed in it.
func TestFloatingChildKeepsItsMargins(t *testing.T) {
	// The margins leave the room {10, 20, 360, 240} of a 400 by 300 Stack.
	margins := []Item{MarginL(10), MarginT(20), MarginR(30), MarginB(40)}
	tests := []struct {
		name          string
		width, height int
		float         Option
		moved         *MoveState
		want          core.Rect
		wantMoved     [2]int
	}{
		{name: "center", width: 100, height: 60, float: FloatCenter(), want: core.Rect{X: 140, Y: 110, W: 100, H: 60}},
		{name: "top right", width: 100, height: 60, float: FloatTR(5, 7), want: core.Rect{X: 263, Y: 25, W: 100, H: 60}},
		{name: "point", width: 100, height: 60, float: FloatAt(5, 7), want: core.Rect{X: 15, Y: 27, W: 100, H: 60}},
		{name: "point past the room's right and bottom", width: 100, height: 60, float: FloatAt(350, 230),
			want: core.Rect{X: 270, Y: 190, W: 100, H: 60}},
		{name: "point before the room", width: 100, height: 60, float: FloatAt(-50, -5), want: core.Rect{X: 10, Y: 20, W: 100, H: 60}},
		{name: "dragged past the left and top", width: 100, height: 60, float: FloatCenter(), moved: &MoveState{X: -1000, Y: -1000},
			want: core.Rect{X: 10, Y: 20, W: 100, H: 60}, wantMoved: [2]int{-130, -90}},
		{name: "dragged past the right and bottom", width: 100, height: 60, float: FloatCenter(), moved: &MoveState{X: 1000, Y: 1000},
			want: core.Rect{X: 270, Y: 200, W: 100, H: 60}, wantMoved: [2]int{130, 90}},
		{name: "larger than the room", width: 900, height: 900, float: FloatCenter(), want: core.Rect{X: 10, Y: 20, W: 360, H: 240}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			items := append([]Item{W(test.width), H(test.height), test.float}, margins...)
			if test.moved != nil {
				items = append(items, Moved(test.moved))
			}
			floating := box(items...)
			Layout(testCtx(t), Stack(floating), 400, 300)
			if floating.Rect() != test.want {
				t.Errorf("rect = %+v, want %+v", floating.Rect(), test.want)
			}
			if test.moved != nil && (test.moved.X != test.wantMoved[0] || test.moved.Y != test.wantMoved[1]) {
				t.Errorf("moved = %d,%d, want %d,%d", test.moved.X, test.moved.Y, test.wantMoved[0], test.wantMoved[1])
			}
		})
	}
}

// Margins wider than the Stack leave no room: the child is empty, never
// negative.
func TestFloatingChildHasNoNegativeSizeWhenItsMarginsExceedTheStack(t *testing.T) {
	for _, float := range []Option{FloatCenter(), FloatTR(0, 0), FloatAt(3, 3)} {
		floating := box(W(100), H(60), Margin(40), float, Moved(&MoveState{X: 9, Y: 9}))
		Layout(testCtx(t), Stack(floating), 50, 30)
		if rect := floating.Rect(); rect.W != 0 || rect.H != 0 {
			t.Errorf("rect = %+v, want it empty", rect)
		}
	}
}

// A floating child with no width of its own is measured at the width of its
// room, so what it wraps fits the width it gets.
func TestFloatingChildIsMeasuredAtTheWidthOfItsRoom(t *testing.T) {
	asked := &widthAsked{}
	floating := Custom(asked, Margin(40), FloatCenter())
	Layout(testCtx(t), Stack(floating), 400, 300)
	if asked.maxW != 320 {
		t.Errorf("measured at %d, want 320", asked.maxW)
	}
}

type widthAsked struct {
	Base
	maxW int
}

func (w *widthAsked) Measure(_ *El, _ *core.Ctx, maxW int) (int, int) {
	w.maxW = maxW
	return 10, 10
}
