package ebitenrender

import (
	"image"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/hajimehoshi/ebiten/v2"
)

// Image is a render.Image made by Device.NewImage.
type Image struct{ img *ebiten.Image }

var _ render.Image = (*Image)(nil)

// Bounds implements render.Image.
func (i *Image) Bounds() image.Rectangle { return i.img.Bounds() }

// Offscreen is a render.Offscreen made by Device.NewOffscreen.
type Offscreen struct{ Canvas }

var _ render.Offscreen = (*Offscreen)(nil)

// Clear implements render.Offscreen.
func (o *Offscreen) Clear() { o.dst.Clear() }

// Release implements render.Offscreen.
func (o *Offscreen) Release() { o.dst.Deallocate() }

// ReadPixels implements render.Offscreen. Ebitengine reads pixels only while
// its run loop runs, and panics before. They come premultiplied already.
func (o *Offscreen) ReadPixels() *image.RGBA {
	bounds := o.Bounds()
	pixels := make([]byte, 4*bounds.Dx()*bounds.Dy())
	o.dst.ReadPixels(pixels)
	return &image.RGBA{Pix: pixels, Stride: 4 * bounds.Dx(), Rect: bounds}
}

// ImageOf is the ebiten image canvas draws on, for a caller that still draws
// with ebiten itself. It reports false for a canvas of another renderer. It is
// part of the bridge that lasts while the toolkit moves onto render.Canvas.
func ImageOf(canvas render.Canvas) (*ebiten.Image, bool) {
	switch surface := canvas.(type) {
	case *Canvas:
		return surface.dst, true
	case *Offscreen:
		return surface.dst, true
	}
	return nil, false
}

// pictureOf is the ebiten image behind img. A canvas that is not an offscreen,
// such as a Sub, is no image of this renderer although it has Bounds. A bare
// *ebiten.Image is one while the toolkit still passes them around (the bridge).
func pictureOf(img render.Image) *ebiten.Image {
	switch picture := img.(type) {
	case *ebiten.Image:
		return picture
	case *Image:
		return picture.img
	case *Offscreen:
		return picture.dst
	}
	panic("render: image from another renderer")
}
