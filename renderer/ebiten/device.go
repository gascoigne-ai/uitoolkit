package ebitenrender

import (
	"bytes"
	"fmt"
	"image"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Device makes what draws with Ebitengine and owns the corner masks the
// canvases made from it share. Make one with NewDevice: the zero value cannot
// build a mask. Like its canvases it is used by one goroutine at a time.
type Device struct{ masks cornerMasks }

var _ render.Device = (*Device)(nil)

// NewDevice returns a device with an empty mask cache.
func NewDevice() *Device {
	return &Device{masks: cornerMasks{make: buildCornerMask, makeRing: buildRingMask}}
}

// NewCanvas returns a canvas on dst, for example the screen that a caller's
// own ebiten.Game is handed.
func (d *Device) NewCanvas(dst *ebiten.Image) *Canvas { return &Canvas{device: d, dst: dst} }

// NewOffscreen returns a transparent offscreen of width by height pixels, both
// at least one.
func (d *Device) NewOffscreen(width, height int) render.Offscreen {
	return &Offscreen{Canvas{device: d, dst: ebiten.NewImage(width, height)}}
}

// NewImage returns a copy of src as an image the canvases can draw.
func (d *Device) NewImage(src image.Image) render.Image {
	return &Image{img: ebiten.NewImageFromImage(src)}
}

// NewFontSource loads a font from the content of a TrueType or OpenType file.
func (d *Device) NewFontSource(data []byte) (render.FontSource, error) {
	source, err := text.NewGoTextFaceSource(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("load font: %w", err)
	}
	return &FontSource{source: source}, nil
}
