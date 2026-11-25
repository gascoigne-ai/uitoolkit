package ebitenrender

import (
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// FontSource is a render.FontSource over a text/v2 font, made by
// Device.NewFontSource. Like its device it is used by one goroutine at a time.
type FontSource struct {
	source *text.GoTextFaceSource
	// faces keeps the face of each size asked for: every label of a frame is
	// measured or drawn through one, so a new face each call is an allocation
	// each call.
	faces map[float64]*text.GoTextFace
}

var _ render.FontSource = (*FontSource)(nil)

// faceLimit bounds FontSource.faces. A program uses a handful of sizes, so the
// limit is reached only by a runaway (a size animated every frame); the whole
// map is then dropped, as the corner masks are.
const faceLimit = 16

// face is the font at size. Only the source and the size are set, and a face
// is never changed once made.
func (f *FontSource) face(size float64) *text.GoTextFace {
	if face, ok := f.faces[size]; ok {
		return face
	}
	if f.faces == nil || len(f.faces) >= faceLimit {
		f.faces = make(map[float64]*text.GoTextFace)
	}
	face := &text.GoTextFace{Source: f.source, Size: size}
	f.faces[size] = face
	return face
}

// Metrics implements render.FontSource.
func (f *FontSource) Metrics(size float64) render.Metrics {
	metrics := f.face(size).Metrics()
	return render.Metrics{Ascent: metrics.HAscent, Descent: metrics.HDescent}
}

// Measure implements render.FontSource.
func (f *FontSource) Measure(content string, size float64) float64 {
	width, _ := text.Measure(content, f.face(size), 0)
	return width
}
