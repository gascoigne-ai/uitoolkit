package widget

import (
	"image"
	"reflect"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
	. "github.com/gascoigne-ai/uitoolkit/view"
)

func TestImageMeasuresItsPicture(t *testing.T) {
	c := testCtx(t)
	im := Image(&rendertest.Image{Rect: image.Rect(0, 0, 80, 40)})
	Layout(c, Row(AlignItems(AlignCenter), im), 200, 100)
	if im.Rect().W != 80 || im.Rect().H != 40 {
		t.Fatalf("natural size %dx%d", im.Rect().W, im.Rect().H)
	}
}

// A loader that failed leaves a nil pointer in the interface, which is not
// equal to nil; it must read as no picture.
func TestImageWithATypedNilPictureMeasuresZero(t *testing.T) {
	c := testCtx(t)
	for name, pic := range map[string]render.Image{"nil": nil, "typed nil": (*rendertest.Image)(nil)} {
		t.Run(name, func(t *testing.T) {
			im := Image(pic)
			root := Row(AlignItems(AlignStart), im)
			canvas := rendertest.NewCanvas(100, 100)
			Layout(c, root, 100, 100)
			Draw(c, root, canvas)
			if im.Rect().W != 0 || im.Rect().H != 0 {
				t.Errorf("size %dx%d, want 0x0", im.Rect().W, im.Rect().H)
			}
			if len(canvas.Ops()) != 0 {
				t.Errorf("ops %v, want none", canvas.Ops())
			}
		})
	}
}

func TestImageDrawsScaledToFitWithLinearFilter(t *testing.T) {
	pic := &rendertest.Image{Rect: image.Rect(0, 0, 80, 40)}
	tests := []struct {
		name       string
		box        int
		left, top  float64 // logical pixels, before the device scale
		scaleToFit float64
	}{
		{"smaller box scales down", 30, 0, 7.5, 0.375},
		{"larger box does not scale up", 300, 110, 130, 1},
	}
	for _, scale := range []float64{1, 1.5, 2} {
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				c := testCtx(t)
				if err := c.SetScale(scale); err != nil {
					t.Fatal(err)
				}
				canvas := rendertest.NewCanvas(1000, 1000)
				root := Row(W(tc.box), H(tc.box), Image(pic, Grow(1)))
				Layout(c, root, tc.box, tc.box)
				Draw(c, root, canvas)
				want := []rendertest.Op{rendertest.DrawImage{
					Clip: image.Rect(0, 0, 1000, 1000), Image: pic,
					Left: tc.left * scale, Top: tc.top * scale, Scale: tc.scaleToFit * scale, Filter: render.FilterLinear,
				}}
				if got := canvas.Ops(); !reflect.DeepEqual(got, want) {
					t.Errorf("scale %v: ops %#v\nwant %#v", scale, got, want)
				}
			})
		}
	}
}
