package render_test

import (
	"image/color"
	"slices"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
)

func TestPathRecordsSegmentsInOrder(t *testing.T) {
	var path render.Path
	if got := path.Segments(); len(got) != 0 {
		t.Fatalf("zero Path has segments %+v", got)
	}
	path.MoveTo(1, 2)
	path.LineTo(3, 4)
	path.CubicTo(5, 6, 7, 8, 9, 10)
	path.Close()
	want := []render.Segment{
		{Verb: render.VerbMove, Coords: [6]float32{1, 2}},
		{Verb: render.VerbLine, Coords: [6]float32{3, 4}},
		{Verb: render.VerbCubic, Coords: [6]float32{5, 6, 7, 8, 9, 10}},
		{Verb: render.VerbClose},
	}
	if got := path.Segments(); !slices.Equal(got, want) {
		t.Fatalf("Segments() = %+v, want %+v", got, want)
	}

	canvas := rendertest.NewCanvas(10, 10)
	canvas.FillPath(&path, color.RGBA{A: 255})
	canvas.StrokePath(&path, render.Stroke{Width: 1}, color.RGBA{A: 255})
	// This breaks the rule that a caller does not modify the segments, on
	// purpose: only an op that holds a copy keeps the old point.
	path.Segments()[0].Coords[0] = 99
	recorded := [][]render.Segment{
		canvas.Ops()[0].(rendertest.FillPath).Segments,
		canvas.Ops()[1].(rendertest.StrokePath).Segments,
	}
	for _, got := range recorded {
		if !slices.Equal(got, want) {
			t.Errorf("recorded segments = %+v, want those at the call %+v", got, want)
		}
	}
}
