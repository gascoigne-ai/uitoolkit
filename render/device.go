package render

import "image"

// Device makes the things of one renderer that a program needs besides the
// canvas it is handed. What a device makes is for the canvases of that
// renderer only.
type Device interface {
	// NewFontSource loads a font from the content of a TrueType or OpenType
	// file.
	NewFontSource(data []byte) (FontSource, error)
	// NewOffscreen returns a transparent offscreen of width by height pixels,
	// both at least one.
	NewOffscreen(width, height int) Offscreen
	// NewImage returns a copy of source as an image the canvases can draw.
	NewImage(source image.Image) Image
}
