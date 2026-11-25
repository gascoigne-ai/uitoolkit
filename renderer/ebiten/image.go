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

// pictureOf is the ebiten image behind img. A canvas that is not an offscreen,
// such as a Sub, is no image of this renderer although it has Bounds, and
// neither is a bare *ebiten.Image, which satisfies render.Image by its Bounds
// but would draw as if it were one.
func pictureOf(img render.Image) *ebiten.Image {
	switch picture := img.(type) {
	case *Image:
		return picture.img
	case *Offscreen:
		return picture.dst
	}
	panic("render: image from another renderer")
}
