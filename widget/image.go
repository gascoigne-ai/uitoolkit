package widget

import (
	"math"
	"reflect"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
	. "github.com/gascoigne-ai/uitoolkit/view"
)

// Image shows a picture scaled down to fit its box, centered, keeping its
// proportions; a picture smaller than the box is not scaled up. A nil picture
// shows nothing and measures 0x0, and so does a nil pointer held in the
// interface (what a renderer's loader leaves when it fails), which is not
// equal to nil.
func Image(img render.Image, opts ...Option) *El {
	if value := reflect.ValueOf(img); value.Kind() == reflect.Pointer && value.IsNil() {
		img = nil
	}
	return Custom(imageB{img: img}, With(nil, opts)...)
}

type imageB struct {
	Base
	img render.Image
}

func (b imageB) Measure(*El, *core.Ctx, int) (int, int) {
	if b.img == nil {
		return 0, 0
	}
	r := b.img.Bounds()
	return r.Dx(), r.Dy()
}

func (b imageB) Draw(element *El, ctx *core.Ctx, dst render.Canvas) {
	if b.img == nil {
		return
	}
	r := element.Rect()
	iw, ih := float64(b.img.Bounds().Dx()), float64(b.img.Bounds().Dy())
	if iw == 0 || ih == 0 || r.W <= 0 || r.H <= 0 {
		return
	}
	// fit in logical pixels, then to device pixels
	k := math.Min(1, math.Min(float64(r.W)/iw, float64(r.H)/ih))
	ox := float64(r.X) + (float64(r.W)-iw*k)/2
	oy := float64(r.Y) + (float64(r.H)-ih*k)/2
	dst.DrawImage(b.img, ox*ctx.Scale, oy*ctx.Scale, k*ctx.Scale, render.FilterLinear)
}
