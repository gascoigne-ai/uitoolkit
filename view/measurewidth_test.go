package view_test

import (
	"slices"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/view"
)

// flowing is a leaf whose height depends on the width it is measured at, as
// wrapped content's does.
type flowing struct {
	view.Base
	measuredAt []int // every width asked: layout asks again at the final one
}

const flowingArea = 6000

func (f *flowing) Measure(_ *view.El, _ *core.Ctx, maxW int) (int, int) {
	f.measuredAt = append(f.measuredAt, maxW)
	return maxW, flowingArea / max(maxW, 1)
}

// Content was measured at the offered width and clamped to the bound only
// afterwards, so wrapped content came out shorter than it was drawn.
func TestContentIsMeasuredAtTheWidthItsParentIsBoundTo(t *testing.T) {
	const pad = 10
	leaf := &flowing{}
	box := view.Col(view.MaxW(100), view.Pad(pad), view.Custom(leaf))
	view.Layout(newCtx(t), view.Col(box), 400, 400)
	const want = 100 - 2*pad
	if !slices.Equal(leaf.measuredAt, []int{want}) {
		t.Errorf("measured at %v, want only %d", leaf.measuredAt, want)
	}
	if got := box.Rect().H; got != flowingArea/want+2*pad {
		t.Errorf("box height %d, want %d", got, flowingArea/want+2*pad)
	}
}

// MinW wider than the room is the width the element gets, so it is also the
// width its content is measured at.
func TestContentIsMeasuredAtAMinWiderThanTheRoom(t *testing.T) {
	leaf := &flowing{}
	box := view.Col(view.MinW(300), view.Custom(leaf))
	view.Layout(newCtx(t), view.Col(box), 100, 400)
	if !slices.Equal(leaf.measuredAt, []int{300}) {
		t.Errorf("measured at %v, want only 300", leaf.measuredAt)
	}
	if got := box.Rect().H; got != flowingArea/300 {
		t.Errorf("box height %d, want %d", got, flowingArea/300)
	}
}
