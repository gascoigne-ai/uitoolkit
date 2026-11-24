package rendertest

import (
	"image"
	"strings"
	"unicode/utf8"

	"github.com/gascoigne-ai/uitoolkit/render"
)

// Offscreen is a render.Offscreen that records like a Canvas. It holds no
// pixels, so ReadPixels returns a transparent image.
type Offscreen struct{ Canvas }

var _ render.Offscreen = (*Offscreen)(nil)

// NewOffscreen returns a live offscreen of width by height pixels.
func NewOffscreen(width, height int) *Offscreen {
	return &Offscreen{Canvas: *NewCanvas(width, height)}
}

// Clear implements render.Offscreen.
func (o *Offscreen) Clear() {
	o.record(Clear{Clip: o.bounds})
}

// Release implements render.Offscreen. From then on a drawing call or Sub, on
// the offscreen or on a canvas taken from it, and Clear, Release and
// ReadPixels panic. Bounds and Ops still answer, so that a test can read what
// was drawn.
func (o *Offscreen) Release() {
	o.mustBeLive()
	o.surface.released = true
}

// ReadPixels implements render.Offscreen.
func (o *Offscreen) ReadPixels() *image.RGBA {
	o.mustBeLive()
	return image.NewRGBA(o.bounds)
}

// Image is a render.Image that is only its bounds.
type Image struct{ Rect image.Rectangle }

var _ render.Image = (*Image)(nil)

// Bounds implements render.Image.
func (i *Image) Bounds() image.Rectangle { return i.Rect }

// FixedFont is a render.FontSource where every rune is equally wide. Its
// fields are in ems: multiples of the size.
type FixedFont struct{ AdvanceEm, AscentEm, DescentEm float64 }

var _ render.FontSource = (*FixedFont)(nil)

// Metrics implements render.FontSource.
func (f *FixedFont) Metrics(size float64) render.Metrics {
	return render.Metrics{Ascent: f.AscentEm * size, Descent: f.DescentEm * size}
}

// Measure implements render.FontSource: the number of runes in the longest
// line times AdvanceEm times size.
func (f *FixedFont) Measure(text string, size float64) float64 {
	longest := 0
	for line := range strings.SplitSeq(text, "\n") {
		longest = max(longest, utf8.RuneCountInString(line))
	}
	return float64(longest) * f.AdvanceEm * size
}
