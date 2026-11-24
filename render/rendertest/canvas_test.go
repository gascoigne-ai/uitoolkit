package rendertest_test

import (
	"image"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
)

func TestSubBoundsIsTheIntersectionWithTheParent(t *testing.T) {
	// The parent is the canvas clipped to (10,10)-(90,70).
	parent := rendertest.NewCanvas(100, 80).Sub(image.Rect(10, 10, 90, 70))
	cases := []struct {
		name string
		clip image.Rectangle
		want image.Rectangle
	}{
		{"inside", image.Rect(20, 20, 30, 30), image.Rect(20, 20, 30, 30)},
		{"past the left", image.Rect(0, 20, 30, 30), image.Rect(10, 20, 30, 30)},
		{"past the top", image.Rect(20, 0, 30, 30), image.Rect(20, 10, 30, 30)},
		{"past the right", image.Rect(20, 20, 99, 30), image.Rect(20, 20, 90, 30)},
		{"past the bottom", image.Rect(20, 20, 30, 79), image.Rect(20, 20, 30, 70)},
		{"containing", image.Rect(0, 0, 100, 80), image.Rect(10, 10, 90, 70)},
		{"equal", image.Rect(10, 10, 90, 70), image.Rect(10, 10, 90, 70)},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := parent.Sub(testCase.clip).Bounds(); got != testCase.want {
				t.Errorf("Sub(%v).Bounds() = %v, want %v", testCase.clip, got, testCase.want)
			}
		})
	}
	t.Run("disjoint", func(t *testing.T) {
		if got := parent.Sub(image.Rect(95, 75, 99, 79)).Bounds(); !got.Empty() {
			t.Errorf("Bounds() = %v, want empty", got)
		}
	})
	t.Run("nested", func(t *testing.T) {
		nested := parent.Sub(image.Rect(0, 0, 50, 50)).Sub(image.Rect(40, 5, 80, 60))
		if got, want := nested.Bounds(), image.Rect(40, 10, 50, 50); got != want {
			t.Errorf("Bounds() = %v, want %v", got, want)
		}
	})
}

func TestSubKeepsTheParentCoordinates(t *testing.T) {
	canvas := rendertest.NewCanvas(100, 80)
	clip := image.Rect(10, 20, 60, 70)
	sub := canvas.Sub(clip)
	if got := sub.Bounds(); got != clip {
		t.Fatalf("Bounds() = %v, want %v: not translated to the origin", got, clip)
	}
	box := render.Box{X0: 12, Y0: 22.5, X1: 30, Y1: 40}
	sub.FillRect(box, red)
	want := []rendertest.Op{rendertest.FillRect{Clip: clip, Box: box, Color: red}}
	assertOps(t, canvas.Ops(), want)
}
