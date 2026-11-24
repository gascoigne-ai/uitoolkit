package rendertest

import (
	"image"
	"image/color"
	"slices"

	"github.com/gascoigne-ai/uitoolkit/render"
)

// Op is one recorded drawing call. The set is closed: there is one type per
// kind of call. Each holds in Clip the bounds of the canvas it was made on,
// and the arguments as given.
type Op interface{ isOp() }

// FillRect records a Canvas.FillRect call.
type FillRect struct {
	Clip  image.Rectangle
	Box   render.Box
	Color color.RGBA
}

// FillRectGradient records a Canvas.FillRectGradient call.
type FillRectGradient struct {
	Clip        image.Rectangle
	Box         render.Box
	Top, Bottom color.RGBA
}

// FillRoundRect records a Canvas.FillRoundRect call.
type FillRoundRect struct {
	Clip        image.Rectangle
	Box         render.Box
	Radius      float32
	Corners     render.Corners
	Top, Bottom color.RGBA
}

// StrokeRect records a Canvas.StrokeRect call.
type StrokeRect struct {
	Clip  image.Rectangle
	Box   render.Box
	Width float32
	Color color.RGBA
}

// StrokeRoundRect records a Canvas.StrokeRoundRect call.
type StrokeRoundRect struct {
	Clip          image.Rectangle
	Box           render.Box
	Radius, Width float32
	Color         color.RGBA
}

// StrokeLine records a Canvas.StrokeLine call.
type StrokeLine struct {
	Clip           image.Rectangle
	X0, Y0, X1, Y1 float32
	Width          float32
	Color          color.RGBA
}

// FillPath records a Canvas.FillPath call, with a copy of the segments the
// path had at the call.
type FillPath struct {
	Clip     image.Rectangle
	Segments []render.Segment
	Color    color.RGBA
}

// StrokePath records a Canvas.StrokePath call, with a copy of the segments the
// path had at the call.
type StrokePath struct {
	Clip     image.Rectangle
	Segments []render.Segment
	Stroke   render.Stroke
	Color    color.RGBA
}

// DrawImage records a Canvas.DrawImage call.
type DrawImage struct {
	Clip             image.Rectangle
	Image            render.Image
	Left, Top, Scale float64
	Filter           render.Filter
}

// DrawMask records a Canvas.DrawMask call.
type DrawMask struct {
	Clip      image.Rectangle
	Mask      render.Image
	Left, Top float64
	Color     color.RGBA
}

// DrawText records a Canvas.DrawText call.
type DrawText struct {
	Clip      image.Rectangle
	Face      render.Face
	Text      string
	Left, Top float64
	Color     color.RGBA
}

// Clear records an Offscreen.Clear call.
type Clear struct {
	Clip image.Rectangle
}

func (FillRect) isOp()         {}
func (FillRectGradient) isOp() {}
func (FillRoundRect) isOp()    {}
func (StrokeRect) isOp()       {}
func (StrokeRoundRect) isOp()  {}
func (StrokeLine) isOp()       {}
func (FillPath) isOp()         {}
func (StrokePath) isOp()       {}
func (DrawImage) isOp()        {}
func (DrawMask) isOp()         {}
func (DrawText) isOp()         {}
func (Clear) isOp()            {}

// surface is what a Canvas and all its Subs share.
type surface struct {
	ops      []Op
	released bool
}

// Canvas is a render.Canvas that draws nothing and records every call, in paint
// order, together with the calls made on the canvases it shares its surface
// with.
type Canvas struct {
	bounds  image.Rectangle
	surface *surface
}

var _ render.Canvas = (*Canvas)(nil)

// NewCanvas returns an empty canvas of width by height pixels at the origin.
func NewCanvas(width, height int) *Canvas {
	return &Canvas{bounds: image.Rect(0, 0, width, height), surface: &surface{}}
}

// Bounds implements render.Canvas.
func (c *Canvas) Bounds() image.Rectangle { return c.bounds }

// Sub implements render.Canvas. The result is a *Canvas.
func (c *Canvas) Sub(clip image.Rectangle) render.Canvas {
	c.mustBeLive()
	return &Canvas{bounds: c.bounds.Intersect(clip), surface: c.surface}
}

// Ops returns every call recorded on the surface. The caller must not modify
// the slice.
func (c *Canvas) Ops() []Op { return c.surface.ops }

// FillRect implements render.Canvas.
func (c *Canvas) FillRect(box render.Box, clr color.RGBA) {
	c.record(FillRect{Clip: c.bounds, Box: box, Color: clr})
}

func (c *Canvas) record(op Op) {
	c.mustBeLive()
	c.surface.ops = append(c.surface.ops, op)
}

func (c *Canvas) mustBeLive() {
	if c.surface.released {
		panic("rendertest: use of a released offscreen")
	}
}

// FillRectGradient implements render.Canvas.
func (c *Canvas) FillRectGradient(box render.Box, top, bottom color.RGBA) {
	c.record(FillRectGradient{Clip: c.bounds, Box: box, Top: top, Bottom: bottom})
}

// FillRoundRect implements render.Canvas.
func (c *Canvas) FillRoundRect(box render.Box, radius float32, corners render.Corners, top, bottom color.RGBA) {
	c.record(FillRoundRect{Clip: c.bounds, Box: box, Radius: radius, Corners: corners, Top: top, Bottom: bottom})
}

// StrokeRect implements render.Canvas.
func (c *Canvas) StrokeRect(box render.Box, width float32, clr color.RGBA) {
	c.record(StrokeRect{Clip: c.bounds, Box: box, Width: width, Color: clr})
}

// StrokeRoundRect implements render.Canvas.
func (c *Canvas) StrokeRoundRect(box render.Box, radius, width float32, clr color.RGBA) {
	c.record(StrokeRoundRect{Clip: c.bounds, Box: box, Radius: radius, Width: width, Color: clr})
}

// StrokeLine implements render.Canvas.
func (c *Canvas) StrokeLine(x0, y0, x1, y1, width float32, clr color.RGBA) {
	c.record(StrokeLine{Clip: c.bounds, X0: x0, Y0: y0, X1: x1, Y1: y1, Width: width, Color: clr})
}

// FillPath implements render.Canvas.
func (c *Canvas) FillPath(path *render.Path, clr color.RGBA) {
	c.record(FillPath{Clip: c.bounds, Segments: slices.Clone(path.Segments()), Color: clr})
}

// StrokePath implements render.Canvas.
func (c *Canvas) StrokePath(path *render.Path, stroke render.Stroke, clr color.RGBA) {
	c.record(StrokePath{Clip: c.bounds, Segments: slices.Clone(path.Segments()), Stroke: stroke, Color: clr})
}

// DrawImage implements render.Canvas.
func (c *Canvas) DrawImage(img render.Image, left, top, scale float64, filter render.Filter) {
	c.record(DrawImage{Clip: c.bounds, Image: img, Left: left, Top: top, Scale: scale, Filter: filter})
}

// DrawMask implements render.Canvas.
func (c *Canvas) DrawMask(mask render.Image, left, top float64, clr color.RGBA) {
	c.record(DrawMask{Clip: c.bounds, Mask: mask, Left: left, Top: top, Color: clr})
}

// DrawText implements render.Canvas.
func (c *Canvas) DrawText(face render.Face, text string, left, top float64, clr color.RGBA) {
	c.record(DrawText{Clip: c.bounds, Face: face, Text: text, Left: left, Top: top, Color: clr})
}
