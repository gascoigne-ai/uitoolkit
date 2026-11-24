package ebitenrender

// These tests moved here with the mask code, from core/cornermask_test.go.

import (
	"image/color"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/hajimehoshi/ebiten/v2"
)

// gradientAt keeps what the square gradient computed per vertex before it was
// extracted: a clamped linear mix of the two ends.
func TestGradientAt(t *testing.T) {
	top, bottom := color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 51}
	tests := []struct {
		name string
		y    float32
		want [4]float32
	}{
		{"top", 10, [4]float32{1, 0, 0, 1}},
		{"bottom", 30, [4]float32{0, 1, 0, 0.2}},
		{"middle", 20, [4]float32{0.5, 0.5, 0, 0.6}},
		{"above clamps to top", 0, [4]float32{1, 0, 0, 1}},
		{"below clamps to bottom", 99, [4]float32{0, 1, 0, 0.2}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, g, b, a := gradientAt(top, bottom, 10, 30, tc.y)
			got := [4]float32{r, g, b, a}
			for i := range got {
				if d := got[i] - tc.want[i]; d > 0.002 || d < -0.002 {
					t.Fatalf("gradientAt = %v, want %v", got, tc.want)
				}
			}
		})
	}
}

func TestGradientAtFlatSpanDoesNotDivideByZero(t *testing.T) {
	r, _, _, _ := gradientAt(color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 0, 255}, 5, 5, 5)
	if r != 1 {
		t.Fatalf("red = %v, want 1", r)
	}
}

func TestSliceQuads(t *testing.T) {
	tests := []struct {
		name           string
		x0, y0, x1, y1 float32
		r              float32
		corners        render.Corners
		count          int
		discCorners    int // quads that sample the disc, not the solid texel
	}{
		{"all corners", 0, 0, 20, 12, 4, render.CornersAll, 9, 4},
		{"top corners only", 0, 0, 20, 12, 4, render.CornersTop, 9, 2},
		{"bottom-left only", 0, 0, 20, 12, 4, render.CornerBL, 9, 1},
		{"no corners", 0, 0, 20, 12, 4, 0, 9, 0},
		{"width below 2r clamps the radius", 0, 0, 6, 12, 4, render.CornersAll, 6, 4},
		{"square clamps to a disc", 0, 0, 8, 8, 4, render.CornersAll, 4, 4},
		{"fractional edges", 2.5, 1.25, 22.5, 13.75, 4, render.CornersAll, 9, 4},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			quads := sliceQuads(tc.x0, tc.y0, tc.x1, tc.y1, tc.r, tc.corners)
			if len(quads) != tc.count {
				t.Fatalf("got %d quads, want %d: %v", len(quads), tc.count, quads)
			}
			assertCoversExactly(t, quads, render.Box{X0: tc.x0, Y0: tc.y0, X1: tc.x1, Y1: tc.y1})
			discs := 0
			for _, q := range quads {
				if q.src.X1 > q.src.X0 {
					discs++
				}
			}
			if discs != tc.discCorners {
				t.Fatalf("%d quads sample the disc, want %d", discs, tc.discCorners)
			}
		})
	}
}

func boxArea(box render.Box) float32 { return (box.X1 - box.X0) * (box.Y1 - box.Y0) }

// assertCoversExactly checks the quads tile rect: their areas add up to its
// area and no two overlap, so there is no gap and no double blend.
func assertCoversExactly(t *testing.T, quads []quad, rect render.Box) {
	t.Helper()
	var sum float32
	for i, q := range quads {
		sum += boxArea(q.dst)
		if q.dst.X0 < rect.X0 || q.dst.Y0 < rect.Y0 || q.dst.X1 > rect.X1 || q.dst.Y1 > rect.Y1 {
			t.Fatalf("quad %v leaves %v", q.dst, rect)
		}
		for _, other := range quads[i+1:] {
			if q.dst.X0 < other.dst.X1 && other.dst.X0 < q.dst.X1 && q.dst.Y0 < other.dst.Y1 && other.dst.Y0 < q.dst.Y1 {
				t.Fatalf("quads overlap: %v and %v", q.dst, other.dst)
			}
		}
	}
	if d := sum - boxArea(rect); d > 0.001 || d < -0.001 {
		t.Fatalf("quad areas add up to %v, want %v", sum, boxArea(rect))
	}
}

// discQuarters are the corner squares of an 8x8 mask, in the order TL, TR, BR,
// BL.
var discQuarters = []render.Box{
	{X0: 0, Y0: 0, X1: 4, Y1: 4}, {X0: 4, Y0: 0, X1: 8, Y1: 4},
	{X0: 4, Y0: 4, X1: 8, Y1: 8}, {X0: 0, Y0: 4, X1: 4, Y1: 8},
}

func TestSliceQuadsDiscQuarters(t *testing.T) {
	quads := sliceQuads(0, 0, 20, 12, 4, render.CornersAll)
	// The corner squares come first; each samples its own quarter of the disc.
	for i, src := range discQuarters {
		if len(quads) < 4 || quads[i].src != src {
			t.Fatalf("corner %d samples %v, want %v", i, quads, src)
		}
	}
}

