package core

import (
	"fmt"
	"image"

	"github.com/gascoigne-ai/uitoolkit/ebitenrender"
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/hajimehoshi/ebiten/v2"
)

// This file is the bridge between the callers that still draw on ebiten images
// and Ctx, which draws through render.Canvas. It lasts until they all hand
// over canvases, and it is the only file of core that knows Ebitengine.

// Surface is what a Ctx draw call draws on: a render.Canvas, or an
// *ebiten.Image while some callers still hold one.
type Surface interface{ Bounds() image.Rectangle }

// ebitenBridge remembers the canvas of the last ebiten image drawn on, since a
// frame makes many draw calls on one image.
type ebitenBridge struct {
	dst    *ebiten.Image
	canvas render.Canvas
}

// CanvasOf is dst as a canvas: a render.Canvas as it is, an *ebiten.Image
// through Device, which must then be an *ebitenrender.Device. Any other dst
// panics.
func (c *Ctx) CanvasOf(dst Surface) render.Canvas {
	if canvas, ok := dst.(render.Canvas); ok {
		return canvas
	}
	picture, ok := dst.(*ebiten.Image)
	if !ok {
		panic(fmt.Sprintf("core: cannot draw on a %T", dst))
	}
	if c.bridge.dst == picture {
		return c.bridge.canvas
	}
	device, ok := c.Device.(*ebitenrender.Device)
	if !ok {
		panic(fmt.Sprintf("core: drawing on an *ebiten.Image needs the device of ebitenrender, not %T", c.Device))
	}
	c.bridge = ebitenBridge{dst: picture, canvas: device.NewCanvas(picture)}
	return c.bridge.canvas
}

// EbitenImage is the ebiten image canvas draws on, for code that still draws
// with ebiten itself. It reports false for a canvas of another renderer.
func (c *Ctx) EbitenImage(canvas render.Canvas) (*ebiten.Image, bool) {
	return ebitenrender.ImageOf(canvas)
}

// Clip returns dst limited to r, so drawing outside r is discarded.
func (c *Ctx) Clip(dst *ebiten.Image, rect Rect) *ebiten.Image {
	b := dst.Bounds()
	x0 := max(int(c.dev(float32(rect.X))), b.Min.X)
	y0 := max(int(c.dev(float32(rect.Y))), b.Min.Y)
	x1 := min(int(c.dev(float32(rect.X+rect.W))+0.5), b.Max.X)
	y1 := min(int(c.dev(float32(rect.Y+rect.H))+0.5), b.Max.Y)
	return dst.SubImage(image.Rect(x0, y0, max(x1, x0), max(y1, y0))).(*ebiten.Image)
}

// Read samples the devices for this frame, including the click count.
func (s *Sampler) Read() Input { return s.Sample(ebitenrender.ReadInput()) }
