package render

// Metrics is the height of a line of text at one size, in device pixels:
// Ascent above the baseline and Descent below it, both positive.
type Metrics struct{ Ascent, Descent float64 }

// FontSource is a font a renderer loaded, usable at any size. Implementations
// are pointer types, so that a Face is comparable and two faces are equal when
// they hold the very same font.
type FontSource interface {
	// Metrics is the line height at size.
	Metrics(size float64) Metrics
	// Measure is the width of text at size in device pixels.
	Measure(text string, size float64) float64
}

// Face is a font at one size. It is a value, written as a literal, and it is
// comparable, because caches use it as a map key. FontSource has no method
// that makes one: a Face is only its two fields.
type Face struct {
	Source FontSource
	Size   float64
}

// Metrics is the line height of the face.
func (f Face) Metrics() Metrics { return f.Source.Metrics(f.Size) }

// Measure is the width of text in the face.
func (f Face) Measure(text string) float64 { return f.Source.Measure(text, f.Size) }
