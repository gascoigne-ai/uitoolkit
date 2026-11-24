package ebitenrender

import (
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// FontSource is a render.FontSource over a text/v2 font, made by
// Device.NewFontSource.
type FontSource struct{ source *text.GoTextFaceSource }

var _ render.FontSource = (*FontSource)(nil)

// face is the font at size. Only the source and the size are set, as on every
// face core makes.
func (f *FontSource) face(size float64) *text.GoTextFace {
	return &text.GoTextFace{Source: f.source, Size: size}
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