func TestCornerMasksBuildOncePerRadius(t *testing.T) {
	built := 0
	masks := &cornerMasks{make: func(int) *ebiten.Image { built++; return nil }}
	for _, radius := range []int{4, 4, 6, 4, 6} {
		masks.get(radius)
	}
	if built != 2 {
		t.Fatalf("built %d masks, want 2", built)
	}
}

func TestCornerMasksAreBounded(t *testing.T) {
	built := 0
	masks := &cornerMasks{make: func(int) *ebiten.Image { built++; return nil }}
	for radius := 1; radius <= maskLimit; radius++ {
		masks.get(radius)
	}
	if len(masks.images) != maskLimit {
		t.Fatalf("holds %d masks, want %d", len(masks.images), maskLimit)
	}
	masks.get(maskLimit + 1)
	if len(masks.images) != 1 {
		t.Fatalf("full cache holds %d masks after one more, want 1 (dropped all)", len(masks.images))
	}
	masks.get(1) // dropped, so built again
	if built != maskLimit+2 {
		t.Fatalf("built %d masks, want %d", built, maskLimit+2)
	}
}

func TestRingQuads(t *testing.T) {
	tests := []struct {
		name           string
		x0, y0, x1, y1 float32
		r, width       float32
		count          int
	}{
		{"corners and four edges", 0, 0, 20, 12, 4, 1, 8},
		{"wider line", 0, 0, 20, 12, 4, 2, 8},
		{"line as wide as the radius", 0, 0, 20, 12, 4, 4, 8},
		{"width below 2r drops the empty edge", 0, 0, 8, 12, 4, 1, 6},
		{"square of 2r is four corners", 0, 0, 8, 8, 4, 1, 4},
		{"fractional origin", 3, 5, 23, 17, 4, 1, 8},
		{"one pixel radius", 0, 0, 2, 2, 1, 1, 4},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			quads := ringQuads(tc.x0, tc.y0, tc.x1, tc.y1, tc.r, tc.width)
			if len(quads) != tc.count {
				t.Fatalf("got %d quads, want %d: %v", len(quads), tc.count, quads)
			}
			assertRingCoverage(t, quads, render.Box{X0: tc.x0, Y0: tc.y0, X1: tc.x1, Y1: tc.y1}, tc.r, tc.width)
		})
	}
}

// assertRingCoverage checks that the quads never overlap, stay inside rect and
// leave the inner rect (rect minus the line width on each side) uncovered
// except for the corner squares, which are covered whole (the mask is clear
// there).
func assertRingCoverage(t *testing.T, quads []quad, rect render.Box, radius, width float32) {
	t.Helper()
	inner := render.Box{X0: rect.X0 + width, Y0: rect.Y0 + width, X1: rect.X1 - width, Y1: rect.Y1 - width}
	radius = min(radius, (rect.X1-rect.X0)/2, (rect.Y1-rect.Y0)/2)
	for y := rect.Y0; y < rect.Y1; y++ {
		for x := rect.X0; x < rect.X1; x++ {
			covers := 0
			for _, q := range quads {
				if x >= q.dst.X0 && x < q.dst.X1 && y >= q.dst.Y0 && y < q.dst.Y1 {
					covers++
				}
			}
			inInner := x >= inner.X0 && x < inner.X1 && y >= inner.Y0 && y < inner.Y1
			inCorner := (x < rect.X0+radius || x >= rect.X1-radius) && (y < rect.Y0+radius || y >= rect.Y1-radius)
			want := 1
			if inInner && !inCorner {
				want = 0
			}
			if covers != want {
				t.Fatalf("pixel %v,%v covered %d times, want %d (quads %v)", x, y, covers, want, quads)
			}
		}
	}
}

func TestRingQuadsSamplesTheMask(t *testing.T) {
	quads := ringQuads(0, 0, 20, 12, 4, 2)
	// Corners first, each its own quarter of the 8x8 ring.
	for i, src := range discQuarters {
		if len(quads) < 8 || quads[i].src != src {
			t.Fatalf("corner %d samples %v, want %v", i, quads, src)
		}
	}
	// The edges are the line width thick and sample the solid texel.
	for _, q := range quads[4:] {
		if q.src.X1 != q.src.X0 || q.src.Y1 != q.src.Y0 {
			t.Fatalf("edge %v samples more than the solid texel", q)
		}
	}
}

func TestRingMasksAreKeyedByWidth(t *testing.T) {
	built := 0
	masks := &cornerMasks{makeRing: func(int, int) *ebiten.Image { built++; return nil }, make: func(int) *ebiten.Image { built++; return nil }}
	masks.getRing(4, 1)
	masks.getRing(4, 1)
	masks.getRing(4, 2)
	masks.get(4)
	if built != 3 {
		t.Fatalf("built %d masks, want 3 (ring w1, ring w2, fill)", built)
	}
}
