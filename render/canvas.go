package render

import (
	"image"
	"image/color"
)

// Box is a rectangle in device pixels, from (X0, Y0) to (X1, Y1). It is not an
// image.Rectangle because shapes may start and end between pixels: a canvas
// draws such an edge anti-aliased, and image.Rectangle holds whole pixels
// only. image.Rectangle is used where whole pixels are meant: bounds and
// clips.
type Box struct{ X0, Y0, X1, Y1 float32 }

// Corners is a set of rectangle corners. For now it duplicates core.Corners;
// task 003 makes core.Corners an alias of this type.
type Corners uint8

// The corners of a rectangle.
const (
	CornerTL Corners = 1 << iota
	CornerTR
	CornerBR
	CornerBL

	CornersTop    = CornerTL | CornerTR
	CornersBottom = CornerBL | CornerBR
	CornersAll    = CornersTop | CornersBottom
)

// Filter is how a scaled image is sampled.
type Filter uint8

// The filters. The zero value is FilterNearest.
const (
	// FilterNearest takes the closest source pixel.
	FilterNearest Filter = iota
	// FilterLinear blends the closest source pixels.
	FilterLinear
)

// Canvas is a surface a renderer draws on.
//
// All coordinates and sizes are device pixels. A canvas never snaps them to
// whole pixels; a caller that wants sharp edges snaps before it draws. Every
// color.RGBA is alpha-premultiplied, as image/color defines it. A canvas, and
// every canvas that shares its surface, is used by one goroutine at a time.
//
// It has 13 methods, more than a port should: each is one thing a GPU backend
// draws in its own way, and a caller needs all of them, so splitting it would
// only make every caller ask for every part (an accepted exception, listed in
// .harness/architecture/toolkit.md).
type Canvas interface {
	// Bounds is the part of the surface that drawing can change.
	Bounds() image.Rectangle
	// Sub is the same surface clipped to Bounds().Intersect(clip). It keeps
	// the coordinates of the surface: nothing is translated, so a point means
	// the same pixel on the canvas and on its Sub.
	Sub(clip image.Rectangle) Canvas
	// FillRect fills box with one color.
	FillRect(box Box, clr color.RGBA)
	// FillRectGradient fills box with a vertical gradient from top, at its
	// upper edge, to bottom, at its lower edge.
	FillRectGradient(box Box, top, bottom color.RGBA)
	// FillRoundRect is FillRectGradient with the given corners rounded by
	// radius, anti-aliased. For one color, pass it as both top and bottom.
	FillRoundRect(box Box, radius float32, corners Corners, top, bottom color.RGBA)
	// StrokeRect draws a border of the given width just inside box.
	StrokeRect(box Box, width float32, clr color.RGBA)
	// StrokeRoundRect is StrokeRect with all corners rounded by radius,
	// anti-aliased.
	StrokeRoundRect(box Box, radius, width float32, clr color.RGBA)
	// StrokeLine draws an anti-aliased line from (x0, y0) to (x1, y1).
	StrokeLine(x0, y0, x1, y1, width float32, clr color.RGBA)
	// FillPath fills the inside of path by the non-zero winding rule,
	// anti-aliased.
	FillPath(path *Path, clr color.RGBA)
	// StrokePath draws the outline of path, anti-aliased.
	StrokePath(path *Path, stroke Stroke, clr color.RGBA)
	// DrawImage draws img scaled by scale, its top-left corner at (left, top).
	DrawImage(img Image, left, top, scale float64, filter Filter)
	// DrawMask paints clr through the alpha of mask, the top-left corner of
	// the mask at (left, top).
	DrawMask(mask Image, left, top float64, clr color.RGBA)
	// DrawText draws text in face and one pass, with the top-left corner of
	// its line box at (left, top).
	DrawText(face Face, text string, left, top float64, clr color.RGBA)
}

// Image is a picture a renderer made. A canvas accepts only the images of its
// own renderer and panics with "render: image from another renderer" for any
// other.
type Image interface {
	// Bounds is the size of the image in device pixels.
	Bounds() image.Rectangle
}

// Offscreen is a surface that is drawn on as a Canvas and then drawn elsewhere
// as an Image. It is live until Release; after Release no method may be
// called, including those of the canvases taken from it with Sub.
type Offscreen interface {
	Canvas
	Image
	// Clear makes every pixel transparent.
	Clear()
	// Release frees what the offscreen holds.
	Release()
	// ReadPixels copies the pixels, alpha-premultiplied, into an image whose
	// Rect equals Bounds().
	ReadPixels() *image.RGBA
}
